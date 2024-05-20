package command

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type ManageFriendRequestAction int

const (
	Reject ManageFriendRequestAction = iota
	Accept
)

type ManageFriendRequest struct {
	UserID string
	ID     uint32
	Action ManageFriendRequestAction
}

type ManageFriendRequestHandler decorator.CommandHandlerNoneResponse[*ManageFriendRequest]

func NewManageFriendRequestHandler(
	logger *zap.Logger,
	userFriendRequestService service.UserFriendRequestDomain,
) ManageFriendRequestHandler {
	return &manageFriendRequestHandler{
		logger:                   logger,
		userFriendRequestService: userFriendRequestService,
	}
}

type manageFriendRequestHandler struct {
	logger                   *zap.Logger
	userFriendRequestService service.UserFriendRequestDomain
}

func (h *manageFriendRequestHandler) Handle(ctx context.Context, cmd *ManageFriendRequest) error {
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

	isHandled, err := h.userFriendRequestService.IsHandled(ctx, cmd.ID)
	if err != nil {
		h.logger.Error("can handle friend request error", zap.Error(err))
		return err
	}
	if isHandled {
		return code.DuplicateOperation.CustomMessage("该好友申请已处理")
	}

	switch {
	case cmd.Action == Reject:
		return h.userFriendRequestService.Reject(ctx, cmd.ID)
	case cmd.Action == Accept:
		return h.userFriendRequestService.Accept(ctx, cmd.ID)
	default:
		return code.InvalidParameter.CustomMessage("无效的操作")
	}
}
