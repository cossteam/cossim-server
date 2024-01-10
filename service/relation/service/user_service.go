package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/cossim/coss-server/service/relation/domain/entity"
	"gorm.io/gorm"
)

func (s *Service) AddFriend(ctx context.Context, request *v1.AddFriendRequest) (*v1.AddFriendResponse, error) {
	resp := &v1.AddFriendResponse{}

	userId := request.GetUserId()
	friendId := request.GetFriendId()
	// Fetch the existing relationship between the user and friend
	relation, err := s.urr.GetRelationByID(userId, friendId)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return resp, fmt.Errorf("failed to add friend: %w", err)
	}

	if relation != nil {
		if relation.Status == entity.RelationStatusPending {
			return resp, fmt.Errorf("friend request already pending")
		} else if relation.Status == entity.RelationStatusAdded {
			return resp, fmt.Errorf("already friends")
		}
	}

	// Create a new UserRelation instance with relation status pending
	relation1 := &entity.UserRelation{
		UserID:   userId,
		FriendID: friendId,
		//Status:   entity.RelationStatusPending,
		Status: entity.RelationStatusAdded,
	}

	// Save the new relationship to the database
	_, err = s.urr.CreateRelation(relation1)
	if err != nil {
		return resp, fmt.Errorf("failed to add friend: %w", err)
	}

	relation2 := &entity.UserRelation{
		UserID:   friendId,
		FriendID: userId,
		Status:   entity.RelationStatusAdded,
	}

	// Save the new relationship to the database
	_, err = s.urr.CreateRelation(relation2)
	if err != nil {
		return resp, fmt.Errorf("failed to add friend: %w", err)
	}

	return resp, nil
}

func (s *Service) ConfirmFriend(ctx context.Context, request *v1.ConfirmFriendRequest) (*v1.ConfirmFriendResponse, error) {
	resp := &v1.ConfirmFriendResponse{}

	userId := request.GetUserId()
	friendId := request.GetFriendId()
	relation, err := s.urr.GetRelationByID(userId, friendId)
	if err != nil {
		return resp, fmt.Errorf("failed to retrieve relation: %w", err)
	}

	if relation == nil {
		return resp, fmt.Errorf("relation not found")
	}

	newRelation := &entity.UserRelation{
		UserID:   friendId,
		FriendID: userId,
		Status:   entity.RelationStatusAdded,
	}

	// Save the new relationship to the database
	_, err = s.urr.CreateRelation(newRelation)
	if err != nil {
		return resp, fmt.Errorf("failed to add friend: %w", err)
	}

	relation.Status = entity.RelationStatusAdded
	if _, err = s.urr.UpdateRelation(relation); err != nil {
		return resp, fmt.Errorf("failed to update relation: %w", err)
	}

	return resp, nil
}

func (s *Service) DeleteFriend(ctx context.Context, request *v1.DeleteFriendRequest) (*v1.DeleteFriendResponse, error) {
	resp := &v1.DeleteFriendResponse{}

	userId := request.GetUserId()
	friendId := request.GetFriendId()
	// Assuming urr is a UserRelationRepository instance in UserService
	if err := s.urr.DeleteRelationByID(userId, friendId); err != nil {
		return resp, fmt.Errorf("failed to delete friend: %w", err)
	}

	if err := s.urr.DeleteRelationByID(friendId, userId); err != nil {
		return resp, fmt.Errorf("failed to delete friend: %w", err)
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
		return resp, fmt.Errorf("failed to retrieve relation: %w", err)
	}

	if relation1 == nil {
		return resp, fmt.Errorf("relation not found")
	}

	relation1.Status = entity.RelationStatusBlocked
	if _, err = s.urr.UpdateRelation(relation1); err != nil {
		return resp, fmt.Errorf("failed to update relation: %w", err)
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
		return resp, fmt.Errorf("failed to retrieve relation: %w", err)
	}

	if relation1 == nil {
		return resp, fmt.Errorf("relation not found")
	}

	relation1.Status = entity.RelationStatusAdded
	if _, err = s.urr.UpdateRelation(relation1); err != nil {
		return resp, fmt.Errorf("failed to update relation: %w", err)
	}

	relation2, err := s.urr.GetRelationByID(friendId, userId)
	if err != nil {
		return resp, fmt.Errorf("failed to retrieve relation: %w", err)
	}

	if relation2 == nil {
		return resp, fmt.Errorf("relation not found")
	}

	relation2.Status = entity.RelationStatusAdded
	if _, err = s.urr.UpdateRelation(relation2); err != nil {
		return resp, fmt.Errorf("failed to update relation: %w", err)
	}

	return resp, nil
}

func (s *Service) GetFriendList(ctx context.Context, request *v1.GetFriendListRequest) (*v1.GetFriendListResponse, error) {
	resp := &v1.GetFriendListResponse{}

	friends, err := s.urr.GetRelationsByUserID(request.GetUserId())
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, fmt.Errorf("未找到用户")
		}
		return resp, fmt.Errorf("获取用户好友信息失败: %w", err)
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
			return resp, fmt.Errorf("未找到用户")
		}
		return resp, fmt.Errorf("获取用户好友信息失败: %w", err)
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
		return resp, nil
	}

	resp.Status = v1.RelationStatus(relation.Status)
	return resp, nil
}
