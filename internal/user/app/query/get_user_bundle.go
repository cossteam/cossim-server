package query

import (
	"context"
	"github.com/cossim/coss-server/internal/user/domain/service"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type GetUserBundle struct {
	UserID string
}

type GetUserBundleResponse struct {
	UserID       string
	SecretBundle string
}

type GetUserBundleHandler decorator.CommandHandler[*GetUserBundle, *GetUserBundleResponse]

func NewGetUserBundleHandler(logger *zap.Logger, userDomain service.UserDomain) GetUserBundleHandler {
	return &getUserBundleHandler{
		logger: logger,
		ud:     userDomain,
	}
}

type getUserBundleHandler struct {
	logger *zap.Logger
	ud     service.UserDomain
}

func (h *getUserBundleHandler) Handle(ctx context.Context, cmd *GetUserBundle) (*GetUserBundleResponse, error) {
	if cmd == nil {
		return nil, code.InvalidParameter
	}

	user, err := h.ud.GetUser(ctx, cmd.UserID)
	if err != nil {
		h.logger.Error("get user failed", zap.Error(err))
		return nil, err
	}

	return &GetUserBundleResponse{
		UserID:       user.ID,
		SecretBundle: user.SecretBundle,
	}, nil
}
