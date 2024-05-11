package rpc

import (
	"context"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	"google.golang.org/grpc"
)

type MsgService interface {
	DeleteMessage(ctx context.Context, msg ...uint32) error
	SendUserTextMessage(ctx context.Context, dialogID uint32, sender, recipient, content string) (uint32, error)
}

var _ MsgService = &msgServiceGrpc{}

type msgServiceGrpc struct {
	client msggrpcv1.MsgServiceClient
}

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

func (s *msgServiceGrpc) SendUserTextMessage(ctx context.Context, dialogID uint32, sender, recipient, content string) (uint32, error) {
	if dialogID == 0 || sender == "" || recipient == "" || content == "" {
		return 0, code.InvalidParameter
	}
	r, err := s.client.SendUserMessage(ctx, &msggrpcv1.SendUserMsgRequest{
		SenderId:   sender,
		ReceiverId: recipient,
		Content:    content,
		DialogId:   dialogID,
		Type:       int32(msggrpcv1.MessageType_Text),
	})

	return r.MsgId, err
}

func (s *msgServiceGrpc) DeleteMessage(ctx context.Context, msg ...uint32) error {
	if len(msg) == 0 {
		return code.InvalidParameter
	}
	_, err := s.client.DeleteUserMessageByIDs(ctx, &msggrpcv1.DeleteUserMessageByIdsRequest{
		IDs: msg,
	})

	return err
}
