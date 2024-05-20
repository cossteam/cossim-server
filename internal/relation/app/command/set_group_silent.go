package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type SetGroupSilent struct {
	UserID  string
	GroupID uint32
	Silent  bool
}

func (cmd *SetGroupSilent) Validate() error {
	if cmd == nil || cmd.UserID == "" || cmd.GroupID == 0 {
		return code.InvalidParameter
	}
	return nil
}

type SetGroupSilentHandler decorator.CommandHandlerNoneResponse[*SetGroupSilent]

func NewSetGroupSilentHandler(
	logger *zap.Logger,
	groupRelationDomain service.GroupRelationDomain,
) SetGroupSilentHandler {
	return &setGroupSilentHandler{
		logger:              logger,
		groupRelationDomain: groupRelationDomain,
	}
}

type setGroupSilentHandler struct {
	logger              *zap.Logger
	groupRelationDomain service.GroupRelationDomain
}

func (h *setGroupSilentHandler) Handle(ctx context.Context, cmd *SetGroupSilent) error {
	//if err := cmd.Validate(); err != nil {
	//	return err
	//}

	if err := h.groupRelationDomain.SetGroupSilent(ctx, cmd.UserID, cmd.GroupID, cmd.Silent); err != nil {
		h.logger.Error("set group silent error", zap.Error(err))
		return err
	}

	return nil
}
