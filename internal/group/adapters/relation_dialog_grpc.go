package adapters

import (
	"context"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"go.uber.org/zap"
)

//var _ command.RelationDialogService = &RelationDialogGrpc{}

const (
	RelationDialogServiceName = "relation_service"
)

func NewRelationDialogGrpc(client relationgrpcv1.DialogServiceClient) *RelationDialogGrpc {
	return &RelationDialogGrpc{client: client}
}

type RelationDialogGrpc struct {
	client relationgrpcv1.DialogServiceClient
	logger *zap.Logger
}

func (s *RelationDialogGrpc) DeleteDialog(ctx context.Context, dialogID uint32) error {
	_, err := s.client.DeleteDialogById(ctx, &relationgrpcv1.DeleteDialogByIdRequest{DialogId: dialogID})
	if err != nil {
		return err
	}

	return nil
}

func (s *RelationDialogGrpc) DeleteDialogRevert(ctx context.Context, dialogID uint32) error {
	_, err := s.client.DeleteDialogByIdRevert(ctx, &relationgrpcv1.DeleteDialogByIdRequest{DialogId: dialogID})
	if err != nil {
		return err
	}

	return nil
}

func (s *RelationDialogGrpc) DeleteUserDialog(ctx context.Context, dialogID uint32) error {
	_, err := s.client.DeleteDialogUsersByDialogID(ctx, &relationgrpcv1.DeleteDialogUsersByDialogIDRequest{DialogId: dialogID})
	if err != nil {
		return err
	}

	return nil
}

func (s *RelationDialogGrpc) DeleteUserDialogRevert(ctx context.Context, dialogID uint32) error {
	_, err := s.client.DeleteDialogUsersByDialogIDRevert(ctx, &relationgrpcv1.DeleteDialogUsersByDialogIDRequest{DialogId: dialogID})
	if err != nil {
		return err
	}

	return nil
}

func (s *RelationDialogGrpc) GetGroupDialogID(ctx context.Context, groupID uint32) (uint32, error) {
	resp, err := s.client.GetDialogByGroupId(ctx, &relationgrpcv1.GetDialogByGroupIdRequest{GroupId: groupID})
	if err != nil {
		return 0, err
	}

	return resp.DialogId, nil
}
