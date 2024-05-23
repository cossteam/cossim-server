package grpc

import (
	"context"
	"errors"
	v1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/internal/relation/infra/persistence"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	// RequestExpiredTime 群聊请求过期时间 7天的毫秒数
	RequestExpiredTime = 7 * 24 * 60 * 60 * 1000
)

var _ v1.GroupRelationServiceServer = &groupServiceServer{}

type groupServiceServer struct {
	repos *persistence.Repositories
}

func (s *groupServiceServer) GetGroupUserIDs(ctx context.Context, request *v1.GroupIDRequest) (*v1.UserIdsResponse, error) {
	resp := &v1.UserIdsResponse{}

	if request.GroupId == 0 {
		return nil, code.WrapCodeToGRPC(code.InvalidParameter.CustomMessage("group id is empty"))
	}

	userGroupIDs, err := s.repos.GroupRepo.GetGroupUserIDs(ctx, request.GroupId)
	if err != nil {
		return nil, code.WrapCodeToGRPC(code.GroupErrGetBatchGroupInfoByIDsFailed.Reason(utils.FormatErrorStack(err)))
	}

	resp.UserIds = userGroupIDs
	return resp, nil
}

func (s *groupServiceServer) GetGroupRelation(ctx context.Context, request *v1.GetGroupRelationRequest) (*v1.GetGroupRelationResponse, error) {
	resp := &v1.GetGroupRelationResponse{}

	rel, err := s.repos.GroupRepo.GetUserGroupByGroupIDAndUserID(ctx, request.GroupId, request.UserId)
	if err != nil {
		return nil, code.WrapCodeToGRPC(code.RelationGroupErrGroupRelationFailed.Reason(utils.FormatErrorStack(err)))
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

	return resp, nil
}

func (s *groupServiceServer) GetBatchGroupRelation(ctx context.Context, request *v1.GetBatchGroupRelationRequest) (*v1.GetBatchGroupRelationResponse, error) {
	resp := &v1.GetBatchGroupRelationResponse{}

	grs, err := s.repos.GroupRepo.GetUsersGroupByGroupIDAndUserIDs(ctx, request.GroupId, request.UserIds)
	if err != nil {
		if errors.Is(err, code.NotFound) {
			return nil, code.WrapCodeToGRPC(code.RelationGroupErrRelationNotFound.Reason(utils.FormatErrorStack(err)))
		}
		return nil, code.WrapCodeToGRPC(code.RelationGroupErrGroupRelationFailed.Reason(utils.FormatErrorStack(err)))
	}

	for _, gr := range grs {
		resp.GroupRelationResponses = append(resp.GroupRelationResponses, &v1.GetGroupRelationResponse{
			UserId:      gr.UserID,
			GroupId:     gr.GroupID,
			Identity:    v1.GroupIdentity(gr.Identity),
			MuteEndTime: gr.MuteEndTime,
		})
	}

	return resp, nil
}

func (s *groupServiceServer) DeleteGroupRelationByGroupId(ctx context.Context, request *v1.GroupIDRequest) (*emptypb.Empty, error) {
	if err := s.repos.GroupRepo.DeleteByGroupID(ctx, request.GroupId); err != nil {
		return nil, code.WrapCodeToGRPC(code.RelationGroupErrDeleteUsersGroupRelationFailed.Reason(utils.FormatErrorStack(err)))
	}

	return &emptypb.Empty{}, nil
}

func (s *groupServiceServer) DeleteGroupRelationByGroupIdRevert(ctx context.Context, request *v1.GroupIDRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}

	if err := s.repos.GroupRepo.UpdateFieldsByGroupID(ctx, request.GroupId, map[string]interface{}{
		"deleted_at": 0,
	}); err != nil {
		return nil, code.WrapCodeToGRPC(code.GroupErrDeleteGroupFailed.Reason(utils.FormatErrorStack(err)))
	}
	return resp, nil
}

func (s *groupServiceServer) GetGroupAdminIds(ctx context.Context, in *v1.GroupIDRequest) (*v1.UserIdsResponse, error) {
	var resp = &v1.UserIdsResponse{}
	ids, err := s.repos.GroupRepo.ListGroupAdmin(ctx, in.GroupId)
	if err != nil {
		return nil, code.WrapCodeToGRPC(code.MyCustomErrorCode.CustomMessage("获取群聊管理员失败").Reason(utils.FormatErrorStack(err)))
	}
	resp.UserIds = ids
	return resp, nil
}

func (s *groupServiceServer) GetUserManageGroupID(ctx context.Context, request *v1.GetUserManageGroupIDRequest) (*v1.GetUserManageGroupIDResponse, error) {
	resp := &v1.GetUserManageGroupIDResponse{}

	ids, err := s.repos.GroupRepo.GetUserManageGroupIDs(ctx, request.UserId)
	if err != nil {
		return nil, code.WrapCodeToGRPC(code.GroupErrGetBatchGroupInfoByIDsFailed.Reason(utils.FormatErrorStack(err)))
	}

	for _, id := range ids {
		resp.GroupIDs = append(resp.GroupIDs, &v1.GroupIDRequest{GroupId: id})
	}

	return resp, nil
}

func (s *groupServiceServer) CreateGroupAndInviteUsers(ctx context.Context, request *v1.CreateGroupAndInviteUsersRequest) (*v1.CreateGroupAndInviteUsersResponse, error) {
	resp := &v1.CreateGroupAndInviteUsersResponse{}

	if err := s.repos.TXRepositories(func(txr *persistence.Repositories) error {
		// 创建群聊对话
		dialog, err2 := txr.DialogRepo.Create(ctx, &repository.CreateDialog{
			Type:    entity.GroupDialog,
			OwnerId: request.UserID,
			GroupId: request.GroupId,
		})
		if err2 != nil {
			return err2
		}

		resp.DialogId = dialog.ID

		// 群主加入对话
		if _, err := txr.DialogUserRepo.Create(ctx, &repository.CreateDialogUser{
			DialogID: dialog.ID,
			UserID:   request.UserID,
		}); err != nil {
			return err
		}

		// 群主加入群聊
		if _, err := txr.GroupRepo.Create(ctx, &entity.CreateGroupRelation{
			GroupID:  request.GroupId,
			UserID:   request.UserID,
			Identity: entity.IdentityOwner,
		}); err != nil {
			return err
		}

		//发送邀请给其他成员
		requests := make([]*entity.GroupJoinRequest, 0)

		at := ptime.Now()
		for _, v := range request.Member {
			req := &entity.GroupJoinRequest{
				UserID:    v,
				GroupID:   request.GroupId,
				Status:    entity.Invitation,
				Inviter:   request.UserID,
				OwnerID:   v,
				InviterAt: at,
				ExpiredAt: at + RequestExpiredTime,
			}
			requests = append(requests, req)
		}

		for _, s2 := range request.Member {
			req := &entity.GroupJoinRequest{
				UserID:    s2,
				GroupID:   request.GroupId,
				Status:    entity.Invitation,
				Inviter:   request.UserID,
				OwnerID:   request.UserID,
				InviterAt: at,
				ExpiredAt: at + RequestExpiredTime,
			}
			requests = append(requests, req)
		}

		if len(requests) > 0 {
			if _, err := txr.GroupJoinRequestRepo.Creates(ctx, requests); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return nil, code.WrapCodeToGRPC(code.MyCustomErrorCode.CustomMessage("创建群聊失败").Reason(utils.FormatErrorStack(err)))
	}

	return resp, nil
}

func (s *groupServiceServer) CreateGroupAndInviteUsersRevert(ctx context.Context, request *v1.CreateGroupAndInviteUsersRequest) (*emptypb.Empty, error) {
	resp := &emptypb.Empty{}
	ids := []string{request.UserID}
	for _, v := range request.Member {
		ids = append(ids, v)
	}

	if err := s.repos.GroupRepo.DeleteByGroupIDAndUserID(ctx, request.GroupId, ids...); err != nil {
		return nil, code.WrapCodeToGRPC(code.GroupErrDeleteGroupFailed.CustomMessage("删除群聊回滚失败").Reason(utils.FormatErrorStack(err)))
	}

	return resp, nil
}
