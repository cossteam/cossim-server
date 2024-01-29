package service

import (
	"context"
	"github.com/cossim/coss-server/interface/relation/api/model"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
)

func (s *Service) OpenOrCloseDialog(ctx context.Context, userId string, request *model.CloseOrOpenDialogRequest) (interface{}, error) {
	_, err := s.dialogClient.GetDialogById(ctx, &relationgrpcv1.GetDialogByIdRequest{
		DialogId: request.DialogId,
	})
	if err != nil {
		return nil, err
	}

	_, err = s.dialogClient.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		DialogId: request.DialogId,
		UserId:   userId,
	})
	if err != nil {
		return nil, err
	}

	_, err = s.dialogClient.CloseOrOpenDialog(ctx, &relationgrpcv1.CloseOrOpenDialogRequest{
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
	_, err := s.dialogClient.GetDialogById(ctx, &relationgrpcv1.GetDialogByIdRequest{
		DialogId: request.DialogId,
	})
	if err != nil {
		return nil, err
	}

	_, err = s.dialogClient.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		DialogId: request.DialogId,
		UserId:   userId,
	})
	if err != nil {
		return nil, err
	}

	_, err = s.dialogClient.TopOrCancelTopDialog(ctx, &relationgrpcv1.TopOrCancelTopDialogRequest{
		DialogId: request.DialogId,
		UserId:   userId,
		Action:   relationgrpcv1.TopOrCancelTopDialogType(request.Action),
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
}
