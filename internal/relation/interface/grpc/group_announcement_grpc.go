package grpc

import (
	"context"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/pkg/code"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Handler) CreateGroupAnnouncement(ctx context.Context, in *relationgrpcv1.CreateGroupAnnouncementRequest) (*relationgrpcv1.CreateGroupAnnouncementResponse, error) {
	announcement := &relationgrpcv1.CreateGroupAnnouncementResponse{}
	if err := s.gar.CreateGroupAnnouncement(&entity.GroupAnnouncement{
		Content: in.Content,
		GroupID: in.GroupId,
		Title:   in.Title,
		UserID:  in.UserId,
	}); err != nil {
		return announcement, status.Error(codes.Code(code.RelationGroupErrCreateGroupAnnouncementFailed.Code()), err.Error())
	}
	return announcement, nil
}

func (s *Handler) GetGroupAnnouncementList(ctx context.Context, in *relationgrpcv1.GetGroupAnnouncementListRequest) (*relationgrpcv1.GetGroupAnnouncementListResponse, error) {
	resp := &relationgrpcv1.GetGroupAnnouncementListResponse{}
	announcements, err := s.gar.GetGroupAnnouncementList(in.GroupId)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrGetGroupAnnouncementListFailed.Code()), err.Error())
	}
	for _, announcement := range announcements {
		resp.AnnouncementList = append(resp.GetAnnouncementList(), &relationgrpcv1.GroupAnnouncementInfo{
			ID:        uint32(announcement.ID),
			Content:   announcement.Content,
			GroupId:   announcement.GroupID,
			Title:     announcement.Title,
			UserId:    announcement.UserID,
			CreatedAt: announcement.CreatedAt,
			UpdatedAt: announcement.UpdatedAt,
		})
	}

	return resp, nil
}

func (s *Handler) GetGroupAnnouncement(ctx context.Context, in *relationgrpcv1.GetGroupAnnouncementRequest) (*relationgrpcv1.GetGroupAnnouncementResponse, error) {
	resp := &relationgrpcv1.GetGroupAnnouncementResponse{}
	announcement, err := s.gar.GetGroupAnnouncement(in.ID)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrGetGroupAnnouncementFailed.Code()), err.Error())
	}
	resp.AnnouncementInfo = &relationgrpcv1.GroupAnnouncementInfo{
		ID:        uint32(announcement.ID),
		Content:   announcement.Content,
		GroupId:   announcement.GroupID,
		Title:     announcement.Title,
		UserId:    announcement.UserID,
		CreatedAt: announcement.CreatedAt,
		UpdatedAt: announcement.UpdatedAt,
	}
	return resp, nil
}

func (s *Handler) UpdateGroupAnnouncement(ctx context.Context, in *relationgrpcv1.UpdateGroupAnnouncementRequest) (*relationgrpcv1.UpdateGroupAnnouncementResponse, error) {
	resp := &relationgrpcv1.UpdateGroupAnnouncementResponse{}
	if err := s.gar.UpdateGroupAnnouncement(&entity.GroupAnnouncement{
		BaseModel: entity.BaseModel{
			ID: uint(in.ID),
		},
		Content: in.Content,
		Title:   in.Title,
	}); err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrUpdateGroupAnnouncementFailed.Code()), err.Error())
	}
	resp.ID = in.ID
	return resp, nil
}

func (s *Handler) DeleteGroupAnnouncement(ctx context.Context, in *relationgrpcv1.DeleteGroupAnnouncementRequest) (*relationgrpcv1.DeleteGroupAnnouncementResponse, error) {
	resp := &relationgrpcv1.DeleteGroupAnnouncementResponse{}
	if err := s.gar.DeleteGroupAnnouncement(in.ID); err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrDeleteGroupAnnouncementFailed.Code()), err.Error())
	}
	resp.ID = in.ID
	return resp, nil
}
