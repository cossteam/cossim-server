package command

import (
	"context"
	"github.com/cossim/coss-server/internal/user/domain/service"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type SetUserPublicKey struct {
	UserID    string
	PublicKey string
}

type SetUserPublicKeyHandler decorator.CommandHandlerNoneResponse[*SetUserPublicKey]

func NewSetUserPublicKeyHandler(logger *zap.Logger, ud service.UserDomain) SetUserPublicKeyHandler {
	return &setUserPublicKeyHandler{
		logger: logger,
		ud:     ud,
	}
}

type setUserPublicKeyHandler struct {
	logger *zap.Logger
	ud     service.UserDomain
}

func (h *setUserPublicKeyHandler) Handle(ctx context.Context, cmd *SetUserPublicKey) error {
	if cmd == nil || cmd.UserID == "" || cmd.PublicKey == "" {
		return code.InvalidParameter
	}

	user, err := h.ud.GetUser(ctx, cmd.UserID)
	if err != nil {
		h.logger.Error("get user error", zap.Error(err))
		return err
	}

	if user.PublicKey != "" {
		return code.MyCustomErrorCode.CustomMessage("用户已设置公钥，如需修改请使用重置公钥")
	}

	if err := h.ud.SetUserPublicKey(ctx, cmd.UserID, cmd.PublicKey); err != nil {
		h.logger.Error("set user public key error", zap.Error(err))
		return err
	}

	return nil
}
