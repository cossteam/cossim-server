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

func (repo *GroupRelationRepo) CreateRelation(ur *entity.GroupRelation) (*entity.GroupRelation, error) {
	if err := repo.db.Create(ur).Error; err != nil {
		return nil, err
	}
	return ur, nil
}

func (repo *GroupRelationRepo) GetUserGroupIDs(gid uint32) ([]string, error) {
	var userGroupIDs []string
	if err := repo.db.Model(&entity.GroupRelation{}).Where("group_id = ?", gid).Pluck("user_id", &userGroupIDs).Error; err != nil {
		return nil, err
	}
	return userGroupIDs, nil
}

func (repo *GroupRelationRepo) UpdateRelation(ur *entity.GroupRelation) (*entity.GroupRelation, error) {
	if err := repo.db.Model(&entity.GroupRelation{}).Where("id = ?", ur.ID).Updates(ur).Error; err != nil {
		return ur, err
	}
	return ur, nil
}

func (repo *GroupRelationRepo) DeleteRelationByID(gid uint32, uid string) error {
	return repo.db.Model(&entity.GroupRelation{}).Where("group_id = ? and user_id = ?", gid, uid).Delete(&entity.GroupRelation{}).Error
}

func (repo *GroupRelationRepo) GetUserGroupByID(gid uint32, uid string) (*entity.GroupRelation, error) {
	var ug entity.GroupRelation
	if err := repo.db.Model(&entity.GroupRelation{}).Where(" group_id = ? and user_id = ?", gid, uid).First(&ug).Error; err != nil {
		return nil, err
	}
	return &ug, nil
}

func (repo *GroupRelationRepo) GetJoinRequestListByID(gid uint32) ([]*entity.GroupRelation, error) {
	var joinRequests []*entity.GroupRelation
	if err := repo.db.Where("group_id = ? AND status = ?", gid, entity.GroupStatusApplying).Find(&joinRequests).Error; err != nil {
		return nil, err
	}
	return joinRequests, nil
}
func (repo *GroupRelationRepo) DeleteGroupRelationByID(gid uint32) error {
	return repo.db.Model(&entity.GroupRelation{}).Where("group_id = ?", gid).Delete(&entity.GroupRelation{}).Error
}

func (repo *GroupRelationRepo) GetGroupAdminIds(gid uint32) ([]string, error) {
	var adminIds []string
	repo.db.Model(&entity.GroupRelation{}).Where(" group_id = ? AND status = ?", gid, entity.IdentityAdmin).Pluck("user_id", &adminIds)
	return adminIds, nil
}
