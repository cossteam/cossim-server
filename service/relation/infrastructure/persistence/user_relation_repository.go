package persistence

import (
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
	if err := u.db.Save(ur).Error; err != nil {
		return nil, err
	}
	return ur, nil
}

func (u *UserRelationRepo) DeleteRelationByID(userId, friendId string) error {
	return u.db.Where("user_id = ? AND friend_id = ?", userId, friendId).Delete(&entity.UserRelation{}).Error
}

func (u *UserRelationRepo) GetRelationByID(userId, friendId string) (*entity.UserRelation, error) {
	var relation entity.UserRelation
	if err := u.db.Where("user_id = ? AND friend_id = ?", userId, friendId).First(&relation).Error; err != nil {
		return nil, err
	}
	return &relation, nil
}

func (u *UserRelationRepo) GetRelationsByUserID(userId string) ([]*entity.UserRelation, error) {
	var relations []*entity.UserRelation
	if err := u.db.Where("user_id = ? AND status = ?", userId, entity.RelationStatusAdded).Find(&relations).Error; err != nil {
		return nil, err
	}
	return relations, nil
}

func (u *UserRelationRepo) GetBlacklistByUserID(userId string) ([]*entity.UserRelation, error) {
	var relations []*entity.UserRelation
	if err := u.db.Where("user_id = ? AND status = ?", userId, entity.RelationStatusBlocked).Find(&relations).Error; err != nil {
		return nil, err
	}
	return relations, nil
}

func (u *UserRelationRepo) GetUserShowSessionUserIds(userId string) ([]string, error) {
	var ids []string
	u.db.Model(entity.UserRelation{}).Where("user_id =? AND session_show =?", userId, entity.IsShow).Pluck("friend_id", &ids)
	return ids, nil
}

func (u *UserRelationRepo) SetUserShowSession(userId, friendId string, showSession entity.ShowSession) error {
	return u.db.Model(entity.UserRelation{}).Where("user_id =? AND friend_id =?", userId, friendId).Update("session_show", showSession).Error
}
