package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type SetUserRemark struct {
	CurrentUserID string
	TargetUserID  string
	Remark        string
}

type SetUserRemarkHandler decorator.CommandHandlerNoneResponse[*SetUserRemark]

func NewSetUserRemarkHandler(
	logger *zap.Logger,
	userRelationDomain service.UserRelationDomain,
) SetUserRemarkHandler {
	return &setUserRemarkHandler{
		logger:             logger,
		userRelationDomain: userRelationDomain,
	}
}

type setUserRemarkHandler struct {
	logger             *zap.Logger
	userRelationDomain service.UserRelationDomain
}

func (h *setUserRemarkHandler) Handle(ctx context.Context, cmd *SetUserRemark) error {
	if cmd == nil || cmd.CurrentUserID == "" || cmd.TargetUserID == "" {
		return code.InvalidParameter
	}

	isFriend, err := h.userRelationDomain.IsFriend(ctx, cmd.CurrentUserID, cmd.TargetUserID)
	if err != nil {
		h.logger.Error("查询好友关系失败", zap.Error(err))
		return err
	}
	if !isFriend {
		return code.RelationUserErrFriendRelationNotFound
	}

	return h.userRelationDomain.SetUserRemark(ctx, cmd.CurrentUserID, cmd.TargetUserID, cmd.Remark)
}
