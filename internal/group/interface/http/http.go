package http

import (
	"context"
	mygrpc "github.com/cossim/coss-server/internal/group/interface/grpc"
	"github.com/cossim/coss-server/internal/group/service"

	"github.com/cossim/coss-server/pkg/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/server"
	"github.com/cossim/coss-server/pkg/version"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net/http"
)

var (
	_ server.HTTPService = &Handler{}
)

type Handler struct {
	redisClient *cache.RedisClient
	logger      *zap.Logger
	enc         encryption.Encryptor
	svc         *service.Service
	server      *http.Server
	engine      *gin.Engine
	GrpcService *mygrpc.Handler
}

func (h *Handler) Init(cfg *pkgconfig.AppConfig) error {
	h.setupRedisClient(cfg)
	h.logger = plog.NewDefaultLogger("group_bff", int8(cfg.Log.Level))
	h.enc = encryption.NewEncryptor([]byte(cfg.Encryption.Passphrase), cfg.Encryption.Name, cfg.Encryption.Email, cfg.Encryption.RsaBits, cfg.Encryption.Enable)
	h.svc = service.New(cfg, h.GrpcService)
	return h.enc.ReadKeyPair()
}

func (h *Handler) Name() string {
	return "group_bff"
}

func (h *Handler) setupRedisClient(cfg *pkgconfig.AppConfig) {
	h.redisClient = cache.NewRedisClient(cfg.Redis.Addr(), cfg.Redis.Password)
}

func (h *Handler) Version() string {
	return version.FullVersion()
}

// @title coss-user服务

func (h *Handler) RegisterRoute(r gin.IRouter) {
	// 添加一些中间件或其他配置
	r.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(h.logger), middleware.EncryptionMiddleware(h.enc), middleware.RecoveryMiddleware())
	g := r.Group("/api/v1/group")
	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("group")))
	g.Use(middleware.AuthMiddleware(h.redisClient))
	// 获取群聊信息
	g.GET("/info", h.getGroupInfoByGid)
	// 创建群聊
	g.POST("/create", h.createGroup)
	// 删除群聊
	g.POST("/delete", h.deleteGroup)
	// 更新群聊信息
	g.POST("/update", h.updateGroup)
	//修改群聊头像
	g.POST("/avatar/modify", h.modifyGroupAvatar)
}

func (h *Handler) Health(r gin.IRouter) string {
	return ""
}

func (h *Handler) Stop(ctx context.Context) error {
	return h.svc.Stop()
}

func (h *Handler) DiscoverServices(services map[string]*grpc.ClientConn) error {
	for k, v := range services {
		if err := h.svc.HandlerGrpcClient(k, v.Target()); err != nil {
			h.logger.Error("handler grpc client error", zap.String("name", k), zap.String("addr", v.Target()), zap.Error(err))
		}
	}
	return nil
}
