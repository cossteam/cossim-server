package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type DeleteGroupRequest struct {
	UserID    string
	GroupID   uint32
	RequestID uint32
}

type DeleteGroupRequestHandler decorator.CommandHandlerNoneResponse[*DeleteGroupRequest]

func NewDeleteGroupRequestHandler(
	logger *zap.Logger,
	groupRequestDomain service.GroupRequestDomain,
) DeleteGroupRequestHandler {
	return &deleteGroupRequestHandler{
		logger:             logger,
		groupRequestDomain: groupRequestDomain,
	}
}

type deleteGroupRequestHandler struct {
	logger             *zap.Logger
	groupRequestDomain service.GroupRequestDomain
}

func (h *deleteGroupRequestHandler) Handle(ctx context.Context, cmd *DeleteGroupRequest) error {

	if err := h.groupRequestDomain.Delete(ctx, cmd.RequestID, cmd.UserID); err != nil {
		h.logger.Error("delete group request error", zap.Error(err))
	}

	return nil
}
