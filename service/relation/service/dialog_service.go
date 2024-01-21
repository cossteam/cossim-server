package service

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/pkg/code"
	v1 "github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/cossim/coss-server/service/relation/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
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

func (s *Service) CreateAndJoinDialogWithGroup(ctx context.Context, request *v1.CreateAndJoinDialogWithGroupRequest) (*v1.CreateAndJoinDialogWithGroupResponse, error) {
	resp := &v1.CreateAndJoinDialogWithGroupResponse{}

	//return resp, status.Error(codes.Aborted, formatErrorMessage(errors.New("测试回滚")))

	err := s.db.Transaction(func(tx *gorm.DB) error {
		dialog, err := s.dr.CreateDialog(request.OwnerId, entity.DialogType(request.Type), uint(request.GroupId))
		if err != nil {
			return status.Error(codes.Code(code.DialogErrCreateDialogFailed.Code()), fmt.Sprintf("failed to create dialog: %s", err.Error()))
		}

		_, err = s.JoinDialog(ctx, &v1.JoinDialogRequest{DialogId: uint32(dialog.ID), UserId: request.OwnerId})
		if err != nil {
			return status.Error(codes.Code(code.DialogErrJoinDialogFailed.Code()), fmt.Sprintf("failed to join dialog: %s", err.Error()))
		}

		return nil
	})

	if err != nil {
		return resp, status.Error(codes.Aborted, fmt.Sprintf("failed to create dialog: %s", err.Error()))
	}

	return resp, nil
}

func (s *Service) CreateAndJoinDialogWithGroupRevert(ctx context.Context, request *v1.CreateAndJoinDialogWithGroupRequest) (*v1.CreateAndJoinDialogWithGroupResponse, error) {
	resp := &v1.CreateAndJoinDialogWithGroupResponse{}

	fmt.Println("CreateAndJoinDialogWithGroupRevert req => ", request)

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := s.dr.DeleteDialogUserByDialogIDAndUserID(uint(request.DialogId), request.OwnerId); err != nil {
			return status.Error(codes.Code(code.DialogErrDeleteDialogUsersFailed.Code()), fmt.Sprintf("failed to delete dialog user revert : %s", err.Error()))
		}
		if err := s.dr.DeleteDialogByDialogID(uint(request.DialogId)); err != nil {
			return status.Error(codes.Code(code.DialogErrDeleteDialogFailed.Code()), fmt.Sprintf("failed to delete dialog revert : %s", err.Error()))
		}

		return nil
	}); err != nil {
		return resp, status.Error(codes.Aborted, fmt.Sprintf("failed to create dialog revert: %s", err.Error()))
	}

	return resp, nil
}

func (s *Service) ConfirmFriendAndJoinDialog(ctx context.Context, request *v1.ConfirmFriendAndJoinDialogRequest) (*v1.ConfirmFriendAndJoinDialogResponse, error) {
	resp := &v1.ConfirmFriendAndJoinDialogResponse{}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		dialog, err := s.dr.CreateDialog(request.OwnerId, entity.DialogType(request.Type), uint(request.GroupId))
		if err != nil {
			return status.Error(codes.Code(code.DialogErrCreateDialogFailed.Code()), err.Error())
		}

		_, err = s.JoinDialog(ctx, &v1.JoinDialogRequest{DialogId: uint32(dialog.ID), UserId: request.OwnerId})
		if err != nil {
			return status.Error(codes.Code(code.DialogErrCreateDialogFailed.Code()), err.Error())
		}

		_, err = s.JoinDialog(ctx, &v1.JoinDialogRequest{DialogId: uint32(dialog.ID), UserId: request.UserId})
		if err != nil {
			return status.Error(codes.Code(code.DialogErrCreateDialogFailed.Code()), err.Error())
		}

		resp.Id = uint32(dialog.ID)
		resp.OwnerId = request.OwnerId
		resp.Type = v1.DialogType(dialog.Type)
		resp.GroupId = uint32(dialog.GroupId)

		return nil
	}); err != nil {
		return resp, err
	}

	return resp, nil
}

func (s *Service) ConfirmFriendAndJoinDialogRevert(ctx context.Context, request *v1.ConfirmFriendAndJoinDialogRevertRequest) (*v1.ConfirmFriendAndJoinDialogRevertResponse, error) {
	resp := &v1.ConfirmFriendAndJoinDialogRevertResponse{}

	if err := s.dr.DeleteDialogUserByDialogIDAndUserID(uint(request.DialogId), request.UserId); err != nil {
		return resp, status.Error(codes.Code(code.DialogErrCreateDialogFailed.Code()), err.Error())
	}

	if err := s.dr.DeleteDialogUserByDialogIDAndUserID(uint(request.DialogId), request.OwnerId); err != nil {
		return resp, status.Error(codes.Code(code.DialogErrCreateDialogFailed.Code()), err.Error())
	}

	if err := s.dr.DeleteDialogByDialogID(uint(request.DialogId)); err != nil {
		return resp, status.Error(codes.Code(code.DialogErrCreateDialogFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s *Service) JoinDialog(ctx context.Context, in *v1.JoinDialogRequest) (*v1.JoinDialogResponse, error) {
	resp := &v1.JoinDialogResponse{}
	_, err := s.dr.JoinDialog(uint(in.DialogId), in.UserId)
	if err != nil {
		return resp, status.Error(codes.Code(code.DialogErrJoinDialogFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) JoinDialogRevert(ctx context.Context, request *v1.JoinDialogRequest) (*v1.JoinDialogResponse, error) {
	fmt.Println("JoinDialogRevert req => ", request)
	resp := &v1.JoinDialogResponse{}
	if err := s.dr.DeleteDialogUserByDialogIDAndUserID(uint(request.DialogId), request.UserId); err != nil {
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
	resp := &v1.DeleteDialogByIdResponse{}

	//return resp, status.Error(codes.Aborted, formatErrorMessage(errors.New("测试回滚")))

	if err := s.dr.DeleteDialogByDialogID(uint(in.DialogId)); err != nil {
		return resp, status.Error(codes.Code(code.DialogErrDeleteDialogFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) DeleteDialogByIdRevert(ctx context.Context, request *v1.DeleteDialogByIdRequest) (*v1.DeleteDialogByIdResponse, error) {
	var resp = &v1.DeleteDialogByIdResponse{}
	fmt.Println("DeleteDialogByIdRevert req => ", request)
	if err := s.dr.UpdateDialogByDialogID(uint(request.DialogId), map[string]interface{}{
		"deleted_at": 0,
	}); err != nil {
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

func (s *Service) DeleteDialogUsersByDialogIDRevert(ctx context.Context, request *v1.DeleteDialogUsersByDialogIDRequest) (*v1.DeleteDialogUsersByDialogIDResponse, error) {
	var resp = &v1.DeleteDialogUsersByDialogIDResponse{}
	if err := s.dr.UpdateDialogUserByDialogID(uint(request.DialogId), map[string]interface{}{
		"deleted_at": 0,
	}); err != nil {
		return resp, status.Error(codes.Code(code.DialogErrDeleteDialogUsersFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) GetDialogUserByDialogIDAndUserID(ctx context.Context, in *v1.GetDialogUserByDialogIDAndUserIdRequest) (*v1.GetDialogUserByDialogIDAndUserIdResponse, error) {
	var resp = &v1.GetDialogUserByDialogIDAndUserIdResponse{}
	user, err := s.dr.GetDialogUserByDialogIDAndUserID(uint(in.DialogId), in.UserId)
	if err != nil {
		return resp, status.Error(codes.Code(code.DialogErrGetDialogUserByDialogIDAndUserIDFailed.Code()), err.Error())
	}
	resp.DialogId = uint32(user.DialogId)
	resp.UserId = user.UserId
	resp.IsShow = uint32(user.IsShow)
	resp.TopAt = uint64(user.TopAt)

	return resp, nil
}

func (s *Service) DeleteDialogUserByDialogIDAndUserID(ctx context.Context, request *v1.DeleteDialogUserByDialogIDAndUserIDRequest) (*v1.DeleteDialogUserByDialogIDAndUserIDResponse, error) {
	var resp = &v1.DeleteDialogUserByDialogIDAndUserIDResponse{}

	err := s.dr.DeleteDialogUserByDialogIDAndUserID(uint(request.DialogId), request.UserId)
	if err != nil {
		return resp, status.Error(codes.Code(code.DialogErrGetDialogUserByDialogIDAndUserIDFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) DeleteDialogUserByDialogIDAndUserIDRevert(ctx context.Context, request *v1.DeleteDialogUserByDialogIDAndUserIDRequest) (*v1.DeleteDialogUserByDialogIDAndUserIDResponse, error) {
	resp := &v1.DeleteDialogUserByDialogIDAndUserIDResponse{}

	if err := s.dr.UpdateDialogUserByDialogIDAndUserID(uint(request.DialogId), request.UserId, map[string]interface{}{
		"deleted_at": 0,
	}); err != nil {
		return resp, status.Error(codes.Code(code.DialogErrGetDialogUserByDialogIDAndUserIDFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s *Service) GetDialogByGroupId(ctx context.Context, in *v1.GetDialogByGroupIdRequest) (*v1.GetDialogByGroupIdResponse, error) {
	var resp = &v1.GetDialogByGroupIdResponse{}
	dialog, err := s.dr.GetDialogByGroupId(uint(in.GroupId))
	if err != nil {
		return resp, err
	}
	resp.DialogId = uint32(dialog.ID)
	resp.GroupId = uint32(dialog.GroupId)
	return resp, nil
}

func (s *Service) GetDialogByGroupIds(ctx context.Context, in *v1.GetDialogByGroupIdsRequest) (*v1.GetDialogByGroupIdsResponse, error) {
	var resp = &v1.GetDialogByGroupIdsResponse{}
	var idlist []uint
	if len(in.GroupId) > 0 {
		for _, id := range in.GroupId {
			idlist = append(idlist, uint(id))
		}
	}

	ids, err := s.dr.GetDialogByGroupIds(idlist)
	if err != nil {
		return resp, err
	}

	if len(ids) > 0 {
		for _, id := range ids {
			resp.Dialogs = append(resp.Dialogs, &v1.GetDialogByGroupIdResponse{
				DialogId: uint32(id.ID),
				GroupId:  uint32(id.GroupId),
			})
		}
	}
	return resp, nil
}
