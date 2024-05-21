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
	model := converter.GroupAnnouncementEntityToPo(announcement)

	if err := m.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, err
	}

	return converter.GroupAnnouncementPoToEntity(model), nil
}

func (m *MySQLRelationGroupAnnouncementRepository) Find(ctx context.Context, query *entity.GroupAnnouncementQuery) ([]*entity.GroupAnnouncement, error) {
	var models []*po.GroupAnnouncement

	if len(query.ID) <= 0 && len(query.GroupID) <= 0 && query.Name == "" {
		return nil, code.InvalidParameter
	}

	db := m.db.WithContext(ctx).Model(&po.GroupAnnouncement{}).Where("deleted_at = 0")

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
		announcements = append(announcements, converter.GroupAnnouncementPoToEntity(model))
	}

	return announcements, nil
}

func (m *MySQLRelationGroupAnnouncementRepository) Get(ctx context.Context, announcementID uint32) (*entity.GroupAnnouncement, error) {
	model := &po.GroupAnnouncement{}
	if err := m.db.WithContext(ctx).
		Where("id = ? AND deleted_at = 0", announcementID).
		First(model).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.NotFound
		}
		return nil, err
	}

	return converter.GroupAnnouncementPoToEntity(model), nil
}

func (m *MySQLRelationGroupAnnouncementRepository) Update(ctx context.Context, announcement *entity.UpdateGroupAnnouncement) error {
	if err := m.db.WithContext(ctx).
		Model(&po.GroupAnnouncement{}).
		Where("id = ? and deleted_at = 0", announcement.ID).
		Update("title", announcement.Title).
		Update("content", announcement.Content).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.NotFound
		}
		return err
	}

	return nil
}

func (m *MySQLRelationGroupAnnouncementRepository) Delete(ctx context.Context, announcementID uint32) error {
	if err := m.db.WithContext(ctx).
		Model(&po.GroupAnnouncement{}).
		Where("id = ?", announcementID).
		Update("deleted_at", ptime.Now()).
		Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return code.NotFound
		}
		return err
	}
	return nil
}

func (m *MySQLRelationGroupAnnouncementRepository) MarkAsRead(ctx context.Context, groupId, announcementId uint32, userIds ...string) error {
	if err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, userId := range userIds {
			announcementRead := &po.GroupAnnouncementRead{
				AnnouncementID: announcementId,
				GroupID:        groupId,
				UserID:         userId,
				ReadAt:         ptime.Now(),
			}
			if err := tx.
				Where("id = ? and user_id = ? and deleted_at = 0", announcementId, userId).
				Assign(announcementRead).
				FirstOrCreate(announcementRead).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return code.NotFound
				}
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
	var users []*po.GroupAnnouncementRead

	if err := m.db.WithContext(ctx).
		Model(&po.GroupAnnouncementRead{}).
		Where("group_id = ? AND announcement_id = ? and deleted_at = 0", groupId, announcementId).
		Find(&users).
		Error; err != nil {
		return nil, err
	}

	var es []*entity.GroupAnnouncementRead
	for _, user := range users {
		es = append(es, converter.GroupAnnouncementReadEntityPoToEntity(user))
	}

	return es, nil
}

func (m *MySQLRelationGroupAnnouncementRepository) GetReadByUserId(ctx context.Context, groupId, announcementId uint32, userId string) (*entity.GroupAnnouncementRead, error) {
	model := &po.GroupAnnouncementRead{}

	if err := m.db.WithContext(ctx).
		Model(&po.GroupAnnouncementRead{}).
		Where("group_id = ? AND announcement_id = ? AND user_id = ? and deleted_at = 0", groupId, announcementId, userId).
		First(model).Error; err != nil {
		return nil, err
	}

	return converter.GroupAnnouncementReadEntityPoToEntity(model), nil
}
