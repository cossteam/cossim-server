package service

import (
	"context"
	v1 "github.com/cossim/coss-server/service/msg/api/v1"
	"github.com/cossim/coss-server/service/msg/domain/entity"
)

func (s *Service) CreateDialog(ctx context.Context, in *v1.CreateDialogRequest) (*v1.CreateDialogResponse, error) {
	resp := &v1.CreateDialogResponse{}

	dialog, err := s.dr.CreateDialog(in.OwnerId, entity.DialogType(in.Type), uint(in.GroupId))
	if err != nil {
		return resp, err
	}
	return &v1.CreateDialogResponse{
		Id:      uint32(dialog.ID),
		OwnerId: dialog.OwnerId,
		GroupId: uint32(dialog.ID),
		Type:    uint32(dialog.Type),
	}, nil
}

func (s *Service) JoinDialog(ctx context.Context, in *v1.JoinDialogRequest) (*v1.Empty, error) {
	resp := &v1.Empty{}
	_, err := s.dr.JoinDialog(uint(in.DialogId), in.UserId)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *Service) GetUserDialogList(ctx context.Context, in *v1.GetUserDialogListRequest) (*v1.GetUserDialogListResponse, error) {
	resp := &v1.GetUserDialogListResponse{}
	ids, err := s.dr.GetUserDialogs(in.UserId)
	if err != nil {
		return resp, err
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
		return resp, err
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
