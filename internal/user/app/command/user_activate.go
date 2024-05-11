package command

import (
	"context"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/service"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type UserActivate struct {
	UserID           string
	VerificationCode string
}

type UserActivateHandler decorator.CommandHandler[*UserActivate, *interface{}]

func NewUserActivateHandler(logger *zap.Logger, ud service.UserDomain, userCache cache.UserCache) UserActivateHandler {
	return &userActivateHandler{
		logger:    logger,
		ud:        ud,
		userCache: userCache,
	}
}

type userActivateHandler struct {
	logger    *zap.Logger
	ud        service.UserDomain
	userCache cache.UserCache
}

func (h *userActivateHandler) Handle(ctx context.Context, cmd *UserActivate) (*interface{}, error) {
	verificationCode, err := h.userCache.GetUserEmailVerificationCode(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	if verificationCode != cmd.VerificationCode {
		return nil, code.MyCustomErrorCode.CustomMessage("验证码不存在或已过期")
	}

	if err := h.ud.ActivateUser(ctx, cmd.UserID); err != nil {
		h.logger.Error("激活用户失败", zap.Error(err))
		return nil, code.UserErrActivateUserFailed
	}

	if err := h.userCache.DeleteUserEmailVerificationCode(ctx, cmd.UserID); err != nil {
		h.logger.Error("删除用户邮箱激活验证码失败", zap.Error(err))
	}

	return nil, nil
}
