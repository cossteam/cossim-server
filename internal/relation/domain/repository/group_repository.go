package repository

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
)

type CreateGroupRelation struct {
	GroupID     uint32
	UserID      string
	Identity    entity.GroupIdentity
	EntryMethod entity.EntryMethod
	Inviter     string
	JoinedAt    int64
}

type GroupRepository interface {
	Get(ctx context.Context, id uint32) (*entity.GroupRelation, error)
	Create(ctx context.Context, createGroupRelation *CreateGroupRelation) (*entity.GroupRelation, error)
	Update(ctx context.Context, ur *entity.GroupRelation) (*entity.GroupRelation, error)
	Delete(ctx context.Context, id uint32) error

	// DeleteByGroupID 删除群聊的所有关系
	DeleteByGroupID(ctx context.Context, gid uint32) error

	// GetGroupUserIDs 获取群聊所有的成员id
	GetGroupUserIDs(ctx context.Context, gid uint32) ([]string, error)

	// GetUserGroupIDs 获取用户加入的所有群聊id
	GetUserGroupIDs(ctx context.Context, uid string) ([]uint32, error)

	// GetUserGroupByGroupIDAndUserID 获取用户加入的群聊信息
	GetUserGroupByGroupIDAndUserID(ctx context.Context, gid uint32, uid string) (*entity.GroupRelation, error)

	// GetUsersGroupByGroupIDAndUserIDs 获取用户加入的群聊信息
	GetUsersGroupByGroupIDAndUserIDs(ctx context.Context, gid uint32, uids []string) ([]*entity.GroupRelation, error)

	// GetUserJoinedGroupIDs 获取用户加入的所有群聊ID
	GetUserJoinedGroupIDs(ctx context.Context, uid string) ([]uint32, error)

	// GetUserManageGroupIDs 获取用户管理的或创建的群聊ID
	GetUserManageGroupIDs(ctx context.Context, uid string) ([]uint32, error)

	// DeleteByGroupIDAndUserID 根据群聊id和用户id删除群聊关系
	DeleteByGroupIDAndUserID(ctx context.Context, gid uint32, uid ...string) error

	// ListJoinRequest 获取群聊的入群请求
	ListJoinRequest(ctx context.Context, gids []uint32) ([]*entity.GroupRelation, error)

	// ListGroupAdmin 获取群聊管理员
	ListGroupAdmin(ctx context.Context, gid uint32) ([]string, error)

	// SetUserGroupRemark 设置用户的群聊备注
	SetUserGroupRemark(ctx context.Context, gid uint32, uid string, remark string) error

	// UpdateIdentity 更新用户身份
	UpdateIdentity(ctx context.Context, gid uint32, uid string, identity entity.GroupIdentity) error

	// UserGroupSilentNotification 设置用户的群聊是否开启免打扰
	UserGroupSilentNotification(ctx context.Context, gid uint32, uid string, silentNotification bool) error

	// UpdateFieldsByGroupAndUser 根据群聊和用户id关系对应的 GroupRelation 对象的多个字段
	UpdateFieldsByGroupAndUser(ctx context.Context, gid uint32, uid string, fields map[string]interface{}) error

	// UpdateFieldsByGroupID 更新这个群聊的所有 GroupRelation 对象的多个字段
	UpdateFieldsByGroupID(ctx context.Context, id uint32, fields map[string]interface{}) error

	// UpdateFields 根据记录id更新 GroupRelation 对象的多个字段
	UpdateFields(ctx context.Context, id uint32, fields map[string]interface{}) error
}
