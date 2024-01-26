package persistence

import (
	"github.com/cossim/coss-server/service/relation/domain/entity"
	"gorm.io/gorm"
)

type UserFriendRequestRepo struct {
	db *gorm.DB
}

func NewUserFriendRequestRepo(db *gorm.DB) *UserFriendRequestRepo {
	return &UserFriendRequestRepo{db: db}
}

func (u UserFriendRequestRepo) AddFriendRequest(entity *entity.UserFriendRequest) error {
	return u.db.Create(entity).Error
}

func (u UserFriendRequestRepo) GetFriendRequestList(userId string) ([]*entity.UserFriendRequest, error) {
	var result []*entity.UserFriendRequest
	err := u.db.Where("receiver_id = ? OR sender_id = ?", userId, userId).Find(&result).Error
	return result, err
}

func (u UserFriendRequestRepo) GetFriendRequestBySenderIDAndReceiverID(senderId string, receiverId string) (*entity.UserFriendRequest, error) {
	var result entity.UserFriendRequest
	return &result, u.db.Where("sender_id = ? AND receiver_id = ? AND status = ?", senderId, receiverId, entity.Pending).First(&result).Error
}

func (u UserFriendRequestRepo) GetFriendRequestByID(id uint) (*entity.UserFriendRequest, error) {
	var result entity.UserFriendRequest
	return &result, u.db.Where("id = ?", id).First(&result).Error
}

func (u UserFriendRequestRepo) UpdateFriendRequestStatus(id uint, status entity.RequestStatus) error {
	return u.db.Model(&entity.UserFriendRequest{}).Where("id = ?", id).Update("status", status).Error
}
