package command

import (
	"context"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type UpdateQRCodeHandler decorator.CommandHandlerNoneResponse[*entity.QRCode]

func NewUpdateQRCodeHandler(
	logger *zap.Logger,
	userCache cache.UserCache,
) UpdateQRCodeHandler {
	return &updateQRCodeHandler{
		logger:    logger,
		userCache: userCache,
	}
}

type updateQRCodeHandler struct {
	logger    *zap.Logger
	userCache cache.UserCache
}

func (h *updateQRCodeHandler) Handle(ctx context.Context, cmd *entity.QRCode) error {
	err := h.userCache.SetQrCode(ctx, cmd)
	if err != nil {
		return err
	}
	return nil
}
