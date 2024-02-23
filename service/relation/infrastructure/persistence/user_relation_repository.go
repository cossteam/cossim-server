package persistence

import (
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/utils/time"
	"github.com/cossim/coss-server/service/relation/domain/entity"
	"gorm.io/gorm"
)

// UserRelationRepo 需要实现UserRelationRepository接口
type UserRelationRepo struct {
	db *gorm.DB
}

func NewUserRelationRepo(db *gorm.DB) *UserRelationRepo {
	return &UserRelationRepo{db: db}
}

func (u *UserRelationRepo) CreateRelation(ur *entity.UserRelation) (*entity.UserRelation, error) {
	if err := u.db.Create(ur).Error; err != nil {
		return nil, err
	}
	return ur, nil
}

func (u *UserRelationRepo) UpdateRelation(ur *entity.UserRelation) (*entity.UserRelation, error) {
	if err := u.db.Model(&entity.UserRelation{}).Where("ID = ?", ur.ID).Updates(ur).Error; err != nil {
		return nil, err
	}
	return ur, nil
}

func (u *UserRelationRepo) DeleteRelationByID(userId, friendId string) error {
	return u.db.Model(&entity.UserRelation{}).Where("user_id = ? AND friend_id = ? AND deleted_at = 0", userId, friendId).Update("status", entity.UserStatusDeleted).Update("deleted_at", time.Now()).Error
}

func (u *UserRelationRepo) UpdateRelationDeleteAtByID(userId, friendId string, deleteAt int64) error {
	return u.db.Model(&entity.UserRelation{}).Where("user_id = ? AND friend_id = ?", userId, friendId).Update("deleted_at", deleteAt).Error
}

func (u *UserRelationRepo) GetRelationByID(userId, friendId string) (*entity.UserRelation, error) {
	var relation entity.UserRelation
	if err := u.db.Where("user_id = ? AND friend_id = ? AND deleted_at = 0", userId, friendId).Order("created_at DESC").First(&relation).Error; err != nil {
		return nil, err
	}
	return &relation, nil
}

func (u *UserRelationRepo) GetRelationByIDs(userId string, friendIds []string) ([]*entity.UserRelation, error) {
	var relations []*entity.UserRelation
	if err := u.db.Where("user_id = ? AND friend_id IN (?) AND status = ? AND deleted_at = 0", userId, friendIds, entity.UserStatusNormal).Find(&relations).Error; err != nil {
		return nil, err
	}
	return relations, nil
}

func (u *UserRelationRepo) GetRelationsByUserID(userId string) ([]*entity.UserRelation, error) {
	var relations []*entity.UserRelation
	if err := u.db.Where("user_id = ? AND status = ? AND deleted_at = 0 AND friend_id NOT IN (?)", userId, entity.UserStatusNormal, constants.SystemUserList).Find(&relations).Error; err != nil {
		return nil, err
	}
	return relations, nil
}

func (u *UserRelationRepo) GetBlacklistByUserID(userId string) ([]*entity.UserRelation, error) {
	var relations []*entity.UserRelation
	if err := u.db.Where("user_id = ? AND status = ? AND deleted_at = 0", userId, entity.UserStatusBlocked).Find(&relations).Error; err != nil {
		return nil, err
	}
	return relations, nil
}

func (u *UserRelationRepo) GetFriendRequestListByUserID(userId string) ([]*entity.UserRelation, error) {
	var urs []*entity.UserRelation
	if err := u.db.Where("user_id = ? AND status NOT IN ? AND deleted_at = 0", userId, []entity.UserRelationStatus{entity.UserStatusBlocked, entity.UserStatusDeleted}).Find(&urs).Error; err != nil {
		return nil, err
	}
	return urs, nil
}

func (u *UserRelationRepo) UpdateRelationColumn(id uint, column string, value interface{}) error {
	return u.db.Model(&entity.UserRelation{}).Where("id = ?", id).Update(column, value).Error
}

func (u *UserRelationRepo) SetUserFriendSilentNotification(uid, friendId string, silentNotification entity.SilentNotification) error {
	return u.db.Model(&entity.UserRelation{}).Where("user_id = ? AND friend_id = ?", uid, friendId).Update("silent_notification", silentNotification).Error
}

func (u *UserRelationRepo) SetUserOpenBurnAfterReading(uid, friendId string, openBurnAfterReading entity.OpenBurnAfterReadingType) error {
	return u.db.Model(&entity.UserRelation{}).Where("user_id = ? AND friend_id = ?", uid, friendId).Update("open_burn_after_reading", openBurnAfterReading).Error
}

func (u *UserRelationRepo) SetFriendRemarkByUserIdAndFriendId(userId, friendId string, remark string) error {
	return u.db.Model(&entity.UserRelation{}).Where("user_id = ? AND friend_id = ?", userId, friendId).Update("remark", remark).Error
}
