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

type GroupJoinRequestModel struct {
	BaseModel
	GroupID   uint32 `gorm:"comment:群聊id"`
	InviterAt int64  `gorm:"comment:邀请时间"`
	ExpiredAt int64  `gorm:"comment:过期时间"`
	UserID    string `gorm:"comment:被邀请人id"`
	Inviter   string `gorm:"comment:邀请人ID"`
	Remark    string `gorm:"comment:邀请备注"`
	OwnerID   string `gorm:"comment:所有者id"`
	Status    uint8  `gorm:"comment:申请记录状态 (0=待处理 1=已接受 2=已拒绝 3=邀请)"`
}

func (m *GroupJoinRequestModel) TableName() string {
	return "group_join_requests"
}

func (m *GroupJoinRequestModel) FromEntity(e *entity.GroupJoinRequest) {
	m.ID = e.ID
	m.GroupID = e.GroupID
	m.InviterAt = e.InviterAt
	m.ExpiredAt = e.ExpiredAt
	m.UserID = e.UserID
	m.Inviter = e.Inviter
	m.Remark = e.Remark
	m.OwnerID = e.OwnerID
	m.Status = uint8(e.Status)
}

func (m *GroupJoinRequestModel) ToEntity() *entity.GroupJoinRequest {
	e := &entity.GroupJoinRequest{}
	e.ID = m.ID
	e.CreatedAt = m.CreatedAt
	e.InviterAt = m.InviterAt
	e.ExpiredAt = m.ExpiredAt
	e.GroupID = m.GroupID
	e.UserID = m.UserID
	e.Inviter = m.Inviter
	e.Remark = m.Remark
	e.OwnerID = m.OwnerID
	e.Status = entity.RequestStatus(m.Status)
	return e
}

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
	var model GroupJoinRequestModel

	if err := m.db.WithContext(ctx).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		First(&model).Error; err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (m *MySQLGroupJoinRequestRepository) Get(ctx context.Context, id uint32) (*entity.GroupJoinRequest, error) {
	var model GroupJoinRequestModel

	if err := m.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (m *MySQLGroupJoinRequestRepository) Create(ctx context.Context, entity *entity.GroupJoinRequest) (*entity.GroupJoinRequest, error) {
	var model GroupJoinRequestModel
	model.FromEntity(entity)

	if err := m.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (m *MySQLGroupJoinRequestRepository) Find(ctx context.Context, query *repository.GroupJoinRequestQuery) ([]*entity.GroupJoinRequest, error) {
	var models []GroupJoinRequestModel

	db := m.db.Model(&GroupJoinRequestModel{})

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
		es = append(es, model.ToEntity())
	}

	return es, nil
}

func (m *MySQLGroupJoinRequestRepository) Creates(ctx context.Context, es []*entity.GroupJoinRequest) ([]*entity.GroupJoinRequest, error) {
	var models []GroupJoinRequestModel

	if len(es) == 0 {
		return nil, errors.New("entity is empty")
	}

	for _, e := range es {
		var model GroupJoinRequestModel
		model.FromEntity(e)
		models = append(models, model)
	}

	if err := m.db.WithContext(ctx).Create(&models).Error; err != nil {
		return nil, err
	}

	var res []*entity.GroupJoinRequest
	for _, model := range models {
		res = append(res, model.ToEntity())
	}

	return res, nil
}

func (m *MySQLGroupJoinRequestRepository) UpdateStatus(ctx context.Context, id uint32, status entity.RequestStatus) error {
	var model GroupJoinRequestModel
	model.Status = uint8(status)

	if err := m.db.WithContext(ctx).
		Model(&GroupJoinRequestModel{}).
		Where("id = ?", id).
		Update("status", model.Status).
		Error; err != nil {
		return err
	}

	return nil
}

func (m *MySQLGroupJoinRequestRepository) Delete(ctx context.Context, id uint32) error {

	if err := m.db.WithContext(ctx).
		Model(&GroupJoinRequestModel{}).
		Where("id = ?", id).
		Update("deleted_at", ptime.Now()).
		Error; err != nil {
		return err
	}

	return nil
}
