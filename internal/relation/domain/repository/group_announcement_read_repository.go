package repository

import "github.com/cossim/coss-server/internal/relation/domain/entity"

type GroupAnnouncementReadRepository interface {
	// 标记公告为已读
	MarkAnnouncementAsRead(groupId, announcementId uint, userIds []string) error

	// 获取已读用户列表
	GetReadUsers(groupId, announcementId uint) ([]*entity.GroupAnnouncementRead, error)

	// 获取已读记录
	GetAnnouncementReadByUserId(groupId, announcementId uint, userId string) (*entity.GroupAnnouncementRead, error)
}
