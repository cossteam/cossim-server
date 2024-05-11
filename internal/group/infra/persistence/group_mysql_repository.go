package persistence

import (
	"context"
	"github.com/cossim/coss-server/internal/group/cache"
	"github.com/cossim/coss-server/internal/group/domain/entity"
	"github.com/cossim/coss-server/internal/group/domain/repository"
	"github.com/cossim/coss-server/internal/group/infra/persistence/converter"
	"github.com/cossim/coss-server/internal/group/infra/persistence/po"
	"github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"log"
)

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
	return m.db.AutoMigrate(&po.Group{})
}

func (m *MySQLGroupRepository) UpdateFields(ctx context.Context, id uint32, fields map[string]interface{}) error {
	if err := m.db.WithContext(ctx).Model(&po.Group{}).Where("id = ?", id).Unscoped().Updates(fields).Error; err != nil {
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

	model := &po.Group{}
	if err := m.db.WithContext(ctx).First(&model, id).Error; err != nil {
		return nil, err
	}

	e := converter.UserPOToEntity(model)

	if m.cache != nil {
		if err := m.cache.SetGroup(ctx, e); err != nil {
			log.Println("Error caching group:", err)
		}
	}

	return e, nil
}

func (m *MySQLGroupRepository) Update(ctx context.Context, group *entity.Group, updateFn func(h *entity.Group) (*entity.Group, error)) error {
	model := converter.GroupEntityToPO(group)

	//// 查询数据库中原始数据
	//originalModel := &po.Group{}
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
		e := converter.UserPOToEntity(model)
		if err != nil {
			return err
		}
		if err := m.cache.SetGroup(ctx, e); err != nil {
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
	model := converter.GroupEntityToPO(group)

	if err := m.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}

	e := converter.UserPOToEntity(model)
	return createFn(e)
}

func (m *MySQLGroupRepository) Delete(ctx context.Context, id uint32) error {
	if err := m.db.WithContext(ctx).Delete(&po.Group{}, id).Error; err != nil {
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
	var models []*po.Group

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
		e := converter.UserPOToEntity(model)
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
