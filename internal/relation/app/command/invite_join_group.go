package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type InviteJoinGroup struct {
	GroupID    uint32
	UserID     string
	TargetUser []string
}

type InviteJoinGroupHandler decorator.CommandHandlerNoneResponse[*InviteJoinGroup]

func NewInviteJoinGroupHandler(
	logger *zap.Logger,
	groupRequestDomain service.GroupRequestDomain,
	pushService rpc.PushService,
) InviteJoinGroupHandler {
	return &inviteJoinGroupHandler{
		logger:             logger,
		groupRequestDomain: groupRequestDomain,
		pushService:        pushService,
	}
}

type inviteJoinGroupHandler struct {
	logger             *zap.Logger
	groupRequestDomain service.GroupRequestDomain
	pushService        rpc.PushService
}

func (h *inviteJoinGroupHandler) Handle(ctx context.Context, cmd *InviteJoinGroup) error {
	if err := h.groupRequestDomain.VerifyGroupInviteCondition(ctx, cmd.GroupID, cmd.UserID, cmd.TargetUser); err != nil {
		h.logger.Error("verify group invite condition error", zap.Error(err))
		return err
	}

	if err := h.groupRequestDomain.InviteJoinGroup(ctx, cmd.GroupID, cmd.UserID, cmd.TargetUser); err != nil {
		h.logger.Error("invite join group error", zap.Error(err))
		return err
	}

	if err := h.pushService.InviteJoinGroupPush(ctx, cmd.GroupID, cmd.UserID, cmd.TargetUser, cmd.TargetUser); err != nil {
		h.logger.Error("invite join group push error", zap.Error(err))
	}

	return nil
}
