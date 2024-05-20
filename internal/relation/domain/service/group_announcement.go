package service

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/pkg/code"
)

type AddGroupAnnouncement struct {
	GroupID uint32
	UserID  string
	Title   string
	Content string
}

type UpdateGroupAnnouncement struct {
	ID      uint32
	Title   string
	Content string
}

type GroupAnnouncementDomain interface {
	IsExist(ctx context.Context, announcementID uint32) (bool, error)
	Update(ctx context.Context, req *UpdateGroupAnnouncement) error
	Delete(ctx context.Context, announcementID uint32) error
	Add(ctx context.Context, req *AddGroupAnnouncement) (*entity.GroupAnnouncement, error)
	SetAnnouncementRead(ctx context.Context, groupID uint32, announcementID uint32, userID string) error
	GetAnnouncement(ctx context.Context, id uint32) (*entity.GroupAnnouncement, error)
	ListAnnouncement(ctx context.Context, groupID uint32) (*entity.GroupAnnouncementList, error)
	ListAnnouncementRead(ctx context.Context, groupID, announcementID uint32) ([]*entity.GroupAnnouncementRead, error)
}

var _ GroupAnnouncementDomain = &groupAnnouncementDomain{}

func NewGroupAnnouncementDomain(repo repository.GroupAnnouncementRepository) GroupAnnouncementDomain {
	return &groupAnnouncementDomain{
		repo: repo,
	}
}

type groupAnnouncementDomain struct {
	repo repository.GroupAnnouncementRepository
}

func (d *groupAnnouncementDomain) IsExist(ctx context.Context, announcementID uint32) (bool, error) {
	a, err := d.repo.Get(ctx, announcementID)
	if err != nil {
		if errors.Is(err, code.NotFound) {
			return false, nil
		}
		return false, err
	}
	return a != nil, nil
}

func (d *groupAnnouncementDomain) Update(ctx context.Context, req *UpdateGroupAnnouncement) error {
	if req.ID == 0 || req.Title == "" {
		return code.InvalidParameter
	}

	return d.repo.Update(ctx, &entity.UpdateGroupAnnouncement{
		ID:      req.ID,
		Title:   req.Title,
		Content: req.Content,
	})
}

func (d *groupAnnouncementDomain) Delete(ctx context.Context, announcementID uint32) error {
	return d.repo.Delete(ctx, announcementID)
}

func (d *groupAnnouncementDomain) Add(ctx context.Context, req *AddGroupAnnouncement) (*entity.GroupAnnouncement, error) {
	return d.repo.Create(ctx, &entity.GroupAnnouncement{
		GroupID: req.GroupID,
		Title:   req.Title,
		Content: req.Content,
		UserID:  req.UserID,
	})
}

func (d *groupAnnouncementDomain) SetAnnouncementRead(ctx context.Context, groupID uint32, announcementID uint32, userID string) error {
	// TODO 可能要先判断一下是否已读了？
	if err := d.repo.MarkAsRead(ctx, groupID, announcementID, userID); err != nil {
		if errors.Is(err, code.NotFound) {
			return code.RelationErrGroupAnnouncementReadFailed.CustomMessage("群聊公告不存在")
		}
		return err
	}
	return nil
}

func (d *groupAnnouncementDomain) GetAnnouncement(ctx context.Context, id uint32) (*entity.GroupAnnouncement, error) {
	return d.repo.Get(ctx, id)
}

func (d *groupAnnouncementDomain) ListAnnouncementRead(ctx context.Context, groupID, announcementID uint32) ([]*entity.GroupAnnouncementRead, error) {
	return d.repo.GetReadUsers(ctx, groupID, announcementID)
}

func (d *groupAnnouncementDomain) ListAnnouncement(ctx context.Context, groupID uint32) (*entity.GroupAnnouncementList, error) {
	as, err := d.repo.Find(ctx, &entity.GroupAnnouncementQuery{
		GroupID: []uint32{groupID},
	})
	if err != nil {
		return nil, err
	}
	return &entity.GroupAnnouncementList{
		List:  as,
		Total: 0,
	}, nil
}
