package persistence

//import (
//	"github.com/cossim/coss-server/internal/group/domain/group"
//	"github.com/cossim/coss-server/pkg/utils/time"
//	"gorm.io/gorm"
//)
//
//type GroupRepo struct {
//	db *gorm.DB
//}
//
//func NewGroupRepo(db *gorm.DB) *GroupRepo {
//	return &GroupRepo{db: db}
//}
//
//func (repo *GroupRepo) GetGroupInfoByGid(gid uint) (*GroupModel, error) {
//	var group GroupModel
//	result := repo.db.First(&group, gid)
//	if result.Error != nil {
//		return nil, result.Error
//	}
//	return &group, nil
//}
//
//func (repo *GroupRepo) GetBatchGetGroupInfoByIDs(groupIds []uint) ([]*GroupModel, error) {
//	var groups []*GroupModel
//	result := repo.db.Find(&groups, groupIds)
//	if result.Error != nil {
//		return nil, result.Error
//	}
//	return groups, nil
//}
//
//func (repo *GroupRepo) UpdateGroup(group *GroupModel) (*GroupModel, error) {
//	result := repo.db.Where("id = ?", group.ID).Updates(group)
//	if result.Error != nil {
//		return nil, result.Error
//	}
//	return group, nil
//}
//
//func (repo *GroupRepo) InsertGroup(group *GroupModel) (*GroupModel, error) {
//	result := repo.db.Create(group)
//	if result.Error != nil {
//		return nil, result.Error
//	}
//	return group, nil
//}
//
//func (repo *GroupRepo) DeleteGroup(gid uint) error {
//	err := repo.db.Model(&GroupModel{}).Where("id = ?", gid).Update("status", group.StatusDeleted).Update("deleted_at", time.Now()).Error
//	return err
//}
//
//func (repo *GroupRepo) UpdateGroupByGroupID(gid uint, updateFields map[string]interfaces{}) error {
//	return repo.db.Model(&GroupModel{}).Where("id = ?", gid).Unscoped().Updates(updateFields).Error
//}
