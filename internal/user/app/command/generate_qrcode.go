package command

import (
	"bytes"
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/infra/remote"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/decorator"
	"github.com/cossim/coss-server/pkg/utils/qr"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
)

type GenerateQRCode struct {
}

type GenerateQRCodeResponse struct {
	Token  string
	QrCode string
}

type GenerateQRCodeHandler decorator.CommandHandler[*GenerateQRCode, *GenerateQRCodeResponse]

func NewGenerateQRCodeHandler(
	logger *zap.Logger,
	storageService remote.StorageService,
	userCache cache.UserCache,
	baseUrl string,
) GenerateQRCodeHandler {
	return &generateQRCodeHandler{
		logger:         logger,
		storageService: storageService,
		baseUrl:        baseUrl,
		userCache:      userCache,
	}
}

type generateQRCodeHandler struct {
	logger         *zap.Logger
	userCache      cache.UserCache
	storageService remote.StorageService
	baseUrl        string
}

func (h *generateQRCodeHandler) Handle(ctx context.Context, cmd *GenerateQRCode) (*GenerateQRCodeResponse, error) {
	token := uuid.New().String()
	qrcode, err := qr.GenQrcode(token)
	if err != nil {
		return nil, err
	}

	reader := bytes.NewReader(qrcode.Bytes())

	code, err := h.storageService.UploadQRCode(ctx, reader, minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return nil, err
	}

	//保存到redis里
	err = h.userCache.SetQrCode(ctx, &entity.QRCode{Token: token, Status: entity.QRCodeStatusNotScanned})
	if err != nil {
		return nil, err
	}

	aUrl := fmt.Sprintf("%s%s", h.baseUrl+constants.DownLoadAddress, code)

	return &GenerateQRCodeResponse{
		Token:  token,
		QrCode: aUrl,
	}, nil
}
