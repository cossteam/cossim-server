package rpc

import (
	"context"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type PushService interface {
	PushWS(ctx context.Context, data []byte) (interface{}, error)
}

func NewPushService(addr string) (PushService, error) {
	var grpcOptions = []grpc.DialOption{grpc.WithInsecure()}
	grpcOptions = append(grpcOptions, grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`))
	conn, err := grpc.Dial(
		addr,
		grpcOptions...,
	)
	if err != nil {
		return nil, err
	}
	return &PushServiceGrpc{client: pushgrpcv1.NewPushServiceClient(conn)}, nil
}

func NewPushServiceWithClient(client pushgrpcv1.PushServiceClient) PushService {
	return &PushServiceGrpc{client: client}
}

//type PushType int32
//
//const (
//	PushWs          PushType = iota // ws推送
//	PushMobile                      // 移动端推送
//	PushEmail                       // 邮件推送
//	PushMessage                     // 短信推送
//	PushBatchWs                     // 批量ws推送
//	PushWsBatchUser                 // 批量ws推送
//)

type PushServiceGrpc struct {
	client pushgrpcv1.PushServiceClient
	logger *zap.Logger
}

func (s *PushServiceGrpc) PushWS(ctx context.Context, data []byte) (interface{}, error) {
	return s.client.Push(ctx, &pushgrpcv1.PushRequest{
		Type: pushgrpcv1.Type_Ws,
		Data: data,
	})
}
