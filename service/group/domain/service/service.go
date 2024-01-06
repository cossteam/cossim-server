package service

import (
	"github.com/cossim/coss-server/service/group/domain/entity"
	"github.com/cossim/coss-server/service/group/domain/repository"
)

type GroupService struct {
	ur repository.GroupRepository
}

func NewGroupService(repo repository.GroupRepository) *GroupService {
	return &GroupService{ur: repo}
}

func (s *GroupService) GetGroupInfoByGid(gid uint) (*entity.Group, error) {
	group, err := s.ur.GetGroupInfoByGid(gid)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (s *GroupService) GetBatchGetGroupInfoByIDs(groupIds []uint) ([]*entity.Group, error) {
	groups, err := s.ur.GetBatchGetGroupInfoByIDs(groupIds)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

func (s *GroupService) UpdateGroup(group *entity.Group) (*entity.Group, error) {
	group, err := s.ur.UpdateGroup(group)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (s *GroupService) InsertGroup(group *entity.Group) (*entity.Group, error) {
	group, err := s.ur.InsertGroup(group)
	if err != nil {
		return nil, err
	}
	return group, nil
}

func (s *GroupService) DeleteGroup(gid uint) error {
	return s.ur.DeleteGroup(gid)
}
