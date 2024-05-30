package query

import (
	"context"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type GetQrCode struct {
	Token string
}

type GetQRCodeHandler decorator.CommandHandler[*GetQrCode, *entity.QRCode]

type getQRCodeHandler struct {
	logger    *zap.Logger
	userCache cache.UserCache
}

func NewGetQRCodeHandler(logger *zap.Logger, userCache cache.UserCache) GetQRCodeHandler {

	return &getQRCodeHandler{logger: logger, userCache: userCache}
}

func (g getQRCodeHandler) Handle(ctx context.Context, cmd *GetQrCode) (*entity.QRCode, error) {

	code, err := g.userCache.GetQrCode(ctx, cmd.Token)
	if err != nil {
		return nil, err
	}

	return code, nil
}
