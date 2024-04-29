package repository

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
)

type GroupAnnouncementQuery struct {
	ID      []uint32 // 群公告 ID 列表
	GroupID []uint32 // 群聊 ID 列表
	Name    string   // 群聊名称
	Limit   int      // 限制结果数量
	Offset  int      // 结果的偏移量
}

type GroupAnnouncementRepository interface {
	Create(ctx context.Context, announcement *entity.GroupAnnouncement) (*entity.GroupAnnouncement, error)
	Find(ctx context.Context, query *GroupAnnouncementQuery) ([]*entity.GroupAnnouncement, error)
	Get(ctx context.Context, announcementID uint32) (*entity.GroupAnnouncement, error)
	Update(ctx context.Context, announcement *entity.UpdateGroupAnnouncement) error
	Delete(ctx context.Context, announcementID uint32) error

	// MarkAsRead 标记公告为已读
	MarkAsRead(ctx context.Context, groupId, announcementId uint32, userIds []string) error

	// GetReadUsers 获取已读用户列表
	GetReadUsers(ctx context.Context, groupId, announcementId uint32) ([]*entity.GroupAnnouncementRead, error)

	// GetReadByUserId 获取已读记录
	GetReadByUserId(ctx context.Context, groupId, announcementId uint32, userId string) (*entity.GroupAnnouncementRead, error)
}
