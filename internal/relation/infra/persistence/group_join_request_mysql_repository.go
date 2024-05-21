package persistence

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/cache"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/internal/relation/infra/persistence/converter"
	"github.com/cossim/coss-server/internal/relation/infra/persistence/po"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var _ repository.GroupRequestRepository = &MySQLGroupJoinRequestRepository{}

func NewMySQLGroupJoinRequestRepository(db *gorm.DB, cache cache.RelationUserCache) *MySQLGroupJoinRequestRepository {
	return &MySQLGroupJoinRequestRepository{
		db: db,
		//cache: cache,
	}
}

type MySQLGroupJoinRequestRepository struct {
	db *gorm.DB
}

func (m *MySQLGroupJoinRequestRepository) GetByGroupIDAndUserID(ctx context.Context, groupID uint32, userID string) (*entity.GroupJoinRequest, error) {
	model := &po.GroupJoinRequest{}

	if err := m.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(model).Error; err != nil {
		return nil, err
	}

	return converter.GroupJoinRequestPoToEntity(model), nil
}

func (m *MySQLGroupJoinRequestRepository) Get(ctx context.Context, id uint32) (*entity.GroupJoinRequest, error) {
	model := &po.GroupJoinRequest{}

	if err := m.db.WithContext(ctx).Where("id = ?", id).First(model).Error; err != nil {
		return nil, err
	}

	return converter.GroupJoinRequestPoToEntity(model), nil
}

func (m *MySQLGroupJoinRequestRepository) Create(ctx context.Context, entity *entity.GroupJoinRequest) (*entity.GroupJoinRequest, error) {
	model := converter.GroupJoinRequestEntityToPo(entity)

	if err := m.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}

	return converter.GroupJoinRequestPoToEntity(model), nil
}

func (m *MySQLGroupJoinRequestRepository) Find(ctx context.Context, query *repository.GroupJoinRequestQuery) ([]*entity.GroupJoinRequest, error) {
	var models []*po.GroupJoinRequest

	db := m.db.WithContext(ctx).Model(&po.GroupJoinRequest{})

	if query.ID != nil && len(query.ID) > 0 {
		db = db.Where("id IN (?)", query.ID)
	}
	if query.GroupID != nil && len(query.GroupID) > 0 {
		db = db.Where("group_id IN (?)", query.GroupID)
	}
	if query.UserID != nil && len(query.UserID) > 0 {
		db = db.Where("owner_id IN (?)", query.UserID)
	}
	if query.PageSize > 0 && query.PageNum > 0 {
		offset := (query.PageNum - 1) * query.PageSize
		db = db.Offset(offset).Limit(query.PageSize)
	}

	if !query.Force {
		db = db.Where("deleted_at = 0")
	}

	result := db.Find(&models)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "failed to find user friend requests")
	}

	var es []*entity.GroupJoinRequest

	for _, model := range models {
		es = append(es, converter.GroupJoinRequestPoToEntity(model))
	}

	return es, nil
}

func (m *MySQLGroupJoinRequestRepository) Creates(ctx context.Context, es []*entity.GroupJoinRequest) ([]*entity.GroupJoinRequest, error) {
	var models []*po.GroupJoinRequest

	if len(es) == 0 {
		return nil, errors.New("entity is empty")
	}

	for _, e := range es {
		model := converter.GroupJoinRequestEntityToPo(e)
		models = append(models, model)
	}

	if err := m.db.WithContext(ctx).Create(&models).Error; err != nil {
		return nil, err
	}

	var res []*entity.GroupJoinRequest
	for _, model := range models {
		res = append(res, converter.GroupJoinRequestPoToEntity(model))
	}

	return res, nil
}

func (m *MySQLGroupJoinRequestRepository) UpdateStatus(ctx context.Context, id uint32, status entity.RequestStatus) error {
	model := &po.GroupJoinRequest{}
	model.Status = uint8(status)

	if err := m.db.WithContext(ctx).
		Model(&po.GroupJoinRequest{}).
		Where("id = ?", id).
		Update("status", model.Status).
		Error; err != nil {
		return err
	}

	return nil
}

func (m *MySQLGroupJoinRequestRepository) Delete(ctx context.Context, id uint32) error {

	if err := m.db.WithContext(ctx).
		Model(&po.GroupJoinRequest{}).
		Where("id = ?", id).
		Update("deleted_at", ptime.Now()).
		Error; err != nil {
		return err
	}

	return nil
}
