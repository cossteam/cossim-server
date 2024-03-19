package persistence

import (
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

type GroupAnnouncementRepository struct {
	db *gorm.DB
}

func NewGroupAnnouncementRepository(db *gorm.DB) *GroupAnnouncementRepository {
	return &GroupAnnouncementRepository{db: db}
}

func (g *GroupAnnouncementRepository) CreateGroupAnnouncement(announcement *entity.GroupAnnouncement) error {
	return g.db.Create(announcement).Error
}

func (g *GroupAnnouncementRepository) GetGroupAnnouncementList(groupID uint32) ([]entity.GroupAnnouncement, error) {
	var announcements []entity.GroupAnnouncement
	err := g.db.Where("group_id = ? AND deleted_at = 0", groupID).Find(&announcements).Error
	if err != nil {
		return nil, err
	}
	return announcements, nil
}

func (g *GroupAnnouncementRepository) GetGroupAnnouncement(announcementID uint32) (*entity.GroupAnnouncement, error) {
	var announcement entity.GroupAnnouncement
	err := g.db.Where("id = ? AND deleted_at = 0", announcementID).First(&announcement).Error
	if err != nil {
		return nil, err
	}
	return &announcement, nil
}

func (g *GroupAnnouncementRepository) UpdateGroupAnnouncement(announcement *entity.GroupAnnouncement) error {
	return g.db.Model(&entity.GroupAnnouncement{}).Where("id = ?", announcement.ID).Updates(announcement).Error
}

func (g *GroupAnnouncementRepository) DeleteGroupAnnouncement(announcementID uint32) error {
	return g.db.Model(&entity.GroupAnnouncement{}).Where("id = ?", announcementID).Update("deleted_at", time.Now()).Error
}
