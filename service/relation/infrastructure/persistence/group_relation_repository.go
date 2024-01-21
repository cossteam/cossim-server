package persistence

import (
	"github.com/cossim/coss-server/service/relation/domain/entity"
	"gorm.io/gorm"
	"time"
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

func (repo *GroupRelationRepo) CreateRelations(urs []*entity.GroupRelation) ([]*entity.GroupRelation, error) {
	return urs, repo.db.CreateInBatches(urs, len(urs)).Error
}

func (repo *GroupRelationRepo) GetGroupUserIDs(gid uint32) ([]string, error) {
	var userGroupIDs []string
	if err := repo.db.Model(&entity.GroupRelation{}).Where("group_id = ? AND status = ? AND deleted_at = 0", gid, entity.GroupStatusJoined).Pluck("user_id", &userGroupIDs).Error; err != nil {
		return nil, err
	}
	return userGroupIDs, nil
}

func (repo *GroupRelationRepo) GetUserGroupIDs(uid string) ([]uint32, error) {
	var groupIDs []uint32
	if err := repo.db.Model(&entity.GroupRelation{}).Where("user_id = ?", uid).Pluck("group_id", &groupIDs).Error; err != nil {
		return groupIDs, err
	}
	return groupIDs, nil
}

func (repo *GroupRelationRepo) GetUserJoinedGroupIDs(uid string) ([]uint32, error) {
	var groupIDs []uint32
	if err := repo.db.Model(&entity.GroupRelation{}).Where("user_id = ? AND status = ?", uid, entity.GroupStatusJoined).Pluck("group_id", &groupIDs).Error; err != nil {
		return groupIDs, err
	}
	return groupIDs, nil
}

func (repo *GroupRelationRepo) GetUserManageGroupIDs(uid string) ([]uint32, error) {
	var groupIDs []uint32
	if err := repo.db.Model(&entity.GroupRelation{}).Distinct("group_id").Where("user_id = ? AND identity IN ?", uid, []entity.GroupIdentity{entity.IdentityAdmin, entity.IdentityOwner}).Pluck("group_id", &groupIDs).Error; err != nil {
		return nil, err
	}
	return groupIDs, nil
}

func (repo *GroupRelationRepo) UpdateRelation(ur *entity.GroupRelation) (*entity.GroupRelation, error) {
	if err := repo.db.Model(&entity.GroupRelation{}).Where("id = ?", ur.ID).Updates(ur).Error; err != nil {
		return ur, err
	}
	return ur, nil
}

func (repo *GroupRelationRepo) UpdateRelationColumnByGroupAndUser(gid uint32, uid string, column string, value interface{}) error {
	return repo.db.Model(&entity.GroupRelation{}).Where("group_id = ? and user_id = ?", gid, uid).Update(column, value).Error
}

func (repo *GroupRelationRepo) DeleteRelationByID(gid uint32, uid string) error {
	return repo.db.Model(&entity.GroupRelation{}).Where("group_id = ? and user_id = ?", gid, uid).Update("status", entity.UserStatusDeleted).Update("deleted_at", time.Now().Unix()).Error
}

func (repo *GroupRelationRepo) GetUserGroupByID(gid uint32, uid string) (*entity.GroupRelation, error) {
	var ug entity.GroupRelation
	if err := repo.db.Model(&entity.GroupRelation{}).Where(" group_id = ? and user_id = ? AND deleted_at = 0", gid, uid).First(&ug).Error; err != nil {
		return nil, err
	}
	return &ug, nil
}

func (repo *GroupRelationRepo) GetUserGroupByIDs(gid uint32, uids []string) ([]*entity.GroupRelation, error) {
	var ugs []*entity.GroupRelation
	if err := repo.db.Model(&entity.GroupRelation{}).Where(" group_id = ? and user_id IN (?) AND deleted_at = 0", gid, uids).Find(&ugs).Error; err != nil {
		return nil, err
	}
	return ugs, nil
}

func (repo *GroupRelationRepo) GetJoinRequestListByID(gid uint32) ([]*entity.GroupRelation, error) {
	var joinRequests []*entity.GroupRelation
	if err := repo.db.Where("group_id = ? AND status = ?", gid, entity.GroupStatusApplying).Find(&joinRequests).Error; err != nil {
		return nil, err
	}
	return joinRequests, nil
}

func (repo *GroupRelationRepo) DeleteGroupRelationByID(gid uint32) error {
	return repo.db.Model(&entity.GroupRelation{}).Where("group_id = ?", gid).Update("status", entity.GroupStatusDeleted).Update("deleted_at", time.Now().Unix()).Error
}

func (repo *GroupRelationRepo) DeleteUserGroupRelationByGroupIDAndUserID(gid uint32, uid string) error {
	return repo.db.Model(&entity.GroupRelation{}).Where("group_id = ? AND user_id = ?", gid, uid).Update("status", entity.GroupStatusDeleted).Update("deleted_at", time.Now().Unix()).Error
}

func (repo *GroupRelationRepo) GetJoinRequestBatchListByID(gids []uint32) ([]*entity.GroupRelation, error) {
	var joinRequests []*entity.GroupRelation
	if err := repo.db.Where("group_id IN (?) AND status NOT IN ?", gids, []entity.GroupRelationStatus{entity.GroupStatusBlocked, entity.GroupStatusDeleted}).Find(&joinRequests).Error; err != nil {
		return joinRequests, err
	}
	return joinRequests, nil
}

func (repo *GroupRelationRepo) GetGroupAdminIds(gid uint32) ([]string, error) {
	var adminIds []string
	repo.db.Model(&entity.GroupRelation{}).Where(" group_id = ? AND status = ? AND deleted_at = 0", gid, entity.IdentityAdmin).Pluck("user_id", &adminIds)
	return adminIds, nil
}

func (repo *GroupRelationRepo) UpdateGroupRelationByGroupID(groupID uint32, updateFields map[string]interface{}) error {
	return repo.db.Model(&entity.GroupRelation{}).Where("group_id = ?", groupID).Updates(updateFields).Error
}

func (repo *GroupRelationRepo) DeleteRelationByGroupIDAndUserIDs(gid uint32, uid []string) error {
	return repo.db.Model(&entity.GroupRelation{}).Where(" group_id = ? AND status = ? AND deleted_at = 0", gid, entity.GroupStatusJoined).Delete(&entity.GroupRelation{}).Error
}
