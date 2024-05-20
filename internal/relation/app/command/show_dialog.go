package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type ShowDialog struct {
	UserID   string
	DialogID uint32
	Show     bool
}

func (cmd *ShowDialog) Validate() error {
	if cmd == nil || cmd.UserID == "" || cmd.DialogID == 0 {
		return code.InvalidParameter
	}
	return nil
}

type ShowDialogHandler decorator.CommandHandlerNoneResponse[*ShowDialog]

func NewShowDialogHandler(
	logger *zap.Logger,
	dialogRelationDomain service.DialogRelationDomain,
) ShowDialogHandler {
	return &showDialogHandler{
		logger:               logger,
		dialogRelationDomain: dialogRelationDomain,
	}
}

type showDialogHandler struct {
	logger               *zap.Logger
	dialogRelationDomain service.DialogRelationDomain
}

func (h *showDialogHandler) Handle(ctx context.Context, cmd *ShowDialog) error {
	//if err := cmd.Validate(); err != nil {
	//	return err
	//}

	if err := h.dialogRelationDomain.OpenOrCloseDialog(ctx, cmd.DialogID, cmd.UserID, cmd.Show); err != nil {
		h.logger.Error("failed to open or close dialog", zap.Error(err))
		return err
	}

	return nil
}
