package http

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/push/service"
	"github.com/cossim/coss-server/pkg/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/version"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net/http"
)

type Handler struct {
	logger      *zap.Logger
	enc         encryption.Encryptor
	userCache   cache.UserCache
	PushService *service.Service
}

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func (h *Handler) Init(cfg *pkgconfig.AppConfig) error {
	h.logger = plog.NewDefaultLogger("push_bff", int8(cfg.Log.Level))
	userCache := cache.NewUserCacheRedisWithClient(redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s", cfg.Redis.Addr()),
		Password: cfg.Redis.Password,
		DB:       0,
	}))
	h.userCache = userCache
	h.enc = encryption.NewEncryptor([]byte(cfg.Encryption.Passphrase), cfg.Encryption.Name, cfg.Encryption.Email, cfg.Encryption.RsaBits, cfg.Encryption.Enable)
	if cfg.Encryption.Enable {
		return h.enc.ReadKeyPair()
	}
	return nil
}

func (h *Handler) Name() string {
	return "push_bff"
}

func (h *Handler) Version() string {
	return version.FullVersion()
}

func (h *Handler) RegisterRoute(r gin.IRouter) {
	u := r.Group("/api/v1/push")
	u.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(h.logger), middleware.EncryptionMiddleware(h.enc), middleware.RecoveryMiddleware())
	u.Use(middleware.AuthMiddleware(h.userCache))
	u.GET("/ws", h.ws)
}

func (h *Handler) Health(r gin.IRouter) string {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) Stop(ctx context.Context) error {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) DiscoverServices(services map[string]*grpc.ClientConn) error {
	for k, v := range services {
		if err := h.PushService.HandlerGrpcClient(k, v); err != nil {
			h.logger.Error("handler grpc client error", zap.String("name", k), zap.String("addr", v.Target()), zap.Error(err))
		}
	}
	return nil
}
