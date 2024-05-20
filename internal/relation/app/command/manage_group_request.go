package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type ManageGroupRequest struct {
	UserID    string
	RequestID uint32
	action    ManageGroupRequestAction
}

type ManageGroupRequestAction uint8

const (
	GroupReject ManageGroupRequestAction = iota
	GroupAccept
)

type ManageGroupRequestHandler decorator.CommandHandlerNoneResponse[*ManageGroupRequest]

func NewManageGroupRequestHandler(
	logger *zap.Logger,
	groupRequestDomain service.GroupRequestDomain,
	groupRelationDomain service.GroupRelationDomain,
	pushService rpc.PushService,
) ManageGroupRequestHandler {
	return &manageGroupRequestHandler{
		logger:              logger,
		groupRequestDomain:  groupRequestDomain,
		groupRelationDomain: groupRelationDomain,
		pushService:         pushService,
	}
}

type manageGroupRequestHandler struct {
	logger              *zap.Logger
	groupRequestDomain  service.GroupRequestDomain
	groupRelationDomain service.GroupRelationDomain
	pushService         rpc.PushService
}

func (h *manageGroupRequestHandler) Handle(ctx context.Context, cmd *ManageGroupRequest) error {
	var action entity.RequestAction
	switch {
	case cmd.action == GroupReject:
		action = entity.Reject
	case cmd.action == GroupAccept:
		action = entity.Accept
	default:
		return code.InvalidParameter
	}

	if err := h.groupRequestDomain.ManageRequest(ctx, cmd.UserID, cmd.RequestID, action); err != nil {
		h.logger.Error("manage group request failed", zap.Error(err))
		return err
	}

	r, err := h.groupRequestDomain.Get(ctx, cmd.RequestID)
	if err != nil {
		return err
	}

	if err := h.pushService.ManageGroupRequestPush(ctx, r.GroupID, cmd.UserID, uint32(action)); err != nil {
		h.logger.Error("push manage group request failed", zap.Error(err))
	}

	return nil
}
