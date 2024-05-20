package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type RemoveGroupMember struct {
	GroupID       uint32
	CurrentUserID string
	RemoveMember  []string
}

type RemoveGroupMemberHandler decorator.CommandHandlerNoneResponse[*RemoveGroupMember]

func NewRemoveGroupMemberHandler(
	logger *zap.Logger,
	groupRelationDomain service.GroupRelationDomain,
) RemoveGroupMemberHandler {
	return &removeGroupMemberHandler{
		logger:              logger,
		groupRelationDomain: groupRelationDomain,
	}
}

type removeGroupMemberHandler struct {
	logger              *zap.Logger
	groupRelationDomain service.GroupRelationDomain
}

func (h *removeGroupMemberHandler) Handle(ctx context.Context, cmd *RemoveGroupMember) error {

	if err := h.groupRelationDomain.RemoveGroupMember(ctx, cmd.GroupID, cmd.CurrentUserID, cmd.RemoveMember...); err != nil {
		h.logger.Error("remove group member error", zap.Error(err))
	}

	return nil
}
