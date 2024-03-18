package repository

import "github.com/cossim/coss-server/internal/group/domain/entity"

type GroupRepository interface {
	GetGroupInfoByGid(gid uint) (*entity.Group, error)
	GetBatchGetGroupInfoByIDs(groupIds []uint) ([]*entity.Group, error)
	UpdateGroup(group *entity.Group) (*entity.Group, error)
	InsertGroup(group *entity.Group) (*entity.Group, error)
	DeleteGroup(gid uint) error
	// UpdateGroupByGroupID 根据群聊id更新群聊信息
	UpdateGroupByGroupID(gid uint, updateFields map[string]interface{}) error
}
