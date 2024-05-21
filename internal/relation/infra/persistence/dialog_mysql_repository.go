package persistence

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/internal/relation/cache"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/internal/relation/infra/persistence/converter"
	"github.com/cossim/coss-server/internal/relation/infra/persistence/po"
	"github.com/cossim/coss-server/pkg/code"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

var _ repository.DialogRepository = &MySQLDialogRepository{}

func NewMySQLMySQLDialogRepository(db *gorm.DB, cache cache.RelationUserCache) *MySQLDialogRepository {
	return &MySQLDialogRepository{
		db: db,
		//cache: cache,
	}
}

type MySQLDialogRepository struct {
	db *gorm.DB
}

func (m *MySQLDialogRepository) GetByGroupID(ctx context.Context, groupID uint32) (*entity.Dialog, error) {
	model := &po.Dialog{}

	if err := m.db.WithContext(ctx).
		Where("group_id = ? AND deleted_at = 0", groupID).
		First(model).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.NotFound
		}
		return nil, err
	}

	return converter.DialogEntityPoToEntity(model), nil
}

func (m *MySQLDialogRepository) Creates(ctx context.Context, dialog []*entity.Dialog) ([]*entity.Dialog, error) {
	var models []*po.Dialog

	for _, dialog := range dialog {
		model := converter.DialogEntityToPo(dialog)
		models = append(models, model)
	}

	if err := m.db.WithContext(ctx).
		Create(&models).
		Error; err != nil {
		return nil, err
	}

	var entities []*entity.Dialog
	for _, model := range models {
		entities = append(entities, converter.DialogEntityPoToEntity(model))
	}

	return entities, nil
}

func (m *MySQLDialogRepository) Get(ctx context.Context, id uint32) (*entity.Dialog, error) {
	model := &po.Dialog{}

	if err := m.db.WithContext(ctx).
		Where("id = ? AND deleted_at = 0", id).
		First(model).
		Error; err != nil {
		return nil, err
	}

	return converter.DialogEntityPoToEntity(model), nil
}

func (m *MySQLDialogRepository) Create(ctx context.Context, createDialog *repository.CreateDialog) (*entity.Dialog, error) {
	model := &po.Dialog{}

	model.Type = uint8(createDialog.Type)
	model.OwnerId = createDialog.OwnerId
	model.GroupId = createDialog.GroupId

	if err := m.db.WithContext(ctx).
		Create(&model).
		Error; err != nil {
		return nil, err
	}

	return converter.DialogEntityPoToEntity(model), nil
}

func (m *MySQLDialogRepository) Update(ctx context.Context, dialog *entity.Dialog) (*entity.Dialog, error) {
	model := converter.DialogEntityToPo(dialog)

	if err := m.db.WithContext(ctx).
		Updates(model).
		Error; err != nil {
		return nil, err
	}

	return converter.DialogEntityPoToEntity(model), nil
}

func (m *MySQLDialogRepository) Delete(ctx context.Context, id ...uint32) error {
	if err := m.db.WithContext(ctx).
		Model(&po.Dialog{}).
		Where("id IN (?)", id).
		Update("deleted_at", ptime.Now()).
		Error; err != nil {
		return err
	}

	return nil
}

func (m *MySQLDialogRepository) Find(ctx context.Context, query *repository.DialogQuery) ([]*entity.Dialog, error) {
	var models []*po.Dialog

	db := m.db.WithContext(ctx).Model(&po.Dialog{})

	if query.DialogID != nil && len(query.DialogID) > 0 {
		db = db.Where("id IN (?)", query.DialogID)
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

	if err := db.Debug().Where("deleted_at = 0").Find(&models).Error; err != nil {
		return nil, err
	}

	var dialogs []*entity.Dialog
	for _, model := range models {
		dialogs = append(dialogs, converter.DialogEntityPoToEntity(model))
	}

	return dialogs, nil
}

func (m *MySQLDialogRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return m.db.WithContext(ctx).
		Model(&po.Dialog{}).
		Where("id = ?", id).
		//Unscoped().
		Updates(fields).Error
}
