package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils/time"
	v1 "github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/cossim/coss-server/service/relation/domain/entity"
	"github.com/cossim/coss-server/service/relation/infrastructure/persistence"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

func (s *Service) GetGroupUserIDs(ctx context.Context, id *v1.GroupIDRequest) (*v1.UserIdsResponse, error) {
	resp := &v1.UserIdsResponse{}

	// 调用持久层方法获取用户群关系列表
	userGroupIDs, err := s.grr.GetGroupUserIDs(id.GetGroupId())
	if err != nil {
		return resp, status.Error(codes.Code(code.GroupErrGetBatchGroupInfoByIDsFailed.Code()), err.Error())
	}

	resp.UserIds = userGroupIDs
	return resp, nil
}

func (s *Service) GetUserGroupIDs(ctx context.Context, request *v1.GetUserGroupIDsRequest) (*v1.GetUserGroupIDsResponse, error) {
	resp := &v1.GetUserGroupIDsResponse{}

	ds, err := s.grr.GetUserJoinedGroupIDs(request.UserId)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrGetGroupIDsFailed.Code()), err.Error())
	}

	for _, v := range ds {
		resp.GroupId = append(resp.GroupId, v)
	}
	return resp, nil
}

func (s *Service) GetUserGroupList(ctx context.Context, request *v1.GetUserGroupListRequest) (*v1.GetUserGroupListResponse, error) {
	resp := &v1.GetUserGroupListResponse{}

	isAdmin := false

	// 获取用户所属群组的ID列表
	groupIDs, err := s.grr.GetUserGroupIDs(request.UserId)
	if err != nil {
		return resp, status.Error(codes.Code(code.GroupErrGetBatchGroupInfoByIDsFailed.Code()), err.Error())
	}

	for _, gid := range groupIDs {
		// 获取群组的申请记录
		reqList, err := s.grr.GetJoinRequestBatchListByID([]uint32{gid})
		if err != nil {
			return resp, status.Error(codes.Code(code.RelationUserErrGetRequestListFailed.Code()), err.Error())
		}

		for _, v := range reqList {
			fmt.Println("id => ", v.UserID)
			// 判断用户是否为群组管理员
			gr, err := s.grr.GetUserGroupByID(gid, v.UserID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return resp, status.Error(codes.Code(code.RelationUserErrGetRequestListFailed.Code()), err.Error())
			}

			if gr.Identity != entity.IdentityUser {
				isAdmin = true
			} else {
				isAdmin = false
			}

			if isAdmin {
				resp.UserGroupRequestList = append(resp.UserGroupRequestList, &v1.UserGroupRequestList{
					GroupId:   gid,
					UserId:    v.UserID,
					Msg:       v.Remark,
					CreatedAt: v.CreatedAt,
				})
			} else if !isAdmin && v.Inviter != "" && v.UserID == request.UserId {
				resp.UserGroupRequestList = append(resp.UserGroupRequestList, &v1.UserGroupRequestList{
					GroupId:   gid,
					UserId:    v.Inviter,
					Msg:       v.Remark,
					CreatedAt: v.CreatedAt,
				})
			}
		}
	}

	return resp, nil
}

func (s *Service) RemoveFromGroup(ctx context.Context, request *v1.RemoveFromGroupRequest) (*v1.RemoveFromGroupResponse, error) {
	resp := &v1.RemoveFromGroupResponse{}

	if err := s.grr.DeleteRelationByID(request.GroupId, request.UserId); err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrRemoveUserFromGroupFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s *Service) LeaveGroup(ctx context.Context, request *v1.LeaveGroupRequest) (*v1.LeaveGroupResponse, error) {
	resp := &v1.LeaveGroupResponse{}
	if err := s.grr.DeleteRelationByID(request.GroupId, request.UserId); err != nil {
		return resp, status.Error(codes.Aborted, err.Error())
	}
	return resp, nil
}

func (s *Service) LeaveGroupRevert(ctx context.Context, request *v1.LeaveGroupRequest) (*v1.LeaveGroupResponse, error) {
	fmt.Println("revert leave group req => ", request)
	resp := &v1.LeaveGroupResponse{}

	if err := s.grr.UpdateRelationColumnByGroupAndUser(request.GroupId, request.UserId, "deleted_at", 0); err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrLeaveGroupFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s *Service) GetGroupRelation(ctx context.Context, request *v1.GetGroupRelationRequest) (*v1.GetGroupRelationResponse, error) {
	resp := &v1.GetGroupRelationResponse{}

	relation, err := s.grr.GetUserGroupByID(request.GroupId, request.UserId)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrGroupRelationFailed.Code()), err.Error())
	}
	resp.GroupId = uint32(relation.GroupID)
	resp.UserId = relation.UserID
	resp.Identity = v1.GroupIdentity(uint32(relation.Identity))
	resp.MuteEndTime = relation.MuteEndTime
	resp.IsSilent = v1.GroupSilentNotificationType(relation.SilentNotification)
	return resp, nil
}

func (s *Service) GetBatchGroupRelation(ctx context.Context, request *v1.GetBatchGroupRelationRequest) (*v1.GetBatchGroupRelationResponse, error) {
	resp := &v1.GetBatchGroupRelationResponse{}

	grs, err := s.grr.GetUserGroupByIDs(request.GroupId, request.UserIds)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationGroupErrRelationNotFound.Code()), code.RelationGroupErrRelationNotFound.Message())
		}
		return resp, status.Error(codes.Code(code.RelationGroupErrGroupRelationFailed.Code()), err.Error())
	}

	for _, gr := range grs {
		resp.GroupRelationResponses = append(resp.GroupRelationResponses, &v1.GetGroupRelationResponse{
			UserId:      gr.UserID,
			GroupId:     uint32(gr.GroupID),
			Identity:    v1.GroupIdentity(uint32(gr.Identity)),
			MuteEndTime: gr.MuteEndTime,
		})
	}
	return resp, nil
}

func (s *Service) DeleteGroupRelationByGroupId(ctx context.Context, in *v1.GroupIDRequest) (*emptypb.Empty, error) {
	err := s.grr.DeleteGroupRelationByID(in.GroupId)
	if err != nil {
		return &emptypb.Empty{}, status.Error(codes.Aborted, err.Error())
	}
	return &emptypb.Empty{}, nil
}

func (s *Service) DeleteGroupRelationByGroupIdRevert(ctx context.Context, request *v1.GroupIDRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}
	fmt.Println("DeleteGroupRelationByGroupIdRevert req => ", request)

	if err := s.grr.UpdateGroupRelationByGroupID(request.GroupId, map[string]interface{}{
		"deleted_at": 0,
	}); err != nil {
		return resp, status.Error(codes.Code(code.GroupErrDeleteGroupFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) GetGroupAdminIds(ctx context.Context, in *v1.GroupIDRequest) (*v1.UserIdsResponse, error) {
	var resp = &v1.UserIdsResponse{}
	ids, err := s.grr.GetGroupAdminIds(in.GroupId)
	if err != nil {
		return resp, err
	}
	resp.UserIds = ids
	return resp, nil
}

func (s *Service) GetUserManageGroupID(ctx context.Context, request *v1.GetUserManageGroupIDRequest) (*v1.GetUserManageGroupIDResponse, error) {
	resp := &v1.GetUserManageGroupIDResponse{}

	ids, err := s.grr.GetUserManageGroupIDs(request.UserId)
	if err != nil {
		return resp, status.Error(codes.Code(code.GroupErrGetBatchGroupInfoByIDsFailed.Code()), err.Error())
	}

	for _, id := range ids {
		resp.GroupIDs = append(resp.GroupIDs, &v1.GroupIDRequest{GroupId: id})
	}

	return resp, nil
}

func (s *Service) DeleteGroupRelationByGroupIdAndUserID(ctx context.Context, request *v1.DeleteGroupRelationByGroupIdAndUserIDRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}

	//return resp, status.Error(codes.Aborted, formatErrorMessage(errors.New("测试回滚")))

	if err := s.grr.DeleteUserGroupRelationByGroupIDAndUserID(request.GroupID, request.UserID); err != nil {
		//return resp, status.Error(codes.Code(code.GroupErrDeleteUserGroupRelationFailed.Code()), err.Error())
		return resp, status.Error(codes.Aborted, fmt.Sprintf(code.GroupErrDeleteUserGroupRelationFailed.Message()+" : "+err.Error()))
	}
	return resp, nil
}

func (s *Service) DeleteGroupRelationByGroupIdAndUserIDRevert(ctx context.Context, request *v1.DeleteGroupRelationByGroupIdAndUserIDRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}
	if err := s.grr.UpdateRelationColumnByGroupAndUser(request.GroupID, request.UserID, "deleted_at", 0); err != nil {
		//return resp, status.Error(codes.Code(code.GroupErrDeleteUserGroupRelationFailed.Code()), err.Error())
		return resp, status.Error(codes.Code(code.GroupErrDeleteUserGroupRelationRevertFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) CreateGroupAndInviteUsers(ctx context.Context, request *v1.CreateGroupAndInviteUsersRequest) (*v1.CreateGroupAndInviteUsersResponse, error) {
	resp := &v1.CreateGroupAndInviteUsersResponse{}

	//return resp, status.Error(codes.Aborted, formatErrorMessage(errors.New("测试回滚")))

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		npo := persistence.NewRepositories(tx)
		//创建群聊对话
		dialog, err := npo.Dr.CreateDialog(request.UserID, entity.GroupDialog, uint(request.GroupId))
		if err != nil {
			return err
		}

		//群主加入对话
		_, err = npo.Dr.JoinDialog(dialog.ID, request.UserID)
		if err != nil {
			return err
		}
		resp.DialogId = uint32(dialog.ID)
		//群主加入群聊
		owner := &entity.GroupRelation{
			UserID:   request.UserID,
			GroupID:  uint(request.GroupId),
			Identity: entity.IdentityOwner,
		}
		_, err = npo.Grr.CreateRelation(owner)
		if err != nil {
			return err
		}

		//发送邀请给其他成员
		requests := make([]*entity.GroupJoinRequest, 0)
		for _, v := range request.Member {
			req := &entity.GroupJoinRequest{
				UserID:      v,
				GroupID:     uint(request.GroupId),
				Status:      entity.Invitation,
				Inviter:     request.UserID,
				InviterTime: time.Now(),
			}
			requests = append(requests, req)
		}
		if _, err := npo.Gjqr.AddJoinRequestBatch(requests); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return resp, status.Error(codes.Aborted, err.Error())
	}
	return resp, nil
}

func (s *Service) CreateGroupAndInviteUsersRevert(ctx context.Context, request *v1.CreateGroupAndInviteUsersRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}
	ids := []string{request.UserID}
	for _, v := range request.Member {
		ids = append(ids, v)
	}
	if err := s.grr.DeleteRelationByGroupIDAndUserIDs(request.GroupId, ids); err != nil {
		return resp, status.Error(codes.Aborted, err.Error())
	}
	return resp, nil
}

func (s *Service) SetGroupSilentNotification(ctx context.Context, in *v1.SetGroupSilentNotificationRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}
	if err := s.grr.SetUserGroupSilentNotification(in.GroupId, in.UserId, entity.SilentNotification(in.IsSilent)); err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrSetUserGroupSilentNotificationFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) RemoveGroupRelationByGroupIdAndUserIDs(ctx context.Context, in *v1.RemoveGroupRelationByGroupIdAndUserIDsRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		npo := persistence.NewRepositories(tx)

		//查询对话信息
		dialog, err := npo.Dr.GetDialogByGroupId(uint(in.GroupId))
		if err != nil {
			return err
		}

		//删除对话用户
		err = npo.Dr.DeleteDialogUserByDialogIDAndUserID(dialog.ID, in.UserIDs)
		if err != nil {
			return err
		}

		//删除关系
		if err := npo.Grr.DeleteRelationByGroupIDAndUserIDs(in.GroupId, in.UserIDs); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return resp, status.Error(codes.Code(code.GroupErrDeleteUserGroupRelationFailed.Code()), err.Error())
	}

	return resp, nil
}
