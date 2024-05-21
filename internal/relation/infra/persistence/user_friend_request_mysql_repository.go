package persistence

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/cache"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/internal/relation/infra/persistence/converter"
	"github.com/cossim/coss-server/internal/relation/infra/persistence/po"
	"github.com/cossim/coss-server/pkg/code"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

var _ repository.UserFriendRequestRepository = &MySQLUserFriendRequestRepository{}

func NewMySQLUserFriendRequestRepository(db *gorm.DB, cache cache.RelationUserCache) *MySQLUserFriendRequestRepository {
	return &MySQLUserFriendRequestRepository{
		db: db,
		//cache: cache,
	}
}

type MySQLUserFriendRequestRepository struct {
	db *gorm.DB
}

func (m *MySQLUserFriendRequestRepository) NewRepository(db *gorm.DB) repository.UserFriendRequestRepository {
	return &MySQLUserFriendRequestRepository{
		db: db,
	}
}

func (m *MySQLUserFriendRequestRepository) GetByUserIdAndFriendId(ctx context.Context, senderId, receiverId string) (*entity.UserFriendRequest, error) {
	model := &po.UserFriendRequest{}

	if err := m.db.WithContext(ctx).
		Where("sender_id = ? AND receiver_id = ? AND status = ? AND deleted_at = 0", senderId, receiverId, entity.Pending).
		First(model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.NotFound
		}
		return nil, err
	}

	return converter.UserFriendRequestPoToEntity(model), nil
}

func (m *MySQLUserFriendRequestRepository) UpdateStatus(ctx context.Context, id uint32, status entity.RequestStatus) error {
	if err := m.db.Model(&po.UserFriendRequest{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return err
	}

	// TODO cache

	return nil
}

func (m *MySQLUserFriendRequestRepository) Get(ctx context.Context, id uint32) (*entity.UserFriendRequest, error) {
	// TODO cache

	model := &po.UserFriendRequest{}

	if err := m.db.WithContext(ctx).Where("id = ? AND deleted_at = 0", id).First(model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.NotFound
		}
		return nil, err
	}

	return converter.UserFriendRequestPoToEntity(model), nil
}

func (m *MySQLUserFriendRequestRepository) Create(ctx context.Context, entity *entity.UserFriendRequest) (*entity.UserFriendRequest, error) {
	model := converter.UserFriendRequestEntityToPo(entity)

	if err := m.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}

	return converter.UserFriendRequestPoToEntity(model), nil
}

func (m *MySQLUserFriendRequestRepository) Find(ctx context.Context, query *repository.UserFriendRequestQuery) (*entity.UserFriendRequestList, error) {
	var userFriendRequests []*po.UserFriendRequest

	db := m.db
	if query.UserID != "" {
		db = db.Where("owner_id = ?", query.UserID)
	}
	if query.SenderId != "" {
		db = db.Where("sender_id = ?", query.SenderId)
	}
	if query.ReceiverId != "" {
		db = db.Where("receiver_id = ?", query.ReceiverId)
	}
	if query.PageSize > 0 && query.PageNum > 0 {
		offset := (query.PageNum - 1) * query.PageSize
		db = db.Offset(offset).Limit(query.PageSize)
	}
	if query.Status != nil {
		db = db.Where("status = ?", query.Status)
	}
	if !query.Force {
		db = db.Where("deleted_at = 0")
	}

	result := db.Find(&userFriendRequests)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "failed to find user friend requests")
	}

	var totalCount int64
	if err := db.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	ufrs := &entity.UserFriendRequestList{
		Total: totalCount,
	}
	for _, model := range userFriendRequests {
		ufrs.List = append(ufrs.List, converter.UserFriendRequestPoToEntity(model))
	}

	return ufrs, nil
}

func (m *MySQLUserFriendRequestRepository) Delete(ctx context.Context, id uint32) error {
	if err := m.db.WithContext(ctx).
		Model(&po.UserFriendRequest{}).
		Where("id = ?", id).
		Update("deleted_at", ptime.Now()).Error; err != nil {
		return err
	}

	// TODO cache

	return nil
}

func (m *MySQLUserFriendRequestRepository) UpdateFields(ctx context.Context, id uint32, fields map[string]interface{}) error {
	if err := m.db.WithContext(ctx).
		Model(&po.UserFriendRequest{}).
		Where("id = ?", id).
		Unscoped().
		Updates(fields).Error; err != nil {
		return err
	}

	// TODO cache

	return nil
}
