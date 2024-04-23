package http

import (
	"context"
	"github.com/cossim/coss-server/internal/live/service"
	"github.com/cossim/coss-server/internal/user/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/manager/server"
	"github.com/cossim/coss-server/pkg/version"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	_ server.HTTPService = &Handler{}
)

type Handler struct {
	svc       *service.Service
	logger    *zap.Logger
	enc       encryption.Encryptor
	userCache cache.UserCache
}

func (h *Handler) Init(cfg *pkgconfig.AppConfig) error {
	if cfg.Encryption.Enable {
		return h.enc.ReadKeyPair()
	}
	userCache, err := cache.NewUserCacheRedis(cfg.Redis.Addr(), cfg.Redis.Password, 0)
	if err != nil {
		return err
	}
	h.userCache = userCache
	h.logger = plog.NewDefaultLogger("live_bff", int8(cfg.Log.Level))
	h.enc = encryption.NewEncryptor([]byte(cfg.Encryption.Passphrase), cfg.Encryption.Name, cfg.Encryption.Email, cfg.Encryption.RsaBits, cfg.Encryption.Enable)
	h.svc = service.New(cfg)
	return nil
}

func (h *Handler) Name() string {
	return "live_bff"
}

func (h *Handler) Version() string {
	return version.FullVersion()
}

// @title CossApi

func (h *Handler) RegisterRoute(r gin.IRouter) {
	r.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(h.logger), middleware.EncryptionMiddleware(h.enc), middleware.RecoveryMiddleware())
	r.Use(middleware.AuthMiddleware(h.userCache))
	u := r.Group("/api/v1/live/user")

	u.GET("/show", h.UserShow)
	u.POST("/create", h.UserCreate)
	u.POST("/join", h.UserJoin)
	u.POST("/reject", h.UserReject)
	u.POST("/leave", h.UserLeave)

	g := r.Group("/api/v1/live/group")
	g.GET("/show", h.GroupShow)
	g.POST("/create", h.GroupCreate)
	g.POST("/join", h.GroupJoin)
	g.POST("/reject", h.GroupReject)
	g.POST("/leave", h.GroupLeave)
}

func (h *Handler) Health(r gin.IRouter) string {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) DiscoverServices(services map[string]*grpc.ClientConn) error {
	for k, v := range services {
		if err := h.svc.HandlerGrpcClient(k, v); err != nil {
			h.logger.Error("handler grpc client error", zap.String("name", k), zap.String("addr", v.Target()), zap.Error(err))
		}
	}
	return nil
}

func (h *Handler) Stop(ctx context.Context) error {
	return nil
}
