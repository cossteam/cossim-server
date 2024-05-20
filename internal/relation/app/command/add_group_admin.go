package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type AddGroupAdmin struct {
	UserID      string
	GroupID     uint32
	TargetUsers []string
}

type AddGroupAdminHandler decorator.CommandHandlerNoneResponse[*AddGroupAdmin]

func NewAddGroupAdminHandler(
	logger *zap.Logger,
	groupRelationDomain service.GroupRelationDomain,
) AddGroupAdminHandler {
	return &addGroupAdminHandler{
		logger:              logger,
		groupRelationDomain: groupRelationDomain,
	}
}

type addGroupAdminHandler struct {
	logger              *zap.Logger
	groupRelationDomain service.GroupRelationDomain
}

func (h *addGroupAdminHandler) Handle(ctx context.Context, cmd *AddGroupAdmin) error {
	if err := h.groupRelationDomain.AddGroupAdmin(ctx, cmd.GroupID, cmd.UserID, cmd.TargetUsers...); err != nil {
		h.logger.Error("add group admin error", zap.Error(err))
		return err
	}
	return nil
}
