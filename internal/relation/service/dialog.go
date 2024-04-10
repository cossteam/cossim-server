package service

import (
	"context"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/api/http/model"
)

func (s *Service) OpenOrCloseDialog(ctx context.Context, userId string, request *model.CloseOrOpenDialogRequest) (interface{}, error) {
	_, err := s.relationDialogService.GetDialogById(ctx, &relationgrpcv1.GetDialogByIdRequest{
		DialogId: request.DialogId,
	})
	if err != nil {
		return nil, err
	}

	_, err = s.relationDialogService.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		DialogId: request.DialogId,
		UserId:   userId,
	})
	if err != nil {
		return nil, err
	}

	_, err = s.relationDialogService.CloseOrOpenDialog(ctx, &relationgrpcv1.CloseOrOpenDialogRequest{
		DialogId: request.DialogId,
		UserId:   userId,
		Action:   relationgrpcv1.CloseOrOpenDialogType(request.Action),
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *Service) TopOrCancelTopDialog(ctx context.Context, userId string, request *model.TopOrCancelTopDialogRequest) (interface{}, error) {
	_, err := s.relationDialogService.GetDialogById(ctx, &relationgrpcv1.GetDialogByIdRequest{
		DialogId: request.DialogId,
	})
	if err != nil {
		return nil, err
	}

	_, err = s.relationDialogService.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		DialogId: request.DialogId,
		UserId:   userId,
	})
	if err != nil {
		return nil, err
	}

	_, err = s.relationDialogService.TopOrCancelTopDialog(ctx, &relationgrpcv1.TopOrCancelTopDialogRequest{
		DialogId: request.DialogId,
		UserId:   userId,
		Action:   relationgrpcv1.TopOrCancelTopDialogType(request.Action),
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}
