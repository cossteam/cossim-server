package persistence

import (
	"github.com/cossim/coss-server/pkg/utils/time"
	"github.com/cossim/coss-server/service/relation/domain/entity"
	"gorm.io/gorm"
)

type UserFriendRequestRepo struct {
	db *gorm.DB
}

func NewUserFriendRequestRepo(db *gorm.DB) *UserFriendRequestRepo {
	return &UserFriendRequestRepo{db: db}
}

func (u UserFriendRequestRepo) AddFriendRequest(ent *entity.UserFriendRequest) (*entity.UserFriendRequest, error) {

	if err := u.db.Create(ent).Error; err != nil {
		return nil, err
	}

	return ent, nil
}

func (u UserFriendRequestRepo) GetFriendRequestList(userId string) ([]*entity.UserFriendRequest, error) {
	var result []*entity.UserFriendRequest
	err := u.db.Where("receiver_id = ? OR sender_id = ? AND deleted_at = 0", userId, userId).Find(&result).Error
	return result, err
}

func (u UserFriendRequestRepo) GetFriendRequestBySenderIDAndReceiverID(senderId string, receiverId string) (*entity.UserFriendRequest, error) {
	var result entity.UserFriendRequest
	return &result, u.db.Where("sender_id = ? AND receiver_id = ? AND status = ?", senderId, receiverId, entity.Pending).Order("created_at DESC").First(&result).Error
}

func (u UserFriendRequestRepo) GetFriendRequestByID(id uint) (*entity.UserFriendRequest, error) {
	var result entity.UserFriendRequest
	return &result, u.db.Where("id = ?", id).First(&result).Error
}

func (u UserFriendRequestRepo) UpdateFriendRequestStatus(id uint, status entity.RequestStatus) error {
	return u.db.Model(&entity.UserFriendRequest{}).Where("id = ?", id).Update("status", status).Error
}

func (u UserFriendRequestRepo) DeleteFriendRequestByUserIdAndFriendIdRequest(userId string, friendId string) error {
	return u.db.Model(&entity.UserFriendRequest{}).Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?) AND status != ?", userId, friendId, friendId, userId, entity.Pending).Update("deleted_at", time.Now()).Error
}
