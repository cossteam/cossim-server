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
)

var _ v1.UserRelationServiceServer = &userServiceServer{}

type userServiceServer struct {
	repos *persistence.Repositories
}

func (s *userServiceServer) AddFriend(ctx context.Context, request *v1.AddFriendRequest) (*v1.AddFriendResponse, error) {
	resp := &v1.AddFriendResponse{}

	rel, err := s.repos.UserRepo.Get(ctx, request.FriendId, request.UserId)
	if err != nil && !errors.Is(err, code.NotFound) {
		return nil, code.WrapCodeToGRPC(code.RelationErrAddFriendFailed.Reason(utils.FormatErrorStack(err)))
	}

	if rel != nil {
		if _, err := s.repos.UserRepo.Create(ctx, &entity.UserRelation{
			UserID:   request.UserId,
			FriendID: request.FriendId,
			Status:   entity.UserRelationStatus(v1.RelationStatus_RELATION_NORMAL),
			DialogId: rel.DialogId,
		}); err != nil {
			return nil, code.WrapCodeToGRPC(code.RelationErrAddFriendFailed.Reason(utils.FormatErrorStack(err)))
		}
		return resp, nil
	}
	// 双方都没有好友关系
	if err := s.repos.TXRepositories(func(txr *persistence.Repositories) error {
		dialog, err := txr.DialogRepo.Create(ctx, &repository.CreateDialog{
			Type:    entity.DialogType(v1.DialogType_USER_DIALOG),
			OwnerId: request.UserId,
			GroupId: 0,
		})
		if err != nil {
			return err
		}

		if err := txr.UserRepo.EstablishFriendship(ctx, dialog.ID, request.UserId, request.FriendId); err != nil {
			return err
		}

		if _, err := txr.DialogUserRepo.Creates(ctx, dialog.ID, []string{request.UserId, request.FriendId}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, code.WrapCodeToGRPC(code.RelationErrAddFriendFailed.Reason(utils.FormatErrorStack(err)))
	}

	return resp, nil
}

func (s *userServiceServer) GetFriendList(ctx context.Context, request *v1.GetFriendListRequest) (*v1.GetFriendListResponse, error) {
	resp := &v1.GetFriendListResponse{}

	friends, err := s.repos.UserRepo.ListFriend(ctx, request.UserId)
	if err != nil {
		return nil, code.WrapCodeToGRPC(code.RelationErrGetFriendListFailed.Reason(utils.FormatErrorStack(err)))
	}

	for _, friend := range friends {
		resp.FriendList = append(resp.FriendList,
			&v1.Friend{
				UserId:                      friend.UserID,
				DialogId:                    friend.DialogID,
				Remark:                      friend.Remark,
				Status:                      v1.RelationStatus(friend.Status),
				IsSilent:                    friend.IsSilent,
				OpenBurnAfterReading:        friend.OpenBurnAfterReading,
				OpenBurnAfterReadingTimeOut: friend.OpenBurnAfterReadingTimeOut,
			})
	}

	return resp, nil
}

func (s *userServiceServer) GetUserRelation(ctx context.Context, request *v1.GetUserRelationRequest) (*v1.GetUserRelationResponse, error) {
	resp := &v1.GetUserRelationResponse{}
	var err error

	// 从数据库中获取关系对象
	entityRelation, err := s.repos.UserRepo.Get(ctx, request.UserId, request.FriendId)
	if err != nil {
		if errors.Is(err, code.NotFound) {
			return nil, code.WrapCodeToGRPC(code.RelationUserErrFriendRelationNotFound.Reason(utils.FormatErrorStack(err)))
		}
		return nil, err
	}

	var stat v1.RelationStatus
	switch entityRelation.Status {
	case entity.UserStatusNormal:
		stat = v1.RelationStatus_RELATION_NORMAL
	case entity.UserStatusBlocked:
		stat = v1.RelationStatus_RELATION_STATUS_BLOCKED
	case entity.UserStatusDeleted:
		stat = v1.RelationStatus_RELATION_STATUS_DELETED
	default:
		stat = v1.RelationStatus_RELATION_UNKNOWN
	}

	resp.Status = stat
	resp.DialogId = entityRelation.DialogId
	resp.UserId = entityRelation.UserID
	resp.FriendId = entityRelation.FriendID
	resp.IsSilent = entityRelation.SilentNotification
	resp.OpenBurnAfterReading = entityRelation.OpenBurnAfterReading
	resp.Remark = entityRelation.Remark
	resp.OpenBurnAfterReadingTimeOut = entityRelation.BurnAfterReadingTimeOut
	return resp, nil
}

func (s *userServiceServer) GetRelationsWithUsers(ctx context.Context, request *v1.GetUserRelationByUserIdsRequest) (*v1.GetUserRelationByUserIdsResponse, error) {
	resp := &v1.GetUserRelationByUserIdsResponse{}

	if request.UserId == "" || request.FriendIds == nil || len(request.FriendIds) == 0 {
		return resp, nil
	}

	relations, err := s.repos.UserRepo.Find(ctx, &repository.UserQuery{
		UserId:   request.UserId,
		FriendId: request.FriendIds,
	})
	if err != nil {
		return nil, code.WrapCodeToGRPC(code.RelationErrGetUserRelationFailed.Reason(utils.FormatErrorStack(err)))
	}

	for _, relation := range relations {
		resp.Users = append(resp.Users, &v1.GetUserRelationResponse{
			UserId:   relation.UserID,
			FriendId: relation.FriendID,
			Status:   v1.RelationStatus(relation.Status),
			DialogId: relation.DialogId,
		})
	}

	return resp, nil
}
