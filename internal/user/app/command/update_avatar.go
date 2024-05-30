package command

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/domain/service"
	"github.com/cossim/coss-server/internal/user/infra/remote"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/decorator"
	"github.com/minio/minio-go/v7"
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
	ud service.UserDomain,
	storageService remote.StorageService,
	baseUrl string,
) UpdateUserAvatarHandler {
	return &updateUserAvatarHandler{
		logger:         logger,
		ud:             ud,
		storageService: storageService,
	}
}

type updateUserAvatarHandler struct {
	logger         *zap.Logger
	ssl            bool
	ud             service.UserDomain
	storageService remote.StorageService
	baseUrl        string
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

	path, err := h.storageService.UploadOther(ctx, reader, minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return nil, err
	}

	aUrl := fmt.Sprintf("%s/%s", h.baseUrl+constants.DownLoadAddress, path)
	_, err = h.ud.UpdateUser(ctx, &entity.User{ID: cmd.UserID, Avatar: aUrl}, true)
	if err != nil {
		return nil, err
	}

	return &UpdateUserAvatarResponse{
		Avatar: aUrl,
	}, nil
}
