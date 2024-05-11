package adapters

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/group/cache"
	"github.com/cossim/coss-server/internal/group/domain/entity"
	"github.com/cossim/coss-server/internal/group/domain/repository"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"log"
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
	SilenceTime     int64  `gorm:"comment:全员禁言结束时间"`
	JoinApprove     bool   `gorm:"default:false;comment:是否开启入群验证"`
	Encrypt         bool   `gorm:"default:false;comment:是否开启群聊加密，只有当群聊类型为私密群时，该字段才有效"`
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

func (m *GroupModel) FromEntity(e *entity.Group) error {
	if m == nil {
		m = &GroupModel{}
	}
	if err := e.Validate(); err != nil {
		return err
	}
	m.ID = e.ID
	m.Type = uint(e.Type)
	m.Status = uint(e.Status)
	m.MaxMembersLimit = e.MaxMembersLimit
	m.CreatorID = e.CreatorID
	m.Name = e.Name
	m.Avatar = e.Avatar
	m.SilenceTime = e.SilenceTime
	m.JoinApprove = e.JoinApprove
	m.Encrypt = e.Encrypt
	return nil
}

func (m *GroupModel) ToEntity() (*entity.Group, error) {
	if m == nil {
		return nil, errors.New("group model is nil")
	}
	return &entity.Group{
		ID:              m.ID,
		CreatedAt:       m.CreatedAt,
		Type:            entity.Type(m.Type),
		Status:          entity.Status(m.Status),
		MaxMembersLimit: m.MaxMembersLimit,
		CreatorID:       m.CreatorID,
		Name:            m.Name,
		Avatar:          m.Avatar,
		SilenceTime:     m.SilenceTime,
		JoinApprove:     m.JoinApprove,
		Encrypt:         m.Encrypt,
	}, nil
}

const mySQLDeadlockErrorCode = 1213

var _ repository.Repository = &MySQLGroupRepository{}

func NewMySQLGroupRepository(db *gorm.DB, cache cache.GroupCache) *MySQLGroupRepository {
	return &MySQLGroupRepository{
		db:    db,
		cache: cache,
	}
}

type MySQLGroupRepository struct {
	db    *gorm.DB
	cache cache.GroupCache
}

func (m *MySQLGroupRepository) Automigrate() error {
	return m.db.AutoMigrate(&GroupModel{})
}

func (m *MySQLGroupRepository) UpdateFields(ctx context.Context, id uint32, fields map[string]interface{}) error {
	if err := m.db.WithContext(ctx).Model(&GroupModel{}).Where("id = ?", id).Unscoped().Updates(fields).Error; err != nil {
		return err
	}

	if m.cache != nil {
		if err := m.cache.DeleteGroup(ctx, id); err != nil {
			log.Println("Error deleting group from cache:", err)
		}
	}

	return nil
}

func (m *MySQLGroupRepository) Get(ctx context.Context, id uint32) (*entity.Group, error) {
	if m.cache != nil {
		cachedGroup, err := m.cache.GetGroup(ctx, id)
		if err == nil && cachedGroup != nil {
			return cachedGroup, nil
		}
	}

	model := &GroupModel{}
	if err := m.db.WithContext(ctx).First(&model, id).Error; err != nil {
		return nil, err
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, err
	}

	if m.cache != nil {
		if err := m.cache.SetGroup(ctx, entity); err != nil {
			log.Println("Error caching group:", err)
		}
	}

	return model.ToEntity()
}

func (m *MySQLGroupRepository) Update(ctx context.Context, group *entity.Group, updateFn func(h *entity.Group) (*entity.Group, error)) error {
	model := &GroupModel{}
	if err := model.FromEntity(group); err != nil {
		return err
	}

	//// 查询数据库中原始数据
	//originalModel := &GroupModel{}
	//if err := m.db.WithContext(ctx).First(originalModel, "id = ?", group.ID).Error; err != nil {
	//	return err
	//}
	//
	//// 获取更新后的字段列表
	//updatedFields := make(map[string]interfaces{})
	//updatedFieldsValue := reflect.ValueOf(model).Elem()
	//for i := 0; i < updatedFieldsValue.NumField(); i++ {
	//	fieldName := updatedFieldsValue.Type().Field(i).Name
	//	fieldValue := updatedFieldsValue.Field(i).Interface()
	//	updatedFields[fieldName] = fieldValue
	//}
	//
	//// 遍历对比字段值并更新
	//for fieldName, updatedValue := range updatedFields {
	//	originalValue := reflect.ValueOf(originalModel).Elem().FieldByName(fieldName).Interface()
	//	if reflect.DeepEqual(updatedValue, originalValue) {
	//		// 如果更新后的值与原始值相同，则使用原始数据的值
	//		reflect.ValueOf(model).Elem().FieldByName(fieldName).Set(reflect.ValueOf(originalValue))
	//	} else {
	//		// 否则使用更新后的值
	//		reflect.ValueOf(model).Elem().FieldByName(fieldName).Set(reflect.ValueOf(updatedValue))
	//	}
	//}

	// 再次保存更新后的数据
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

	if m.cache != nil {
		entity, err := model.ToEntity()
		if err != nil {
			return err
		}
		if err := m.cache.SetGroup(ctx, entity); err != nil {
			log.Println("Error caching group:", err)
		}
	}

	return nil
}

func (m *MySQLGroupRepository) Create(ctx context.Context, group *entity.Group, createFn func(h *entity.Group) (*entity.Group, error)) error {
	for {
		_, err := m.create(ctx, group, createFn)

		if val, ok := errors.Cause(err).(*mysql.MySQLError); ok && val.Number == mySQLDeadlockErrorCode {
			continue
		}

		return err
	}
}

func (m *MySQLGroupRepository) create(ctx context.Context, group *entity.Group, createFn func(h *entity.Group) (*entity.Group, error)) (*entity.Group, error) {
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
	if err := m.db.WithContext(ctx).Delete(&GroupModel{}, id).Error; err != nil {
		return err
	}

	if m.cache != nil {
		if err := m.cache.DeleteGroup(ctx, id); err != nil {
			log.Println("Error caching group:", err)
		}
	}

	return nil
}

// Find retrieves groups based on the provided query.
func (m *MySQLGroupRepository) Find(ctx context.Context, query repository.Query) ([]*entity.Group, error) {
	if query.Cache && m.cache != nil {
		cacheKey := m.generateCacheKey(query)
		cachedGroups, err := m.cache.GetGroups(ctx, cacheKey)
		if err == nil && len(cachedGroups) > 0 {
			return cachedGroups, nil
		}
	}

	groups, err := m.findWithoutCache(ctx, query)
	if err != nil {
		return nil, err
	}

	if err := m.cache.SetGroup(ctx, groups...); err != nil {
		log.Println("Error caching groups:", err)
	}

	return groups, nil
}

// findWithoutCache executes the query directly without using cache.
func (m *MySQLGroupRepository) findWithoutCache(ctx context.Context, query repository.Query) ([]*entity.Group, error) {
	// Perform query directly on the database
	var models []*GroupModel

	db := m.db.WithContext(ctx)

	// Apply filters
	if len(query.ID) > 0 {
		db = db.Or("id IN (?)", query.ID)
	}
	if query.Name != "" {
		db = db.Or("name LIKE ?", "%"+query.Name+"%")
	}

	if len(query.UserID) > 0 {
		db = db.Joins("JOIN group_members ON groups.id = group_members.group_id").
			Where("group_members.user_id IN (?)", query.UserID)
	}
	if query.CreateAt != nil {
		db = db.Where("created_at >= ?", query.CreateAt)
	}
	if query.UpdateAt != nil {
		db = db.Where("updated_at >= ?", query.UpdateAt)
	}
	if query.Limit > 0 {
		db = db.Limit(query.Limit)
	}
	if query.Offset > 0 {
		db = db.Offset(query.Offset)
	}

	if err := db.Find(&models).Error; err != nil {
		return nil, err
	}

	// Convert models to entities
	groups := make([]*entity.Group, len(models))
	for i, model := range models {
		e, err := model.ToEntity()
		if err != nil {
			return nil, err
		}
		groups[i] = e
	}

	return groups, nil
}

// generateCacheKey generates a cache key based on the query parameters.
func (m *MySQLGroupRepository) generateCacheKey(query repository.Query) []uint32 {
	// Example: concatenate query parameters to form a cache key
	// This is a simplistic approach; you might need to customize it based on your specific requirements
	cacheKey := make([]uint32, 0, len(query.ID))
	cacheKey = append(cacheKey, query.ID...)
	return cacheKey
}
