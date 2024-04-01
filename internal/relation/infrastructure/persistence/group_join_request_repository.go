package persistence

import (
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

type GroupJoinRequestRepo struct {
	db *gorm.DB
}

var _ repository.GroupJoinRequestRepository = &GroupJoinRequestRepo{}

func NewGroupJoinRequestRepo(db *gorm.DB) *GroupJoinRequestRepo {
	return &GroupJoinRequestRepo{db: db}
}

func (g *GroupJoinRequestRepo) AddJoinRequest(en *entity.GroupJoinRequest) (*entity.GroupJoinRequest, error) {
	err := g.db.Create(en).Error
	if err != nil {
		return nil, err
	}
	return en, nil
}

func (g *GroupJoinRequestRepo) GetJoinRequestListByID(userId string) ([]*entity.GroupJoinRequest, error) {
	var result []*entity.GroupJoinRequest
	if err := g.db.Where("id = ? AND deleted_at = 0", userId).Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (g *GroupJoinRequestRepo) AddJoinRequestBatch(en []*entity.GroupJoinRequest) ([]*entity.GroupJoinRequest, error) {
	if len(en) == 0 {
		return nil, nil
	}
	if err := g.db.Create(&en).Error; err != nil {
		return nil, err
	}
	return en, nil
}

func (g *GroupJoinRequestRepo) GetGroupJoinRequestListByUserId(userID string) ([]*entity.GroupJoinRequest, error) {
	var result []*entity.GroupJoinRequest
	if err := g.db.Where("owner_id = ? AND deleted_at = 0", userID).Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (g *GroupJoinRequestRepo) GetGroupJoinRequestByGroupIdAndUserId(groupID uint, userID string) (*entity.GroupJoinRequest, error) {
	var result *entity.GroupJoinRequest
	if err := g.db.Where("group_id = ? AND user_id = ? AND deleted_at = 0", groupID, userID).Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (g *GroupJoinRequestRepo) ManageGroupJoinRequestByID(gid uint, uid string, status entity.RequestStatus) error {
	return g.db.Model(&entity.GroupJoinRequest{}).Where("group_id = ? AND user_id = ?  AND deleted_at = 0", gid, uid).Update("status", status).Error
}

func (g *GroupJoinRequestRepo) GetGroupJoinRequestByRequestID(id uint) (*entity.GroupJoinRequest, error) {
	var result *entity.GroupJoinRequest
	if err := g.db.Where("id = ? AND deleted_at = 0", id).Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (g *GroupJoinRequestRepo) GetJoinRequestBatchListByGroupIDs(gids []uint, uid string) ([]*entity.GroupJoinRequest, error) {
	var result []*entity.GroupJoinRequest
	if err := g.db.Where("group_id IN (?) AND owner_id = ? AND deleted_at = 0", gids, uid).Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (g *GroupJoinRequestRepo) GetJoinRequestListByRequestIDs(ids []uint) ([]*entity.GroupJoinRequest, error) {
	var result []*entity.GroupJoinRequest
	if err := g.db.Where("id IN (?) AND deleted_at = 0", ids).Find(&result).Error; err != nil {
		return nil, err
	}
	return result, nil
}

func (g *GroupJoinRequestRepo) DeleteJoinRequestByID(id uint) error {
	return g.db.Model(&entity.GroupJoinRequest{}).Where("id = ?", id).Update("deleted_at", time.Now()).Error
}
