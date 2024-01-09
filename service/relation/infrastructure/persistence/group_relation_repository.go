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

func (repo *GroupRelationRepo) GetUserGroupShowSessionGroupIds(userId string) ([]uint, error) {
	var groupIDs []uint
	repo.db.Model(&entity.UserGroup{}).Where("uid =? AND session_show = ?", userId, entity.IsShow).Pluck("group_id", &groupIDs)
	return groupIDs, nil
}

func (repo *GroupRelationRepo) SetUserGroupShowSession(userId string, groupId uint, showSession entity.ShowSession) error {
	return repo.db.Model(&entity.UserGroup{}).Where("uid =? AND group_id = ?", userId, groupId).Update("session_show", showSession).Error
}
