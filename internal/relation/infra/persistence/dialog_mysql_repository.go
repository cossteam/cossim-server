package persistence

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/cache"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

type DialogModel struct {
	BaseModel
	OwnerId string `gorm:"comment:用户id"`
	Type    uint8  `gorm:"comment:对话类型"`
	GroupId uint32 `gorm:"comment:群组id"`
}

func (m *DialogModel) FromEntity(e *entity.Dialog) error {
	m.ID = e.ID
	m.OwnerId = e.OwnerId
	m.Type = uint8(e.Type)
	m.GroupId = e.GroupId
	return nil
}

func (m *DialogModel) ToEntity() *entity.Dialog {
	return &entity.Dialog{
		ID:        m.ID,
		CreatedAt: m.CreatedAt,
		OwnerId:   m.OwnerId,
		Type:      entity.DialogType(m.Type),
		GroupId:   m.GroupId,
	}
}

func (m *DialogModel) TableName() string {
	return "dialogs"
}

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
	var model DialogModel

	if err := m.db.WithContext(ctx).
		Where("group_id = ? AND deleted_at = 0", groupID).
		First(&model).
		Error; err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (m *MySQLDialogRepository) Creates(ctx context.Context, dialog []*entity.Dialog) ([]*entity.Dialog, error) {
	var models []*DialogModel

	for _, dialog := range dialog {
		var model DialogModel
		if err := model.FromEntity(dialog); err != nil {
			return nil, err
		}
		models = append(models, &model)
	}

	if err := m.db.WithContext(ctx).
		Create(&models).
		Error; err != nil {
		return nil, err
	}

	var entities []*entity.Dialog
	for _, model := range models {
		entities = append(entities, model.ToEntity())
	}

	return entities, nil
}

func (m *MySQLDialogRepository) Get(ctx context.Context, id uint32) (*entity.Dialog, error) {
	var model DialogModel

	if err := m.db.WithContext(ctx).
		Where("id = ? AND deleted_at = 0", id).
		First(&model).
		Error; err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (m *MySQLDialogRepository) Create(ctx context.Context, createDialog *repository.CreateDialog) (*entity.Dialog, error) {
	var model DialogModel

	model.Type = uint8(createDialog.Type)
	model.OwnerId = createDialog.OwnerId
	model.GroupId = createDialog.GroupId

	if err := m.db.WithContext(ctx).
		Create(&model).
		Error; err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (m *MySQLDialogRepository) Update(ctx context.Context, dialog *entity.Dialog) (*entity.Dialog, error) {
	var model DialogModel

	if err := model.FromEntity(dialog); err != nil {
		return nil, err
	}

	if err := m.db.WithContext(ctx).
		Updates(model).
		Error; err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (m *MySQLDialogRepository) Delete(ctx context.Context, id ...uint32) error {
	if err := m.db.WithContext(ctx).
		Model(&DialogModel{}).
		Where("id IN (?)", id).
		Update("deleted_at", ptime.Now()).
		Error; err != nil {
		return err
	}

	return nil
}

func (m *MySQLDialogRepository) Find(ctx context.Context, query *repository.DialogQuery) ([]*entity.Dialog, error) {
	var models []DialogModel

	db := m.db.Model(&DialogModel{})

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
		dialogs = append(dialogs, model.ToEntity())
	}

	return dialogs, nil
}

func (m *MySQLDialogRepository) UpdateFields(ctx context.Context, id uint, fields map[string]interface{}) error {
	return m.db.WithContext(ctx).
		Model(&DialogModel{}).
		Where("id = ?", id).
		//Unscoped().
		Updates(fields).Error
}
