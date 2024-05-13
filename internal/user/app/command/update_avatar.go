package command

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/domain/service"
	"github.com/cossim/coss-server/internal/user/infra/remote"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	httputil "github.com/cossim/coss-server/pkg/utils/http"
	"go.uber.org/zap"
	"io/ioutil"
	"mime/multipart"
)

type UpdateUserAvatar struct {
	UserID string
	Avatar multipart.File
}

type UpdateUserAvatarResponse struct {
	Avatar string
}

type UpdateUserAvatarHandler decorator.CommandHandler[*UpdateUserAvatar, *UpdateUserAvatarResponse]

func NewUpdateUserAvatarHandler(logger *zap.Logger,
	ssl bool,
	gatewayAddress string,
	ud service.UserDomain,
	storageService remote.StorageService,
) UpdateUserAvatarHandler {
	return &updateUserAvatarHandler{
		logger:         logger,
		ssl:            ssl,
		gatewayAddress: gatewayAddress,
		ud:             ud,
		storageService: storageService,
	}
}

type updateUserAvatarHandler struct {
	logger         *zap.Logger
	ssl            bool
	gatewayAddress string
	downloadURL    string
	ud             service.UserDomain
	storageService remote.StorageService
}

func (h *updateUserAvatarHandler) Handle(ctx context.Context, cmd *UpdateUserAvatar) (*UpdateUserAvatarResponse, error) {
	if cmd == nil || cmd.UserID == "" || cmd.Avatar == nil {
		return nil, code.InvalidParameter
	}

	user, err := h.ud.GetUser(ctx, cmd.UserID)
	if err != nil {
		h.logger.Error("get user error", zap.Error(err))
		return nil, err
	}

	if user.Status != entity.UserStatusNormal {
		return nil, code.UserErrStatusException.CustomMessage(fmt.Sprintf("user status is %s", user.Status.String()))
	}

	data, err := ioutil.ReadAll(cmd.Avatar)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(data)

	path, err := h.storageService.UploadAvatar(ctx, reader)
	if err != nil {
		return nil, err
	}

	aUrl := fmt.Sprintf("http://%s%s/%s", h.gatewayAddress, h.downloadURL, path)
	if h.ssl {
		aUrl, err = httputil.ConvertToHttps(aUrl)
		if err != nil {
			return nil, err
		}
	}

	_, err = h.ud.UpdateUser(ctx, &entity.User{ID: cmd.UserID, Avatar: aUrl}, true)
	if err != nil {
		return nil, err
	}

	return &UpdateUserAvatarResponse{
		Avatar: aUrl,
	}, nil
}
