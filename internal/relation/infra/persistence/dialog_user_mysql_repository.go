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

var _ repository.DialogUserRepository = &MySQLDialogUserRepository{}

func NewMySQLDialogUserRepository(db *gorm.DB, cache cache.RelationUserCache) *MySQLDialogUserRepository {
	return &MySQLDialogUserRepository{
		db: db,
		//cache: cache,
	}
}

type MySQLDialogUserRepository struct {
	db *gorm.DB
}

func (m *MySQLDialogUserRepository) GetByDialogIDAndUserID(ctx context.Context, dialogID uint32, userID string) (*entity.DialogUser, error) {
	model := &po.DialogUser{}

	if err := m.db.WithContext(ctx).
		Where(&po.DialogUser{DialogId: dialogID, UserId: userID}).
		First(model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.NotFound
		}
	}

	return converter.DialogUserPoToEntity(model), nil
}

func (m *MySQLDialogUserRepository) UpdateDialogStatus(ctx context.Context, param *repository.UpdateDialogStatusParam) error {
	fields := map[string]interface{}{}
	db := m.db.WithContext(ctx)
	if param.UserID != nil && len(param.UserID) > 0 {
		db = db.Where("user_id IN (?)", param.UserID)
	}
	if param.IsShow != nil {
		fields["is_show"] = *param.IsShow
	}
	if param.TopAt != nil {
		fields["top_at"] = *param.TopAt
	}
	if param.DeletedAt != nil {
		fields["deleted_at"] = *param.DeletedAt
	}
	if len(fields) == 0 {
		return nil
	}
	return m.db.WithContext(ctx).
		Model(&po.DialogUser{}).
		Where("dialog_id = ?", param.DialogID).
		Updates(fields).
		Error
}

func (m *MySQLDialogUserRepository) Creates(ctx context.Context, dialogID uint32, userIDs []string) ([]*entity.DialogUser, error) {
	var models []*po.DialogUser

	for _, userID := range userIDs {
		models = append(models, &po.DialogUser{
			DialogId: dialogID,
			UserId:   userID,
			IsShow:   true,
		})
	}

	if err := m.db.WithContext(ctx).
		Create(&models).
		Error; err != nil {
		return nil, err
	}

	var dialogUsers []*entity.DialogUser
	for _, model := range models {
		dialogUsers = append(dialogUsers, converter.DialogUserPoToEntity(model))
	}

	return dialogUsers, nil
}

func (m *MySQLDialogUserRepository) Get(ctx context.Context, id uint32) (*entity.DialogUser, error) {
	model := &po.DialogUser{}

	if err := m.db.WithContext(ctx).
		Where("id = ?", id).
		First(model).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.NotFound
		}
		return nil, err
	}

	return converter.DialogUserPoToEntity(model), nil
}

func (m *MySQLDialogUserRepository) Create(ctx context.Context, createDialogUser *repository.CreateDialogUser) (*entity.DialogUser, error) {
	model := &po.DialogUser{
		DialogId: createDialogUser.DialogID,
		UserId:   createDialogUser.UserID,
		IsShow:   true,
	}

	if err := m.db.WithContext(ctx).
		Create(model).
		Error; err != nil {
		return nil, err
	}

	return converter.DialogUserPoToEntity(model), nil
}

func (m *MySQLDialogUserRepository) Update(ctx context.Context, dialog *entity.DialogUser) (*entity.DialogUser, error) {
	model := &po.DialogUser{}

	if err := m.db.WithContext(ctx).
		Where("id = ?", dialog.ID).
		Updates(model).
		Error; err != nil {
		return nil, err
	}

	return converter.DialogUserPoToEntity(model), nil
}

func (m *MySQLDialogUserRepository) Delete(ctx context.Context, id ...uint32) error {
	return m.db.WithContext(ctx).
		Model(&po.DialogUser{}).
		Where("id in (?)", id).
		Update("deleted_at", ptime.Now()).
		Error
}

func (m *MySQLDialogUserRepository) Find(ctx context.Context, query *repository.DialogUserQuery) ([]*entity.DialogUser, error) {
	if query == nil || (query.DialogID == nil && query.UserID == nil) {
		return nil, code.InvalidParameter
	}

	var models []*po.DialogUser

	db := m.db.WithContext(ctx).Model(&po.DialogUser{})

	if query.DialogID != nil && len(query.DialogID) > 0 {
		db = db.Where("dialog_id IN (?)", query.DialogID)
	}

	if query.UserID != nil && len(query.UserID) > 0 {
		db = db.Where("user_id IN (?)", query.UserID)
	}

	if !query.Force {
		db = db.Where("deleted_at = 0")
	}

	if query.IsShow {
		db = db.Where(&po.DialogUser{IsShow: true})
	}

	if query.PageSize > 0 && query.PageNum > 0 {
		offset := (query.PageNum - 1) * query.PageSize
		db = db.Offset(offset).Limit(query.PageSize)
	}

	if err := db.Debug().Find(&models).Error; err != nil {
		return nil, err
	}

	var dialogs []*entity.DialogUser
	for _, model := range models {
		dialogs = append(dialogs, converter.DialogUserPoToEntity(model))
	}

	return dialogs, nil
}

func (m *MySQLDialogUserRepository) ListByDialogID(ctx context.Context, dialogID uint32) ([]*entity.DialogUser, error) {
	var models []*po.DialogUser
	if err := m.db.
		Model(&po.DialogUser{}).
		Where("dialog_id =?", dialogID).
		Find(&models).Error; err != nil {
		return nil, err
	}

	var dialogUsers = make([]*entity.DialogUser, 0)
	for _, model := range models {
		dialogUsers = append(dialogUsers, converter.DialogUserPoToEntity(model))
	}

	return dialogUsers, nil
}

func (m *MySQLDialogUserRepository) DeleteByDialogID(ctx context.Context, dialogID uint32) error {
	if err := m.db.WithContext(ctx).
		Model(&po.DialogUser{}).
		Where("dialog_id = ?", dialogID).
		Unscoped().
		Update("deleted_at", ptime.Now()).
		Error; err != nil {
		return err
	}

	return nil
}

func (m *MySQLDialogUserRepository) DeleteByDialogIDAndUserID(ctx context.Context, dialogID uint32, userID ...string) error {
	return m.db.WithContext(ctx).
		Model(&po.DialogUser{}).
		Where("dialog_id = ? AND user_id IN (?)", dialogID, userID).
		Unscoped().
		Update("deleted_at", ptime.Now()).
		Error
}

func (m *MySQLDialogUserRepository) UpdateFields(ctx context.Context, id uint32, updateFields map[string]interface{}) error {
	return m.db.WithContext(ctx).
		Model(&po.DialogUser{}).
		Where("dialog_id = ?", id).
		Updates(updateFields).
		Error
}
