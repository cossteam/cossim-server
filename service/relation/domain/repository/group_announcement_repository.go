package repository

import "github.com/cossim/coss-server/service/relation/domain/entity"

type GroupAnnouncementRepository interface {
	CreateGroupAnnouncement(announcement *entity.GroupAnnouncement) error
	GetGroupAnnouncementList(groupID uint32) ([]entity.GroupAnnouncement, error)
	GetGroupAnnouncement(announcementID uint32) (*entity.GroupAnnouncement, error)
	UpdateGroupAnnouncement(announcement *entity.GroupAnnouncement) error
	DeleteGroupAnnouncement(announcementID uint32) error
}
