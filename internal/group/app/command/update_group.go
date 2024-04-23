package command

import (
	"context"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	"github.com/cossim/coss-server/internal/group/cache"
	"github.com/cossim/coss-server/internal/group/domain/group"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type UpdateGroup struct {
	ID     uint32 `json:"id"`
	Type   uint32 `json:"type"`
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

const (
	DefaultGroup   = 1000 //默认群
	EncryptedGroup = 500  //加密群
)

type UpdateGroupResponse struct {
}

type UpdateGroupHandler decorator.CommandHandler[UpdateGroup, *UpdateGroupResponse]

var _ decorator.CommandHandler[UpdateGroup, *UpdateGroupResponse] = &updateGroupHandler{}

type updateGroupHandler struct {
	groupRepo group.Repository
	logger    *zap.Logger

	cache       cache.GroupCache
	cacheEnable bool

	relationGroupService RelationGroupService
	groupService         GroupService
}

func NewUpdateGroupHandler(
	repo group.Repository,
	logger *zap.Logger,
	dtmGrpcServer string,
	relationGroupService RelationGroupService,
	groupService GroupService,
	cacheEnable bool,
	cache cache.GroupCache,
) UpdateGroupHandler {
	if repo == nil {
		panic("nil repo")
	}

	h := &updateGroupHandler{
		groupRepo:            repo,
		logger:               logger,
		relationGroupService: relationGroupService,
		groupService:         groupService,
		cacheEnable:          cacheEnable,
		cache:                cache,
	}
	return h
}

func (h *updateGroupHandler) Handle(ctx context.Context, cmd UpdateGroup) (*UpdateGroupResponse, error) {
	isValidGroupType := func(value groupgrpcv1.GroupType) bool {
		return value == groupgrpcv1.GroupType_TypeEncrypted || value == groupgrpcv1.GroupType_TypeDefault
	}
	if !isValidGroupType(groupgrpcv1.GroupType(cmd.Type)) {
		return nil, code.InvalidParameter
	}

	owner, err := h.relationGroupService.IsGroupOwner(ctx, cmd.ID, cmd.UserID)
	if err != nil {
		return nil, err
	}

	if !owner {
		return nil, code.Forbidden
	}

	r, err := h.groupService.Get(ctx, cmd.ID)
	if err != nil {
		return nil, err
	}

	if r == nil {
		return nil, code.GroupErrGroupNotFound
	}

	var MaxMembersLimit int

	switch cmd.Type {
	case uint32(groupgrpcv1.GroupType_TypeEncrypted):
		MaxMembersLimit = EncryptedGroup
	default:
		MaxMembersLimit = DefaultGroup
	}

	if r.Name != cmd.Name && cmd.Name != "" {
		r.Name = cmd.Name
	}
	if r.Avatar != cmd.Avatar && cmd.Avatar != "" {
		r.Avatar = cmd.Avatar
	}
	if r.Type != cmd.Type {
		r.Type = cmd.Type
	}

	if err := h.groupRepo.Update(ctx, &group.Group{
		ID:              r.ID,
		Type:            group.Type(r.Type),
		MaxMembersLimit: MaxMembersLimit,
		Name:            r.Name,
		Avatar:          r.Avatar,
		CreatorID:       r.CreatorID,
	}, func(e *group.Group) (*group.Group, error) {
		if h.cacheEnable {
			if err := h.cache.DeleteGroup(ctx, e.ID); err != nil {
				h.logger.Error("delete group cache error", zap.Error(err))
			}
		}
		return nil, nil
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

//func (h *updateGroupHandler) HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) {
//	addr := conn.Target()
//	switch serviceName {
//	case app.UserServiceName:
//		//h.userService = adapters.NewUserGrpc(usergrpcv1.NewUserServiceClient(conn))
//	case app.GroupServiceName:
//		h.groupService = adapters.NewGroupGrpc(groupgrpcv1.NewGroupServiceClient(conn))
//	case app.RelationGroupServiceName:
//		h.relationGroupService = adapters.NewRelationGroupGrpc(relationgrpcv1.NewGroupRelationServiceClient(conn))
//	case app.PushServiceName:
//		//h.pushService = adapters.NewPushService(pushgrpcv1.NewPushServiceClient(conn))
//	default:
//	}
//	h.logger.Info("gRPC client service initialized", zap.String("service", serviceName), zap.String("addr", addr))
//}
