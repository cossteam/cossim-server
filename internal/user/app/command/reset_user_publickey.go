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
	"go.uber.org/zap"
)

type ResetUserPublicKey struct {
	UserID    string
	PublicKey string
	Code      string
}

type ResetUserPublicKeyHandler decorator.CommandHandlerNoneResponse[*ResetUserPublicKey]

func NewResetUserPublicKeyHandler(
	logger *zap.Logger,
	ud service.UserDomain,
	userCache cache.UserCache,
	smtpService remote.SmtpService,
) ResetUserPublicKeyHandler {
	return &resetUserPublicKeyHandler{
		logger:      logger,
		ud:          ud,
		userCache:   userCache,
		smtpService: smtpService,
	}
}

type resetUserPublicKeyHandler struct {
	logger      *zap.Logger
	userCache   cache.UserCache
	ud          service.UserDomain
	smtpService remote.SmtpService
}

func (h *resetUserPublicKeyHandler) Handle(ctx context.Context, cmd *ResetUserPublicKey) error {
	h.logger.Info("重置用户pgp公钥", zap.Any("cmd", cmd))
	if cmd == nil || cmd.PublicKey == "" || cmd.Code == "" || cmd.UserID == "" {
		return code.InvalidParameter
	}

	user, err := h.ud.GetUserWithOpts(ctx, entity.WithUserID(cmd.UserID))
	if err != nil {
		if errors.Is(code.Cause(err), code.NotFound) {
			return code.UserErrNotExist
		}
		h.logger.Error("get user error", zap.Error(err))
		return err
	}

	verificationCode, err := h.userCache.GetUserVerificationCode(ctx, cmd.UserID, cmd.Code)
	if err != nil {
		if errors.Is(code.Cause(err), code.NotFound) {
			return code.MyCustomErrorCode.CustomMessage("验证码不存在或已过期")
		}
		return err
	}
	if verificationCode != cmd.Code {
		return code.MyCustomErrorCode.CustomMessage("验证码不存在或已过期")
	}

	if user.PublicKey == "" {
		return code.MyCustomErrorCode.CustomMessage("用户未设置pgp公钥，不允许重置")
	}

	if err = h.ud.SetUserPublicKey(ctx, cmd.UserID, cmd.PublicKey); err != nil {
		h.logger.Info("重置用户pgp公钥失败", zap.Error(err))
		return err
	}

	if err := h.userCache.DeleteUserVerificationCode(ctx, cmd.UserID, cmd.Code); err != nil {
		h.logger.Error("删除用户邮箱验证码失败", zap.Error(err))
	}

	h.logger.Info("重置用户pgp公钥", zap.String("user_id", cmd.UserID), zap.String("email", user.Email), zap.String("code", verificationCode))

	return nil
}
