package command

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/domain/service"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type UpdateUser struct {
	UserID    string
	Avatar    string
	CossID    string
	Nickname  string
	Signature string
	Tel       string
}

type UpdateUserHandler decorator.CommandHandler[*UpdateUser, *interface{}]

func NewUpdateUserHandler(logger *zap.Logger, ud service.UserDomain, userCache cache.UserCache) UpdateUserHandler {
	return &updateUserHandler{
		logger:    logger,
		ud:        ud,
		userCache: userCache,
	}
}

type updateUserHandler struct {
	logger    *zap.Logger
	ud        service.UserDomain
	userCache cache.UserCache
}

func (h *updateUserHandler) Handle(ctx context.Context, cmd *UpdateUser) (*interface{}, error) {
	h.logger.Info("update user handler", zap.Any("cmd", cmd))
	if cmd == nil {
		return nil, errors.New("cmd is nil")
	}

	user, err := h.ud.GetUser(ctx, cmd.UserID)
	if err != nil && user == nil {
		return nil, err
	}

	_, err = h.ud.UpdateUser(ctx, &entity.User{
		ID:        cmd.UserID,
		CossID:    cmd.CossID,
		Tel:       cmd.Tel,
		NickName:  cmd.Nickname,
		Avatar:    cmd.Avatar,
		Signature: cmd.Signature,
	}, true)
	if err != nil {
		h.logger.Error("update user error", zap.Error(err))
		return nil, err
	}

	return nil, nil
}
