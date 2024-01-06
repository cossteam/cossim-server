package service

import (
	"github.com/cossim/coss-server/service/relation/domain/entity"
	"github.com/cossim/coss-server/service/relation/domain/repository"
)

type UserGroupService struct {
	urr repository.GroupRelationRepository
}

func NewUserGroupService(urr repository.GroupRelationRepository) *UserGroupService {
	return &UserGroupService{
		urr: urr,
	}
}

func (s *UserGroupService) InsertUserGroup(userID string, groupID uint) (*entity.UserGroup, error) {
	// 创建领域模型
	userGroup := &entity.UserGroup{
		UID:     userID,
		GroupID: groupID,
	}

	// 调用持久层方法插入用户群关系
	insertedUserGroup, err := s.urr.InsertUserGroup(userGroup)
	if err != nil {
		return nil, err
	}

	return insertedUserGroup, nil
}

func (s *UserGroupService) GetUserGroupIDs(groupID uint) ([]string, error) {
	// 调用持久层方法获取用户群关系列表
	userGroupIDs, err := s.urr.GetUserGroupIDs(groupID)
	if err != nil {
		return nil, err
	}

	return userGroupIDs, nil
}
