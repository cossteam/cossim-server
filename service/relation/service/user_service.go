package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/cossim/coss-server/service/relation/domain/entity"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func (s *Service) AddFriend(ctx context.Context, request *v1.AddFriendRequest) (*v1.AddFriendResponse, error) {
	resp := &v1.AddFriendResponse{}

	userId := request.GetUserId()
	friendId := request.GetFriendId()
	// Fetch the existing relationship between the user and friend
	relation, err := s.urr.GetRelationByID(userId, friendId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		//if errors.Is(err, gorm.ErrRecordNotFound) {
		//	return resp, status.Error(codes.Code(code.RelationErrUserNotFound.Code()), err.Error())
		//}
		return resp, status.Error(codes.Code(code.RelationErrAddFriendFailed.Code()), err.Error())
	}

	if relation != nil {
		if relation.Status == entity.UserStatusPending {
			return resp, status.Error(codes.Code(code.RelationErrFriendRequestAlreadyPending.Code()), "好友状态处于申请中")
		} else if relation.Status == entity.UserStatusAdded {
			return resp, status.Error(codes.Code(code.RelationErrAlreadyFriends.Code()), "已经是好友")
		}
	}

	// Create a new UserRelation instance with relation status pending
	relation1 := &entity.UserRelation{
		UserID:   userId,
		FriendID: friendId,
		Remark:   request.Msg,
		Status:   entity.UserStatusPending,
		//Status: entity.UserStatusAdded,
	}
	//
	//if userId == friendId {
	//	relation1.Status = entity.UserStatusAdded
	//}

	// Save the new relationship to the database
	_, err = s.urr.CreateRelation(relation1)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrAddFriendFailed.Code()), err.Error())
	}

	relation2 := &entity.UserRelation{
		UserID:   friendId,
		FriendID: userId,
		Status:   entity.UserStatusApplying,
		//Status: entity.UserStatusAdded,
	}

	//if userId == friendId {
	//	relation2.Status = entity.UserStatusAdded
	//}

	// Save the new relationship to the database
	_, err = s.urr.CreateRelation(relation2)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrAddFriendFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s *Service) ConfirmFriend(ctx context.Context, request *v1.ConfirmFriendRequest) (*v1.ConfirmFriendResponse, error) {
	resp := &v1.ConfirmFriendResponse{}

	userId := request.GetUserId()
	friendId := request.GetFriendId()
	relation1, err := s.urr.GetRelationByID(userId, friendId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationUserErrNoFriendRequestRecords.Code()), err.Error())
		}
		return resp, status.Error(codes.Code(code.RelationErrConfirmFriendFailed.Code()), fmt.Sprintf("failed to get relation: %v", err))
	}

	if relation1.Status == entity.UserStatusAdded {
		return resp, status.Error(codes.Code(code.RelationErrAlreadyFriends.Code()), "已经是好友")
	}

	relation1.Status = entity.UserStatusAdded
	relation1.DialogId = uint(request.DialogId)
	_, err = s.urr.UpdateRelation(relation1)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrConfirmFriendFailed.Code()), fmt.Sprintf("failed to update relation: %v", err))
	}

	relation2, err := s.urr.GetRelationByID(friendId, userId)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrConfirmFriendFailed.Code()), fmt.Sprintf("failed to get relation: %v", err))
	}

	relation2.Status = entity.UserStatusAdded
	relation2.DialogId = uint(request.DialogId)
	_, err = s.urr.UpdateRelation(relation2)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrConfirmFriendFailed.Code()), fmt.Sprintf("failed to update relation: %v", err))
	}

	return resp, nil
}

func (s *Service) DeleteFriend(ctx context.Context, request *v1.DeleteFriendRequest) (*v1.DeleteFriendResponse, error) {
	resp := &v1.DeleteFriendResponse{}

	userId := request.GetUserId()
	friendId := request.GetFriendId()
	// Assuming urr is a UserRelationRepository instance in UserService
	if err := s.urr.DeleteRelationByID(userId, friendId); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrDeleteFriendFailed.Code()), fmt.Sprintf("failed to delete relation: %v", err))
	}

	if err := s.urr.DeleteRelationByID(friendId, userId); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrDeleteFriendFailed.Code()), fmt.Sprintf("failed to delete relation: %v", err))
	}

	return resp, nil
}

func (s *Service) AddBlacklist(ctx context.Context, request *v1.AddBlacklistRequest) (*v1.AddBlacklistResponse, error) {
	resp := &v1.AddBlacklistResponse{}

	userId := request.GetUserId()
	friendId := request.GetFriendId()
	// Assuming urr is a UserRelationRepository instance in UserService
	relation1, err := s.urr.GetRelationByID(userId, friendId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationUserErrFriendRelationNotFound.Code()), fmt.Sprintf("failed to retrieve relation: %v", err))
		}
		return resp, status.Error(codes.Code(code.RelationErrAddBlacklistFailed.Code()), fmt.Sprintf("failed to retrieve relation: %v", err))
	}

	if relation1.Status != entity.UserStatusAdded {
		return resp, code.RelationUserErrFriendRelationNotFound
	}

	relation1.Status = entity.UserStatusBlocked
	if _, err = s.urr.UpdateRelation(relation1); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrAddBlacklistFailed.Code()), fmt.Sprintf("failed to update relation: %v", err))
	}

	return resp, nil
}

func (s *Service) DeleteBlacklist(ctx context.Context, request *v1.DeleteBlacklistRequest) (*v1.DeleteBlacklistResponse, error) {
	resp := &v1.DeleteBlacklistResponse{}

	userId := request.GetUserId()
	friendId := request.GetFriendId()
	// Assuming urr is a UserRelationRepository instance in UserService
	relation1, err := s.urr.GetRelationByID(userId, friendId)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrDeleteBlacklistFailed.Code()), fmt.Sprintf("failed to retrieve relation: %v", err))
	}

	relation1.Status = entity.UserStatusAdded
	if _, err = s.urr.UpdateRelation(relation1); err != nil {
		return resp, status.Error(codes.Code(code.RelationErrDeleteBlacklistFailed.Code()), fmt.Sprintf("failed to update relation: %v", err))
	}

	//relation2, err := s.urr.GetRelationByID(friendId, userId)
	//if err != nil {
	//	return resp, code.RelationErrDeleteBlacklistFailed.Reason(fmt.Errorf("failed to retrieve relation: %w", err))
	//}
	//
	//relation2.Status = entity.UserStatusAdded
	//if _, err = s.urr.UpdateRelation(relation2); err != nil {
	//	return resp, code.RelationErrDeleteBlacklistFailed.Reason(fmt.Errorf("failed to update relation: %w", err))
	//}

	return resp, nil
}

func (s *Service) GetFriendList(ctx context.Context, request *v1.GetFriendListRequest) (*v1.GetFriendListResponse, error) {
	resp := &v1.GetFriendListResponse{}

	friends, err := s.urr.GetRelationsByUserID(request.GetUserId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationErrUserNotFound.Code()), fmt.Sprintf("failed to get friend list: %v", err))
		}
		return resp, status.Error(codes.Code(code.RelationErrGetFriendListFailed.Code()), fmt.Sprintf("failed to get friend list: %v", err))
	}

	for _, friend := range friends {
		resp.FriendList = append(resp.FriendList, &v1.Friend{UserId: friend.FriendID})
	}

	return resp, nil
}

func (s *Service) GetBlacklist(ctx context.Context, request *v1.GetBlacklistRequest) (*v1.GetBlacklistResponse, error) {
	resp := &v1.GetBlacklistResponse{}

	blacklist, err := s.urr.GetBlacklistByUserID(request.GetUserId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.RelationErrUserNotFound.Code()), fmt.Sprintf("failed to get black list: %v", err))
		}
		return resp, status.Error(codes.Code(code.RelationErrGetBlacklistFailed.Code()), fmt.Sprintf("failed to get black list: %v", err))
	}

	for _, black := range blacklist {
		resp.Blacklist = append(resp.Blacklist, &v1.Blacklist{UserId: black.UserID})
	}

	return resp, nil
}

func (s *Service) GetUserRelation(ctx context.Context, request *v1.GetUserRelationRequest) (*v1.GetUserRelationResponse, error) {
	resp := &v1.GetUserRelationResponse{}

	relation, err := s.urr.GetRelationByID(request.GetUserId(), request.GetFriendId())
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrGetUserRelationFailed.Code()), fmt.Sprintf("failed to get user relation: %v", err))
	}

	resp.Status = v1.RelationStatus(relation.Status)
	return resp, nil
}

func (s *Service) GetFriendRequestList(ctx context.Context, request *v1.GetFriendRequestListRequest) (*v1.GetFriendRequestListResponse, error) {
	resp := &v1.GetFriendRequestListResponse{}

	friends, err := s.urr.GetFriendRequestListByUserID(request.UserId)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrGetJoinRequestListFailed.Code()), err.Error())
	}

	for _, friend := range friends {
		fmt.Println("GetFriendRequestList => ", friend)
		resp.FriendRequestList = append(resp.FriendRequestList, &v1.FriendRequestList{
			UserId: friend.FriendID,
			Msg:    friend.Remark,
		})
	}

	return resp, nil
}
