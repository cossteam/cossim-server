package persistence

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/cache"
	"github.com/cossim/coss-server/internal/relation/domain/relation"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

type DialogUserModel struct {
	BaseModel
	DialogId uint32 `gorm:"default:0;comment:对话ID"`
	UserId   string `gorm:"default:0;comment:会员ID"`
	IsShow   bool   `gorm:"default:false;comment:对话是否显示"`
	TopAt    int64  `gorm:"comment:置顶时间"`
}

func (m *DialogUserModel) TableName() string {
	return "dialog_users"
}

func (m *DialogUserModel) FromEntity(e *relation.DialogUser) error {
	m.ID = e.ID
	m.DialogId = e.DialogId
	m.UserId = e.UserId
	m.IsShow = e.IsShow
	m.TopAt = e.TopAt
	return nil
}

func (m *DialogUserModel) ToEntity() *relation.DialogUser {
	return &relation.DialogUser{
		ID:        m.ID,
		CreatedAt: m.CreatedAt,
		DialogId:  m.DialogId,
		UserId:    m.UserId,
		IsShow:    m.IsShow,
		TopAt:     m.TopAt,
	}
}

var _ relation.DialogUserRepository = &MySQLDialogUserRepository{}

func NewMySQLDialogUserRepository(db *gorm.DB, cache cache.RelationUserCache) *MySQLDialogUserRepository {
	return &MySQLDialogUserRepository{
		db: db,
		//cache: cache,
	}
}

type MySQLDialogUserRepository struct {
	db *gorm.DB
}

func (m *MySQLDialogUserRepository) UpdateDialogStatus(ctx context.Context, param *relation.UpdateDialogStatusParam) error {
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
	return m.db.Model(&DialogUserModel{}).
		Where("dialog_id = ?", param.DialogID).
		Updates(fields).
		Error
}

func (m *MySQLDialogUserRepository) Creates(ctx context.Context, dialogID uint32, userIDs []string) ([]*relation.DialogUser, error) {
	var models []DialogUserModel

	for _, userID := range userIDs {
		models = append(models, DialogUserModel{
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

	var dialogUsers []*relation.DialogUser
	for _, model := range models {
		dialogUsers = append(dialogUsers, model.ToEntity())
	}

	return dialogUsers, nil
}

func (m *MySQLDialogUserRepository) Get(ctx context.Context, id uint32) (*relation.DialogUser, error) {
	var model DialogUserModel

	if err := m.db.WithContext(ctx).
		Where("id = ?", id).
		First(&model).
		Error; err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (m *MySQLDialogUserRepository) Create(ctx context.Context, createDialogUser *relation.CreateDialogUser) (*relation.DialogUser, error) {
	model := DialogUserModel{
		DialogId: createDialogUser.DialogID,
		UserId:   createDialogUser.UserID,
		IsShow:   true,
	}

	if err := m.db.WithContext(ctx).
		Create(&model).
		Error; err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (m *MySQLDialogUserRepository) Update(ctx context.Context, dialog *relation.DialogUser) (*relation.DialogUser, error) {
	var model DialogUserModel

	if err := m.db.WithContext(ctx).
		Where("id = ?", dialog.ID).
		Updates(&model).
		Error; err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (m *MySQLDialogUserRepository) Delete(ctx context.Context, id ...uint32) error {
	return m.db.WithContext(ctx).
		Model(&DialogUserModel{}).
		Where("id in (?)", id).
		Update("deleted_at", ptime.Now()).
		Error
}

func (m *MySQLDialogUserRepository) Find(ctx context.Context, query *relation.DialogUserQuery) ([]*relation.DialogUser, error) {
	var models []DialogUserModel

	db := m.db.Model(&DialogUserModel{})

	if query.DialogID != nil && len(query.DialogID) > 0 {
		db = db.Where("dialog_id IN (?)", query.DialogID)
	}

	if query.UserID != nil && len(query.UserID) > 0 {
		db = db.Where("user_id IN (?)", query.UserID)
	}

	if !query.Force {
		db = db.Where("deleted_at = 0")
	}

	if query.PageSize > 0 && query.PageNum > 0 {
		offset := (query.PageNum - 1) * query.PageSize
		db = db.Offset(offset).Limit(query.PageSize)
	}

	if err := db.Debug().Find(&models).Error; err != nil {
		return nil, err
	}

	var dialogs []*relation.DialogUser
	for _, model := range models {
		dialogs = append(dialogs, model.ToEntity())
	}

	return dialogs, nil
}

func (m *MySQLDialogUserRepository) ListByDialogID(ctx context.Context, dialogID uint32) ([]*relation.DialogUser, error) {
	var models []DialogUserModel
	if err := m.db.
		Model(&DialogUserModel{}).
		Where("dialog_id =?", dialogID).
		Find(&models).Error; err != nil {
		return nil, err
	}

	var dialogUsers = make([]*relation.DialogUser, 0)
	for _, model := range models {
		dialogUsers = append(dialogUsers, model.ToEntity())
	}

	return dialogUsers, nil
}

func (m *MySQLDialogUserRepository) DeleteByDialogID(ctx context.Context, dialogID uint32) error {
	if err := m.db.WithContext(ctx).
		Model(&DialogUserModel{}).
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
		Model(&DialogUserModel{}).
		Where("dialog_id = ? AND user_id IN (?)", dialogID, userID).
		Unscoped().
		Update("deleted_at", ptime.Now()).
		Error
}

func (m *MySQLDialogUserRepository) UpdateFields(ctx context.Context, id uint32, updateFields map[string]interface{}) error {
	return m.db.WithContext(ctx).
		Model(&DialogUserModel{}).
		Where("dialog_id = ?", id).
		Updates(updateFields).
		Error
}
