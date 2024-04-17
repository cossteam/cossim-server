package adapters

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/group/domain/group"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint32 `gorm:"primaryKey;autoIncrement;"`
	CreatedAt int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt int64  `gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt int64  `gorm:"default:0;comment:删除时间"`
}

type GroupModel struct {
	BaseModel
	Type            uint   `gorm:"default:0;comment:群聊类型(0=私密群, 1=公开群)"`
	Status          uint   `gorm:"comment:群聊状态(0=正常状态, 1=锁定状态, 2=删除状态)"`
	MaxMembersLimit int    `gorm:"comment:群聊人数限制"`
	CreatorID       string `gorm:"type:varchar(64);comment:创建者id"`
	Name            string `gorm:"comment:群聊名称"`
	Avatar          string `gorm:"default:'';comment:头像（群）"`
}

func (bm *BaseModel) TableName() string {
	fmt.Println("table name")
	return "groups"
}

func (bm *BaseModel) BeforeCreate(tx *gorm.DB) error {
	now := ptime.Now()
	bm.CreatedAt = now
	bm.UpdatedAt = now
	return nil
}

func (bm *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	bm.UpdatedAt = ptime.Now()
	return nil
}

func (m *GroupModel) FromEntity(e *group.Group) error {
	if m == nil {
		m = &GroupModel{}
	}
	if err := e.Validate(); err != nil {
		return err
	}
	m.ID = e.ID
	m.CreatedAt = e.CreatedAt
	m.Type = uint(e.Type)
	m.Status = uint(e.Status)
	m.MaxMembersLimit = e.MaxMembersLimit
	m.CreatorID = e.CreatorID
	m.Name = e.Name
	m.Avatar = e.Avatar
	return nil
}

func (m *GroupModel) ToEntity() (*group.Group, error) {
	if m == nil {
		return nil, errors.New("group model is nil")
	}
	return &group.Group{
		ID:              m.ID,
		CreatedAt:       m.CreatedAt,
		Type:            group.Type(m.Type),
		Status:          group.Status(m.Status),
		MaxMembersLimit: m.MaxMembersLimit,
		CreatorID:       m.CreatorID,
		Name:            m.Name,
		Avatar:          m.Avatar,
	}, nil
}

var _ group.Repository = &MySQLGroupRepository{}

func NewMySQLGroupRepository(db *gorm.DB) *MySQLGroupRepository {
	return &MySQLGroupRepository{
		db: db,
	}
}

type MySQLGroupRepository struct {
	db *gorm.DB

	// groupFactory
}

func (m *MySQLGroupRepository) Automigrate() error {
	return m.db.AutoMigrate(&GroupModel{})
}

func (m *MySQLGroupRepository) UpdateFields(ctx context.Context, id uint32, fields map[string]interface{}) error {
	return m.db.WithContext(ctx).Where("id = ?", id).Unscoped().Updates(fields).Error
}

func (m *MySQLGroupRepository) Get(ctx context.Context, id uint32) (*group.Group, error) {
	model := &GroupModel{}

	if err := m.db.WithContext(ctx).First(&model, id).Error; err != nil {
		return nil, err
	}

	return model.ToEntity()
}

func (m *MySQLGroupRepository) Update(ctx context.Context, group *group.Group, updateFn func(h *group.Group) (*group.Group, error)) error {
	model := &GroupModel{}
	if err := model.FromEntity(group); err != nil {
		return err
	}
	if err := m.db.WithContext(ctx).Save(model).Error; err != nil {
		return err
	}

	if updateFn == nil {
		return nil
	}

	_, err := updateFn(group)
	if err != nil {
		return err
	}

	return nil
}

const mySQLDeadlockErrorCode = 1213

func (m *MySQLGroupRepository) Create(ctx context.Context, group *group.Group, createFn func(h *group.Group) (*group.Group, error)) error {
	for {
		_, err := m.create(ctx, group, createFn)

		if val, ok := errors.Cause(err).(*mysql.MySQLError); ok && val.Number == mySQLDeadlockErrorCode {
			continue
		}

		return err
	}
}

func (m *MySQLGroupRepository) create(ctx context.Context, group *group.Group, createFn func(h *group.Group) (*group.Group, error)) (*group.Group, error) {
	model := &GroupModel{}
	if err := model.FromEntity(group); err != nil {
		return nil, err
	}

	if err := m.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, err
	}

	return createFn(entity)
}

func (m *MySQLGroupRepository) Delete(ctx context.Context, id uint32) error {
	err := m.db.WithContext(ctx).Delete(&GroupModel{}, id).Error
	return err
}

func (m *MySQLGroupRepository) Find(ctx context.Context, query group.Query) ([]*group.Group, error) {
	var groups []*group.Group

	var models []*GroupModel

	db := m.db.WithContext(ctx)

	// 根据 ID 列表进行筛选
	if len(query.ID) > 0 {
		db = db.Where("id IN (?)", query.ID)
	}

	// 根据名称进行筛选
	if query.Name != "" {
		db = db.Where("name LIKE ?", "%"+query.Name+"%")
	}

	// 根据用户 ID 列表进行筛选
	if len(query.UserID) > 0 {
		db = db.Joins("JOIN group_members ON groups.id = group_members.group_id").
			Where("group_members.user_id IN (?)", query.UserID)
	}

	// 根据创建时间范围进行筛选
	if query.CreateAt != nil {
		db = db.Where("created_at >= ?", query.CreateAt)
	}

	// 根据更新时间范围进行筛选
	if query.UpdateAt != nil {
		db = db.Where("updated_at >= ?", query.UpdateAt)
	}

	// 设置分页参数
	if query.Limit > 0 {
		db = db.Limit(query.Limit)
	}
	if query.Offset > 0 {
		db = db.Offset(query.Offset)
	}

	// 执行查询
	if err := db.Find(&models).Error; err != nil {
		return nil, err
	}

	for _, model := range models {
		entity, err := model.ToEntity()
		if err != nil {
			return nil, err
		}
		groups = append(groups, entity)
	}

	return groups, nil
}
