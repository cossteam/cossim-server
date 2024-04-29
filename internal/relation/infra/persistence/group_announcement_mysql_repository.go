package persistence

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/cache"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

type GroupAnnouncementModel struct {
	BaseModel
	GroupID uint32 `gorm:"column:group_id"`
	Title   string `gorm:"column:title;comment:公告标题"`
	Content string `gorm:"column:content;comment:公告内容"`
	UserID  string `gorm:"column:user_id"`
}

func (m *GroupAnnouncementModel) TableName() string {
	return "group_announcements"
}

func (m *GroupAnnouncementModel) FromEntity(e *entity.GroupAnnouncement) error {
	m.ID = e.ID
	m.GroupID = e.GroupID
	m.Title = e.Title
	m.Content = e.Content
	m.UserID = e.UserID
	return nil
}

func (m *GroupAnnouncementModel) ToEntity() *entity.GroupAnnouncement {
	return &entity.GroupAnnouncement{
		ID:        m.ID,
		GroupID:   m.GroupID,
		Title:     m.Title,
		Content:   m.Content,
		UserID:    m.UserID,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}
}

type GroupAnnouncementReadModel struct {
	BaseModel
	AnnouncementID uint32 `gorm:"comment:公告ID" json:"announcement_id"`
	DialogID       uint32 `gorm:"default:0;comment:对话ID" json:"dialog_id"`
	GroupID        uint32 `gorm:"comment:群聊id" json:"group_id"`
	ReadAt         int64  `gorm:"comment:已读时间" json:"read_at"`
	UserID         string `gorm:"comment:用户ID" json:"user_id"`
}

func (m *GroupAnnouncementReadModel) TableName() string {
	return "group_announcement_reads"
}

func (m *GroupAnnouncementReadModel) FromEntity(e *entity.GroupAnnouncementRead) error {
	m.ID = e.ID
	m.AnnouncementID = e.AnnouncementId
	m.DialogID = e.DialogId
	m.GroupID = e.GroupID
	m.ReadAt = e.ReadAt
	m.UserID = e.UserId
	return nil
}

func (m *GroupAnnouncementReadModel) ToEntity() *entity.GroupAnnouncementRead {
	return &entity.GroupAnnouncementRead{
		ID:             m.ID,
		ReadAt:         m.ReadAt,
		UserId:         m.UserID,
		DialogId:       m.DialogID,
		GroupID:        m.GroupID,
		AnnouncementId: m.AnnouncementID,
		CreatedAt:      m.CreatedAt,
	}
}

var _ repository.GroupAnnouncementRepository = &MySQLRelationGroupAnnouncementRepository{}

func NewMySQLRelationGroupAnnouncementRepository(db *gorm.DB, cache cache.RelationUserCache) *MySQLRelationGroupAnnouncementRepository {
	return &MySQLRelationGroupAnnouncementRepository{
		db: db,
		//cache: cache,
	}
}

type MySQLRelationGroupAnnouncementRepository struct {
	db *gorm.DB
}

func (m *MySQLRelationGroupAnnouncementRepository) Create(ctx context.Context, announcement *entity.GroupAnnouncement) (*entity.GroupAnnouncement, error) {
	var model GroupAnnouncementModel

	if err := model.FromEntity(announcement); err != nil {
		return nil, err
	}

	if err := m.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (m *MySQLRelationGroupAnnouncementRepository) Find(ctx context.Context, query *repository.GroupAnnouncementQuery) ([]*entity.GroupAnnouncement, error) {
	var models []GroupAnnouncementModel

	db := m.db.Model(&GroupAnnouncementModel{})

	if len(query.ID) > 0 {
		db = db.Where("id IN (?)", query.ID)
	}
	if len(query.GroupID) > 0 {
		db = db.Where("group_id IN (?)", query.GroupID)
	}
	if query.Name != "" {
		db = db.Where("name = ?", query.Name)
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

	var announcements = make([]*entity.GroupAnnouncement, 0)
	for _, model := range models {
		announcements = append(announcements, model.ToEntity())
	}

	return announcements, nil
}

func (m *MySQLRelationGroupAnnouncementRepository) Get(ctx context.Context, announcementID uint32) (*entity.GroupAnnouncement, error) {
	var model GroupAnnouncementModel
	if err := m.db.WithContext(ctx).
		Where("id = ? AND deleted_at = 0", announcementID).
		First(&model).
		Error; err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (m *MySQLRelationGroupAnnouncementRepository) Update(ctx context.Context, announcement *entity.UpdateGroupAnnouncement) error {
	var model GroupAnnouncementModel

	if err := m.db.WithContext(ctx).
		Model(&GroupRelationModel{}).
		Where("id = ?", model.ID).
		Update("title", announcement.Title).
		Update("content", announcement.Content).
		Error; err != nil {
		return err
	}

	return nil
}

func (m *MySQLRelationGroupAnnouncementRepository) Delete(ctx context.Context, announcementID uint32) error {
	if err := m.db.WithContext(ctx).
		Model(&GroupRelationModel{}).
		Where("id = ?", announcementID).
		Update("deleted_at", ptime.Now()).
		Error; err != nil {
		return err
	}
	return nil
}

func (m *MySQLRelationGroupAnnouncementRepository) MarkAsRead(ctx context.Context, groupId, announcementId uint32, userIds []string) error {
	if err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, userId := range userIds {
			announcementRead := &GroupAnnouncementReadModel{
				AnnouncementID: announcementId,
				GroupID:        groupId,
				UserID:         userId,
				ReadAt:         ptime.Now(),
			}
			if err := tx.
				Where("id = ? and user_id = ? ", announcementId, userId).
				Assign(announcementRead).
				FirstOrCreate(announcementRead).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
}

func (m *MySQLRelationGroupAnnouncementRepository) GetReadUsers(ctx context.Context, groupId, announcementId uint32) ([]*entity.GroupAnnouncementRead, error) {
	var users []*GroupAnnouncementReadModel

	if err := m.db.WithContext(ctx).
		Model(&GroupAnnouncementReadModel{}).
		Where("group_id = ? AND announcement_id = ?", groupId, announcementId).
		Find(&users).
		Error; err != nil {
		return nil, err
	}

	var es []*entity.GroupAnnouncementRead
	for _, user := range users {
		es = append(es, user.ToEntity())
	}

	return es, nil
}

func (m *MySQLRelationGroupAnnouncementRepository) GetReadByUserId(ctx context.Context, groupId, announcementId uint32, userId string) (*entity.GroupAnnouncementRead, error) {
	var model GroupAnnouncementReadModel

	if err := m.db.WithContext(ctx).
		Model(&GroupAnnouncementReadModel{}).
		Where("group_id = ? AND announcement_id = ? AND user_id = ?", groupId, announcementId, userId).
		First(&model).Error; err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}
