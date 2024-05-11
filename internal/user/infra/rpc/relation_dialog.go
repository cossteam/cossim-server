package rpc

import (
	"context"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"google.golang.org/grpc"
)

type RelationDialogService interface {
	ToggleDialog(ctx context.Context, dialogID uint32, userID string, open bool) error
	IsDialogClosed(ctx context.Context, dialogID uint32, userID string) (bool, error)
}

var _ RelationDialogService = &relationDialogGrpc{}

type relationDialogGrpc struct {
	client relationgrpcv1.DialogServiceClient
}

func NewRelationDialogGrpcWithClient(client relationgrpcv1.DialogServiceClient) RelationDialogService {
	return &relationDialogGrpc{client: client}
}

func NewRelationDialogGrpc(addr string) (RelationDialogService, error) {
	var grpcOptions = []grpc.DialOption{grpc.WithInsecure()}
	grpcOptions = append(grpcOptions, grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`))
	conn, err := grpc.Dial(
		addr,
		grpcOptions...,
	)
	if err != nil {
		return nil, err
	}

	return &relationDialogGrpc{client: relationgrpcv1.NewDialogServiceClient(conn)}, nil
}

func (s *relationDialogGrpc) ToggleDialog(ctx context.Context, dialogID uint32, userID string, open bool) error {
	var action relationgrpcv1.CloseOrOpenDialogType
	if open {
		action = relationgrpcv1.CloseOrOpenDialogType_OPEN
	} else {
		action = relationgrpcv1.CloseOrOpenDialogType_CLOSE
	}
	_, err := s.client.CloseOrOpenDialog(ctx, &relationgrpcv1.CloseOrOpenDialogRequest{
		DialogId: dialogID,
		Action:   action,
		UserId:   userID,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *relationDialogGrpc) IsDialogClosed(ctx context.Context, dialogID uint32, userID string) (bool, error) {
	r, err := s.client.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		DialogId: dialogID,
		UserId:   userID,
	})
	if err != nil {
		return false, err
	}

	if r.IsShow != uint32(relationgrpcv1.CloseOrOpenDialogType_OPEN) {
		return true, nil
	}

	return false, nil
}
