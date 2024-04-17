package adapters

import (
	"context"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	"go.uber.org/zap"
)

//var _ command.PushService = &PushServiceGrpc{}

func NewPushService(client pushgrpcv1.PushServiceClient) *PushServiceGrpc {
	return &PushServiceGrpc{client: client}
}

type PushServiceGrpc struct {
	client pushgrpcv1.PushServiceClient
	logger *zap.Logger
}

func (s *PushServiceGrpc) Push(ctx context.Context, t int32, data []byte) (interface{}, error) {
	return s.client.Push(ctx, &pushgrpcv1.PushRequest{
		Type: pushgrpcv1.Type_Ws,
		Data: data,
	})
}
