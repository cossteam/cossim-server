package persistence

import (
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

type GroupAnnouncementRepository struct {
	db *gorm.DB
}

//func NewGroupAnnouncementRepository(db *gorm.DB) *GroupAnnouncementRepository {
//	return &GroupAnnouncementRepository{db: db}
//}

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

func (g *GroupAnnouncementRepository) MarkAnnouncementAsRead(groupId, announcementId uint, userIds []string) error {
	err := g.db.Transaction(func(tx *gorm.DB) error {
		for _, userId := range userIds {
			announcementRead := entity.GroupAnnouncementRead{
				GroupID:        groupId,
				AnnouncementId: announcementId,
				UserId:         userId,
				ReadAt:         time.Now(),
			}
			if err := tx.Where(entity.GroupAnnouncementRead{AnnouncementId: announcementId, UserId: userId}).Assign(announcementRead).FirstOrCreate(&announcementRead).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (g *GroupAnnouncementRepository) GetReadUsers(groupId, announcementId uint) ([]*entity.GroupAnnouncementRead, error) {
	var users []*entity.GroupAnnouncementRead
	if err := g.db.Model(&entity.GroupAnnouncementRead{}).
		Where("group_id = ? AND announcement_id = ?", groupId, announcementId).Find(&users).
		Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (g *GroupAnnouncementRepository) GetAnnouncementReadByUserId(groupId, announcementId uint, userId string) (*entity.GroupAnnouncementRead, error) {
	resp := &entity.GroupAnnouncementRead{}
	if err := g.db.Model(&entity.GroupAnnouncementRead{}).
		Where("group_id = ? AND announcement_id = ? AND user_id = ?", groupId, announcementId, userId).First(resp).Error; err != nil {
		return nil, err
	}
	return resp, nil
}
