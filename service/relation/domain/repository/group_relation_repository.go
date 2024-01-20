package repository

import "github.com/cossim/coss-server/service/relation/domain/entity"

type GroupRelationRepository interface {
	CreateRelation(ur *entity.GroupRelation) (*entity.GroupRelation, error)
	UpdateRelation(ur *entity.GroupRelation) (*entity.GroupRelation, error)
	UpdateRelationColumnByGroupAndUser(gid uint32, uid string, column string, value interface{}) error
	DeleteRelationByID(gid uint32, uid string) error
	DeleteUserGroupRelationByGroupIDAndUserID(gid uint32, uid string) error
	GetGroupUserIDs(gid uint32) ([]string, error)
	GetUserGroupIDs(uid string) ([]uint32, error)
	GetUserGroupByID(gid uint32, uid string) (*entity.GroupRelation, error)
	GetUserJoinedGroupIDs(uid string) ([]uint32, error) // 获取用户加入的所有群聊ID
	GetUserManageGroupIDs(uid string) ([]uint32, error) // 获取用户管理的或创建的群聊ID
	GetJoinRequestBatchListByID(gids []uint32) ([]*entity.GroupRelation, error)
	GetJoinRequestListByID(gid uint32) ([]*entity.GroupRelation, error)
	DeleteGroupRelationByID(gid uint32) error
	GetGroupAdminIds(gid uint32) ([]string, error)

	// UpdateGroupRelationByGroupID 根据群聊id更新群聊的所有用户信息
	UpdateGroupRelationByGroupID(dialogID uint32, updateFields map[string]interface{}) error
}
