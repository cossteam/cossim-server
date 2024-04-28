package relation

import "context"

type UserFriendRequestQuery struct {
	UserID     string
	SenderId   string
	ReceiverId string
	Force      bool
	Status     *RequestStatus
	PageSize   int
	PageNum    int
}

type UserFriendRequestRepository interface {
	Get(ctx context.Context, id uint32) (*UserFriendRequest, error)
	Create(ctx context.Context, entity *UserFriendRequest) (*UserFriendRequest, error)
	// Find 根据条件查询好友申请列表
	Find(ctx context.Context, query *UserFriendRequestQuery) (*UserFriendRequestList, error)
	// Delete 根据记录id删除好友申请记录
	Delete(ctx context.Context, id uint32) error
	// UpdateFields 根据 ID 更新 UserFriendRequest 对象的多个字段
	UpdateFields(ctx context.Context, id uint32, fields map[string]interface{}) error

	// UpdateStatus 更新申请记录状态
	UpdateStatus(ctx context.Context, id uint32, status RequestStatus) error

	GetByUserIdAndFriendId(ctx context.Context, senderId, receiverId string) (*UserFriendRequest, error)
}
