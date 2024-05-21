package converter

import (
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/infra/persistence/po"
)

func GroupAnnouncementEntityToPo(e *entity.GroupAnnouncement) *po.GroupAnnouncement {
	m := &po.GroupAnnouncement{}
	m.ID = e.ID
	m.GroupID = e.GroupID
	m.Title = e.Title
	m.Content = e.Content
	m.UserID = e.UserID
	return m
}

func GroupAnnouncementPoToEntity(po *po.GroupAnnouncement) *entity.GroupAnnouncement {
	return &entity.GroupAnnouncement{
		ID:        po.ID,
		GroupID:   po.GroupID,
		Title:     po.Title,
		Content:   po.Content,
		UserID:    po.UserID,
		CreatedAt: po.CreatedAt,
		UpdatedAt: po.UpdatedAt,
	}
}

func GroupAnnouncementReadEntityToPo(e *entity.GroupAnnouncementRead) *po.GroupAnnouncementRead {
	m := &po.GroupAnnouncementRead{}
	m.ID = e.ID
	m.AnnouncementID = e.AnnouncementId
	m.DialogID = e.DialogId
	m.GroupID = e.GroupID
	m.ReadAt = e.ReadAt
	m.UserID = e.UserId
	return m
}

func GroupAnnouncementReadEntityPoToEntity(po *po.GroupAnnouncementRead) *entity.GroupAnnouncementRead {
	return &entity.GroupAnnouncementRead{
		ID:             po.ID,
		ReadAt:         po.ReadAt,
		UserId:         po.UserID,
		DialogId:       po.DialogID,
		GroupID:        po.GroupID,
		AnnouncementId: po.AnnouncementID,
		CreatedAt:      po.CreatedAt,
	}
}
