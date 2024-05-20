package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type TopDialog struct {
	UserID   string
	DialogID uint32
	Show     bool
}

func (cmd *TopDialog) Validate() error {
	if cmd == nil || cmd.UserID == "" || cmd.DialogID == 0 {
		return code.InvalidParameter
	}
	return nil
}

type TopDialogHandler decorator.CommandHandlerNoneResponse[*TopDialog]

func NewTopDialogHandler(
	logger *zap.Logger,
	dialogRelationDomain service.DialogRelationDomain,
) TopDialogHandler {
	return &topDialogHandler{
		logger:               logger,
		dialogRelationDomain: dialogRelationDomain,
	}
}

type topDialogHandler struct {
	logger               *zap.Logger
	dialogRelationDomain service.DialogRelationDomain
}

func (h *topDialogHandler) Handle(ctx context.Context, cmd *TopDialog) error {
	//if err := cmd.Validate(); err != nil {
	//	return err
	//}

	if err := h.dialogRelationDomain.TopDialog(ctx, cmd.DialogID, cmd.UserID, cmd.Show); err != nil {
		h.logger.Error("failed to open or close dialog", zap.Error(err))
		return err
	}

	return nil
}
