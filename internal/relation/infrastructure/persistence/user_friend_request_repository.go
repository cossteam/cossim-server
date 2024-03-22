package persistence

import (
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

var _ repository.UserFriendRequestRepository = &UserFriendRequestRepo{}

type UserFriendRequestRepo struct {
	db *gorm.DB
}

func NewUserFriendRequestRepo(db *gorm.DB) *UserFriendRequestRepo {
	return &UserFriendRequestRepo{db: db}
}

func (u *UserFriendRequestRepo) AddFriendRequest(ent *entity.UserFriendRequest) (*entity.UserFriendRequest, error) {
	if err := u.db.Create(ent).Error; err != nil {
		return nil, err
	}

	return ent, nil
}

// GetFriendRequestList 获取指定用户 ID 的用户好友请求列表。
// deleted_by记录的是删除这条记录的用户ID列表，如果userId存在于deleted_by那么将查询不到这条记录
func (u *UserFriendRequestRepo) GetFriendRequestList(userId string) ([]*entity.UserFriendRequest, error) {
	var result []*entity.UserFriendRequest
	err := u.db.Where("(receiver_id = ? OR sender_id = ?) AND deleted_at = 0 AND deleted_by not like ?", userId, userId, "%"+userId+"%").Find(&result).Error
	return result, err
}

func (u *UserFriendRequestRepo) GetFriendRequestBySenderIDAndReceiverID(senderId string, receiverId string) (*entity.UserFriendRequest, error) {
	var result entity.UserFriendRequest
	return &result, u.db.Where("sender_id = ? AND receiver_id = ? AND status = ?", senderId, receiverId, entity.Pending).Order("created_at DESC").First(&result).Error
}

func (u *UserFriendRequestRepo) GetFriendRequestByID(id uint) (*entity.UserFriendRequest, error) {
	var result entity.UserFriendRequest
	return &result, u.db.Where("id = ? AND deleted_at = 0", id).First(&result).Error
}

func (u *UserFriendRequestRepo) UpdateFriendRequestStatus(id uint, status entity.RequestStatus) error {
	return u.db.Model(&entity.UserFriendRequest{}).Where("id = ?", id).Update("status", status).Error
}

func (u *UserFriendRequestRepo) DeleteFriendRequestByUserIdAndFriendIdRequest(userId string, friendId string) error {
	return u.db.Model(&entity.UserFriendRequest{}).Where("(sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?) AND status != ?", userId, friendId, friendId, userId, entity.Pending).Update("deleted_at", time.Now()).Error
}

func (u *UserFriendRequestRepo) DeletedById(id uint32) error {
	return u.db.Model(&entity.UserFriendRequest{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}

func (u *UserFriendRequestRepo) UpdateUserColumnById(id uint32, columns map[string]interface{}) error {
	return u.db.Model(&entity.UserFriendRequest{}).Where("id = ?", id).Unscoped().Updates(columns).Error

}
