package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type AddGroupRequest struct {
	GroupID uint32
	UserID  string
	Remark  string
}

type AddGroupRequestHandler decorator.CommandHandlerNoneResponse[*AddGroupRequest]

func NewAddGroupRequestHandler(
	logger *zap.Logger,
	groupRequestDomain service.GroupRequestDomain,
	groupRelationDomain service.GroupRelationDomain,
	pushService rpc.PushService,
) AddGroupRequestHandler {
	return &addGroupRequestHandler{
		logger:              logger,
		groupRequestDomain:  groupRequestDomain,
		groupRelationDomain: groupRelationDomain,
		pushService:         pushService,
	}
}

type addGroupRequestHandler struct {
	logger              *zap.Logger
	groupRequestDomain  service.GroupRequestDomain
	groupRelationDomain service.GroupRelationDomain
	pushService         rpc.PushService
}

func (h *addGroupRequestHandler) Handle(ctx context.Context, cmd *AddGroupRequest) error {
	if err := h.groupRelationDomain.IsUserInActiveGroup(ctx, cmd.GroupID, cmd.UserID); err != nil {
		h.logger.Error("IsUserInActiveGroup error", zap.Error(err))
		return err
	}

	if err := h.groupRequestDomain.AddGroupRequest(ctx, cmd.GroupID, cmd.UserID, cmd.Remark); err != nil {
		h.logger.Error("AddGroupRequest error", zap.Error(err))
		return err
	}

	admins, err := h.groupRelationDomain.ListGroupAdminID(ctx, cmd.GroupID)
	if err != nil {
		h.logger.Error("ListGroupAdminID error", zap.Error(err))
		return err
	}

	if err := h.pushService.AddGroupRequestPush(ctx, cmd.GroupID, admins, cmd.UserID); err != nil {
		h.logger.Error("AddGroupRequestPush error", zap.Error(err))
	}

	return nil
}
