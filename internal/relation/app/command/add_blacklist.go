package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type AddBlacklist struct {
	CurrentUserID string
	TargetUserID  string
}

type AddBlacklistHandler decorator.CommandHandlerNoneResponse[*AddBlacklist]

func NewAddBlacklistHandler(
	logger *zap.Logger,
	urd service.UserRelationDomain,
	userService rpc.UserService,
) AddBlacklistHandler {
	return &addBlacklistHandler{
		logger:      logger,
		urd:         urd,
		userService: userService,
	}
}

type addBlacklistHandler struct {
	logger      *zap.Logger
	urd         service.UserRelationDomain
	userService rpc.UserService
}

func (h *addBlacklistHandler) Handle(ctx context.Context, cmd *AddBlacklist) error {
	if cmd == nil {
		return code.InvalidParameter
	}

	if err := h.urd.AddBlacklist(ctx, cmd.CurrentUserID, cmd.TargetUserID); err != nil {
		h.logger.Error("add blacklist failed", zap.Error(err))
		return err
	}

	return nil
}
