package http

import (
	"context"
	"github.com/cossim/coss-server/interface/relation/service"
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
	redisClient *cache.RedisClient
	logger      *zap.Logger
	svc         *service.Service
	enc         encryption.Encryptor
}

func (h *Handler) Init(cfg *pkgconfig.AppConfig) error {
	h.setupRedisClient(cfg)
	h.logger = plog.NewDefaultLogger("relation_bff", int8(cfg.Log.Level))
	h.enc = encryption.NewEncryptor([]byte(cfg.Encryption.Passphrase), cfg.Encryption.Name, cfg.Encryption.Email, cfg.Encryption.RsaBits, cfg.Encryption.Enable)
	h.svc = service.New(cfg)
	return h.enc.ReadKeyPair()
}

func (h *Handler) Name() string {
	return "relation_bff"
}

func (h *Handler) Version() string {
	return version.FullVersion()
}

func (h *Handler) setupRedisClient(cfg *pkgconfig.AppConfig) {
	h.redisClient = cache.NewRedisClient(cfg.Redis.Addr(), cfg.Redis.Password)
}

// @title relation服务

func (h *Handler) RegisterRoute(r gin.IRouter) {
	gin.SetMode(gin.ReleaseMode)
	// 添加一些中间件或其他配置
	r.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(h.logger), middleware.EncryptionMiddleware(h.enc), middleware.RecoveryMiddleware())
	// 添加不同的中间件给不同的路由组
	// 比如除了swagger路径外其他的路径添加了身份验证中间件
	api := r.Group("/api/v1/relation")
	// 为Swagger路径添加不需要身份验证的中间件
	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("relation")))
	api.Use(middleware.AuthMiddleware(h.redisClient))

	u := api.Group("/user")
	u.GET("/friend_list", h.friendList)
	u.GET("/blacklist", h.blackList)
	u.GET("/request_list", h.userRequestList)
	u.POST("/add_friend", h.addFriend)
	u.POST("/manage_friend", h.manageFriend)
	u.POST("/delete_friend", h.deleteFriend)
	u.POST("/add_blacklist", h.addBlacklist)
	u.POST("/delete_blacklist", h.deleteBlacklist)
	u.POST("/switch/e2e/key", h.switchUserE2EPublicKey)
	//设置用户静默通知
	u.POST("/silent", h.setUserSilentNotification)
	u.POST("/burn/open", h.openUserBurnAfterReading)
	u.POST("/burn/timeout/set", h.setUserOpenBurnAfterReadingTimeOut)

	u.POST("/remark/set", h.setUserFriendRemark) //设置好友备注

	g := api.Group("/group")
	g.GET("/member", h.getGroupMember)
	g.GET("/request_list", h.groupRequestList)
	// 邀请好友加入群聊
	g.POST("/invite", h.inviteGroup)
	// 申请加入群聊
	g.POST("/join", h.joinGroup)
	// 用户加入或拒绝群聊
	g.POST("/manage_join", h.manageJoinGroup)
	//获取用户群聊列表
	g.GET("/list", h.getUserGroupList)
	// 退出群聊
	g.POST("quit", h.quitGroup)
	//群聊设置消息静默
	g.POST("/silent", h.setGroupSilentNotification)
	//关闭或打开阅后即焚消息
	g.POST("/burn/open", h.openGroupBurnAfterReading)
	g.POST("/burn/timeout/set", h.setGroupOpenBurnAfterReadingTimeOut)

	//获取群聊公告列表
	g.GET("/announcement/list", h.getGroupAnnouncementList)
	//获取群公告详情
	g.GET("/announcement/detail", h.getGroupAnnouncementDetail)
	//设置群公告为已读
	g.POST("/announcement/read", h.readGroupAnnouncement)
	//获取群公告已读列表
	g.GET("/announcement/read/list", h.getReadGroupAnnouncementList)

	gg := api.Group("/group/admin")
	// 管理员管理群聊申请
	gg.POST("/manage/join", h.adminManageJoinGroup)
	// 管理员移除群聊成员
	gg.POST("/manage/remove", h.removeUserFromGroup)
	//创建群公告
	gg.POST("/announcement", h.createGroupAnnouncement)
	//修改群公告
	gg.POST("/announcement/update", h.updateGroupAnnouncement)
	//删除群公告
	gg.POST("/announcement/delete", h.deleteGroupAnnouncement)

	d := api.Group("/dialog")
	d.POST("/top", h.topOrCancelTopDialog)
	d.POST("/show", h.closeOrOpenDialog)
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
