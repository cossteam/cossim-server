package rpc

import (
	"context"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	"google.golang.org/grpc"
)

type MsgService interface {
	CleanDialogMessages(ctx context.Context, dialog uint32) error
}

var _ MsgService = &msgServiceGrpc{}

func NewMsgServiceGrpc(addr string) (MsgService, error) {
	var grpcOptions = []grpc.DialOption{grpc.WithInsecure()}
	grpcOptions = append(grpcOptions, grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`))
	conn, err := grpc.Dial(
		addr,
		grpcOptions...,
	)
	if err != nil {
		return nil, err
	}
	return &msgServiceGrpc{client: msggrpcv1.NewMsgServiceClient(conn)}, nil
}

func NewMsgServiceGrpcWithClient(client msggrpcv1.MsgServiceClient) MsgService {
	return &msgServiceGrpc{client: client}
}

type msgServiceGrpc struct {
	client msggrpcv1.MsgServiceClient
}

func (s *msgServiceGrpc) CleanDialogMessages(ctx context.Context, dialog uint32) error {
	_, err := s.client.ConfirmDeleteUserMessageByDialogId(ctx, &msggrpcv1.DeleteUserMsgByDialogIdRequest{DialogId: dialog})
	if err != nil {
		return err
	}

	return nil
}
