package decorator

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"strings"
)

func ApplyCommandDecorators[H any, R any](handler CommandHandler[H, R], logger *zap.Logger, metricsClient MetricsClient) CommandHandler[H, R] {
	return commandLoggingDecorator[H, R]{
		base: commandMetricsDecorator[H, R]{
			base:   handler,
			client: metricsClient,
		},
		logger: logger,
	}
}

type CommandHandler[C any, R any] interface {
	Handle(ctx context.Context, cmd C) (R, error)
}

type CommandHandlerNoneResponse[C any] interface {
	Handle(ctx context.Context, cmd C) error
}

func generateActionName(handler any) string {
	return strings.Split(fmt.Sprintf("%T", handler), ".")[1]
}
