package interfaces

import (
	"context"
	"github.com/cossim/coss-server/internal/group/app"
	"github.com/cossim/coss-server/internal/group/app/command"
	"github.com/cossim/coss-server/internal/group/app/query"
	"github.com/cossim/coss-server/internal/user/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	"github.com/cossim/coss-server/pkg/http/response"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/manager/server"
	"github.com/cossim/coss-server/pkg/version"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var _ ServerInterface = &HttpServer{}

var _ server.HTTPService = &HttpServer{}

const (
	HttpServiceName = "group_bff"
)

func NewHttpServer(application app.Application) *HttpServer {
	return &HttpServer{
		app: application,
	}
}

type HttpServer struct {
	app       app.Application
	logger    *zap.Logger
	enc       encryption.Encryptor
	userCache cache.UserCache
}

func (h *HttpServer) Init(cfg *pkgconfig.AppConfig) error {
	h.logger = plog.NewDefaultLogger(HttpServiceName, int8(cfg.Log.Level))
	if cfg.Encryption.Enable {
		return h.enc.ReadKeyPair()
	}
	userCache, err := cache.NewUserCacheRedis(cfg.Redis.Addr(), cfg.Redis.Password, 0)
	if err != nil {
		panic(err)
	}
	h.userCache = userCache
	h.enc = encryption.NewEncryptor([]byte(cfg.Encryption.Passphrase), cfg.Encryption.Name, cfg.Encryption.Email, cfg.Encryption.RsaBits, cfg.Encryption.Enable)
	return nil
}

func (h *HttpServer) Name() string {
	return HttpServiceName
}

func (h *HttpServer) Version() string {
	return version.FullVersion()
}

func (h *HttpServer) RegisterRoute(r gin.IRouter) {
	// 添加一些中间件或其他配置
	r.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(h.logger), middleware.EncryptionMiddleware(h.enc))
	r.Use(middleware.AuthMiddleware(h.userCache))
	RegisterHandlers(r.Group("/api/v1"), h)
}

func (h *HttpServer) Health(r gin.IRouter) string {
	return ""
}

func (h *HttpServer) Stop(ctx context.Context) error {
	return nil
}

func (h *HttpServer) DiscoverServices(services map[string]*grpc.ClientConn) error {
	return nil
}

func (h *HttpServer) GetAllGroup(c *gin.Context) {

}

// CreateGroup
// @Summary 创建群聊
// @Description 创建一个新的群聊
// @Tags group
// @Accept json
// @Produce json
// @Param Authorization header string true "Authentication header"
// @Param requestBody body CreateGroupRequest true "创建一个新的群聊"
// @Security ApiKeyAuth
// @Success 200 {object} Group "成功创建群聊"
// @Router /group [post]
func (h *HttpServer) CreateGroup(c *gin.Context) {
	req := &CreateGroupRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.Error(err)
		return
	}

	uid := c.Value(constants.UserID).(string)
	createGroup, err := h.app.Commands.CreateGroup.Handle(c, command.CreateGroup{
		CreateID: uid,
		Name:     req.Name,
		Avatar:   req.Avatar,
		Type:     req.Type,
		Member:   req.Member,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "创建群聊成功", createGroupToResponse(&createGroup))
}

func createGroupToResponse(createGroup *command.CreateGroupResponse) *CreateGroupResponse {
	return &CreateGroupResponse{
		Avatar:          createGroup.Avatar,
		CreatorId:       createGroup.CreatorID,
		DialogId:        createGroup.DialogID,
		Id:              createGroup.ID,
		MaxMembersLimit: createGroup.MaxMembersLimit,
		Name:            createGroup.Name,
		Status:          createGroup.Status,
		Type:            createGroup.Type,
	}
}

// DeleteGroup
// @Summary 删除群聊
// @Description 群主解散创建的群聊
// @Tags group
// @Accept json
// @Produce json
// @Param id path integer true "要删除的群聊ID" format(uint32)
// @Security ApiKeyAuth
// @Success 200 {string} string "成功删除群聊"
// @Router /group/{id} [delete]
func (h *HttpServer) DeleteGroup(c *gin.Context, id uint32) {
	uid := c.Value(constants.UserID).(string)
	resp, err := h.app.Commands.DeleteGroup.Handle(c, command.DeleteGroup{
		ID:     id,
		UserID: uid,
	})
	if err != nil {
		c.Error(err)
		return
	}
	response.SetSuccess(c, "删除群聊成功", resp)
}

// GetGroup
// @Summary 获取群聊信息
// @Description 根据群聊ID获取群聊的详细信息
// @Tags group
// @Param id path integer true "群聊ID" format(uint32)
// @Success 200 {object} GroupInfo "返回群聊信息"
// @Router /group/{id} [get]
func (h *HttpServer) GetGroup(c *gin.Context, id uint32) {
	getGroup, err := h.app.Queries.GetGroup.Handle(c, query.GetGroup{
		ID:     id,
		UserID: c.Value(constants.UserID).(string),
	})
	if err != nil {
		c.Error(err)
		return
	}
	response.SetSuccess(c, "获取群聊信息成功", getGroupToResponse(getGroup))
}

func getGroupToResponse(getGroup *query.GroupInfo) *GroupInfo {
	return &GroupInfo{
		Avatar:          getGroup.Avatar,
		CreatorId:       getGroup.CreatorId,
		DialogId:        getGroup.DialogId,
		Id:              getGroup.Id,
		MaxMembersLimit: int(getGroup.MaxMembersLimit),
		Name:            getGroup.Name,
		Status:          getGroup.Status,
		Type:            getGroup.Type,
		Preferences: &Preferences{
			EntryMethod:          getGroup.Preferences.EntryMethod,
			Identity:             getGroup.Preferences.Identity,
			Inviter:              getGroup.Preferences.Inviter,
			JoinedAt:             getGroup.Preferences.JoinedAt,
			MuteEndTime:          getGroup.Preferences.MuteEndTime,
			OpenBurnAfterReading: getGroup.Preferences.OpenBurnAfterReading,
			Remark:               getGroup.Preferences.Remark,
			SilentNotification:   getGroup.Preferences.SilentNotification,
		},
	}
}

// UpdateGroup
// @Summary 更新群聊信息
// @Description 更新现有群聊的信息
// @Tags group
// @Accept json
// @Produce json
// @Param id path uint32 true "要更新的群聊ID"
// @Param requestBody body UpdateGroupRequest true "更新现有群聊的信息"
// @Success 200 {object} string "成功更新群聊信息"
// @Router /group/{id} [put]
func (h *HttpServer) UpdateGroup(c *gin.Context, id uint32) {
	req := &UpdateGroupRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.Error(err)
		return
	}

	uid := c.Value(constants.UserID).(string)
	resp, err := h.app.Commands.UpdateGroup.Handle(c, command.UpdateGroup{
		ID:     id,
		Type:   req.Type,
		UserID: uid,
		Name:   req.Name,
		Avatar: req.Avatar,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "更新群聊信息成功", resp)
}
