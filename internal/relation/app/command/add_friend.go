package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type AddFriend struct {
	CurrentUserID string
	TargetUserID  string
	Remark        string
	E2ePublicKey  string
}

type AddFriendHandler decorator.CommandHandlerNoneResponse[*AddFriend]

func NewAddFriendHandler(
	logger *zap.Logger,
	userRelationService service.UserRelationDomain,
	userFriendRequestDomain service.UserFriendRequestDomain,
	pushService rpc.PushService,
) AddFriendHandler {
	return &addFriendHandler{
		logger:                  logger,
		userRelationDomain:      userRelationService,
		userFriendRequestDomain: userFriendRequestDomain,
		pushService:             pushService,
	}
}

type addFriendHandler struct {
	logger                  *zap.Logger
	userRelationDomain      service.UserRelationDomain
	userFriendRequestDomain service.UserFriendRequestDomain
	pushService             rpc.PushService
}

func (h *addFriendHandler) Handle(ctx context.Context, cmd *AddFriend) error {
	if cmd == nil {
		return code.InvalidParameter
	}

	h.logger.Info("add friend command", zap.Any("cmd", cmd))

	isFriend, err := h.userRelationDomain.IsFriend(ctx, cmd.CurrentUserID, cmd.TargetUserID)
	if err != nil {
		h.logger.Error("check friend error", zap.Error(err))
		return err
	}
	if isFriend {
		return code.RelationErrAlreadyFriends
	}

	fmt.Println(cmd.TargetUserID, cmd.CurrentUserID)
	is1, err := h.userRelationDomain.IsFriend(ctx, cmd.TargetUserID, cmd.CurrentUserID)
	if err != nil {
		h.logger.Error("check friend error", zap.Error(err))
		return err
	}
	// 如果对方存在好友关系，说明是当前用户单删了，重新恢复好友关系
	if is1 {
		if err := h.userRelationDomain.AddFriendAfterDelete(ctx, cmd.CurrentUserID, cmd.TargetUserID); err != nil {
			h.logger.Error("单删添加好友失败", zap.Error(err))
			return err
		}
		return nil
	}

	_, err = h.userFriendRequestDomain.CanSendRequest(ctx, cmd.CurrentUserID, cmd.TargetUserID)
	if err != nil {
		if !errors.Is(err, code.NotFound) {
			h.logger.Error("check can send friend request error", zap.Error(err))
			return err
		}
	}

	// 创建好友申请
	if err := h.userFriendRequestDomain.CreateFriendRequest(ctx, cmd.CurrentUserID, cmd.TargetUserID); err != nil {
		return err
	}

	// 推送ws事件
	if err := h.pushService.AddFriendPush(ctx, cmd.CurrentUserID, cmd.TargetUserID, cmd.Remark, cmd.E2ePublicKey); err != nil {
		h.logger.Error("push ws error", zap.Error(err))
	}

	return nil
}
