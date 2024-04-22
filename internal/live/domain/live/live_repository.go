package live

import (
	"context"
	"time"
)

type Repository interface {
	// CreateRoom 创建房间
	CreateRoom(ctx context.Context, room *Room) error
	// GetRoom 获取房间
	GetRoom(ctx context.Context, roomID string) (*Room, error)
	// UpdateRoom 更新房间
	UpdateRoom(ctx context.Context, room *Room) error
	// UpdateRoomWithExpiration 更新房间并设置过期时间
	UpdateRoomWithExpiration(ctx context.Context, room *Room, expiration time.Duration) error
	// SetRoomPersist 设置房间永久不过期
	SetRoomPersist(ctx context.Context, roomID string) error
	// DeleteRoom 删除房间
	DeleteRoom(ctx context.Context, roomID string) error
	// ListRooms 获取所有房间
	ListRooms(ctx context.Context, roomIDs []string) ([]*Room, error)
	// GetParticipant 获取房间参与者信息
	GetParticipant(ctx context.Context, roomID string, participantID string) (*ParticipantInfo, error)
	// ListRoomsByCreator 获取指定用户创建的所有房间
	ListRoomsByCreator(ctx context.Context, creator string) ([]*Room, error)
	// ListRoomsByOwner 获取指定用户拥有的所有房间
	ListRoomsByOwner(ctx context.Context, owner string) ([]*Room, error)
	// GetUserRooms 获取用户正在通话的房间
	GetUserRooms(ctx context.Context, userID string) ([]*Room, error)

	CreateUsersLive(ctx context.Context, roomID string, userIDs ...string) error
	DeleteUsersLive(ctx context.Context, userIDs ...string) error
	UpdateUserLiveExpiration(ctx context.Context, userID string, expiration time.Duration) error
	SetUserLivePersist(ctx context.Context, userID string) error

	CreateGroupLive(ctx context.Context, roomID string, groupID string) error
	DeleteGroupLive(ctx context.Context, groupID string) error
	GetGroupRoom(ctx context.Context, groupID string) (*Room, error)
	UpdateGroupLiveExpiration(ctx context.Context, groupID string, expiration time.Duration) error
	SetGroupLivePersist(ctx context.Context, userID string) error
}
