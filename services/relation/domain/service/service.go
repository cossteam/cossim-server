package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/services/relation/domain/entity"
	"github.com/cossim/coss-server/services/relation/domain/repository"
	"gorm.io/gorm"
)

type UserService struct {
	urr repository.UserRelationRepository
}

func NewUserService(urr repository.UserRelationRepository) *UserService {
	return &UserService{
		urr: urr,
	}
}

func (u *UserService) AddFriend(ctx context.Context, userId, friendId string) (*entity.UserRelation, error) {
	// Fetch the existing relationship between the user and friend
	relation, err := u.urr.GetRelationByID(userId, friendId)
	if err != nil {
		return nil, fmt.Errorf("failed to add friend: %w", err)
	}

	if relation != nil {
		if relation.Status == entity.RelationStatusPending {
			return nil, fmt.Errorf("friend request already pending")
		} else if relation.Status == entity.RelationStatusAdded {
			return nil, fmt.Errorf("already friends")
		}
	}

	// Create a new UserRelation instance with relation status pending
	newRelation := &entity.UserRelation{
		UserID:   userId,
		FriendID: friendId,
		Status:   entity.RelationStatusPending,
	}

	// Save the new relationship to the database
	savedRelation, err := u.urr.CreateRelation(newRelation)
	if err != nil {
		return nil, fmt.Errorf("failed to add friend: %w", err)
	}

	return savedRelation, nil
}

func (u *UserService) ConfirmFriend(ctx context.Context, userId string, friendId string) (interface{}, error) {
	relation, err := u.urr.GetRelationByID(userId, friendId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve relation: %w", err)
	}

	if relation == nil {
		return nil, fmt.Errorf("relation not found")
	}

	relation.Status = entity.RelationStatusAdded
	if _, err = u.urr.UpdateRelation(relation); err != nil {
		return nil, fmt.Errorf("failed to update relation: %w", err)
	}

	return nil, nil
}

func (u *UserService) DeleteFriend(ctx context.Context, userId, friendId string) (interface{}, error) {
	// Assuming urr is a UserRelationRepository instance in UserService
	if err := u.urr.DeleteRelationByID(userId, friendId); err != nil {
		return nil, fmt.Errorf("failed to delete friend: %w", err)
	}

	if err := u.urr.DeleteRelationByID(friendId, userId); err != nil {
		return nil, fmt.Errorf("failed to delete friend: %w", err)
	}

	return nil, nil
}

func (u *UserService) AddBlacklist(ctx context.Context, userId, friendId string) (interface{}, error) {
	// Assuming urr is a UserRelationRepository instance in UserService
	relation1, err := u.urr.GetRelationByID(userId, friendId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve relation: %w", err)
	}

	if relation1 == nil {
		return nil, fmt.Errorf("relation not found")
	}

	relation1.Status = entity.RelationStatusBlocked
	if _, err = u.urr.UpdateRelation(relation1); err != nil {
		return nil, fmt.Errorf("failed to update relation: %w", err)
	}

	relation2, err := u.urr.GetRelationByID(friendId, userId)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve relation: %w", err)
	}

	if relation2 == nil {
		return nil, fmt.Errorf("relation not found")
	}

	relation2.Status = entity.RelationStatusBlocked
	if _, err = u.urr.UpdateRelation(relation2); err != nil {
		return nil, fmt.Errorf("failed to update relation: %w", err)
	}

	return nil, nil
}

func (u *UserService) DeleteBlacklist(ctx context.Context, userId, friendId string) (interface{}, error) {
	// Assuming urr is a UserRelationRepository instance in UserService
	if err := u.urr.DeleteRelationByID(userId, friendId); err != nil {
		return nil, fmt.Errorf("failed to delete from blacklist: %w", err)
	}

	if err := u.urr.DeleteRelationByID(friendId, userId); err != nil {
		return nil, fmt.Errorf("failed to delete from blacklist: %w", err)
	}

	return nil, nil
}

func (u *UserService) GetFriendList(ctx context.Context, userId string) ([]*entity.UserRelation, error) {
	friends, err := u.urr.GetRelationsByUserID(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("未找到用户")
		}
		return nil, fmt.Errorf("获取用户好友信息失败: %w", err)
	}

	return friends, nil
}
