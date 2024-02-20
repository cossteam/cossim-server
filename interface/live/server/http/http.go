package http

import (
	"context"
	"github.com/cossim/coss-server/interface/live/service"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/server"
	"github.com/cossim/coss-server/pkg/version"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	_ server.HTTPService = &Handler{}
)

type Handler struct {
	svc         *service.Service
	redisClient *redis.Client
	logger      *zap.Logger
	enc         encryption.Encryptor
}

func (h *Handler) Init(cfg *pkgconfig.AppConfig) error {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password, // no password set
		DB:       0,                  // use default DB
		//Protocol: cfg,
	})
	h.redisClient = rdb
	h.logger = plog.NewDevLogger("live_bff")
	h.enc = encryption.NewEncryptor([]byte(cfg.Encryption.Passphrase), cfg.Encryption.Name, cfg.Encryption.Email, cfg.Encryption.RsaBits, cfg.Encryption.Enable)
	h.svc = service.New(cfg)
	return h.enc.ReadKeyPair()
}

func (h *Handler) Name() string {
	return "live_bff"
}

func (h *Handler) Version() string {
	return version.FullVersion()
}

// @title live服务

func (h *Handler) RegisterRoute(r gin.IRouter) {
	u := r.Group("/api/v1/live/user")
	// 为Swagger路径添加不需要身份验证的中间件
	u.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("live")))

	u.Use(middleware.AuthMiddleware(h.redisClient))
	u.GET("/show", h.UserShow)
	u.POST("/create", h.UserCreate)
	u.POST("/join", h.UserJoin)
	u.POST("/reject", h.UserReject)
	u.POST("/leave", h.UserLeave)

	g := r.Group("/api/v1/live/group")
	g.Use(middleware.AuthMiddleware(h.redisClient))
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
