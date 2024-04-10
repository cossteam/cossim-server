package http

import (
	"context"
	"fmt"
	grpcHandler "github.com/cossim/coss-server/internal/msg/interface/grpc"
	"github.com/cossim/coss-server/internal/msg/service"
	"github.com/cossim/coss-server/pkg/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/manager/server"
	"github.com/cossim/coss-server/pkg/version"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
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
	MsgClient *grpcHandler.Handler
	userCache cache.UserCache
}

func (h *Handler) Init(cfg *pkgconfig.AppConfig) error {
	h.logger = plog.NewDefaultLogger("msg_bff", int8(cfg.Log.Level))
	userCache := cache.NewUserCacheRedisWithClient(redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s", cfg.Redis.Addr()),
		Password: cfg.Redis.Password,
		DB:       0,
	}))
	h.userCache = userCache
	h.enc = encryption.NewEncryptor([]byte(cfg.Encryption.Passphrase), cfg.Encryption.Name, cfg.Encryption.Email, cfg.Encryption.RsaBits, cfg.Encryption.Enable)
	h.svc = service.New(cfg, h.MsgClient)
	if cfg.Encryption.Enable {
		return h.enc.ReadKeyPair()
	}
	return nil
}

func (h *Handler) Name() string {
	return "msg_bff"
}

func (h *Handler) Version() string {
	return version.FullVersion()
}

// @title CossApi

func (h *Handler) RegisterRoute(r gin.IRouter) {
	u := r.Group("/api/v1/msg")
	u.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(h.logger), middleware.EncryptionMiddleware(h.enc), middleware.RecoveryMiddleware())
	u.Use(middleware.AuthMiddleware(h.userCache))

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
	u.POST("/read/group", h.setGroupMessagesRead)
	//获取群聊消息阅读者
	u.GET("/read/group", h.getGroupMessageReaders)
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
