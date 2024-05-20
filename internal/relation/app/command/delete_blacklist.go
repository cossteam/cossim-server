package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type DeleteBlacklist struct {
	CurrentUserID string
	TargetUserID  string
}

type DeleteBlacklistHandler decorator.CommandHandlerNoneResponse[*DeleteBlacklist]

func NewDeleteBlacklistHandler(
	logger *zap.Logger,
	urd service.UserRelationDomain,
	userService rpc.UserService,
) DeleteBlacklistHandler {
	return &deleteBlacklistHandler{
		logger:      logger,
		urd:         urd,
		userService: userService,
	}
}

type deleteBlacklistHandler struct {
	logger      *zap.Logger
	urd         service.UserRelationDomain
	userService rpc.UserService
}

func (h *deleteBlacklistHandler) Handle(ctx context.Context, cmd *DeleteBlacklist) error {
	if cmd == nil {
		return code.InvalidParameter
	}

	if err := h.urd.DeleteBlacklist(ctx, cmd.CurrentUserID, cmd.TargetUserID); err != nil {
		h.logger.Error("delete blacklist failed", zap.Error(err))
		return err
	}

	return nil
}
