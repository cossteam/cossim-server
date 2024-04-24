package command

import (
	"context"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
)

const (
	UserRelationNormal = uint(relationgrpcv1.RelationStatus_RELATION_NORMAL)
)

type UserRelationship struct {
	ID     string
	Status uint
	Remark string
}

type RelationUserService interface {
	// GetUserRelationships retrieves the relationship status between the current user and the given users.
	GetUserRelationships(ctx context.Context, currentUserID string, userIDs []string) (map[string]UserRelationship, error)

	// IsFriendsWithAll checks if the current user is friends with all of the given user IDs.
	IsFriendsWithAll(ctx context.Context, currentUserID string, userIDs []string) (bool, error)
}

type GroupRelationship struct {
	UserId                      string
	ID                          uint32
	Status                      int32
	Identity                    uint
	JoinMethod                  uint
	JoinTime                    int64
	MuteEndTime                 int64
	IsSilent                    uint
	Inviter                     string
	Remark                      string
	OpenBurnAfterReading        uint
	OpenBurnAfterReadingTimeOut uint
}

type RelationGroupService interface {
	CreateGroup(ctx context.Context, groupID uint32, currentUserID string, memberIDs []string) (uint32, error)
	CreateGroupRevert(ctx context.Context, groupID uint32, currentUserID string, memberIDs []string) error
	IsGroupOwner(ctx context.Context, groupID uint32, userID string) (bool, error)
	GetGroupMembers(ctx context.Context, groupID uint32) ([]string, error)
	// DeleteGroupRelations 删除群聊所有用户关系
	DeleteGroupRelations(ctx context.Context, groupID uint32) error
	DeleteGroupRelationsRevert(ctx context.Context, groupID uint32) error

	// GetRelation 获取群聊用户关系
	GetRelation(ctx context.Context, groupID uint32, userID string) (*GroupRelationship, error)
}

type RelationDialogService interface {
	GetGroupDialogID(ctx context.Context, groupID uint32) (uint32, error)
	// DeleteUserDialog 删除所有的用户对话记录
	DeleteUserDialog(ctx context.Context, dialogID uint32) error
	DeleteUserDialogRevert(ctx context.Context, dialogID uint32) error

	DeleteDialog(ctx context.Context, dialogID uint32) error
	DeleteDialogRevert(ctx context.Context, dialogID uint32) error
}

type User struct {
	ID       string
	NickName string
}

type UserService interface {
	GetUserInfo(ctx context.Context, userID string) (*User, error)
}

type PushService interface {
	Push(ctx context.Context, t int32, data []byte) (interface{}, error)
}

type Group struct {
	ID              uint32
	Type            uint
	MaxMembersLimit int32
	Name            string
	Avatar          string
	CreatorID       string
	SilenceTime     int64
	JoinApprove     bool
	Encrypt         bool
}

type GroupService interface {
	Get(ctx context.Context, id uint32) (*Group, error)
}
