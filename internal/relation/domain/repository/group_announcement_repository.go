package repository

import "github.com/cossim/coss-server/internal/relation/domain/entity"

type GroupAnnouncementRepository interface {
	CreateGroupAnnouncement(announcement *entity.GroupAnnouncement) error
	GetGroupAnnouncementList(groupID uint32) ([]entity.GroupAnnouncement, error)
	GetGroupAnnouncement(announcementID uint32) (*entity.GroupAnnouncement, error)
	UpdateGroupAnnouncement(announcement *entity.GroupAnnouncement) error
	DeleteGroupAnnouncement(announcementID uint32) error

	// 标记公告为已读
	MarkAnnouncementAsRead(groupId, announcementId uint, userIds []string) error

	// 获取已读用户列表
	GetReadUsers(groupId, announcementId uint) ([]*entity.GroupAnnouncementRead, error)

	// 获取已读记录
	GetAnnouncementReadByUserId(groupId, announcementId uint, userId string) (*entity.GroupAnnouncementRead, error)
}
