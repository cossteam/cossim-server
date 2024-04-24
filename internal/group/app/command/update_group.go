package command

import (
	"context"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	"github.com/cossim/coss-server/internal/group/domain/group"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
	"time"
)

type UpdateGroup struct {
	ID          uint32
	Type        *uint
	UserID      string
	Name        *string
	Avatar      *string
	SilenceTime *int64
	Encrypt     *bool
	JoinApprove *bool
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
	groupRepo            group.Repository
	logger               *zap.Logger
	relationGroupService RelationGroupService
	groupService         GroupService
}

func NewUpdateGroupHandler(
	repo group.Repository,
	logger *zap.Logger,
	dtmGrpcServer string,
	relationGroupService RelationGroupService,
	groupService GroupService,
) UpdateGroupHandler {
	if repo == nil {
		panic("nil repo")
	}

	h := &updateGroupHandler{
		groupRepo:            repo,
		logger:               logger,
		relationGroupService: relationGroupService,
		groupService:         groupService,
	}
	return h
}

func (h *updateGroupHandler) Handle(ctx context.Context, cmd UpdateGroup) (*UpdateGroupResponse, error) {
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

	updateFields := map[string]interface{}{}
	if cmd.Encrypt != nil && *cmd.Type == uint(groupgrpcv1.GroupType_Private) {
		updateFields["encrypt"] = *cmd.Encrypt
	}
	if cmd.Name != nil {
		updateFields["name"] = *cmd.Name
	}
	if cmd.Avatar != nil {
		updateFields["name"] = *cmd.Avatar
	}
	if cmd.SilenceTime != nil {
		if *cmd.SilenceTime < time.Now().Unix() {
			return nil, code.MyCustomErrorCode.CustomMessage("silence_time cannot be in the past")
		}
		updateFields["silence_time"] = *cmd.SilenceTime
	}
	if cmd.JoinApprove != nil {
		updateFields["join_approve"] = *cmd.JoinApprove
	}

	if len(updateFields) == 0 {
		return nil, code.MyCustomErrorCode.CustomMessage("no update fields")
	}

	if err := h.groupRepo.UpdateFields(ctx, r.ID, updateFields); err != nil {
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
