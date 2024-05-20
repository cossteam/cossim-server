package repository

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"gorm.io/gorm"
)

type UserFriendRequestQuery struct {
	UserID     string
	SenderId   string
	ReceiverId string
	Force      bool
	Status     *entity.RequestStatus
	PageSize   int
	PageNum    int
}

type UserFriendRequestRepository interface {
	NewRepository(db *gorm.DB) UserFriendRequestRepository

	Get(ctx context.Context, id uint32) (*entity.UserFriendRequest, error)
	Create(ctx context.Context, entity *entity.UserFriendRequest) (*entity.UserFriendRequest, error)
	// Find 根据条件查询好友申请列表
	Find(ctx context.Context, query *UserFriendRequestQuery) (*entity.UserFriendRequestList, error)
	// Delete 根据记录id删除好友申请记录
	Delete(ctx context.Context, id uint32) error
	// UpdateFields 根据 ID 更新 UserFriendRequest 对象的多个字段
	UpdateFields(ctx context.Context, id uint32, fields map[string]interface{}) error

	// UpdateStatus 更新申请记录状态
	UpdateStatus(ctx context.Context, id uint32, status entity.RequestStatus) error

	GetByUserIdAndFriendId(ctx context.Context, senderId, receiverId string) (*entity.UserFriendRequest, error)
}
