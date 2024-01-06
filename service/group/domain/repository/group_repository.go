package repository

import "github.com/cossim/coss-server/service/group/domain/entity"

type GroupRepository interface {
	GetGroupInfoByGid(gid uint) (*entity.Group, error)
	GetBatchGetGroupInfoByIDs(groupIds []uint) ([]*entity.Group, error)
	UpdateGroup(group *entity.Group) (*entity.Group, error)
	InsertGroup(group *entity.Group) (*entity.Group, error)
	DeleteGroup(gid uint) error
}
