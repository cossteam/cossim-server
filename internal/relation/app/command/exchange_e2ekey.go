package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type ExchangeE2EKey struct {
	CurrentUserID string
	TargetUserID  string
	PublicKey     string
}

func (cmd *ExchangeE2EKey) Validate() error {
	if cmd == nil || cmd.CurrentUserID == "" || cmd.TargetUserID == "" || cmd.PublicKey == "" {
		return code.InvalidParameter
	}
	return nil
}

type ExchangeE2EKeyHandler decorator.CommandHandlerNoneResponse[*ExchangeE2EKey]

func NewExchangeE2EKeyHandler(
	logger *zap.Logger,
	pushService rpc.PushService,
) ExchangeE2EKeyHandler {
	return &exchangeE2EKeyHandler{
		logger:      logger,
		pushService: pushService,
	}
}

type exchangeE2EKeyHandler struct {
	logger      *zap.Logger
	pushService rpc.PushService
}

func (h *exchangeE2EKeyHandler) Handle(ctx context.Context, cmd *ExchangeE2EKey) error {
	if err := cmd.Validate(); err != nil {
		h.logger.Error("exchange e2e key command validate failed", zap.Error(err))
	}

	if err := h.pushService.ExchangeE2eKeyPush(ctx, cmd.CurrentUserID, cmd.TargetUserID, cmd.PublicKey); err != nil {
		h.logger.Error("exchange e2e key push failed", zap.Error(err))
	}

	return nil
}
