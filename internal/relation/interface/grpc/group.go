package grpc

import (
	"context"
	"errors"
	v1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/relation"
	"github.com/cossim/coss-server/internal/relation/infra/persistence"
	"github.com/cossim/coss-server/pkg/cache"
	"github.com/cossim/coss-server/pkg/code"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"log"
)

var _ v1.GroupRelationServiceServer = &groupServiceServer{}

type groupServiceServer struct {
	db          *gorm.DB
	cache       cache.RelationGroupCache
	cacheEnable bool
	grr         relation.GroupRepository
	dr          relation.DialogRepository
}

func (s *groupServiceServer) AddGroupAdmin(ctx context.Context, request *v1.AddGroupAdminRequest) (*v1.AddGroupAdminResponse, error) {
	resp := &v1.AddGroupAdminResponse{}

	r1, err := s.grr.GetUserGroupByGroupIDAndUserID(ctx, request.GroupID, request.UserID)
	if err != nil {
		return nil, err
	}

	if r1.Identity != relation.IdentityAdmin {
		return nil, code.Forbidden
	}

	for _, v := range request.UserIDs {
		if err := s.grr.UpdateIdentity(ctx, r1.GroupID, v, relation.IdentityAdmin); err != nil {
			return nil, code.RelationErrGroupAddAdmin.Reason(err)
		}
	}

	return resp, nil
}

func (s *groupServiceServer) GetGroupUserIDs(ctx context.Context, request *v1.GroupIDRequest) (*v1.UserIdsResponse, error) {
	resp := &v1.UserIdsResponse{}

	userGroupIDs, err := s.grr.GetGroupUserIDs(ctx, request.GroupId)
	if err != nil {
		return resp, status.Error(codes.Code(code.GroupErrGetBatchGroupInfoByIDsFailed.Code()), err.Error())
	}

	resp.UserIds = userGroupIDs
	return resp, nil
}

func (s *groupServiceServer) GetUserGroupIDs(ctx context.Context, request *v1.GetUserGroupIDsRequest) (*v1.GetUserGroupIDsResponse, error) {
	resp := &v1.GetUserGroupIDsResponse{}

	//if s.cacheEnable {
	//	ds, err := s.cache.GetUserGroupIDs(ctx, request.UserId)
	//	if err == nil && ds != nil {
	//		for _, v := range ds {
	//			resp.GroupId = append(resp.GroupId, v)
	//		}
	//		return resp, nil
	//	}
	//}

	ds, err := s.grr.GetUserJoinedGroupIDs(ctx, request.UserId)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrGetGroupIDsFailed.Code()), err.Error())
	}

	for _, v := range ds {
		resp.GroupId = append(resp.GroupId, v)
	}

	//if s.cacheEnable {
	//	if err := s.cache.SetUserGroupIDs(ctx, request.UserId, ds); err != nil {
	//		log.Printf("set user group id cache failed, err: %v", err)
	//	}
	//}

	return resp, nil
}

func (s *groupServiceServer) GetUserGroupList(ctx context.Context, request *v1.GetUserGroupListRequest) (*v1.GetUserGroupListResponse, error) {
	resp := &v1.GetUserGroupListResponse{}

	isAdmin := false

	// 获取用户所属群组的ID列表
	groupIDs, err := s.grr.GetUserGroupIDs(ctx, request.UserId)
	if err != nil {
		return resp, status.Error(codes.Code(code.GroupErrGetBatchGroupInfoByIDsFailed.Code()), err.Error())
	}

	for _, gid := range groupIDs {
		// 获取群组的申请记录
		reqList, err := s.grr.ListJoinRequest(ctx, []uint32{gid})
		if err != nil {
			return resp, status.Error(codes.Code(code.RelationUserErrGetRequestListFailed.Code()), err.Error())
		}

		//reqList, err := s.grr.GetJoinRequestBatchListByID([]uint32{gid})
		//if err != nil {
		//	return resp, status.Error(codes.Code(code.RelationUserErrGetRequestListFailed.Code()), err.Error())
		//}

		for _, v := range reqList {
			// 判断用户是否为群组管理员
			gr, err := s.grr.GetUserGroupByGroupIDAndUserID(ctx, gid, v.UserID)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					continue
				}
				return resp, status.Error(codes.Code(code.RelationUserErrGetRequestListFailed.Code()), err.Error())
			}

			if gr.Identity != relation.IdentityUser {
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

func (s *groupServiceServer) RemoveFromGroup(ctx context.Context, request *v1.RemoveFromGroupRequest) (*v1.RemoveFromGroupResponse, error) {
	resp := &v1.RemoveFromGroupResponse{}

	gr, err := s.grr.GetUserGroupByGroupIDAndUserID(ctx, request.GroupId, request.UserId)
	if err != nil {
		//if errors.Is(err, gorm.ErrRecordNotFound) {
		//	return resp, status.Error(codes.Code(code.RelationUserErrGetRequestListFailed.Code()), err.Error())
		//}
		return resp, status.Error(codes.Code(code.RelationUserErrGetRequestListFailed.Code()), err.Error())
	}

	if err := s.grr.Delete(ctx, gr.ID); err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrRemoveUserFromGroupFailed.Code()), err.Error())
	}

	//if s.cacheEnable {
	//	if err := s.cache.DeleteRelation(ctx, request.UserId, request.GroupId); err != nil {
	//		log.Printf("delete relation cache failed, err: %v", err)
	//	}
	//	if err := s.cache.DeleteUserGroupIDs(ctx, request.UserId); err != nil {
	//		log.Printf("delete user group id cache failed, err: %v", err)
	//	}
	//}

	return resp, nil
}

func (s *groupServiceServer) LeaveGroup(ctx context.Context, request *v1.LeaveGroupRequest) (*v1.LeaveGroupResponse, error) {
	resp := &v1.LeaveGroupResponse{}

	gr, err := s.grr.GetUserGroupByGroupIDAndUserID(ctx, request.GroupId, request.UserId)
	if err != nil {
		//if errors.Is(err, gorm.ErrRecordNotFound) {
		//	return resp, status.Error(codes.Code(code.RelationUserErrGetRequestListFailed.Code()), err.Error())
		//}
		return resp, status.Error(codes.Code(code.RelationUserErrGetRequestListFailed.Code()), err.Error())
	}

	if err := s.grr.Delete(ctx, gr.ID); err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrLeaveGroupFailed.Code()), err.Error())
	}

	//if s.cacheEnable {
	//	if err := s.cache.DeleteRelation(ctx, request.UserId, request.GroupId); err != nil {
	//		log.Printf("delete relation cache failed, err: %v", err)
	//	}
	//	if err := s.cache.DeleteUserGroupIDs(ctx, request.UserId); err != nil {
	//		log.Printf("delete user group id cache failed, err: %v", err)
	//	}
	//}
	return resp, nil
}

func (s *groupServiceServer) LeaveGroupRevert(ctx context.Context, request *v1.LeaveGroupRequest) (*v1.LeaveGroupResponse, error) {
	resp := &v1.LeaveGroupResponse{}

	gr, err := s.grr.GetUserGroupByGroupIDAndUserID(ctx, request.GroupId, request.UserId)
	if err != nil {
		//if errors.Is(err, gorm.ErrRecordNotFound) {
		//	return resp, status.Error(codes.Code(code.RelationUserErrGetRequestListFailed.Code()), err.Error())
		//}
		return resp, status.Error(codes.Code(code.RelationUserErrGetRequestListFailed.Code()), err.Error())
	}

	if err := s.grr.UpdateFields(ctx, gr.ID, map[string]interface{}{
		"deleted_at": 0,
	}); err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrLeaveGroupFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s *groupServiceServer) GetGroupRelation(ctx context.Context, request *v1.GetGroupRelationRequest) (*v1.GetGroupRelationResponse, error) {
	resp := &v1.GetGroupRelationResponse{}

	//if s.cacheEnable {
	//	if rel, err := s.cache.GetRelation(ctx, request.UserId, request.GroupId); err == nil && relation != nil {
	//		return rel, nil
	//	}
	//}

	rel, err := s.grr.GetUserGroupByGroupIDAndUserID(ctx, request.GroupId, request.UserId)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrGroupRelationFailed.Code()), err.Error())
	}

	var silentNotification uint
	if rel.SilentNotification {
		silentNotification = 1
	} else {
		silentNotification = 0
	}

	resp.GroupId = rel.GroupID
	resp.UserId = rel.UserID
	resp.Identity = v1.GroupIdentity(rel.Identity)
	resp.MuteEndTime = rel.MuteEndTime
	resp.IsSilent = v1.GroupSilentNotificationType(silentNotification)
	resp.OpenBurnAfterReading = v1.OpenBurnAfterReadingType(silentNotification)
	resp.JoinTime = rel.JoinedAt
	resp.Remark = rel.Remark
	resp.JoinMethod = uint32(rel.EntryMethod)
	resp.Inviter = rel.Inviter

	//if s.cacheEnable {
	//	if err := s.cache.SetRelation(ctx, request.UserId, request.GroupId, resp, cache.RelationExpireTime); err != nil {
	//		log.Printf("set relation cache failed, err: %v", err)
	//	}
	//}

	return resp, nil
}

func (s *groupServiceServer) GetBatchGroupRelation(ctx context.Context, request *v1.GetBatchGroupRelationRequest) (*v1.GetBatchGroupRelationResponse, error) {
	resp := &v1.GetBatchGroupRelationResponse{}

	//if s.cacheEnable {
	//	if gr, err := s.cache.GetUsersGroupRelation(ctx, request.UserIds, request.GroupId); err == nil && gr != nil {
	//		return gr, nil
	//	}
	//}

	grs, err := s.grr.GetUsersGroupByGroupIDAndUserIDs(ctx, request.GroupId, request.UserIds)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationGroupErrRelationNotFound.Code()), code.RelationGroupErrRelationNotFound.Message())
		}
		return resp, status.Error(codes.Code(code.RelationGroupErrGroupRelationFailed.Code()), err.Error())
	}

	for _, gr := range grs {
		resp.GroupRelationResponses = append(resp.GroupRelationResponses, &v1.GetGroupRelationResponse{
			UserId:      gr.UserID,
			GroupId:     gr.GroupID,
			Identity:    v1.GroupIdentity(gr.Identity),
			MuteEndTime: gr.MuteEndTime,
		})
	}

	//if s.cacheEnable {
	//	for _, v := range resp.GroupRelationResponses {
	//		if err := s.cache.SetRelation(ctx, v.UserId, request.GroupId, v, cache.RelationExpireTime); err != nil {
	//			log.Printf("set relationcache failed, err: %v", err)
	//		}
	//	}
	//}

	return resp, nil
}

func (s *groupServiceServer) DeleteGroupRelationByGroupId(ctx context.Context, request *v1.GroupIDRequest) (*emptypb.Empty, error) {
	if err := s.grr.DeleteByGroupID(ctx, request.GroupId); err != nil {
		return nil, code.RelationGroupErrDeleteUsersGroupRelationFailed.Reason(err)
	}

	//_, err := s.grr.GetGroupUserIDs(ctx, request.GroupId)
	//if err != nil {
	//	return nil, err
	//}

	//if s.cacheEnable {
	//	for _, uid := range uids {
	//		if err := s.cache.DeleteRelation(ctx, uid, request.GroupId); err != nil {
	//			log.Printf("delete relation cache failed, err: %v", err)
	//		}
	//	}
	//	if err := s.cache.DeleteUserGroupIDs(ctx, uids...); err != nil {
	//		log.Printf("delete user group id cache failed, err: %v", err)
	//	}
	//
	//}

	return &emptypb.Empty{}, nil
}

func (s *groupServiceServer) DeleteGroupRelationByGroupIdRevert(ctx context.Context, request *v1.GroupIDRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}

	if err := s.grr.UpdateFieldsByGroupID(ctx, request.GroupId, map[string]interface{}{
		"deleted_at": 0,
	}); err != nil {
		return resp, status.Error(codes.Code(code.GroupErrDeleteGroupFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *groupServiceServer) GetGroupAdminIds(ctx context.Context, in *v1.GroupIDRequest) (*v1.UserIdsResponse, error) {
	var resp = &v1.UserIdsResponse{}
	ids, err := s.grr.ListGroupAdmin(ctx, in.GroupId)
	if err != nil {
		return resp, err
	}
	resp.UserIds = ids
	return resp, nil
}

func (s *groupServiceServer) GetUserManageGroupID(ctx context.Context, request *v1.GetUserManageGroupIDRequest) (*v1.GetUserManageGroupIDResponse, error) {
	resp := &v1.GetUserManageGroupIDResponse{}

	ids, err := s.grr.GetUserManageGroupIDs(ctx, request.UserId)
	if err != nil {
		return resp, status.Error(codes.Code(code.GroupErrGetBatchGroupInfoByIDsFailed.Code()), err.Error())
	}

	for _, id := range ids {
		resp.GroupIDs = append(resp.GroupIDs, &v1.GroupIDRequest{GroupId: id})
	}

	return resp, nil
}

func (s *groupServiceServer) DeleteGroupRelationByGroupIdAndUserID(ctx context.Context, request *v1.DeleteGroupRelationByGroupIdAndUserIDRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}

	if err := s.grr.DeleteByGroupIDAndUserID(ctx, request.GroupID, request.UserID); err != nil {
		return resp, status.Error(codes.Code(code.GroupErrDeleteUserGroupRelationFailed.Code()), err.Error())
	}

	//if err := s.grr.DeleteUserGroupRelationByGroupIDAndUserID(request.GroupID, request.UserID); err != nil {
	//	//return resp, status.Error(codes.Code(code.GroupErrDeleteUserGroupRelationFailed.Code()), err.Error())
	//	return resp, status.Error(codes.Aborted, fmt.Sprintf(code.GroupErrDeleteUserGroupRelationFailed.Message()+" : "+err.Error()))
	//}

	//if s.cacheEnable {
	//	if err := s.cache.DeleteRelation(ctx, request.UserID, request.GroupID); err != nil {
	//		log.Printf("delete relation cache failed, err: %v", err)
	//	}
	//}

	return resp, nil
}

func (s *groupServiceServer) DeleteGroupRelationByGroupIdAndUserIDRevert(ctx context.Context, request *v1.DeleteGroupRelationByGroupIdAndUserIDRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}

	gr, err := s.grr.GetUserGroupByGroupIDAndUserID(ctx, request.GroupID, request.UserID)
	if err != nil {
		//if errors.Is(err, gorm.ErrRecordNotFound) {
		//	return resp, status.Error(codes.Code(code.RelationUserErrGetRequestListFailed.Code()), err.Error())
		//}
		return resp, status.Error(codes.Code(code.RelationUserErrGetRequestListFailed.Code()), err.Error())
	}

	if err := s.grr.UpdateFields(ctx, gr.ID, map[string]interface{}{
		"deleted_at": 0,
	}); err != nil {
		return resp, status.Error(codes.Code(code.GroupErrDeleteUserGroupRelationFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s *groupServiceServer) CreateGroupAndInviteUsers(ctx context.Context, request *v1.CreateGroupAndInviteUsersRequest) (*v1.CreateGroupAndInviteUsersResponse, error) {
	resp := &v1.CreateGroupAndInviteUsersResponse{}

	//return resp, status.Error(codes.Aborted, formatErrorMessage(errors.New("测试回滚")))

	if err := s.db.Transaction(func(tx *gorm.DB) error {
		npo := persistence.NewRepositories(tx)
		// 创建群聊对话
		//dialog, err := npo.Dr.CreateDialog(request.UserID, entity.GroupDialog, uint(request.GroupId))
		//if err != nil {
		//	return err
		//}

		dialog, err2 := npo.Dr.Create(ctx, &relation.CreateDialog{
			Type:    entity.GroupDialog,
			OwnerId: request.UserID,
			GroupId: request.GroupId,
		})
		if err2 != nil {
			return err2
		}

		resp.DialogId = dialog.ID

		// 群主加入对话
		//_, err = npo.Dr.JoinDialog(dialog.ID, request.UserID)
		//if err != nil {
		//	return err
		//}

		if _, err := npo.Dur.Create(ctx, &relation.CreateDialogUser{
			DialogID: dialog.ID,
			UserID:   request.UserID,
		}); err != nil {
			return err
		}

		// 群主加入群聊
		//owner := &entity.GroupRelation{
		//	UserID:   request.UserID,
		//	GroupID:  uint(request.GroupId),
		//	Identity: entity.IdentityOwner,
		//}
		//_, err := npo.Grr.CreateRelation(owner)
		//if err != nil {
		//	return err
		//}

		// 群主加入群聊
		if _, err := npo.Grr.Create(ctx, &relation.CreateGroupRelation{
			GroupID:  request.GroupId,
			UserID:   request.UserID,
			Identity: relation.IdentityOwner,
		}); err != nil {
			return err
		}

		//发送邀请给其他成员
		requests := make([]*relation.GroupJoinRequest, 0)

		//gjr1 := &entity.GroupJoinRequest{
		//	UserID:      request.UserID,
		//	GroupID:     uint(request.GroupId),
		//	OwnerID:     request.UserID,
		//	InviterTime: time.Now(),
		//	Status:      entity.Invitation,
		//}
		//requests = append(requests, gjr1)

		for _, v := range request.Member {
			req := &relation.GroupJoinRequest{
				UserID:      v,
				GroupID:     request.GroupId,
				Status:      relation.Invitation,
				Inviter:     request.UserID,
				OwnerID:     v,
				InviterTime: ptime.Now(),
			}
			requests = append(requests, req)
		}

		for _, s2 := range request.Member {
			req := &relation.GroupJoinRequest{
				UserID:      s2,
				GroupID:     request.GroupId,
				Status:      relation.Invitation,
				Inviter:     request.UserID,
				OwnerID:     request.UserID,
				InviterTime: ptime.Now(),
			}
			requests = append(requests, req)
		}

		//if _, err := npo.Gjqr.AddJoinRequestBatch(requests); err != nil {
		//	return err
		//}

		if _, err := npo.Gjqr.Creates(ctx, requests); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return resp, status.Error(codes.Aborted, err.Error())
	}

	if s.cacheEnable {
		ids := append([]string{request.UserID}, request.Member...)
		if err := s.cache.DeleteGroupJoinRequestListByUser(ctx, ids...); err != nil {
			log.Printf("delete relation cache failed, err: %v", err)
			//return nil, err
		}
		//if err := s.cache.DeleteRelationByGroupID(ctx, request.GroupId); err != nil {
		//	log.Printf("delete relation cache failed, err: %v", err)
		//}
	}

	return resp, nil
}

func (s *groupServiceServer) CreateGroupAndInviteUsersRevert(ctx context.Context, request *v1.CreateGroupAndInviteUsersRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}
	ids := []string{request.UserID}
	for _, v := range request.Member {
		ids = append(ids, v)
	}

	if err := s.grr.DeleteByGroupIDAndUserID(ctx, request.GroupId, ids...); err != nil {
		return resp, status.Error(codes.Aborted, err.Error())
	}

	return resp, nil
}

func (s *groupServiceServer) SetGroupSilentNotification(ctx context.Context, request *v1.SetGroupSilentNotificationRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}

	if err := s.grr.UserGroupSilentNotification(ctx, request.GroupId, request.UserId, request.IsSilent); err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrSetUserGroupSilentNotificationFailed.Code()), err.Error())
	}

	//if s.cacheEnable {
	//	if err := s.cache.DeleteRelation(ctx, request.UserId, request.GroupId); err != nil {
	//		log.Printf("delete relation cache failed, err: %v", err)
	//	}
	//}

	return resp, nil
}

func (s *groupServiceServer) RemoveGroupRelationByGroupIdAndUserIDs(ctx context.Context, request *v1.RemoveGroupRelationByGroupIdAndUserIDsRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		npo := persistence.NewRepositories(tx)

		// 查询对话信息
		//dialog, err := npo.Dr.GetDialogByGroupId(uint(request.GroupId))
		//if err != nil {
		//	return err
		//}

		dialog, err := npo.Dr.GetByGroupID(ctx, request.GroupId)
		if err != nil {
			return err
		}

		if err := npo.Dur.DeleteByDialogIDAndUserID(ctx, dialog.ID, request.UserIDs...); err != nil {
			return err
		}

		// 删除对话用户
		//err = npo.Dr.DeleteDialogUserByDialogIDAndUserID(dialog.ID, request.UserIDs)
		//if err != nil {
		//	return err
		//}

		// 删除关系
		//if err := npo.Grr.DeleteRelationByGroupIDAndUserIDs(request.GroupId, request.UserIDs); err != nil {
		//	return err
		//}

		if err := npo.Grr.DeleteByGroupIDAndUserID(ctx, request.GroupId, request.UserIDs...); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return resp, status.Error(codes.Code(code.GroupErrDeleteUserGroupRelationFailed.Code()), err.Error())
	}

	//if s.cacheEnable {
	//if err := s.cache.DeleteRelationByGroupID(ctx, request.GroupId); err != nil {
	//	log.Printf("delete relation cache failed, err: %v", err)
	//}
	//}

	return resp, nil
}

func (s *groupServiceServer) SetGroupOpenBurnAfterReading(ctx context.Context, request *v1.SetGroupOpenBurnAfterReadingRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}
	//if err := s.grr.SetUserGroupOpenBurnAfterReading(request.GroupId, request.UserId, entity.OpenBurnAfterReadingType(request.OpenBurnAfterReading)); err != nil {
	//	return resp, status.Error(codes.Code(code.RelationGroupRrrSetUserGroupOpenBurnAfterReadingFailed.Code()), err.Error())
	//}
	return resp, nil
}

func (s *groupServiceServer) SetGroupOpenBurnAfterReadingTimeOut(ctx context.Context, request *v1.SetGroupOpenBurnAfterReadingTimeOutRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}
	//if err := s.grr.SetUserGroupOpenBurnAfterReadingTimeOUt(request.GroupId, request.UserId, request.OpenBurnAfterReadingTimeOut); err != nil {
	//	return resp, status.Error(codes.Code(code.RelationGroupErrSetUserGroupOpenBurnAfterReadingTimeOutFailed.Code()), err.Error())
	//}
	return resp, nil
}

func (s *groupServiceServer) SetGroupUserRemark(ctx context.Context, request *v1.SetGroupUserRemarkRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}
	if err := s.grr.SetUserGroupRemark(ctx, request.GroupId, request.UserId, request.Remark); err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrSetUserGroupRemarkFailed.Code()), err.Error())
	}

	if s.cacheEnable {
		if err := s.cache.DeleteRelation(ctx, request.UserId, request.GroupId); err != nil {
			log.Printf("delete relation cache failed, err: %v", err)
		}
	}

	return resp, nil
}
