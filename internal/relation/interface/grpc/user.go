package grpc

import (
	"context"
	"errors"
	"fmt"
	v1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/internal/relation/infrastructure/persistence"
	"github.com/cossim/coss-server/pkg/cache"
	"github.com/cossim/coss-server/pkg/code"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	emptypb "google.golang.org/protobuf/types/known/emptypb"
	"gorm.io/gorm"
	"log"
)

var _ v1.UserRelationServiceServer = &userServiceServer{}

type userServiceServer struct {
	db          *gorm.DB
	cache       cache.RelationUserCache
	cacheEnable bool
	urr         repository.UserRelationRepository
	dr          repository.DialogRepository
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

	//查询是否单删
	relation, err := s.urr.GetRelationByID(request.FriendId, request.UserId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return resp, status.Error(codes.Code(code.RelationErrAddFriendFailed.Code()), formatErrorMessage(err))
	}
	if relation != nil {
		if _, err := s.urr.CreateRelation(&entity.UserRelation{
			UserID:   request.UserId,
			FriendID: request.FriendId,
			Status:   entity.UserRelationStatus(v1.RelationStatus_RELATION_NORMAL),
			DialogId: relation.DialogId,
		}); err != nil {
			return resp, status.Error(codes.Code(code.RelationErrAddFriendFailed.Code()), formatErrorMessage(err))
		}
		return resp, nil
	}
	//双方都没有好友关系
	if err := s.db.Transaction(func(tx *gorm.DB) error {
		npo := persistence.NewRepositories(tx)

		dialog, err := npo.Dr.CreateDialog(request.UserId, entity.DialogType(v1.DialogType_USER_DIALOG), 0)
		if err != nil {
			return err
		}

		if _, err := npo.Urr.CreateRelation(&entity.UserRelation{
			UserID:   request.UserId,
			FriendID: request.FriendId,
			Status:   entity.UserRelationStatus(v1.RelationStatus_RELATION_NORMAL),
			DialogId: dialog.ID,
		}); err != nil {
			return err
		}
		if _, err := npo.Urr.CreateRelation(&entity.UserRelation{
			UserID:   request.FriendId,
			FriendID: request.UserId,
			Status:   entity.UserRelationStatus(v1.RelationStatus_RELATION_NORMAL),
			DialogId: dialog.ID,
		}); err != nil {
			return err
		}
		_, err = npo.Dr.JoinDialog(dialog.ID, request.UserId)
		if err != nil {
			return err
		}

		_, err = npo.Dr.JoinDialog(dialog.ID, request.FriendId)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrAddFriendFailed.Code()), formatErrorMessage(err))
	}

	if s.cacheEnable {
		if err := s.cache.DeleteFriendIDs(ctx, request.UserId, request.FriendId); err != nil {
			log.Printf("failed to delete cache friend ids: %v", err)
		}
	}

	return resp, nil
}

func (s *userServiceServer) DeleteFriend(ctx context.Context, request *v1.DeleteFriendRequest) (*v1.DeleteFriendResponse, error) {
	resp := &v1.DeleteFriendResponse{}

	if err := s.urr.DeleteRelationByID(request.UserId, request.FriendId); err != nil {
		//return resp, status.Error(codes.Code(code.RelationErrDeleteFriendFailed.Code()), fmt.Sprintf("failed to delete relation: %v", err))
		return resp, status.Error(codes.Aborted, fmt.Sprintf("failed to delete relation: %v", err))
	}

	if s.cacheEnable {
		if err := s.cache.DeleteRelation(ctx, request.UserId, request.FriendId); err != nil {
			return resp, status.Error(codes.Aborted, fmt.Sprintf("failed to delete cache relation: %v", err))
		}
		if err := s.cache.DeleteFriendList(ctx, request.UserId); err != nil {
			return resp, status.Error(codes.Aborted, fmt.Sprintf("failed to delete cache friend list: %v", err))
		}
	}

	return resp, nil
}

func (s *userServiceServer) DeleteFriendRevert(ctx context.Context, request *v1.DeleteFriendRequest) (*v1.DeleteFriendResponse, error) {
	resp := &v1.DeleteFriendResponse{}
	if err := s.urr.UpdateRelationDeleteAtByID(request.UserId, request.FriendId, 0); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrDeleteFriendFailed.Code()), fmt.Sprintf("DeleteFriendRevert failed to update relation: %v", err))
	}
	return resp, nil
}

func (s *userServiceServer) AddBlacklist(ctx context.Context, request *v1.AddBlacklistRequest) (*v1.AddBlacklistResponse, error) {
	resp := &v1.AddBlacklistResponse{}

	if request.UserId == request.FriendId {
		return resp, status.Error(codes.Code(code.RelationErrAddBlacklistFailed.Code()), "user cannot add themselves to the blacklist")
	}

	// urr is a UserRelationRepository instance in UserService
	relation1, err := s.urr.GetRelationByID(request.UserId, request.FriendId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationUserErrFriendRelationNotFound.Code()), fmt.Sprintf("failed to retrieve relation: %v", err))
		}
		return resp, status.Error(codes.Code(code.RelationErrAddBlacklistFailed.Code()), fmt.Sprintf("failed to retrieve relation: %v", err))
	}

	if relation1.Status == entity.UserStatusBlocked {
		return resp, code.RelationErrAlreadyBlacklist
	}

	if relation1.Status != entity.UserStatusNormal {
		return resp, code.RelationUserErrFriendRelationNotFound
	}

	relation1.Status = entity.UserStatusBlocked
	if err = s.urr.UpdateRelationColumn(relation1.ID, "status", entity.UserStatusBlocked); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrAddBlacklistFailed.Code()), fmt.Sprintf("failed to update relation: %v", err))
	}

	if s.cacheEnable {
		if err := s.cache.DeleteBlacklist(ctx, request.UserId); err != nil {
			log.Printf("failed to delete cache blacklist: %v", err)
		}
		if err := s.cache.DeleteRelation(ctx, request.UserId, request.FriendId); err != nil {
			log.Printf("failed to delete cache relation: %v", err)
		}
		if err := s.cache.DeleteFriendList(ctx, request.UserId); err != nil {
			log.Printf("failed to delete cache friend list: %v", err)
		}
	}

	return resp, nil
}

func (s *userServiceServer) DeleteBlacklist(ctx context.Context, request *v1.DeleteBlacklistRequest) (*v1.DeleteBlacklistResponse, error) {
	resp := &v1.DeleteBlacklistResponse{}

	if request.UserId == request.FriendId {
		return resp, status.Error(codes.Code(code.RelationErrDeleteBlacklistFailed.Code()), "user cannot delete themselves from the blacklist")
	}

	// Assuming urr is a UserRelationRepository instance in UserService
	relation1, err := s.urr.GetRelationByID(request.UserId, request.FriendId)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrDeleteBlacklistFailed.Code()), fmt.Sprintf("failed to retrieve relation: %v", err))
	}

	relation1.Status = entity.UserStatusNormal
	if _, err = s.urr.UpdateRelation(relation1); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrDeleteBlacklistFailed.Code()), fmt.Sprintf("failed to update relation: %v", err))
	}

	if s.cacheEnable {
		if err := s.cache.DeleteBlacklist(ctx, request.UserId); err != nil {
			log.Printf("failed to delete cache blacklist: %v", err)
		}
		if err := s.cache.DeleteRelation(ctx, request.UserId, request.FriendId); err != nil {
			log.Printf("failed to delete cache relation: %v", err)
		}
		if err := s.cache.DeleteFriendList(ctx, request.UserId); err != nil {
			log.Printf("failed to delete cache friend list: %v", err)
		}
	}

	return resp, nil
}

func (s *userServiceServer) GetFriendList(ctx context.Context, request *v1.GetFriendListRequest) (*v1.GetFriendListResponse, error) {
	resp := &v1.GetFriendListResponse{}

	if s.cacheEnable {
		// 从缓存中获取关系对象
		r, err := s.cache.GetFriendList(ctx, request.UserId)
		if err == nil && r != nil {
			fmt.Println("获取好友列表 缓存存在返回缓存结果")
			// 如果缓存中存在，则直接返回缓存的结果
			return r, nil
		}
		fmt.Println("获取好友列表 缓存不存在，从数据库中获取关系对象")
	}

	friends, err := s.urr.GetRelationsByUserID(request.GetUserId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationErrUserNotFound.Code()), fmt.Sprintf("failed to get friend list: %v", err))
		}
		return resp, status.Error(codes.Code(code.RelationErrGetFriendListFailed.Code()), fmt.Sprintf("failed to get friend list: %v", err))
	}

	for _, friend := range friends {
		resp.FriendList = append(resp.FriendList,
			&v1.Friend{
				UserId:                      friend.FriendID,
				DialogId:                    uint32(friend.DialogId),
				Remark:                      friend.Remark,
				Status:                      v1.RelationStatus(friend.Status),
				IsSilent:                    v1.UserSilentNotificationType(friend.SilentNotification),
				OpenBurnAfterReading:        v1.OpenBurnAfterReadingType(friend.OpenBurnAfterReading),
				OpenBurnAfterReadingTimeOut: friend.BurnAfterReadingTimeOut,
			})
	}

	if s.cacheEnable {
		if err := s.cache.SetFriendList(ctx, request.UserId, resp, cache.RelationExpireTime); err != nil {
			log.Printf("failed to set get friend list cache: %v", err)
		}
	}

	return resp, nil
}

func (s *userServiceServer) GetBlacklist(ctx context.Context, request *v1.GetBlacklistRequest) (*v1.GetBlacklistResponse, error) {
	resp := &v1.GetBlacklistResponse{}

	if s.cacheEnable {
		// 尝试从缓存中获取黑名单列表
		cachedList, err := s.cache.GetBlacklist(ctx, request.GetUserId())
		if err == nil && cachedList != nil {
			// 如果缓存中存在，则直接返回缓存的结果
			return cachedList, nil
		}
	}

	blacklist, err := s.urr.GetBlacklistByUserID(request.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationErrUserNotFound.Code()), fmt.Sprintf("failed to get black list: %v", err))
		}
		return resp, status.Error(codes.Code(code.RelationErrGetBlacklistFailed.Code()), fmt.Sprintf("failed to get black list: %v", err))
	}

	for _, black := range blacklist {
		resp.Blacklist = append(resp.Blacklist, &v1.Blacklist{UserId: black.FriendID})
	}

	if s.cacheEnable {
		if err := s.cache.SetBlacklist(ctx, request.UserId, resp, cache.RelationExpireTime); err != nil {
			log.Printf("failed to set blacklist cache: %v", err)
		}
	}

	return resp, nil
}

func (s *userServiceServer) GetUserRelation(ctx context.Context, request *v1.GetUserRelationRequest) (*v1.GetUserRelationResponse, error) {
	resp := &v1.GetUserRelationResponse{}
	var err error

	if s.cacheEnable {
		// 从缓存中获取关系对象
		relation, err := s.cache.GetRelation(ctx, request.UserId, request.FriendId)
		if err == nil && relation != nil {
			// 如果缓存中存在，则直接返回缓存的结果
			return relation, nil
		}
	}

	// 从数据库中获取关系对象
	entityRelation, err := s.urr.GetRelationByID(request.GetUserId(), request.GetFriendId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationUserErrFriendRelationNotFound.Code()), err.Error())
		}
		return resp, err
	}

	// 将entity.UserRelation的字段值复制到v1.GetUserRelationResponse
	resp.Status = v1.RelationStatus(entityRelation.Status)
	resp.DialogId = uint32(entityRelation.DialogId)
	resp.UserId = entityRelation.UserID
	resp.FriendId = entityRelation.FriendID
	resp.IsSilent = v1.UserSilentNotificationType(entityRelation.SilentNotification)
	resp.OpenBurnAfterReading = v1.OpenBurnAfterReadingType(entityRelation.OpenBurnAfterReading)
	resp.Remark = entityRelation.Remark
	resp.IsSilent = v1.UserSilentNotificationType(entityRelation.SilentNotification)
	resp.OpenBurnAfterReadingTimeOut = entityRelation.BurnAfterReadingTimeOut

	if s.cacheEnable {
		// 将关系对象存储到缓存中
		err = s.cache.SetRelation(ctx, request.UserId, request.FriendId, resp, cache.RelationExpireTime)
		if err != nil {
			log.Printf("set relation cache failed: %v", err)
		}
	}

	return resp, nil
}

func (s *userServiceServer) GetUserRelationByUserIds(ctx context.Context, request *v1.GetUserRelationByUserIdsRequest) (*v1.GetUserRelationByUserIdsResponse, error) {
	resp := &v1.GetUserRelationByUserIdsResponse{}

	if request.FriendIds == nil || len(request.FriendIds) == 0 {
		return resp, nil
	}

	if s.cacheEnable {
		r, err := s.cache.GetRelations(ctx, request.UserId, request.FriendIds)
		if err != nil {
			return nil, err
		}
		if err == nil && r != nil {
			resp.Users = r
			return resp, nil
		}
	}

	relations, err := s.urr.GetRelationByIDs(request.UserId, request.FriendIds)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrGetUserRelationFailed.Code()), err.Error())
	}

	for _, relation := range relations {
		resp.Users = append(resp.Users, &v1.GetUserRelationResponse{
			UserId:   relation.UserID,
			FriendId: relation.FriendID,
			Status:   v1.RelationStatus(relation.Status),
			DialogId: uint32(relation.DialogId),
		})
	}

	go func() {
		if s.cacheEnable {
			for _, v := range resp.Users {
				if err := s.cache.SetRelation(ctx, request.UserId, v.FriendId, v, cache.RelationExpireTime); err != nil {
					log.Printf("failed to set get user relation cache: %v", err)
				}
			}
		}
	}()

	return resp, nil
}

func (s *userServiceServer) SetFriendSilentNotification(ctx context.Context, request *v1.SetFriendSilentNotificationRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}
	if err := s.urr.SetUserFriendSilentNotification(request.UserId, request.FriendId, entity.SilentNotification(request.IsSilent)); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrSetUserFriendSilentNotificationFailed.Code()), err.Error())
	}
	if s.cacheEnable {
		if err := s.cache.DeleteRelation(ctx, request.UserId, request.FriendId); err != nil {
			log.Printf("delete relation cache failed: %v", err)
		}
		if err := s.cache.DeleteFriendList(ctx, request.UserId); err != nil {
			log.Printf("delete friend request list cache failed: %v", err)
		}
	}
	return resp, nil
}

func (s *userServiceServer) SetUserOpenBurnAfterReading(ctx context.Context, request *v1.SetUserOpenBurnAfterReadingRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}
	if err := s.urr.SetUserOpenBurnAfterReading(
		request.UserId,
		request.FriendId,
		entity.OpenBurnAfterReadingType(request.OpenBurnAfterReading),
		request.TimeOut,
	); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrSetUserOpenBurnAfterReadingFailed.Code()), err.Error())
	}
	if s.cacheEnable {
		if err := s.cache.DeleteRelation(ctx, request.UserId, request.FriendId); err != nil {
			log.Printf("delete relation cache failed: %v", err)
		}
		if err := s.cache.DeleteFriendList(ctx, request.UserId); err != nil {
			log.Printf("delete friend request list cache failed: %v", err)
		}
	}
	return resp, nil
}

func (s *userServiceServer) SetFriendRemark(ctx context.Context, request *v1.SetFriendRemarkRequest) (*emptypb.Empty, error) {
	var resp = &emptypb.Empty{}
	if err := s.urr.SetFriendRemarkByUserIdAndFriendId(request.UserId, request.FriendId, request.Remark); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrSetFriendRemarkFailed.Code()), err.Error())
	}
	if s.cacheEnable {
		if err := s.cache.DeleteRelation(ctx, request.UserId, request.FriendId); err != nil {
			log.Printf("delete relation cache failed: %v", err)
		}
		if err := s.cache.DeleteFriendList(ctx, request.UserId, request.FriendId); err != nil {
			log.Printf("delete friend request list cache failed: %v", err)
		}
	}
	return resp, nil
}
