package http

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/live/api/http/v1"
	"github.com/cossim/coss-server/internal/live/app"
	"github.com/cossim/coss-server/internal/live/app/command"
	"github.com/cossim/coss-server/internal/live/app/query"
	"github.com/cossim/coss-server/internal/user/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	"github.com/cossim/coss-server/pkg/http/response"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/manager/server"
	"github.com/cossim/coss-server/pkg/version"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	omiddleware "github.com/oapi-codegen/gin-middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"os"
	"time"
)

var _ v1.ServerInterface = &HttpServer{}

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
	logger    *zap.Logger
	enc       encryption.Encryptor
	userCache cache.UserCache
	app       app.Application
}

func (h *HttpServer) Init(cfg *pkgconfig.AppConfig) error {
	if cfg.Encryption.Enable {
		return h.enc.ReadKeyPair()
	}
	userCache, err := cache.NewUserCacheRedis(cfg.Redis.Addr(), cfg.Redis.Password, 0)
	if err != nil {
		return err
	}
	h.userCache = userCache
	h.logger = plog.NewDefaultLogger(HttpServiceName, int8(cfg.Log.Level))
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
	r.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(h.logger), middleware.EncryptionMiddleware(h.enc))
	r.Use(middleware.AuthMiddleware(h.userCache))

	swagger, err := v1.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}
	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	validatorOptions := &omiddleware.Options{
		ErrorHandler: func(c *gin.Context, message string, statusCode int) {
			response.SetFail(c, message, nil)
		},
	}

	validatorOptions.Options.AuthenticationFunc = func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		return nil
	}

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	r.Use(omiddleware.OapiRequestValidatorWithOptions(swagger, validatorOptions))

	v1.RegisterHandlers(r, h)
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

// GetGroupRoom
// @Summary 获取群聊通话
// @Description 获取群聊通话信息
// @Tags live
// @Security bearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} Room "获取群聊通话信息成功"
// @Router /live/group/{id} [get]
func (h *HttpServer) GetGroupRoom(c *gin.Context, id uint32) {
	uid := c.Value(constants.UserID).(string)
	rooms, err := h.app.Queries.LiveHandler.GetGroupLive(c, &query.GetGroupLive{UserID: uid, GroupID: id})
	if err != nil {
		c.Error(err)
		return
	}
	if len(rooms) == 0 {
		response.SetSuccess(c, "获取群聊通话信息成功", nil)
		return
	}

	response.SetSuccess(c, "获取群聊通话信息成功", getRoomToResponse(rooms[0]))
}

// GetUserRoom
// @Summary 获取用户通话
// @Description 获取用户通话信息
// @Tags live
// @Security bearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} Room "获取用户通话信息成功"
// @Router /live/user [get]
func (h *HttpServer) GetUserRoom(c *gin.Context) {
	uid := c.Value(constants.UserID).(string)
	rooms, err := h.app.Queries.LiveHandler.GetUserLive(c, &query.GetUserLive{UserID: uid})
	if err != nil {
		c.Error(err)
		return
	}
	if len(rooms) == 0 {
		response.SetSuccess(c, "获取用户通话信息成功", nil)
		return
	}

	response.SetSuccess(c, "获取用户通话信息成功", getRoomToResponse(rooms[0]))
}

// CreateRoom
// @Summary 创建通话
// @Description 创建通话
// @Tags live
// @Security bearerAuth
// @Accept json
// @Produce json
// @Param requestBody body CreateRoomRequest true "请求体参数"
// @Success 200 {object} CreateRoomResponse "创建通话成功"
// @Router /live [post]
func (h *HttpServer) CreateRoom(c *gin.Context) {
	req := &v1.CreateRoomRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.Error(err)
		return
	}

	uid := c.Value(constants.UserID).(string)
	createRoom, err := h.app.Commands.LiveHandler.CreateRoom(c, &command.CreateRoom{
		Creator:      uid,
		Type:         string(req.Type),
		Participants: req.Member,
		GroupID:      req.GroupId,
		Option: command.RoomOption{
			VideoEnabled: req.Option.VideoEnabled,
			AudioEnabled: req.Option.AudioEnabled,
			Resolution:   req.Option.Resolution,
			FrameRate:    req.Option.FrameRate,
			Codec:        req.Option.Codec,
		},
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "创建通话成功", createRoomToResponse(createRoom))
}

func createRoomToResponse(room *command.CreateRoomResponse) *v1.CreateRoomResponse {
	return &v1.CreateRoomResponse{
		Url:  room.Url,
		Room: room.Room,
	}
}

// DeleteRoom
// @Summary 删除通话
// @Description 用户退出当前通话
// @Tags live
// @Security bearerAuth
// @Param id path string true "要退出的通话房间ID"
// @Success 200 {object} Response "成功退出通话"
// @Router /live/{id} [delete]
func (h *HttpServer) DeleteRoom(c *gin.Context, id string) {
	uid := c.Value(constants.UserID).(string)
	did := c.Value(constants.DriverID).(string)
	err := h.app.Commands.LiveHandler.DeleteRoom(c, &command.DeleteRoom{
		Room:     id,
		UserID:   uid,
		DriverID: did,
	})
	if err != nil {
		c.Error(err)
		return
	}
	response.SetSuccess(c, "退出通话成功", nil)
}

// GetRoom
// @Summary 获取通话房间信息
// @Description 获取通话房间信息
// @Tags live
// @Security bearerAuth
// @Param id path string true "要退出的通话房间ID"
// @Produce json
// @Success 200 {object} Room "获取通话信息成功"
// @Router /live/{id} [get]
func (h *HttpServer) GetRoom(c *gin.Context, id string) {
	uid := c.Value(constants.UserID).(string)
	getRoom, err := h.app.Queries.LiveHandler.GetRoom(c, &query.GetRoom{
		Room:   id,
		UserID: uid,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取通话信息成功", getRoomToResponse(getRoom))
}

func getRoomToResponse(room *query.Room) *v1.Room {
	if room == nil {
		return nil
	}
	var participant []v1.ParticipantInfo

	for _, v := range room.Participant {
		participant = append(participant, v1.ParticipantInfo{
			//Uid:         v.Uid,
			Identity:    v.Identity,
			IsPublisher: v.IsPublisher,
			JoinedAt:    v.JoinedAt,
			Name:        v.Name,
			//Room:        v.Room,
			State: v.State,
		})
	}

	return &v1.Room{
		Room:        room.ID,
		Type:        room.Type,
		Duration:    int64(time.Since(time.Unix(room.StartAt, 0)).Seconds()),
		Participant: participant,
		StartAt:     room.StartAt,
	}
}

// JoinRoom
// @Summary 加入通话
// @Description 加入已存在的通话房间
// @Tags live
// @Security bearerAuth
// @Accept json
// @Produce json
// @Param id path string true "要加入的通话房间ID"
// @Param requestBody body JoinRoomRequest true "请求体参数"
// @Success 200 {object} JoinRoomResponse "成功加入通话"
// @Router /live/{id}/join [post]
func (h *HttpServer) JoinRoom(c *gin.Context, id string) {
	uid := c.Value(constants.UserID).(string)
	did := c.Value(constants.DriverID).(string)
	joinRoom, err := h.app.Commands.LiveHandler.JoinRoom(c, &command.JoinRoom{
		Room:     id,
		UserID:   uid,
		DriverID: did,
		Option:   command.RoomOption{},
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "加入通话成功", &v1.JoinRoomResponse{
		Url:   joinRoom.Url,
		Token: joinRoom.Token,
	})
}

// RejectRoom
// @Summary 拒绝通话
// @Description 拒绝加入通话
// @Tags live
// @Security bearerAuth
// @Accept json
// @Produce json
// @Param id path string true "要拒绝的通话房间ID"
// @Success 200 {object} Response "成功拒绝通话"
// @Router /live/{id}/reject [post]
func (h *HttpServer) RejectRoom(c *gin.Context, id string) {
	uid := c.Value(constants.UserID).(string)
	did := c.Value(constants.DriverID).(string)
	rejectLive, err := h.app.Commands.LiveHandler.RejectLive(c, &command.RejectLive{
		Room:     id,
		UserID:   uid,
		DriverID: did,
		Option:   command.RoomOption{},
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "拒绝通话成功", rejectLive)
}
