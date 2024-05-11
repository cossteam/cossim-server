package command

import (
	"context"
	"github.com/cossim/coss-server/internal/user/domain/service"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type UpdateUserBundle struct {
	UserID       string
	SecretBundle string
}

type UpdateUserBundleHandler decorator.CommandHandlerNoneResponse[*UpdateUserBundle]

func NewUpdateUserBundle(logger *zap.Logger, ud service.UserDomain) UpdateUserBundleHandler {
	return &updateUserBundleHandler{
		logger: logger,
		ud:     ud,
	}
}

type updateUserBundleHandler struct {
	logger *zap.Logger
	ud     service.UserDomain
}

func (u *updateUserBundleHandler) Handle(ctx context.Context, cmd *UpdateUserBundle) error {
	if cmd == nil {
		return code.InvalidParameter
	}

	return u.ud.UpdateBundle(ctx, cmd.UserID, cmd.SecretBundle)
}
