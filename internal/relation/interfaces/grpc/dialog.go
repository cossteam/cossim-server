package grpc

import (
	"context"
	"errors"
	v1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/internal/relation/infra/persistence"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils"
	"google.golang.org/protobuf/types/known/emptypb"
)

var _ v1.DialogServiceServer = &dialogServiceServer{}

type dialogServiceServer struct {
	repos *persistence.Repositories
}

func (s *dialogServiceServer) GetUserDialogList(ctx context.Context, request *v1.GetUserDialogListRequest) (*v1.GetUserDialogListResponse, error) {
	resp := &v1.GetUserDialogListResponse{}

	dialogs, err := s.repos.DialogUserRepo.Find(ctx, &repository.DialogUserQuery{
		UserID:   []string{request.UserId},
		PageSize: int(request.PageSize),
		PageNum:  int(request.PageNum),
		IsShow:   true,
	})
	if err != nil {
		return nil, code.WrapCodeToGRPC(code.DialogErrGetUserDialogListFailed.Reason(utils.FormatErrorStack(err)))
	}

	nids := make([]uint32, 0)
	if len(dialogs) > 0 {
		for _, dialog := range dialogs {
			nids = append(nids, dialog.DialogId)

		}
	}
	resp.Total = uint64(len(dialogs))
	resp.DialogIds = nids
	return resp, nil
}

func (s *dialogServiceServer) GetDialogByIds(ctx context.Context, request *v1.GetDialogByIdsRequest) (*v1.GetDialogByIdsResponse, error) {
	resp := &v1.GetDialogByIdsResponse{}

	nids := make([]uint32, 0)
	for _, id := range request.DialogIds {
		nids = append(nids, id)
	}

	infos, err := s.repos.DialogRepo.Find(ctx, &repository.DialogQuery{
		DialogID: nids,
	})
	if err != nil {
		return nil, code.WrapCodeToGRPC(code.DialogErrGetUserDialogListFailed.Reason(utils.FormatErrorStack(err)))
	}

	var dialogInfos []*v1.Dialog
	if len(infos) > 0 {
		for _, info := range infos {
			dialogInfos = append(dialogInfos, &v1.Dialog{
				Id:       info.ID,
				OwnerId:  info.OwnerId,
				GroupId:  info.GroupId,
				Type:     uint32(info.Type),
				CreateAt: info.CreatedAt,
			})
		}
	}
	resp.Dialogs = dialogInfos
	return resp, nil
}

func (s *dialogServiceServer) GetDialogUsersByDialogID(ctx context.Context, request *v1.GetDialogUsersByDialogIDRequest) (*v1.GetDialogUsersByDialogIDResponse, error) {
	resp := &v1.GetDialogUsersByDialogIDResponse{}

	users, err := s.repos.DialogUserRepo.ListByDialogID(ctx, request.DialogId)
	if err != nil {
		return nil, code.WrapCodeToGRPC(code.DialogErrGetUserDialogListFailed.Reason(utils.FormatErrorStack(err)))
	}

	var ids []string
	for _, id := range users {
		ids = append(ids, id.UserId)
	}
	resp.UserIds = ids
	return resp, nil
}

func (s *dialogServiceServer) DeleteDialogById(ctx context.Context, request *v1.DeleteDialogByIdRequest) (*v1.DeleteDialogByIdResponse, error) {
	resp := &v1.DeleteDialogByIdResponse{}

	if err := s.repos.DialogRepo.Delete(ctx, request.DialogId); err != nil {
		return nil, code.WrapCodeToGRPC(code.DialogErrDeleteDialogFailed.Reason(utils.FormatErrorStack(err)))
	}

	return resp, nil
}

func (s *dialogServiceServer) DeleteDialogByIdRevert(ctx context.Context, request *v1.DeleteDialogByIdRequest) (*v1.DeleteDialogByIdResponse, error) {
	resp := &v1.DeleteDialogByIdResponse{}

	if err := s.repos.DialogRepo.UpdateFields(ctx, uint(request.DialogId), map[string]interface{}{
		"deleted_at": 0,
	}); err != nil {
		return nil, code.WrapCodeToGRPC(code.DialogErrDeleteDialogFailed.Reason(utils.FormatErrorStack(err)))
	}

	return resp, nil
}

func (s *dialogServiceServer) DeleteDialogUsersByDialogID(ctx context.Context, request *v1.DeleteDialogUsersByDialogIDRequest) (*v1.DeleteDialogUsersByDialogIDResponse, error) {
	resp := &v1.DeleteDialogUsersByDialogIDResponse{}

	if err := s.repos.DialogUserRepo.Delete(ctx, request.DialogId); err != nil {
		return nil, code.WrapCodeToGRPC(code.DialogErrDeleteDialogUsersFailed.Reason(utils.FormatErrorStack(err)))
	}

	return resp, nil
}

func (s *dialogServiceServer) DeleteDialogUsersByDialogIDRevert(ctx context.Context, request *v1.DeleteDialogUsersByDialogIDRequest) (*v1.DeleteDialogUsersByDialogIDResponse, error) {
	resp := &v1.DeleteDialogUsersByDialogIDResponse{}

	if err := s.repos.DialogUserRepo.UpdateFields(ctx, request.DialogId, map[string]interface{}{
		"deleted_at": 0,
	}); err != nil {
		return nil, code.WrapCodeToGRPC(code.DialogErrDeleteDialogFailed.Reason(utils.FormatErrorStack(err)))
	}

	return resp, nil
}

// TODO
func (s *dialogServiceServer) GetDialogUserByDialogIDAndUserID(ctx context.Context, request *v1.GetDialogUserByDialogIDAndUserIdRequest) (*v1.GetDialogUserByDialogIDAndUserIdResponse, error) {
	resp := &v1.GetDialogUserByDialogIDAndUserIdResponse{}

	users, err := s.repos.DialogUserRepo.Find(ctx, &repository.DialogUserQuery{
		DialogID: []uint32{request.DialogId},
		UserID:   []string{request.UserId},
		Force:    true,
	})
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return nil, code.WrapCodeToGRPC(code.DialogErrGetDialogUserByDialogIDAndUserIDFailed.CustomMessage("对话用户列表为空"))
	}

	var isShow uint32

	if users[0].IsShow == true {
		isShow = 1
	}

	resp.DialogId = users[0].DialogId
	resp.UserId = users[0].UserId
	resp.IsShow = isShow
	resp.TopAt = uint64(users[0].TopAt)

	return resp, nil
}

func (s *dialogServiceServer) GetDialogByGroupId(ctx context.Context, request *v1.GetDialogByGroupIdRequest) (*v1.GetDialogByGroupIdResponse, error) {
	resp := &v1.GetDialogByGroupIdResponse{}

	dialog, err := s.repos.DialogRepo.GetByGroupID(ctx, request.GroupId)
	if err != nil {
		if errors.Is(err, code.NotFound) {
			return nil, code.WrapCodeToGRPC(code.DialogErrGetDialogByIdFailed.CustomMessage("对话不存在").Reason(utils.FormatErrorStack(err)))
		}
		return nil, code.WrapCodeToGRPC(code.DialogErrGetDialogByIdFailed.Reason(utils.FormatErrorStack(err)))
	}

	resp.DialogId = dialog.ID
	resp.GroupId = dialog.GroupId
	resp.Type = uint32(dialog.Type)
	resp.CreateAt = dialog.CreatedAt
	return resp, nil
}

// TODO
func (s *dialogServiceServer) CloseOrOpenDialog(ctx context.Context, request *v1.CloseOrOpenDialogRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}

	var isShow bool
	switch request.Action {
	case v1.CloseOrOpenDialogType_CLOSE:
		isShow = false
	case v1.CloseOrOpenDialogType_OPEN:
		isShow = true
	}

	if err := s.repos.DialogUserRepo.UpdateDialogStatus(ctx, &repository.UpdateDialogStatusParam{
		DialogID: request.DialogId,
		UserID:   []string{request.UserId},
		IsShow:   &isShow,
	}); err != nil {
		return nil, code.WrapCodeToGRPC(code.MyCustomErrorCode.CustomMessage("更新对话状态失败").Reason(utils.FormatErrorStack(err)))
	}

	return resp, nil
}

func (s *dialogServiceServer) GetDialogById(ctx context.Context, request *v1.GetDialogByIdRequest) (*v1.Dialog, error) {
	resp := &v1.Dialog{}

	dialog, err := s.repos.DialogRepo.Get(ctx, request.DialogId)
	if err != nil {
		return nil, code.WrapCodeToGRPC(code.DialogErrGetDialogByIdFailed.Reason(utils.FormatErrorStack(err)))
	}

	resp.Id = dialog.ID
	resp.GroupId = dialog.GroupId
	resp.Type = uint32(dialog.Type)
	resp.CreateAt = dialog.CreatedAt
	resp.OwnerId = dialog.OwnerId

	return resp, nil
}

func (s *dialogServiceServer) GetAllUsersInConversation(ctx context.Context, in *v1.GetAllUsersInConversationRequest) (*v1.GetAllUsersInConversationResponse, error) {
	resp := &v1.GetAllUsersInConversationResponse{}

	users, err := s.repos.DialogUserRepo.ListByDialogID(ctx, uint32(uint(in.DialogId)))
	if err != nil {
		return nil, code.WrapCodeToGRPC(code.DialogErrGetUserDialogListFailed.Reason(utils.FormatErrorStack(err)))
	}

	var ids []string
	for _, id := range users {
		ids = append(ids, id.UserId)
	}
	resp.UserIds = ids
	return resp, nil
}

func (s *dialogServiceServer) BatchCloseOrOpenDialog(ctx context.Context, request *v1.BatchCloseOrOpenDialogRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}

	var isShow bool
	switch request.Action {
	case v1.CloseOrOpenDialogType_CLOSE:
		isShow = false
	case v1.CloseOrOpenDialogType_OPEN:
		isShow = true
	}

	if err := s.repos.DialogUserRepo.UpdateDialogStatus(ctx, &repository.UpdateDialogStatusParam{
		DialogID: request.DialogId,
		UserID:   request.UserIds,
		IsShow:   &isShow,
	}); err != nil {
		return nil, code.WrapCodeToGRPC(code.MyCustomErrorCode.CustomMessage("更新对话状态失败").Reason(utils.FormatErrorStack(err)))
	}

	return resp, nil
}
