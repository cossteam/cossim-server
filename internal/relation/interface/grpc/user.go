package grpc

import (
	"context"
	"errors"
	"fmt"
	v1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/internal/relation/infra/persistence"
	"github.com/cossim/coss-server/pkg/code"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
)

var _ v1.UserRelationServiceServer = &userServiceServer{}

type userServiceServer struct {
	repos *persistence.Repositories
}

func (s *userServiceServer) AddFriendAfterDelete(ctx context.Context, request *v1.AddFriendAfterDeleteRequest) (*v1.AddFriendAfterDeleteResponse, error) {
	if err := s.repos.TXRepositories(func(txr *persistence.Repositories) error {
		r1, err := txr.UserRepo.Get(ctx, request.FriendId, request.UserId)
		if err != nil {
			return err
		}
		if r1.Status != entity.UserStatusNormal {
			return code.RelationUserErrFriendRelationNotFound
		}

		if err := txr.UserRepo.RestoreFriendship(ctx, r1.DialogId, request.UserId, request.FriendId); err != nil {
			return err
		}
		var dat int64 = 0
		if err := txr.DialogUserRepo.UpdateDialogStatus(ctx, &repository.UpdateDialogStatusParam{
			DialogID:  r1.DialogId,
			UserID:    []string{request.UserId},
			DeletedAt: &dat,
		}); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, status.Error(codes.Code(code.RelationErrAddFriendFailed.Code()), formatErrorMessage(err))
	}

	return &v1.AddFriendAfterDeleteResponse{}, nil
}

func (s *userServiceServer) ManageFriend(ctx context.Context, request *v1.ManageFriendRequest) (*v1.ManageFriendResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *userServiceServer) ManageFriendRevert(ctx context.Context, request *v1.ManageFriendRequest) (*v1.ManageFriendResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (s *userServiceServer) AddFriend(ctx context.Context, request *v1.AddFriendRequest) (*v1.AddFriendResponse, error) {
	resp := &v1.AddFriendResponse{}

	rel, err := s.repos.UserRepo.Get(ctx, request.FriendId, request.UserId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Println("err => ", err)
		return resp, status.Error(codes.Code(code.RelationErrAddFriendFailed.Code()), formatErrorMessage(err))
	}

	if rel != nil {
		if _, err := s.repos.UserRepo.Create(ctx, &entity.UserRelation{
			UserID:   request.UserId,
			FriendID: request.FriendId,
			Status:   entity.UserRelationStatus(v1.RelationStatus_RELATION_NORMAL),
			DialogId: rel.DialogId,
		}); err != nil {
			return resp, status.Error(codes.Code(code.RelationErrAddFriendFailed.Code()), formatErrorMessage(err))
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
		return resp, status.Error(codes.Code(code.RelationErrAddFriendFailed.Code()), formatErrorMessage(err))
	}

	//if s.cacheEnable {
	//	if err := s.cache.DeleteFriendRequestList(ctx, request.UserID, request.FriendId); err != nil {
	//		log.Printf("failed to delete cache friend request list: %v", err)
	//	}
	//}

	return resp, nil
}

func (s *userServiceServer) DeleteFriend(ctx context.Context, request *v1.DeleteFriendRequest) (*v1.DeleteFriendResponse, error) {
	resp := &v1.DeleteFriendResponse{}

	if err := s.repos.UserRepo.Delete(ctx, request.UserId, request.FriendId); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrDeleteFriendFailed.Code()), fmt.Sprintf("failed to delete relation: %v", err))
	}

	//if s.cacheEnable {
	//	if err := s.cache.DeleteRelation(ctx, request.UserID, request.FriendId); err != nil {
	//		log.Printf("failed to delete cache relation: %v", err)
	//	}
	//	if err := s.cache.DeleteFriendList(ctx, request.UserID); err != nil {
	//		log.Printf("failed to delete cache friend list: %v", err)
	//	}
	//}

	return resp, nil
}

func (s *userServiceServer) DeleteFriendRevert(ctx context.Context, request *v1.DeleteFriendRequest) (*v1.DeleteFriendResponse, error) {
	resp := &v1.DeleteFriendResponse{}

	rel, err := s.repos.UserRepo.Get(ctx, request.FriendId, request.UserId)
	if err != nil {
		return nil, status.Error(codes.Code(code.RelationErrGetUserRelationFailed.Code()), formatErrorMessage(err))
	}

	if err := s.repos.UserRepo.UpdateStatus(ctx, rel.ID, entity.UserStatusDeleted); err != nil {
		return nil, status.Error(codes.Code(code.RelationErrDeleteFriendFailed.Code()), fmt.Sprintf("DeleteFriendRevert failed to update relation: %v", err))
	}

	//if err := s.repos.UserRepo.UpdateFields(ctx, rel.ID, map[string]interfaces{}{
	//	"deleted_at": 0,
	//}); err != nil {
	//	return resp, status.Error(codes.Code(code.RelationErrDeleteFriendFailed.Code()), fmt.Sprintf("DeleteFriendRevert failed to update relation: %v", err))
	//}
	return resp, nil
}

func (s *userServiceServer) AddBlacklist(ctx context.Context, request *v1.AddBlacklistRequest) (*v1.AddBlacklistResponse, error) {
	resp := &v1.AddBlacklistResponse{}

	if request.UserId == request.FriendId {
		return resp, status.Error(codes.Code(code.RelationErrAddBlacklistFailed.Code()), "user cannot add themselves to the blacklist")
	}

	rel, err := s.repos.UserRepo.Get(ctx, request.UserId, request.FriendId)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrGetUserRelationFailed.Code()), formatErrorMessage(err))
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationUserErrFriendRelationNotFound.Code()), fmt.Sprintf("failed to retrieve relation: %v", err))
		}
		return resp, status.Error(codes.Code(code.RelationErrAddBlacklistFailed.Code()), fmt.Sprintf("failed to retrieve relation: %v", err))
	}

	if rel.Status == entity.UserStatusBlocked {
		return resp, code.RelationErrAlreadyBlacklist
	}

	if rel.Status != entity.UserStatusNormal {
		return resp, code.RelationUserErrFriendRelationNotFound
	}

	//rel.Status = relation.UserStatusBlocked
	if err = s.repos.UserRepo.UpdateStatus(ctx, rel.ID, entity.UserStatusBlocked); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrAddBlacklistFailed.Code()), fmt.Sprintf("failed to update relation: %v", err))
	}

	//if s.cacheEnable {
	//	if err := s.cache.DeleteBlacklist(ctx, request.UserID); err != nil {
	//		log.Printf("failed to delete cache blacklist: %v", err)
	//	}
	//	if err := s.cache.DeleteRelation(ctx, request.UserID, request.FriendId); err != nil {
	//		log.Printf("failed to delete cache relation: %v", err)
	//	}
	//	if err := s.cache.DeleteFriendList(ctx, request.UserID); err != nil {
	//		log.Printf("failed to delete cache friend list: %v", err)
	//	}
	//}

	return resp, nil
}

func (s *userServiceServer) DeleteBlacklist(ctx context.Context, request *v1.DeleteBlacklistRequest) (*v1.DeleteBlacklistResponse, error) {
	resp := &v1.DeleteBlacklistResponse{}

	if request.UserId == request.FriendId {
		return resp, status.Error(codes.Code(code.RelationErrDeleteBlacklistFailed.Code()), "user cannot delete themselves from the blacklist")
	}

	// Assuming urr is a UserRelationRepository instance in UserService
	rel, err := s.repos.UserRepo.Get(ctx, request.UserId, request.FriendId)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrDeleteBlacklistFailed.Code()), fmt.Sprintf("failed to retrieve relation: %v", err))
	}

	if err = s.repos.UserRepo.UpdateStatus(ctx, rel.ID, entity.UserStatusNormal); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrDeleteBlacklistFailed.Code()), fmt.Sprintf("failed to update relation: %v", err))
	}

	//rel.Status = relation.UserStatusNormal
	//if _, err = s.urr.UpdateRelation(relation1); err != nil {
	//	return resp, status.Error(codes.Code(code.RelationErrDeleteBlacklistFailed.Code()), fmt.Sprintf("failed to update relation: %v", err))
	//}

	//if s.cacheEnable {
	//	if err := s.cache.DeleteBlacklist(ctx, request.UserID); err != nil {
	//		log.Printf("failed to delete cache blacklist: %v", err)
	//	}
	//	if err := s.cache.DeleteRelation(ctx, request.UserID, request.FriendId); err != nil {
	//		log.Printf("failed to delete cache relation: %v", err)
	//	}
	//	if err := s.cache.DeleteFriendList(ctx, request.UserID); err != nil {
	//		log.Printf("failed to delete cache friend list: %v", err)
	//	}
	//}

	return resp, nil
}

func (s *userServiceServer) GetFriendList(ctx context.Context, request *v1.GetFriendListRequest) (*v1.GetFriendListResponse, error) {
	resp := &v1.GetFriendListResponse{}

	//if s.cacheEnable {
	//	// 从缓存中获取关系对象
	//	r, err := s.cache.GetFriendList(ctx, request.UserID)
	//	if err == nil && r != nil {
	//		// 如果缓存中存在，则直接返回缓存的结果
	//		return r, nil
	//	}
	//}

	//friends, err := s.urr.Find(ctx, &relation.UserQuery{
	//	UserID:   request.UserID,
	//	FriendId: nil,
	//	Status:   &st,
	//})
	friends, err := s.repos.UserRepo.FriendRequestList(ctx, request.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationErrUserNotFound.Code()), fmt.Sprintf("failed to get friend list: %v", err))
		}
		return resp, status.Error(codes.Code(code.RelationErrGetFriendListFailed.Code()), fmt.Sprintf("failed to get friend list: %v", err))
	}

	for _, friend := range friends {
		resp.FriendList = append(resp.FriendList,
			&v1.Friend{
				UserId:                      friend.UserId,
				DialogId:                    friend.DialogId,
				Remark:                      friend.Remark,
				Status:                      v1.RelationStatus(friend.Status),
				IsSilent:                    friend.IsSilent,
				OpenBurnAfterReading:        friend.OpenBurnAfterReading,
				OpenBurnAfterReadingTimeOut: friend.OpenBurnAfterReadingTimeOut,
			})
	}

	//if s.cacheEnable {
	//	if err := s.cache.SetFriendList(ctx, request.UserID, resp, cache.RelationExpireTime); err != nil {
	//		log.Printf("failed to set get friend list cache: %v", err)
	//	}
	//}

	return resp, nil
}

func (s *userServiceServer) GetBlacklist(ctx context.Context, request *v1.GetBlacklistRequest) (*v1.GetBlacklistResponse, error) {
	resp := &v1.GetBlacklistResponse{}
	//
	//if s.cacheEnable {
	//	// 尝试从缓存中获取黑名单列表
	//	cachedList, err := s.cache.GetBlacklist(ctx, request.GetUserId())
	//	if err == nil && cachedList != nil {
	//		// 如果缓存中存在，则直接返回缓存的结果
	//		return cachedList, nil
	//	}
	//}

	blacklist, err := s.repos.UserRepo.Blacklist(ctx, request.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationErrUserNotFound.Code()), fmt.Sprintf("failed to get black list: %v", err))
		}
		return resp, status.Error(codes.Code(code.RelationErrGetBlacklistFailed.Code()), fmt.Sprintf("failed to get black list: %v", err))
	}

	for _, id := range blacklist.List {
		resp.Blacklist = append(resp.Blacklist, &v1.Blacklist{UserId: id})
	}

	//if s.cacheEnable {
	//	if err := s.cache.SetBlacklist(ctx, request.UserID, resp, cache.RelationExpireTime); err != nil {
	//		log.Printf("failed to set blacklist cache: %v", err)
	//	}
	//}

	return resp, nil
}

func (s *userServiceServer) GetUserRelation(ctx context.Context, request *v1.GetUserRelationRequest) (*v1.GetUserRelationResponse, error) {
	resp := &v1.GetUserRelationResponse{}
	var err error

	//if s.cacheEnable {
	//	// 从缓存中获取关系对象
	//	relation, err := s.cache.GetRelation(ctx, request.UserID, request.FriendId)
	//	if err == nil && relation != nil {
	//		// 如果缓存中存在，则直接返回缓存的结果
	//		return relation, nil
	//	}
	//}

	// 从数据库中获取关系对象
	entityRelation, err := s.repos.UserRepo.Get(ctx, request.UserId, request.FriendId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationUserErrFriendRelationNotFound.Code()), err.Error())
		}
		return resp, err
	}

	// 将entity.UserRelation的字段值复制到v1.GetUserRelationResponse
	resp.Status = v1.RelationStatus(entityRelation.Status)
	resp.DialogId = entityRelation.DialogId
	resp.UserId = entityRelation.UserID
	resp.FriendId = entityRelation.FriendID
	resp.IsSilent = entityRelation.SilentNotification
	resp.OpenBurnAfterReading = entityRelation.OpenBurnAfterReading
	resp.Remark = entityRelation.Remark
	resp.OpenBurnAfterReadingTimeOut = entityRelation.BurnAfterReadingTimeOut

	//if s.cacheEnable {
	//	if err := s.cache.SetRelation(ctx, request.UserID, request.FriendId, resp, cache.RelationExpireTime); err != nil {
	//		log.Printf("set relation cache failed: %v", err)
	//	}
	//}

	return resp, nil
}

func (s *userServiceServer) GetUserRelationByUserIds(ctx context.Context, request *v1.GetUserRelationByUserIdsRequest) (*v1.GetUserRelationByUserIdsResponse, error) {
	resp := &v1.GetUserRelationByUserIdsResponse{}

	if request.FriendIds == nil || len(request.FriendIds) == 0 {
		return resp, nil
	}

	//if s.cacheEnable {
	//	r, err := s.cache.GetRelations(ctx, request.UserID, request.FriendIds)
	//	if err == nil && r != nil {
	//		resp.Users = r
	//		return resp, nil
	//	}
	//}

	relations, err := s.repos.UserRepo.Find(ctx, &repository.UserQuery{
		UserId:   request.UserId,
		FriendId: request.FriendIds,
	})
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrGetUserRelationFailed.Code()), err.Error())
	}

	for _, relation := range relations {
		resp.Users = append(resp.Users, &v1.GetUserRelationResponse{
			UserId:   relation.UserID,
			FriendId: relation.FriendID,
			Status:   v1.RelationStatus(relation.Status),
			DialogId: relation.DialogId,
		})
	}

	//// TODO 考虑使用异步的方式，缓存设置失败了，重试或回滚
	//if s.cacheEnable {
	//	for _, v := range resp.Users {
	//		if err := s.cache.SetRelation(ctx, request.UserID, v.FriendId, v, cache.RelationExpireTime); err != nil {
	//			log.Printf("failed to set get user relation cache: %v", err)
	//		}
	//	}
	//}

	return resp, nil
}

func (s *userServiceServer) SetFriendSilentNotification(ctx context.Context, request *v1.SetFriendSilentNotificationRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}

	if err := s.repos.UserRepo.SetUserFriendSilentNotification(ctx, request.UserId, request.FriendId, request.IsSilent); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrSetUserFriendSilentNotificationFailed.Code()), err.Error())
	}

	//// TODO 考虑使用异步的方式，缓存设置失败了，重试或回滚
	//if s.cacheEnable {
	//	if err := s.cache.DeleteRelation(ctx, request.UserID, request.FriendId); err != nil {
	//		log.Printf("delete relation cache failed: %v", err)
	//	}
	//	if err := s.cache.DeleteFriendList(ctx, request.UserID); err != nil {
	//		log.Printf("delete friend request list cache failed: %v", err)
	//	}
	//}

	return resp, nil
}

func (s *userServiceServer) SetUserOpenBurnAfterReading(ctx context.Context, request *v1.SetUserOpenBurnAfterReadingRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}

	if err := s.repos.UserRepo.SetUserOpenBurnAfterReading(
		ctx,
		request.UserId,
		request.FriendId,
		request.OpenBurnAfterReading,
		request.TimeOut,
	); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrSetUserOpenBurnAfterReadingFailed.Code()), err.Error())
	}

	//// TODO 考虑使用异步的方式，缓存设置失败了，重试或回滚
	//if s.cacheEnable {
	//	if err := s.cache.DeleteRelation(ctx, request.UserID, request.FriendId); err != nil {
	//		log.Printf("delete relation cache failed: %v", err)
	//	}
	//	if err := s.cache.DeleteFriendList(ctx, request.UserID); err != nil {
	//		log.Printf("delete friend request list cache failed: %v", err)
	//	}
	//}

	return resp, nil
}

func (s *userServiceServer) SetFriendRemark(ctx context.Context, request *v1.SetFriendRemarkRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}
	if err := s.repos.UserRepo.SetFriendRemark(ctx, request.UserId, request.FriendId, request.Remark); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrSetFriendRemarkFailed.Code()), err.Error())
	}

	//if s.cacheEnable {
	//	if err := s.cache.DeleteRelation(ctx, request.UserID, request.FriendId); err != nil {
	//		log.Printf("delete relation cache failed: %v", err)
	//	}
	//	if err := s.cache.DeleteFriendList(ctx, request.UserID, request.FriendId); err != nil {
	//		log.Printf("delete friend request list cache failed: %v", err)
	//	}
	//}

	return resp, nil
}
