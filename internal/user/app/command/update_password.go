package command

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/domain/service"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"github.com/cossim/coss-server/pkg/utils"
	"go.uber.org/zap"
	"regexp"
)

type UpdatePassword struct {
	UserID          string
	OldPassword     string
	ConfirmPassword string
	NewPassword     string
}

type UpdatePasswordHandler decorator.CommandHandler[*UpdatePassword, *interface{}]

func NewUpdatePasswordHandler(logger *zap.Logger, ud service.UserDomain) UpdatePasswordHandler {
	return &updatePasswordHandler{
		logger: logger,
		ud:     ud,
	}
}

type updatePasswordHandler struct {
	logger *zap.Logger
	ud     service.UserDomain
}

func (h *updatePasswordHandler) Handle(ctx context.Context, cmd *UpdatePassword) (*interface{}, error) {
	if cmd == nil || cmd.UserID == "" || cmd.NewPassword == "" || cmd.OldPassword == "" {
		return nil, code.InvalidParameter
	}

	if cmd.NewPassword != cmd.ConfirmPassword {
		return nil, code.InvalidParameter.CustomMessage("password and confirm password not match")
	}

	if regexp.MustCompile(`\s`).MatchString(cmd.NewPassword) {
		return nil, code.InvalidParameter.CustomMessage("password cannot contain spaces")
	}

	newPassword := utils.HashString(cmd.NewPassword)
	oldPassword := utils.HashString(cmd.OldPassword)

	user, err := h.ud.GetUserWithOpts(ctx, entity.WithUserID(cmd.UserID), entity.WithPassword(oldPassword))
	if err != nil {
		if errors.Is(err, code.NotFound) {
			return nil, code.NotFound.CustomMessage("Incorrect old password")
		}
		return nil, err
	}

	if user.Password == newPassword {
		return nil, code.InvalidParameter.CustomMessage("New password cannot be the same as the old password")
	}

	_, err = h.ud.UpdatePassword(ctx, cmd.UserID, newPassword)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
