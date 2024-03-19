package repository

import "github.com/cossim/coss-server/internal/relation/domain/entity"

type UserFriendRequestRepository interface {
	AddFriendRequest(entity *entity.UserFriendRequest) (*entity.UserFriendRequest, error)
	GetFriendRequestList(userId string) ([]*entity.UserFriendRequest, error)
	GetFriendRequestBySenderIDAndReceiverID(senderId string, receiverId string) (*entity.UserFriendRequest, error)
	GetFriendRequestByID(id uint) (*entity.UserFriendRequest, error)
	UpdateFriendRequestStatus(id uint, status entity.RequestStatus) error
	DeleteFriendRequestByUserIdAndFriendIdRequest(userId string, friendId string) error
}
