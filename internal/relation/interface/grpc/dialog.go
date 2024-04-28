package grpc

import (
	"context"
	"fmt"
	v1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/domain/relation"
	"github.com/cossim/coss-server/internal/relation/infra/persistence"
	"github.com/cossim/coss-server/pkg/code"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

var _ v1.DialogServiceServer = &dialogServiceServer{}

type dialogServiceServer struct {
	db  *gorm.DB
	dr  relation.DialogRepository
	dur relation.DialogUserRepository
}

func (s *dialogServiceServer) CreateDialog(ctx context.Context, request *v1.CreateDialogRequest) (*v1.CreateDialogResponse, error) {
	//resp := &v1.CreateDialogResponse{}
	//
	//dialog, err := s.dr.CreateDialog(in.OwnerId, entity.DialogType(in.Type), uint(in.GroupId))
	//if err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrCreateDialogFailed.Code()), err.Error())
	//}

	dialog, err := s.dr.Create(ctx, &relation.CreateDialog{
		Type:    relation.DialogType(request.Type),
		OwnerId: request.OwnerId,
		GroupId: request.GroupId,
	})
	if err != nil {
		return nil, err
	}

	return &v1.CreateDialogResponse{
		Id:      dialog.ID,
		OwnerId: dialog.OwnerId,
		GroupId: dialog.ID,
		Type:    uint32(dialog.Type),
	}, nil
}

func (s *dialogServiceServer) CreateAndJoinDialogWithGroup(ctx context.Context, request *v1.CreateAndJoinDialogWithGroupRequest) (*v1.CreateAndJoinDialogWithGroupResponse, error) {
	resp := &v1.CreateAndJoinDialogWithGroupResponse{}

	//return resp, status.Error(codes.Aborted, formatErrorMessage(errors.New("测试回滚")))

	ids := []string{request.OwnerId}
	for _, v := range request.UserIds {
		ids = append(ids, v)
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		npo := persistence.NewRepositories(tx)
		//dialog, err := npo.Dr.CreateDialog(request.OwnerId, entity.DialogType(request.Type), uint(request.GroupId))
		//if err != nil {
		//	return status.Error(codes.Code(code.DialogErrCreateDialogFailed.Code()), fmt.Sprintf("failed to create dialog: %s", err.Error()))
		//}

		dialog, err := npo.Dr.Create(ctx, &relation.CreateDialog{
			Type:    relation.DialogType(request.Type),
			OwnerId: request.OwnerId,
			GroupId: request.GroupId,
		})
		if err != nil {
			return status.Error(codes.Code(code.DialogErrCreateDialogFailed.Code()), fmt.Sprintf("failed to create dialog: %s", err.Error()))
		}

		//_, err = npo.Dr.JoinBatchDialog(dialog.ID, ids)
		//if err != nil {
		//	return status.Error(codes.Code(code.DialogErrJoinDialogFailed.Code()), fmt.Sprintf("failed to join dialog: %s", err.Error()))
		//}
		_, err = npo.Dur.Creates(ctx, dialog.ID, ids)
		if err != nil {
			return status.Error(codes.Code(code.DialogErrJoinDialogFailed.Code()), fmt.Sprintf("failed to join dialog: %s", err.Error()))
		}
		resp.Id = dialog.ID
		resp.OwnerId = dialog.OwnerId
		resp.GroupId = dialog.ID
		resp.Type = uint32(dialog.Type)
		return nil
	}); err != nil {
		return resp, status.Error(codes.Aborted, fmt.Sprintf("failed to create dialog: %s", err.Error()))
	}

	return resp, nil
}

func (s *dialogServiceServer) CreateAndJoinDialogWithGroupRevert(ctx context.Context, request *v1.CreateAndJoinDialogWithGroupRequest) (*v1.CreateAndJoinDialogWithGroupResponse, error) {
	resp := &v1.CreateAndJoinDialogWithGroupResponse{}

	ids := []string{request.OwnerId}
	for _, id := range request.UserIds {
		ids = append(ids, id)
	}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		npo := persistence.NewRepositories(tx)
		//if err := npo.Dr.DeleteDialogUserByDialogIDAndUserID(uint(request.DialogId), ids); err != nil {
		//	return status.Error(codes.Code(code.DialogErrDeleteDialogUsersFailed.Code()), fmt.Sprintf("failed to delete dialog user revert : %s", err.Error()))
		//}

		if err := npo.Dur.DeleteByDialogIDAndUserID(ctx, request.DialogId, ids...); err != nil {
			return status.Error(codes.Code(code.DialogErrDeleteDialogUsersFailed.Code()), fmt.Sprintf("failed to delete dialog user revert : %s", err.Error()))
		}

		if err := npo.Dr.Delete(ctx, request.DialogId); err != nil {
			return status.Error(codes.Code(code.DialogErrDeleteDialogFailed.Code()), fmt.Sprintf("failed to delete dialog revert : %s", err.Error()))
		}

		return nil
	}); err != nil {
		return resp, status.Error(codes.Aborted, fmt.Sprintf("failed to create dialog revert: %s", err.Error()))
	}

	return resp, nil
}

func (s *dialogServiceServer) ConfirmFriendAndJoinDialog(ctx context.Context, request *v1.ConfirmFriendAndJoinDialogRequest) (*v1.ConfirmFriendAndJoinDialogResponse, error) {
	resp := &v1.ConfirmFriendAndJoinDialogResponse{}

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		npo := persistence.NewRepositories(tx)
		//dialog, err := npo.Dr.CreateDialog(request.OwnerId, entity.DialogType(request.Type), uint(request.GroupId))
		//if err != nil {
		//	return status.Error(codes.Code(code.DialogErrCreateDialogFailed.Code()), err.Error())
		//}

		dialog, err := npo.Dr.Create(ctx, &relation.CreateDialog{
			Type:    relation.DialogType(request.Type),
			OwnerId: request.OwnerId,
			GroupId: request.GroupId,
		})
		if err != nil {
			return status.Error(codes.Code(code.DialogErrCreateDialogFailed.Code()), err.Error())
		}

		//_, err = npo.Dr.JoinDialog(dialog.ID, request.OwnerId)
		//if err != nil {
		//	return status.Error(codes.Code(code.DialogErrJoinDialogFailed.Code()), err.Error())
		//}

		_, err = npo.Dur.Create(ctx, &relation.CreateDialogUser{
			DialogID: dialog.ID,
			UserID:   request.UserId,
		})
		if err != nil {
			return status.Error(codes.Code(code.DialogErrJoinDialogFailed.Code()), err.Error())
		}
		resp.Id = dialog.ID
		resp.OwnerId = request.OwnerId
		resp.Type = v1.DialogType(dialog.Type)
		resp.GroupId = dialog.GroupId

		return nil
	}); err != nil {
		return resp, err
	}

	return resp, nil
}

func (s *dialogServiceServer) ConfirmFriendAndJoinDialogRevert(ctx context.Context, request *v1.ConfirmFriendAndJoinDialogRevertRequest) (*v1.ConfirmFriendAndJoinDialogRevertResponse, error) {
	resp := &v1.ConfirmFriendAndJoinDialogRevertResponse{}

	if err := s.dur.DeleteByDialogIDAndUserID(ctx, request.DialogId, request.OwnerId, request.UserId); err != nil {
		return nil, err
	}

	if err := s.dr.Delete(ctx, request.DialogId); err != nil {
		return nil, err
	}

	//if err := s.dr.DeleteDialogUserByDialogIDAndUserID(uint(request.DialogId), []string{request.UserId}); err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrCreateDialogFailed.Code()), err.Error())
	//}
	//
	//if err := s.dr.DeleteDialogUserByDialogIDAndUserID(uint(request.DialogId), []string{request.OwnerId}); err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrCreateDialogFailed.Code()), err.Error())
	//}

	//if err := s.dr.DeleteDialogByDialogID(uint(request.DialogId)); err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrCreateDialogFailed.Code()), err.Error())
	//}

	return resp, nil
}

func (s *dialogServiceServer) JoinDialog(ctx context.Context, request *v1.JoinDialogRequest) (*v1.JoinDialogResponse, error) {
	resp := &v1.JoinDialogResponse{}
	//_, err := s.dr.JoinDialog(uint(in.DialogId), in.UserId)
	//if err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrJoinDialogFailed.Code()), err.Error())
	//}

	_, err := s.dur.Create(ctx, &relation.CreateDialogUser{
		DialogID: request.DialogId,
		UserID:   request.UserId,
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *dialogServiceServer) JoinDialogRevert(ctx context.Context, request *v1.JoinDialogRequest) (*v1.JoinDialogResponse, error) {
	resp := &v1.JoinDialogResponse{}
	//if err := s.dr.DeleteDialogUserByDialogIDAndUserID(uint(request.DialogId), []string{request.UserId}); err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrJoinDialogFailed.Code()), err.Error())
	//}

	if err := s.dur.DeleteByDialogIDAndUserID(ctx, request.DialogId, request.UserId); err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *dialogServiceServer) GetUserDialogList(ctx context.Context, request *v1.GetUserDialogListRequest) (*v1.GetUserDialogListResponse, error) {
	resp := &v1.GetUserDialogListResponse{}
	//ids, total, err := s.dr.GetUserDialogs(in.UserId, int(in.PageSize), int(in.PageNum))
	//if err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrGetUserDialogListFailed.Code()), err.Error())
	//}

	//ids, total, err := s.dr.GetUserDialogs(request.UserId, int(request.PageSize), int(request.PageNum))
	//if err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrGetUserDialogListFailed.Code()), err.Error())
	//}

	dialogs, err := s.dur.Find(ctx, &relation.DialogUserQuery{
		UserID:   []string{request.UserId},
		PageSize: int(request.PageSize),
		PageNum:  int(request.PageNum),
	})
	if err != nil {
		return nil, err
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

	//if len(request.DialogIds) == 0 || len(request.DialogIds) < 0 {
	//	return resp, status.Error(codes.Code(code.DialogErrGetUserDialogListFailed.Code()), "dialog ids is empty")
	//}

	nids := make([]uint32, 0)
	for _, id := range request.DialogIds {
		nids = append(nids, id)
	}
	//infos, err := s.dr.GetDialogsByIDs(nids)
	//if err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrGetUserDialogListFailed.Code()), err.Error())
	//}

	infos, err := s.dr.Find(ctx, &relation.DialogQuery{
		DialogID: nids,
	})
	if err != nil {
		return resp, status.Error(codes.Code(code.DialogErrGetUserDialogListFailed.Code()), err.Error())
	}

	fmt.Println("infos => ", infos)

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
	//users, err := s.dr.GetDialogUsersByDialogID(uint(request.DialogId))
	//if err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrGetUserDialogListFailed.Code()), err.Error())
	//}

	users, err := s.dur.ListByDialogID(ctx, request.DialogId)
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

func (s *dialogServiceServer) DeleteDialogByIds(ctx context.Context, request *v1.DeleteDialogByIdsRequest) (*v1.DeleteDialogByIdsResponse, error) {
	var resp = &v1.DeleteDialogByIdsResponse{}
	uintIds := make([]uint32, 0)
	if len(request.DialogIds) > 0 {
		for _, id := range request.DialogIds {
			uintIds = append(uintIds, id)
		}
	}
	//if err := s.dr.DeleteDialogByIds(uintIds); err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrDeleteDialogFailed.Code()), err.Error())
	//}

	if err := s.dr.Delete(ctx, uintIds...); err != nil {
		return resp, status.Error(codes.Code(code.DialogErrDeleteDialogFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s *dialogServiceServer) DeleteDialogById(ctx context.Context, request *v1.DeleteDialogByIdRequest) (*v1.DeleteDialogByIdResponse, error) {
	resp := &v1.DeleteDialogByIdResponse{}

	//if err := s.dr.DeleteDialogByDialogID(uint(request.DialogId)); err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrDeleteDialogFailed.Code()), err.Error())
	//}

	if err := s.dr.Delete(ctx, request.DialogId); err != nil {
		return resp, status.Error(codes.Code(code.DialogErrDeleteDialogFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s *dialogServiceServer) DeleteDialogByIdRevert(ctx context.Context, request *v1.DeleteDialogByIdRequest) (*v1.DeleteDialogByIdResponse, error) {
	resp := &v1.DeleteDialogByIdResponse{}

	//if err := s.dr.UpdateDialogByDialogID(uint(request.DialogId), map[string]interface{}{
	//	"deleted_at": 0,
	//}); err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrDeleteDialogFailed.Code()), err.Error())
	//}

	if err := s.dr.UpdateFields(ctx, uint(request.DialogId), map[string]interface{}{
		"deleted_at": 0,
	}); err != nil {
		return resp, status.Error(codes.Code(code.DialogErrDeleteDialogFailed.Code()), err.Error())
	}

	//if err := s.dr.UpdateDialogByDialogID(uint(request.DialogId), map[string]interface{}{
	//	"deleted_at": 0,
	//}); err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrDeleteDialogFailed.Code()), err.Error())
	//}

	return resp, nil
}

func (s *dialogServiceServer) DeleteDialogUsersByDialogID(ctx context.Context, request *v1.DeleteDialogUsersByDialogIDRequest) (*v1.DeleteDialogUsersByDialogIDResponse, error) {
	resp := &v1.DeleteDialogUsersByDialogIDResponse{}
	//if err := s.dr.DeleteDialogUserByDialogID(uint(in.DialogId)); err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrDeleteDialogUsersFailed.Code()), err.Error())
	//}

	if err := s.dur.Delete(ctx, request.DialogId); err != nil {
		return resp, status.Error(codes.Code(code.DialogErrDeleteDialogUsersFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s *dialogServiceServer) DeleteDialogUsersByDialogIDRevert(ctx context.Context, request *v1.DeleteDialogUsersByDialogIDRequest) (*v1.DeleteDialogUsersByDialogIDResponse, error) {
	resp := &v1.DeleteDialogUsersByDialogIDResponse{}
	//if err := s.dr.UpdateDialogUserByDialogID(uint(request.DialogId), map[string]interface{}{
	//	"deleted_at": 0,
	//}); err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrDeleteDialogUsersFailed.Code()), err.Error())
	//}

	if err := s.dur.UpdateFields(ctx, request.DialogId, map[string]interface{}{
		"deleted_at": 0,
	}); err != nil {
		return resp, status.Error(codes.Code(code.DialogErrDeleteDialogFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s *dialogServiceServer) GetDialogUserByDialogIDAndUserID(ctx context.Context, request *v1.GetDialogUserByDialogIDAndUserIdRequest) (*v1.GetDialogUserByDialogIDAndUserIdResponse, error) {
	resp := &v1.GetDialogUserByDialogIDAndUserIdResponse{}
	//user, err := s.dr.GetDialogUserByDialogIDAndUserID(uint(request.DialogId), request.UserId)
	//if err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrGetDialogUserByDialogIDAndUserIDFailed.Code()), err.Error())
	//}

	users, err := s.dur.Find(ctx, &relation.DialogUserQuery{
		DialogID: []uint32{request.DialogId},
		UserID:   []string{request.UserId},
		Force:    true,
	})
	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return resp, status.Error(codes.Code(code.DialogErrGetDialogUserByDialogIDAndUserIDFailed.Code()), "对话用户列表为空")
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

func (s *dialogServiceServer) DeleteDialogUserByDialogIDAndUserID(ctx context.Context, request *v1.DeleteDialogUserByDialogIDAndUserIDRequest) (*v1.DeleteDialogUserByDialogIDAndUserIDResponse, error) {
	var resp = &v1.DeleteDialogUserByDialogIDAndUserIDResponse{}

	//err := s.dr.DeleteDialogUserByDialogIDAndUserID(uint(request.DialogId), []string{request.UserId})
	//if err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrGetDialogUserByDialogIDAndUserIDFailed.Code()), err.Error())
	//}

	if err := s.dur.DeleteByDialogIDAndUserID(ctx, request.DialogId, request.UserId); err != nil {
		return resp, status.Error(codes.Code(code.DialogErrGetDialogUserByDialogIDAndUserIDFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s *dialogServiceServer) DeleteDialogUserByDialogIDAndUserIDRevert(ctx context.Context, request *v1.DeleteDialogUserByDialogIDAndUserIDRequest) (*v1.DeleteDialogUserByDialogIDAndUserIDResponse, error) {
	resp := &v1.DeleteDialogUserByDialogIDAndUserIDResponse{}

	//if err := s.dr.UpdateDialogUserColumnByDialogIDAndUserId(uint(request.DialogId), request.UserId, "deleted_at", 0); err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrGetDialogUserByDialogIDAndUserIDFailed.Code()), err.Error())
	//}

	finds, err := s.dur.Find(ctx, &relation.DialogUserQuery{
		DialogID: []uint32{request.DialogId},
		UserID:   []string{request.UserId},
	})
	if err != nil {
		return nil, err
	}

	if len(finds) == 0 {
		return nil, nil
	}

	if err := s.dur.UpdateFields(ctx, finds[0].ID, map[string]interface{}{
		"deleted_at": 0,
	}); err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *dialogServiceServer) GetDialogByGroupId(ctx context.Context, request *v1.GetDialogByGroupIdRequest) (*v1.GetDialogByGroupIdResponse, error) {
	var resp = &v1.GetDialogByGroupIdResponse{}
	//dialog, err := s.dr.GetDialogByGroupId(uint(in.GroupId))
	//if err != nil {
	//	return resp, err
	//}

	dialog, err := s.dr.GetByGroupID(ctx, request.GroupId)
	if err != nil {
		return resp, err
	}

	resp.DialogId = dialog.ID
	resp.GroupId = dialog.GroupId
	resp.Type = uint32(dialog.Type)
	resp.CreateAt = dialog.CreatedAt
	return resp, nil
}

func (s *dialogServiceServer) GetDialogByGroupIds(ctx context.Context, request *v1.GetDialogByGroupIdsRequest) (*v1.GetDialogByGroupIdsResponse, error) {
	resp := &v1.GetDialogByGroupIdsResponse{}

	if len(request.GroupId) == 0 {
		return resp, nil
	}

	//ids, err := s.dr.GetDialogByGroupIds(idlist)
	//if err != nil {
	//	return resp, err
	//}

	ids, err := s.dr.Find(ctx, &relation.DialogQuery{
		GroupID: request.GroupId,
	})
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return resp, nil
	}

	for _, id := range ids {
		resp.Dialogs = append(resp.Dialogs, &v1.GetDialogByGroupIdResponse{
			DialogId: id.ID,
			GroupId:  id.GroupId,
		})
	}
	return resp, nil
}

func (s *dialogServiceServer) CloseOrOpenDialog(ctx context.Context, request *v1.CloseOrOpenDialogRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}
	//if err := s.dr.UpdateDialogUserColumnByDialogIDAndUserId(uint(request.DialogId), request.UserId, "is_show", request.Action); err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrCloseOrOpenDialogFailed.Code()), err.Error())
	//}

	var isShow bool
	switch request.Action {
	case v1.CloseOrOpenDialogType_CLOSE:
		isShow = false
	case v1.CloseOrOpenDialogType_OPEN:
		isShow = true
	}

	if err := s.dur.UpdateDialogStatus(ctx, &relation.UpdateDialogStatusParam{
		DialogID: request.DialogId,
		UserID:   []string{request.UserId},
		IsShow:   &isShow,
	}); err != nil {
		return nil, err
	}

	//if err := s.dr.UpdateDialogUserColumnByDialogIDAndUserId(uint(request.DialogId), request.UserId, "is_show", request.Action); err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrCloseOrOpenDialogFailed.Code()), err.Error())
	//}

	return resp, nil
}

func (s *dialogServiceServer) TopOrCancelTopDialog(ctx context.Context, request *v1.TopOrCancelTopDialogRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}
	var topAt int64
	switch request.Action {
	case v1.TopOrCancelTopDialogType_CANCEL_TOP:
		topAt = 0
		//if err := s.dr.UpdateDialogUserColumnByDialogIDAndUserId(uint(request.DialogId), request.UserId, "top_at", 0); err != nil {
		//	return resp, status.Error(codes.Code(code.DialogErrCloseOrOpenDialogFailed.Code()), err.Error())
		//}
	case v1.TopOrCancelTopDialogType_TOP:
		topAt = ptime.Now()
		//if err := s.dr.UpdateDialogUserColumnByDialogIDAndUserId(uint(request.DialogId), request.UserId, "top_at", ptime.Now()); err != nil {
		//	return resp, status.Error(codes.Code(code.DialogErrCloseOrOpenDialogFailed.Code()), err.Error())
		//}
	}

	if err := s.dur.UpdateDialogStatus(ctx, &relation.UpdateDialogStatusParam{
		DialogID: request.DialogId,
		UserID:   []string{request.UserId},
		TopAt:    &topAt,
	}); err != nil {
		return resp, status.Error(codes.Code(code.DialogErrCloseOrOpenDialogFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s *dialogServiceServer) GetDialogById(ctx context.Context, request *v1.GetDialogByIdRequest) (*v1.Dialog, error) {
	resp := &v1.Dialog{}
	//dialog, err := s.dr.GetDialogById(request.DialogId)
	//if err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrGetDialogByIdFailed.Code()), err.Error())
	//}

	dialog, err := s.dr.Get(ctx, request.DialogId)
	if err != nil {
		return resp, status.Error(codes.Code(code.DialogErrGetDialogByIdFailed.Code()), err.Error())
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
	//users, err := s.dr.GetDialogAllUsers(uint(in.DialogId))
	//if err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrGetUserDialogListFailed.Code()), err.Error())
	//}

	users, err := s.dur.ListByDialogID(ctx, uint32(uint(in.DialogId)))
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

func (s *dialogServiceServer) BatchCloseOrOpenDialog(ctx context.Context, request *v1.BatchCloseOrOpenDialogRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}
	//err := s.dr.UpdateDialogUserByDialogIDAndUserIds(uint(request.DialogId), request.UserIds, "is_show", request.Action)
	//if err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrCloseOrOpenDialogFailed.Code()), err.Error())
	//}

	var isShow bool
	switch request.Action {
	case v1.CloseOrOpenDialogType_CLOSE:
		isShow = false
	case v1.CloseOrOpenDialogType_OPEN:
		isShow = true
	}

	if err := s.dur.UpdateDialogStatus(ctx, &relation.UpdateDialogStatusParam{
		DialogID: request.DialogId,
		UserID:   request.UserIds,
		IsShow:   &isShow,
	}); err != nil {
		return resp, err
	}

	return resp, nil
}

func (s *dialogServiceServer) GetDialogTargetUserId(ctx context.Context, request *v1.GetDialogTargetUserIdRequest) (*v1.GetDialogTargetUserIdResponse, error) {
	resp := &v1.GetDialogTargetUserIdResponse{}
	//userIDs, err := s.dr.GetDialogTargetUserId(uint(request.DialogId), request.UserId)
	//if err != nil {
	//	return resp, status.Error(codes.Code(code.DialogErrGetTargetIdFailed.Code()), err.Error())
	//}

	userIDs, err := s.dur.ListByDialogID(ctx, request.DialogId)
	if err != nil {
		return resp, status.Error(codes.Code(code.DialogErrGetTargetIdFailed.Code()), err.Error())
	}

	var ids []string
	for _, userID := range userIDs {
		if userID.UserId != request.UserId {
			ids = append(ids, userID.UserId)
		}
	}

	resp.UserIds = ids
	return resp, nil
}
