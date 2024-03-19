package persistence

import (
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

type GroupAnnouncementReadRepo struct {
	db *gorm.DB
}

func NewGroupAnnouncementReadRepo(db *gorm.DB) *GroupAnnouncementReadRepo {
	return &GroupAnnouncementReadRepo{db: db}
}

func (g *GroupAnnouncementReadRepo) MarkAnnouncementAsRead(groupId, announcementId uint, userIds []string) error {
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

func (g *GroupAnnouncementReadRepo) GetReadUsers(groupId, announcementId uint) ([]*entity.GroupAnnouncementRead, error) {
	var users []*entity.GroupAnnouncementRead
	if err := g.db.Model(&entity.GroupAnnouncementRead{}).
		Where("group_id = ? AND announcement_id = ?", groupId, announcementId).Find(&users).
		Error; err != nil {
		return nil, err
	}
	return users, nil
}

func (g *GroupAnnouncementReadRepo) GetAnnouncementReadByUserId(groupId, announcementId uint, userId string) (*entity.GroupAnnouncementRead, error) {
	resp := &entity.GroupAnnouncementRead{}
	if err := g.db.Model(&entity.GroupAnnouncementRead{}).
		Where("group_id = ? AND announcement_id = ? AND user_id = ?", groupId, announcementId, userId).First(resp).Error; err != nil {
		return nil, err
	}
	return resp, nil
}
