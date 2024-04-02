package http

import (
	"context"
	grpchandler "github.com/cossim/coss-server/internal/relation/interface/grpc"
	"github.com/cossim/coss-server/internal/relation/service"
	"github.com/cossim/coss-server/pkg/cache"
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
	redisClient     *cache.RedisClient
	logger          *zap.Logger
	svc             *service.Service
	enc             encryption.Encryptor
	RelationService *grpchandler.RelationServiceServer
}

func (h *Handler) Init(cfg *pkgconfig.AppConfig) error {
	h.setupRedisClient(cfg)
	h.logger = plog.NewDefaultLogger("relation_bff", int8(cfg.Log.Level))
	h.enc = encryption.NewEncryptor([]byte(cfg.Encryption.Passphrase), cfg.Encryption.Name, cfg.Encryption.Email, cfg.Encryption.RsaBits, cfg.Encryption.Enable)
	h.svc = service.New(cfg, h.RelationService)
	//return h.enc.ReadKeyPair()
	return nil
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

// @title CossApi

func (h *Handler) RegisterRoute(r gin.IRouter) {
	gin.SetMode(gin.ReleaseMode)
	r.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(h.logger), middleware.EncryptionMiddleware(h.enc), middleware.RecoveryMiddleware())
	api := r.Group("/api/v1/relation")
	api.Use(middleware.AuthMiddleware(h.redisClient))

	u := api.Group("/user")
	u.GET("/friend_list", h.friendList)
	u.GET("/blacklist", h.blackList)
	u.GET("/request_list", h.userRequestList)
	u.POST("/add_friend", h.addFriend)
	u.POST("/manage_friend", h.manageFriend)
	u.POST("/delete_friend", h.deleteFriend)
	u.POST("/delete_request_record", h.deleteUserRequestRecord)
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
	g.POST("/delete_request_record", h.deleteGroupRequestRecord)
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

	g.POST("/remark/set", h.setGroupUserRemark)

	gg := api.Group("/group/admin")
	// 设置群聊管理员
	gg.POST("/add", h.addGroupAdmin)
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
