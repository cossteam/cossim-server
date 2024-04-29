package persistence

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/cache"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

type UserFriendRequestModel struct {
	BaseModel
	SenderID   string `gorm:"comment:发送者id"`
	ReceiverID string `gorm:"comment:接收者id"`
	Remark     string `gorm:"comment:添加备注"`
	OwnerID    string `gorm:"comment:所有者id"`
	Status     uint   `gorm:"comment:申请记录状态 (0=申请中 1=已同意 2=已拒绝)"`
}

func (m *UserFriendRequestModel) TableName() string {
	return "user_friend_requests"
}

func (m *UserFriendRequestModel) FromEntity(u *entity.UserFriendRequest) error {
	m.ID = u.ID
	m.SenderID = u.SenderID
	m.ReceiverID = u.ReceiverID
	m.Remark = u.Remark
	m.OwnerID = u.OwnerID
	m.Status = uint(u.Status)
	return nil
}

func (m *UserFriendRequestModel) ToEntity() (*entity.UserFriendRequest, error) {
	return &entity.UserFriendRequest{
		ID:         uint32(m.ID),
		CreatedAt:  m.CreatedAt,
		SenderID:   m.SenderID,
		ReceiverID: m.ReceiverID,
		Remark:     m.Remark,
		OwnerID:    m.OwnerID,
		Status:     entity.RequestStatus(m.Status),
	}, nil
}

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

func (m *MySQLUserFriendRequestRepository) GetByUserIdAndFriendId(ctx context.Context, senderId, receiverId string) (*entity.UserFriendRequest, error) {
	var model UserFriendRequestModel

	if err := m.db.WithContext(ctx).
		Where("sender_id = ? AND receiver_id = ? AND status = ? AND deleted_at = 0", senderId, receiverId, entity.Pending).
		First(&model).Error; err != nil {
		return nil, err
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (m *MySQLUserFriendRequestRepository) UpdateStatus(ctx context.Context, id uint32, status entity.RequestStatus) error {
	if err := m.db.Model(&UserFriendRequestModel{}).Where("id = ?", id).Update("status", status).Error; err != nil {
		return err
	}

	// TODO cache

	return nil
}

func (m *MySQLUserFriendRequestRepository) Get(ctx context.Context, id uint32) (*entity.UserFriendRequest, error) {
	// TODO cache

	var model UserFriendRequestModel

	if err := m.db.WithContext(ctx).Where("id = ? AND deleted_at = 0", id).First(&model).Error; err != nil {
		return nil, err
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (m *MySQLUserFriendRequestRepository) Create(ctx context.Context, entity *entity.UserFriendRequest) (*entity.UserFriendRequest, error) {
	var model UserFriendRequestModel

	if err := model.FromEntity(entity); err != nil {
		return nil, err
	}

	if err := m.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (m *MySQLUserFriendRequestRepository) Find(ctx context.Context, query *repository.UserFriendRequestQuery) (*entity.UserFriendRequestList, error) {
	var userFriendRequests []*UserFriendRequestModel

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
	if query.Force {
		db = db.Where("deleted_at != 0")
	} else {
		db = db.Where("deleted_at = 0")
	}

	result := db.Find(&userFriendRequests)
	if result.Error != nil {
		return nil, errors.Wrap(result.Error, "failed to find user friend requests")
	}

	ufrs := &entity.UserFriendRequestList{
		Total: result.RowsAffected,
	}
	for _, model := range userFriendRequests {
		entity, err := model.ToEntity()
		if err != nil {
			return nil, err
		}
		ufrs.List = append(ufrs.List, entity)
	}

	return ufrs, nil
}

func (m *MySQLUserFriendRequestRepository) Delete(ctx context.Context, id uint32) error {
	if err := m.db.Model(&UserFriendRequestModel{}).
		Where("id = ?", id).
		Update("deleted_at", ptime.Now()).Error; err != nil {
		return err
	}

	// TODO cache

	return nil
}

func (m *MySQLUserFriendRequestRepository) UpdateFields(ctx context.Context, id uint32, fields map[string]interface{}) error {
	if err := m.db.Model(&UserFriendRequestModel{}).WithContext(ctx).
		Where("id = ?", id).
		Unscoped().
		Updates(fields).Error; err != nil {
		return err
	}

	// TODO cache

	return nil
}
