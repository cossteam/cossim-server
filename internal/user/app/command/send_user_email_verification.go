package command

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/domain/service"
	"github.com/cossim/coss-server/internal/user/infra/remote"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"github.com/cossim/coss-server/pkg/utils"
	"go.uber.org/zap"
)

type SendUserEmailVerification struct {
	UserID string
	Email  string
	Type   string
}

type SendUserEmailVerificationHandler decorator.CommandHandlerNoneResponse[*SendUserEmailVerification]

func NewSendUserEmailVerificationHandler(
	logger *zap.Logger,
	ud service.UserDomain,
	userCache cache.UserCache,
	smtpService remote.SmtpService,
) SendUserEmailVerificationHandler {
	return &sendUserEmailVerificationHandler{
		logger:      logger,
		ud:          ud,
		userCache:   userCache,
		smtpService: smtpService,
	}
}

type sendUserEmailVerificationHandler struct {
	logger      *zap.Logger
	userCache   cache.UserCache
	ud          service.UserDomain
	smtpService remote.SmtpService
}

func (h *sendUserEmailVerificationHandler) Handle(ctx context.Context, cmd *SendUserEmailVerification) error {
	if cmd == nil || cmd.Email == "" {
		return code.InvalidParameter
	}

	user, err := h.ud.GetUserWithOpts(ctx, entity.WithEmail(cmd.Email))
	if err != nil {
		h.logger.Error("get user error", zap.Error(err))
		if errors.Is(code.Cause(err), code.NotFound) {
			return code.UserErrNotExist
		}
		return err
	}

	verifCode := utils.RandomNum()
	if err := h.userCache.SetUserVerificationCode(ctx, user.ID, verifCode, cache.UserVerificationCodeExpireTime); err != nil {
		h.logger.Error("发送邮箱验证码失败", zap.Error(err))
		return err
	}

	//if err := h.smtpService.SendEmail(cmd.Email, "重置pgp验证码(请妥善保管,有效时间5分钟)", verifCode); err != nil {
	//	h.logger.Error("发送邮箱验证码失败", zap.Error(err))
	//	return err
	//}

	h.logger.Info("发送邮箱验证码成功", zap.String("email", cmd.Email), zap.String("code", verifCode))

	return nil
}
