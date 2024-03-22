package http

import (
	"context"
	grpcHandler "github.com/cossim/coss-server/internal/msg/interface/grpc"
	"github.com/cossim/coss-server/internal/msg/service"
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
)

var (
	_ server.HTTPService = &Handler{}
)

type Handler struct {
	svc         *service.Service
	redisClient *cache.RedisClient
	logger      *zap.Logger
	enc         encryption.Encryptor
	MsgClient   *grpcHandler.Handler
}

func (h *Handler) Init(cfg *pkgconfig.AppConfig) error {
	h.setupRedisClient(cfg)
	h.logger = plog.NewDefaultLogger("msg_bff", int8(cfg.Log.Level))
	h.enc = encryption.NewEncryptor([]byte(cfg.Encryption.Passphrase), cfg.Encryption.Name, cfg.Encryption.Email, cfg.Encryption.RsaBits, cfg.Encryption.Enable)
	h.svc = service.New(cfg, h.MsgClient)

	return h.enc.ReadKeyPair()
}

func (h *Handler) Name() string {
	return "msg_bff"
}

func (h *Handler) Version() string {
	return version.FullVersion()
}

func (h *Handler) setupRedisClient(cfg *pkgconfig.AppConfig) {
	h.redisClient = cache.NewRedisClient(cfg.Redis.Addr(), cfg.Redis.Password)
}

// @title CossApi

func (h *Handler) RegisterRoute(r gin.IRouter) {
	u := r.Group("/api/v1/msg")
	u.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(h.logger), middleware.EncryptionMiddleware(h.enc), middleware.RecoveryMiddleware())
	u.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("msg")))
	u.Use(middleware.AuthMiddleware(h.redisClient))

	u.GET("/ws", h.ws)
	u.POST("/send/user", h.sendUserMsg)
	u.POST("/send/group", h.sendGroupMsg)
	u.GET("/list/user", h.getUserMsgList)
	u.GET("/list/group", h.getGroupMsgList)
	u.GET("/dialog/list", h.getUserDialogList)
	u.POST("/recall/group", h.recallGroupMsg)
	u.POST("/recall/user", h.recallUserMsg)
	u.POST("/edit/group", h.editGroupMsg)
	u.POST("/edit/user", h.editUserMsg)
	u.POST("/read/user", h.readUserMsgs)

	//群聊标注消息
	u.POST("/label/group", h.labelGroupMessage)
	u.GET("/label/group", h.getGroupLabelMsgList)
	//私聊标注消息
	u.POST("/label/user", h.labelUserMessage)
	u.GET("/label/user", h.getUserLabelMsgList)
	u.POST("/after/get", h.getDialogAfterMsg)
	//群聊设置消息已读
	u.POST("/group/read/set", h.setGroupMessagesRead)
	//获取群聊消息阅读者
	u.GET("/group/read/get", h.getGroupMessageReaders)
}

func (h *Handler) Health(r gin.IRouter) string {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) Stop(ctx context.Context) error {
	return h.svc.Stop(ctx)
}

func (h *Handler) DiscoverServices(services map[string]*grpc.ClientConn) error {
	for k, v := range services {
		if err := h.svc.HandlerGrpcClient(k, v); err != nil {
			h.logger.Error("handler grpc client error", zap.String("name", k), zap.String("addr", v.Target()), zap.Error(err))
		}
	}
	return nil
}
