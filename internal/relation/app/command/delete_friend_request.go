package command

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type DeleteFriendRequest struct {
	UserID string
	ID     uint32
}

type DeleteFriendRequestHandler decorator.CommandHandlerNoneResponse[*DeleteFriendRequest]

func NewDeleteFriendRequestHandler(
	logger *zap.Logger,
	userFriendRequestService service.UserFriendRequestDomain,
) DeleteFriendRequestHandler {
	return &deleteFriendRequestHandler{
		logger:                   logger,
		userFriendRequestService: userFriendRequestService,
	}
}

type deleteFriendRequestHandler struct {
	logger                   *zap.Logger
	userFriendRequestService service.UserFriendRequestDomain
}

func (h *deleteFriendRequestHandler) Handle(ctx context.Context, cmd *DeleteFriendRequest) error {
	if cmd == nil {
		return code.InvalidParameter
	}

	is, err := h.userFriendRequestService.IsMy(ctx, cmd.ID, cmd.UserID)
	if err != nil {
		if errors.Is(code.Cause(err), code.NotFound) {
			return code.NotFound.CustomMessage("好友申请不存在")
		}
		return err
	}

	if !is {
		return code.Forbidden
	}

	if err := h.userFriendRequestService.Delete(ctx, cmd.ID); err != nil {
		h.logger.Error("delete friend request error", zap.Error(err))
		return err
	}

	return nil
}
