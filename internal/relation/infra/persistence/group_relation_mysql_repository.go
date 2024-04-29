package persistence

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/cache"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
	"time"
)

type GroupRelationModel struct {
	BaseModel
	CreatedAt          int64
	UpdatedAt          int64
	DeletedAt          int64
	GroupID            uint32   `gorm:"comment:群组ID" json:"group_id"`
	Identity           uint8    `gorm:"comment:身份 (0=普通用户, 1=管理员, 2=群主)"`
	EntryMethod        uint8    `gorm:"comment:入群方式"`
	JoinedAt           int64    `gorm:"comment:加入时间"`
	MuteEndTime        int64    `gorm:"comment:禁言结束时间"`
	UserID             string   `gorm:"type:varchar(64);comment:用户ID"`
	Inviter            string   `gorm:"type:varchar(64);comment:邀请人id"`
	Remark             string   `gorm:"type:varchar(255);comment:添加群聊备注"`
	Label              []string `gorm:"type:varchar(255);comment:标签"`
	SilentNotification bool     `gorm:"comment:是否开启静默通知"`
	PrivacyMode        bool     `gorm:"comment:隐私模式"`
}

func (m *GroupRelationModel) TableName() string {
	return "group_relations"
}

func (m *GroupRelationModel) FromEntity(u *entity.GroupRelation) error {
	m.ID = u.ID
	m.GroupID = u.GroupID
	m.Identity = uint8(u.Identity)
	m.EntryMethod = uint8(u.EntryMethod)
	m.JoinedAt = u.JoinedAt
	m.MuteEndTime = u.MuteEndTime
	m.UserID = u.UserID
	m.Inviter = u.Inviter
	m.Remark = u.Remark
	m.Label = u.Label
	m.SilentNotification = u.SilentNotification
	m.PrivacyMode = u.PrivacyMode
	return nil
}

func (m *GroupRelationModel) ToEntity() *entity.GroupRelation {
	return &entity.GroupRelation{
		ID:                 m.ID,
		CreatedAt:          m.CreatedAt,
		GroupID:            m.GroupID,
		Identity:           entity.GroupIdentity(m.Identity),
		EntryMethod:        entity.EntryMethod(m.EntryMethod),
		JoinedAt:           m.JoinedAt,
		MuteEndTime:        m.MuteEndTime,
		UserID:             m.UserID,
		Inviter:            m.Inviter,
		Remark:             m.Remark,
		Label:              m.Label,
		SilentNotification: m.SilentNotification,
		PrivacyMode:        m.PrivacyMode,
	}
}

var _ repository.GroupRepository = &MySQLRelationGroupRepository{}

func NewMySQLRelationGroupRepository(db *gorm.DB, cache cache.RelationUserCache) *MySQLRelationGroupRepository {
	return &MySQLRelationGroupRepository{
		db: db,
		//cache: cache,
	}
}

type MySQLRelationGroupRepository struct {
	db *gorm.DB
}

func (m *MySQLRelationGroupRepository) UpdateFieldsByGroupID(ctx context.Context, id uint32, fields map[string]interface{}) error {
	if err := m.db.WithContext(ctx).Model(&GroupRelationModel{}).
		Where("id = ?", id).
		Updates(fields).
		Error; err != nil {
		return err
	}

	return nil
}

func (m *MySQLRelationGroupRepository) Get(ctx context.Context, id uint32) (*entity.GroupRelation, error) {
	var model GroupRelationModel
	if err := m.db.WithContext(ctx).
		Where("id = ?", id).
		First(&model).
		Error; err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (m *MySQLRelationGroupRepository) Create(ctx context.Context, createGroupRelation *repository.CreateGroupRelation) (*entity.GroupRelation, error) {
	model := &GroupRelationModel{
		GroupID:     createGroupRelation.GroupID,
		Identity:    uint8(createGroupRelation.Identity),
		EntryMethod: uint8(createGroupRelation.EntryMethod),
		JoinedAt:    createGroupRelation.JoinedAt,
		UserID:      createGroupRelation.UserID,
		Inviter:     createGroupRelation.Inviter,
	}

	if err := m.db.WithContext(ctx).
		Create(&model).
		Error; err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (m *MySQLRelationGroupRepository) Update(ctx context.Context, ur *entity.GroupRelation) (*entity.GroupRelation, error) {
	var model GroupRelationModel

	if err := model.FromEntity(ur); err != nil {
		return nil, err
	}

	if err := m.db.WithContext(ctx).
		Where("id = ?", ur.ID).
		Updates(&model).
		Error; err != nil {
		return ur, err
	}

	return model.ToEntity(), nil
}

func (m *MySQLRelationGroupRepository) Delete(ctx context.Context, id uint32) error {
	if err := m.db.WithContext(ctx).
		Model(&GroupRelationModel{}).
		Where("id = ?", id).
		Update("deleted_at", ptime.Now()).Error; err != nil {
		return err
	}

	return nil
}

func (m *MySQLRelationGroupRepository) GetGroupUserIDs(ctx context.Context, gid uint32) ([]string, error) {
	var userGroupIDs []string
	if err := m.db.WithContext(ctx).
		Model(&GroupRelationModel{}).
		Where("group_id = ?  AND deleted_at = 0", gid).
		Pluck("user_id", &userGroupIDs).Error; err != nil {
		return nil, err
	}
	return userGroupIDs, nil
}

func (m *MySQLRelationGroupRepository) GetUserGroupIDs(ctx context.Context, uid string) ([]uint32, error) {
	var groupIDs []uint32
	if err := m.db.WithContext(ctx).
		Model(&GroupRelationModel{}).
		Where("user_id = ? AND deleted_at = 0", uid).
		Pluck("group_id", &groupIDs).Error; err != nil {
		return groupIDs, err
	}
	return groupIDs, nil
}

func (m *MySQLRelationGroupRepository) GetUserGroupByGroupIDAndUserID(ctx context.Context, gid uint32, uid string) (*entity.GroupRelation, error) {
	var model GroupRelationModel
	if err := m.db.WithContext(ctx).
		Model(&GroupRelationModel{}).
		Where(" group_id = ? and user_id = ? AND deleted_at = 0", gid, uid).
		First(&model).Error; err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (m *MySQLRelationGroupRepository) GetUsersGroupByGroupIDAndUserIDs(ctx context.Context, gid uint32, uids []string) ([]*entity.GroupRelation, error) {
	var ugs []*GroupRelationModel
	if err := m.db.WithContext(ctx).
		Where(" group_id = ? and user_id IN (?) AND deleted_at = 0", gid, uids).
		Find(&ugs).Error; err != nil {
		return nil, err
	}

	var ugs2 = make([]*entity.GroupRelation, 0)
	for _, ug := range ugs {
		ugs2 = append(ugs2, ug.ToEntity())
	}

	return ugs2, nil
}

func (m *MySQLRelationGroupRepository) GetUserJoinedGroupIDs(ctx context.Context, uid string) ([]uint32, error) {
	var groupIDs []uint32
	if err := m.db.WithContext(ctx).
		Model(&GroupRelationModel{}).
		Where("user_id = ? AND deleted_at = 0", uid).
		Pluck("group_id", &groupIDs).Error; err != nil {
		return groupIDs, err
	}
	return groupIDs, nil
}

func (m *MySQLRelationGroupRepository) GetUserManageGroupIDs(ctx context.Context, uid string) ([]uint32, error) {
	var groupIDs []uint32
	if err := m.db.WithContext(ctx).
		Model(&GroupRelationModel{}).
		Distinct("group_id").
		Where("user_id = ? AND identity IN ?", uid, []entity.GroupIdentity{entity.IdentityAdmin, entity.IdentityOwner}).
		Pluck("group_id", &groupIDs).Error; err != nil {
		return nil, err
	}
	return groupIDs, nil
}

func (m *MySQLRelationGroupRepository) DeleteByGroupIDAndUserID(ctx context.Context, gid uint32, uid ...string) error {
	if err := m.db.WithContext(ctx).
		Model(&GroupRelationModel{}).
		Where(" group_id = ? AND user_id IN (?) AND deleted_at = 0", gid, uid).
		Update("deleted_at", time.Now()).Error; err != nil {
		return err
	}

	return nil
}

func (m *MySQLRelationGroupRepository) ListJoinRequest(ctx context.Context, gids []uint32) ([]*entity.GroupRelation, error) {
	var joinRequests []*GroupRelationModel

	if err := m.db.WithContext(ctx).
		Model(&GroupRelationModel{}).
		Where("group_id IN (?)", gids).
		Find(&joinRequests).Error; err != nil {
		return nil, err
	}

	var joinRequests2 = make([]*entity.GroupRelation, 0)
	for _, joinRequest := range joinRequests {
		joinRequests2 = append(joinRequests2, joinRequest.ToEntity())
	}

	return joinRequests2, nil
}

func (m *MySQLRelationGroupRepository) ListGroupAdmin(ctx context.Context, gid uint32) ([]string, error) {
	var adminIds []string
	if err := m.db.WithContext(ctx).
		Model(&GroupRelationModel{}).
		Where("(identity = ? or identity = ?)AND group_id = ? AND deleted_at = 0", entity.IdentityAdmin, entity.IdentityOwner, gid).
		Pluck("user_id", &adminIds).Error; err != nil {
		return nil, err
	}
	return adminIds, nil
}

func (m *MySQLRelationGroupRepository) SetUserGroupRemark(ctx context.Context, gid uint32, uid string, remark string) error {
	return m.db.WithContext(ctx).
		Model(&GroupRelationModel{}).
		Where(" group_id = ? AND user_id = ?", gid, uid).
		Update("remark", remark).Error
}

func (m *MySQLRelationGroupRepository) UpdateIdentity(ctx context.Context, gid uint32, uid string, identity entity.GroupIdentity) error {
	var model GroupRelationModel
	model.Identity = uint8(identity)
	return m.db.WithContext(ctx).
		Model(&GroupRelationModel{}).
		Where(" group_id = ? AND user_id = ?", gid, uid).
		Update("identity", model.Identity).Error
}

func (m *MySQLRelationGroupRepository) UserGroupSilentNotification(ctx context.Context, gid uint32, uid string, silentNotification bool) error {
	return m.db.WithContext(ctx).
		Model(&GroupRelationModel{}).
		Where(" group_id = ? AND user_id = ?", gid, uid).
		Update("silent_notification", silentNotification).Error
}

func (m *MySQLRelationGroupRepository) UpdateFieldsByGroupAndUser(ctx context.Context, gid uint32, uid string, fields map[string]interface{}) error {
	return m.db.WithContext(ctx).
		Model(&UserRelationModel{}).
		Where(" group_id = ? AND user_id = ?", gid, uid).
		//Unscoped().
		Updates(fields).Error
}

func (m *MySQLRelationGroupRepository) DeleteByGroupID(ctx context.Context, gid uint32) error {
	return m.db.WithContext(ctx).
		Model(&GroupRelationModel{}).
		Where("group_id = ?", gid).
		Update("deleted_at", time.Now()).Error
}

func (m *MySQLRelationGroupRepository) UpdateFields(ctx context.Context, id uint32, fields map[string]interface{}) error {
	return m.db.WithContext(ctx).
		Model(&UserRelationModel{}).
		Where("id = ?", id).
		//Unscoped().
		Updates(fields).Error
}
