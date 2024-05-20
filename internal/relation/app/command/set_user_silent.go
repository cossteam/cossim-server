package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type SetUserSilent struct {
	CurrentUserID string
	TargetUserID  string
	Silent        bool
}

func (cmd *SetUserSilent) Validate() error {
	if cmd == nil || cmd.CurrentUserID == "" || cmd.TargetUserID == "" {
		return code.InvalidParameter
	}
	return nil
}

type SetUserSilentHandler decorator.CommandHandlerNoneResponse[*SetUserSilent]

func NewSetUserSilentHandler(
	logger *zap.Logger,
	userRelationDomain service.UserRelationDomain,
) SetUserSilentHandler {
	return &setUserSilentHandler{
		logger:             logger,
		userRelationDomain: userRelationDomain,
	}
}

type setUserSilentHandler struct {
	logger             *zap.Logger
	userRelationDomain service.UserRelationDomain
}

func (h *setUserSilentHandler) Handle(ctx context.Context, cmd *SetUserSilent) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	return h.userRelationDomain.SetUserSilent(ctx, cmd.CurrentUserID, cmd.TargetUserID, cmd.Silent)
}
