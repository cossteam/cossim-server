package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type SetUserBurn struct {
	CurrentUserID string
	TargetUserID  string
	Burn          bool
	Timeout       uint32
}

type SetUserBurnHandler decorator.CommandHandlerNoneResponse[*SetUserBurn]

func NewSetUserBurnHandler(
	logger *zap.Logger,
	userRelationDomain service.UserRelationDomain,
) SetUserBurnHandler {
	return &setUserBurnHandler{
		logger:             logger,
		userRelationDomain: userRelationDomain,
	}
}

type setUserBurnHandler struct {
	logger             *zap.Logger
	userRelationDomain service.UserRelationDomain
}

func (h *setUserBurnHandler) Handle(ctx context.Context, cmd *SetUserBurn) error {
	if cmd == nil || cmd.CurrentUserID == "" || cmd.TargetUserID == "" {
		return code.InvalidParameter
	}
	if cmd.Burn && cmd.Timeout == 0 {
		return code.InvalidParameter.CustomMessage("设置消息销毁时间不能为0")
	}

	isFriend, err := h.userRelationDomain.IsFriend(ctx, cmd.CurrentUserID, cmd.TargetUserID)
	if err != nil {
		h.logger.Error("查询好友关系失败", zap.Error(err))
		return err
	}
	if !isFriend {
		return code.RelationUserErrFriendRelationNotFound
	}

	return h.userRelationDomain.SetUserBurn(ctx, cmd.CurrentUserID, cmd.TargetUserID, cmd.Burn, cmd.Timeout)
}
