package repository

import "github.com/cossim/coss-server/internal/relation/domain/entity"

type UserFriendRequestRepository interface {
	AddFriendRequest(entity *entity.UserFriendRequest) (*entity.UserFriendRequest, error)
	GetFriendRequestList(userId string) ([]*entity.UserFriendRequest, error)
	GetFriendRequestBySenderIDAndReceiverID(senderId string, receiverId string) (*entity.UserFriendRequest, error)
	GetFriendRequestByID(id uint) (*entity.UserFriendRequest, error)
	UpdateFriendRequestStatus(id uint, status entity.RequestStatus) error
	DeleteFriendRequestByUserIdAndFriendIdRequest(userId string, friendId string) error
	// DeletedById 根据id删除好友申请记录
	DeletedById(id uint32) error
	// UpdateUserColumnById 根据id更新指定字段的值
	UpdateUserColumnById(id uint32, columns map[string]interface{}) error
}
