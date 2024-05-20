package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type QuitGroup struct {
	UserID  string
	GroupID uint32
}

type QuitGroupHandler decorator.CommandHandlerNoneResponse[*QuitGroup]

func NewQuitGroupHandler(
	logger *zap.Logger,
	groupRelationDomain service.GroupRelationDomain,
) QuitGroupHandler {
	return &quitGroupHandler{
		logger:              logger,
		groupRelationDomain: groupRelationDomain,
	}
}

type quitGroupHandler struct {
	logger              *zap.Logger
	groupRelationDomain service.GroupRelationDomain
}

func (h *quitGroupHandler) Handle(ctx context.Context, cmd *QuitGroup) error {

	if err := h.groupRelationDomain.QuitGroup(ctx, cmd.GroupID, cmd.UserID); err != nil {
		h.logger.Error("error quit group", zap.Error(err))
		return err
	}

	return nil
}
