package repository

import "github.com/cossim/coss-server/service/relation/domain/entity"

type UserFriendRequestRepository interface {
	AddFriendRequest(entity *entity.UserFriendRequest) error
	GetFriendRequestList(userId string) ([]*entity.UserFriendRequest, error)
	GetFriendRequestBySenderIDAndReceiverID(senderId string, receiverId string) (*entity.UserFriendRequest, error)
	GetFriendRequestByID(id uint) (*entity.UserFriendRequest, error)
	UpdateFriendRequestStatus(id uint, status entity.RequestStatus) error
}
