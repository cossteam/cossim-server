package command

import (
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	"github.com/cossim/coss-server/internal/live/domain/live"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/decorator"
	lksdk "github.com/livekit/server-sdk-go"
	"go.uber.org/zap"
	"time"
)

const (
	// PermanentExpiration 永久过期时间
	PermanentExpiration = 0
	// PreviousExpiration 使用之前的过期时间
	PreviousExpiration = -1
)

type CreateRoomHandler decorator.CommandHandler[*CreateRoom, *CreateRoomResponse]

type LiveHandler struct {
	logger   *zap.Logger
	liveRepo live.Repository

	webRtcUrl     string
	liveApiKey    string
	liveApiSecret string
	liveTimeout   time.Duration
	roomService   *lksdk.RoomServiceClient

	msgService           msggrpcv1.MsgServiceClient
	userService          usergrpcv1.UserServiceClient
	pushService          pushgrpcv1.PushServiceClient
	groupService         groupgrpcv1.GroupServiceClient
	relationGroupService relationgrpcv1.GroupRelationServiceClient
	relationUserService  relationgrpcv1.UserRelationServiceClient
}

type LiveHandlerOption func(*LiveHandler)

func WithRepo(repo live.Repository) LiveHandlerOption {
	return func(h *LiveHandler) {
		h.liveRepo = repo
	}
}

func WithLogger(logger *zap.Logger) LiveHandlerOption {
	return func(h *LiveHandler) {
		h.logger = logger
	}
}

func WithLiveKit(c config.LivekitConfig) LiveHandlerOption {
	return func(h *LiveHandler) {
		h.webRtcUrl = c.Url
		h.liveApiKey = c.ApiKey
		h.liveApiSecret = c.ApiSecret
		h.liveTimeout = c.Timeout
		h.roomService = lksdk.NewRoomServiceClient(c.Url, c.ApiKey, c.ApiSecret)
	}
}

func WithMsgService(service msggrpcv1.MsgServiceClient) LiveHandlerOption {
	return func(h *LiveHandler) {
		h.msgService = service
	}
}

func WithUserService(service usergrpcv1.UserServiceClient) LiveHandlerOption {
	return func(h *LiveHandler) {
		h.userService = service
	}
}

func WithPushService(service pushgrpcv1.PushServiceClient) LiveHandlerOption {
	return func(h *LiveHandler) {
		h.pushService = service
	}
}

func WithGroupService(service groupgrpcv1.GroupServiceClient) LiveHandlerOption {
	return func(h *LiveHandler) {
		h.groupService = service
	}
}

func WithRelationGroupService(service relationgrpcv1.GroupRelationServiceClient) LiveHandlerOption {
	return func(h *LiveHandler) {
		h.relationGroupService = service
	}
}

func WithRelationUserService(service relationgrpcv1.UserRelationServiceClient) LiveHandlerOption {
	return func(h *LiveHandler) {
		h.relationUserService = service
	}
}

func NewLiveHandler(options ...LiveHandlerOption) *LiveHandler {
	h := &LiveHandler{}
	for _, option := range options {
		option(h)
	}
	return h
}
