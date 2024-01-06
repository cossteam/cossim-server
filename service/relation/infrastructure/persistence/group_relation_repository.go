package persistence

import (
	"github.com/cossim/coss-server/service/relation/domain/entity"
	"gorm.io/gorm"
)

type GroupRelationRepo struct {
	db *gorm.DB
}

func NewGroupRelationRepo(db *gorm.DB) *GroupRelationRepo {
	return &GroupRelationRepo{db: db}
}

func (repo *GroupRelationRepo) InsertUserGroup(ur *entity.UserGroup) (*entity.UserGroup, error) {
	if err := repo.db.Create(ur).Error; err != nil {
		return nil, err
	}
	return ur, nil
}

func (repo *GroupRelationRepo) GetUserGroupIDs(gid uint) ([]string, error) {
	var userGroupIDs []string
	if err := repo.db.Model(&entity.UserGroup{}).Where("group_id = ?", gid).Pluck("uid", &userGroupIDs).Error; err != nil {
		return nil, err
	}
	return userGroupIDs, nil
}
