package repository

import (
	"context"
	"github.com/cossim/coss-server/internal/live/domain/entity"
	"time"
)

type Repository interface {
	// CreateRoom 创建房间
	CreateRoom(ctx context.Context, room *entity.Room) error
	// GetRoom 获取房间
	GetRoom(ctx context.Context, roomID string) (*entity.Room, error)
	// UpdateRoom 更新房间
	UpdateRoom(ctx context.Context, room *entity.Room) error
	// UpdateRoomWithExpiration 更新房间并设置过期时间
	UpdateRoomWithExpiration(ctx context.Context, room *entity.Room, expiration time.Duration) error
	// SetRoomPersist 设置房间永久不过期
	SetRoomPersist(ctx context.Context, roomID string) error
	// DeleteRoom 删除房间
	DeleteRoom(ctx context.Context, roomID string) error
	// ListRooms 获取所有房间
	ListRooms(ctx context.Context, roomIDs []string) ([]*entity.Room, error)
	// GetParticipant 获取房间参与者信息
	GetParticipant(ctx context.Context, roomID string, participantID string) (*entity.ParticipantInfo, error)
	// ListRoomsByCreator 获取指定用户创建的所有房间
	ListRoomsByCreator(ctx context.Context, creator string) ([]*entity.Room, error)
	// ListRoomsByOwner 获取指定用户拥有的所有房间
	ListRoomsByOwner(ctx context.Context, owner string) ([]*entity.Room, error)
	// GetUserRooms 获取用户正在通话的房间
	GetUserRooms(ctx context.Context, userID string) ([]*entity.Room, error)

	CreateUsersLive(ctx context.Context, roomID string, userIDs ...string) error
	DeleteUsersLive(ctx context.Context, userIDs ...string) error
	UpdateUserLiveExpiration(ctx context.Context, userID string, expiration time.Duration) error
	SetUserLivePersist(ctx context.Context, userID string) error

	CreateGroupLive(ctx context.Context, roomID string, groupID string) error
	DeleteGroupLive(ctx context.Context, groupID string) error
	GetGroupRoom(ctx context.Context, groupID string) (*entity.Room, error)
	UpdateGroupLiveExpiration(ctx context.Context, groupID string, expiration time.Duration) error
	SetGroupLivePersist(ctx context.Context, userID string) error
}
