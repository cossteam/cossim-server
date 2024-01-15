package service

import (
	"context"
	"github.com/cossim/coss-server/pkg/code"
	v1 "github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/cossim/coss-server/service/relation/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) CreateDialog(ctx context.Context, in *v1.CreateDialogRequest) (*v1.CreateDialogResponse, error) {
	resp := &v1.CreateDialogResponse{}

	dialog, err := s.dr.CreateDialog(in.OwnerId, entity.DialogType(in.Type), uint(in.GroupId))
	if err != nil {
		return resp, status.Error(codes.Code(code.DialogErrCreateDialogFailed.Code()), err.Error())
	}
	return &v1.CreateDialogResponse{
		Id:      uint32(dialog.ID),
		OwnerId: dialog.OwnerId,
		GroupId: uint32(dialog.ID),
		Type:    uint32(dialog.Type),
	}, nil
}

func (s *Service) JoinDialog(ctx context.Context, in *v1.JoinDialogRequest) (*v1.JoinDialogResponse, error) {
	resp := &v1.JoinDialogResponse{}
	_, err := s.dr.JoinDialog(uint(in.DialogId), in.UserId)
	if err != nil {
		return resp, status.Error(codes.Code(code.DialogErrJoinDialogFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) GetUserDialogList(ctx context.Context, in *v1.GetUserDialogListRequest) (*v1.GetUserDialogListResponse, error) {
	resp := &v1.GetUserDialogListResponse{}
	ids, err := s.dr.GetUserDialogs(in.UserId)
	if err != nil {
		return resp, status.Error(codes.Code(code.DialogErrGetUserDialogListFailed.Code()), err.Error())
	}
	nids := make([]uint32, 0)
	if len(ids) > 0 {
		for _, id := range ids {
			nids = append(nids, uint32(id))
		}
	}
	resp.DialogIds = nids
	return resp, nil
}

func (s *Service) GetDialogByIds(ctx context.Context, in *v1.GetDialogByIdsRequest) (*v1.GetDialogByIdsResponse, error) {
	resp := &v1.GetDialogByIdsResponse{}
	nids := make([]uint, 0)
	if len(in.DialogIds) > 0 {
		for _, id := range in.DialogIds {
			nids = append(nids, uint(id))
		}
	}
	infos, err := s.dr.GetDialogsByIDs(nids)
	if err != nil {
		return resp, status.Error(codes.Code(code.DialogErrGetUserDialogListFailed.Code()), err.Error())
	}
	var dialogInfos []*v1.Dialog
	if len(infos) > 0 {
		for _, info := range infos {
			dialogInfos = append(dialogInfos, &v1.Dialog{
				Id:      uint32(info.ID),
				OwnerId: info.OwnerId,
				GroupId: uint32(info.GroupId),
				Type:    uint32(info.Type),
			})
		}
	}
	resp.Dialogs = dialogInfos
	return resp, nil
}

func (s *Service) GetDialogUsersByDialogID(ctx context.Context, in *v1.GetDialogUsersByDialogIDRequest) (*v1.GetDialogUsersByDialogIDResponse, error) {
	resp := &v1.GetDialogUsersByDialogIDResponse{}
	users, err := s.dr.GetDialogUsersByDialogID(uint(in.DialogId))
	if err != nil {
		return resp, status.Error(codes.Code(code.DialogErrGetUserDialogListFailed.Code()), err.Error())
	}
	var ids []string
	for _, id := range users {
		ids = append(ids, id.UserId)
	}
	resp.UserIds = ids
	return resp, nil
}

func (s *Service) DeleteDialogByIds(ctx context.Context, in *v1.DeleteDialogByIdsRequest) (*v1.DeleteDialogByIdsResponse, error) {
	var resp = &v1.DeleteDialogByIdsResponse{}
	uintIds := make([]uint, 0)
	if len(in.DialogIds) > 0 {
		for _, id := range in.DialogIds {
			uintIds = append(uintIds, uint(id))
		}
	}
	if err := s.dr.DeleteDialogByIds(uintIds); err != nil {
		return resp, status.Error(codes.Code(code.DialogErrDeleteDialogFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) DeleteDialogById(ctx context.Context, in *v1.DeleteDialogByIdRequest) (*v1.DeleteDialogByIdResponse, error) {
	var resp = &v1.DeleteDialogByIdResponse{}
	if err := s.dr.DeleteDialogByDialogID(uint(in.DialogId)); err != nil {
		return resp, status.Error(codes.Code(code.DialogErrDeleteDialogFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) DeleteDialogUsersByDialogID(ctx context.Context, in *v1.DeleteDialogUsersByDialogIDRequest) (*v1.DeleteDialogUsersByDialogIDResponse, error) {
	var resp = &v1.DeleteDialogUsersByDialogIDResponse{}
	if err := s.dr.DeleteDialogUserByDialogID(uint(in.DialogId)); err != nil {
		return resp, status.Error(codes.Code(code.DialogErrDeleteDialogUsersFailed.Code()), err.Error())
	}
	return resp, nil
}
