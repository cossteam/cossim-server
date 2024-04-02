package grpc

import (
	"context"
	v1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/pkg/cache"
	"github.com/cossim/coss-server/pkg/code"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

var _ v1.GroupAnnouncementServiceServer = &groupAnnouncementServer{}

type groupAnnouncementServer struct {
	db    *gorm.DB
	cache cache.RelationUserCache
	gar   repository.GroupAnnouncementRepository
}

func (s *groupAnnouncementServer) CreateGroupAnnouncement(ctx context.Context, in *v1.CreateGroupAnnouncementRequest) (*v1.CreateGroupAnnouncementResponse, error) {
	announcement := &v1.CreateGroupAnnouncementResponse{}
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

func (s *groupAnnouncementServer) GetGroupAnnouncementList(ctx context.Context, in *v1.GetGroupAnnouncementListRequest) (*v1.GetGroupAnnouncementListResponse, error) {
	resp := &v1.GetGroupAnnouncementListResponse{}
	announcements, err := s.gar.GetGroupAnnouncementList(in.GroupId)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrGetGroupAnnouncementListFailed.Code()), err.Error())
	}
	for _, announcement := range announcements {
		resp.AnnouncementList = append(resp.GetAnnouncementList(), &v1.GroupAnnouncementInfo{
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

func (s *groupAnnouncementServer) GetGroupAnnouncement(ctx context.Context, in *v1.GetGroupAnnouncementRequest) (*v1.GetGroupAnnouncementResponse, error) {
	resp := &v1.GetGroupAnnouncementResponse{}
	announcement, err := s.gar.GetGroupAnnouncement(in.ID)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrGetGroupAnnouncementFailed.Code()), err.Error())
	}
	resp.AnnouncementInfo = &v1.GroupAnnouncementInfo{
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

func (s *groupAnnouncementServer) UpdateGroupAnnouncement(ctx context.Context, in *v1.UpdateGroupAnnouncementRequest) (*v1.UpdateGroupAnnouncementResponse, error) {
	resp := &v1.UpdateGroupAnnouncementResponse{}
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

func (s *groupAnnouncementServer) DeleteGroupAnnouncement(ctx context.Context, in *v1.DeleteGroupAnnouncementRequest) (*v1.DeleteGroupAnnouncementResponse, error) {
	resp := &v1.DeleteGroupAnnouncementResponse{}
	if err := s.gar.DeleteGroupAnnouncement(in.ID); err != nil {
		return resp, status.Error(codes.Code(code.RelationGroupErrDeleteGroupAnnouncementFailed.Code()), err.Error())
	}
	resp.ID = in.ID
	return resp, nil
}

func (s *groupAnnouncementServer) MarkAnnouncementAsRead(ctx context.Context, request *v1.MarkAnnouncementAsReadRequest) (*v1.MarkAnnouncementAsReadResponse, error) {
	resp := &v1.MarkAnnouncementAsReadResponse{}
	err := s.gar.MarkAnnouncementAsRead(uint(request.GroupId), uint(request.AnnouncementId), request.UserIds)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrGroupAnnouncementReadFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *groupAnnouncementServer) GetReadUsers(ctx context.Context, request *v1.GetReadUsersRequest) (*v1.GetReadUsersResponse, error) {
	resp := &v1.GetReadUsersResponse{}
	list, err := s.gar.GetReadUsers(uint(request.GroupId), uint(request.AnnouncementId))
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrGetGroupAnnouncementReadUsersFailed.Code()), err.Error())
	}
	if len(list) > 0 {
		var reads []*v1.AnnouncementRead
		for _, v := range list {
			reads = append(reads, &v1.AnnouncementRead{
				UserId:         v.UserId,
				ReadAt:         uint64(v.ReadAt),
				GroupId:        uint32(v.GroupID),
				AnnouncementId: uint32(v.AnnouncementId),
				ID:             uint32(v.ID),
			})
		}
		resp.AnnouncementReadUsers = reads
	}
	return resp, nil
}

func (s *groupAnnouncementServer) GetAnnouncementReadByUserId(ctx context.Context, request *v1.GetAnnouncementReadByUserIdRequest) (*v1.GetAnnouncementReadByUserIdResponse, error) {
	resp := &v1.GetAnnouncementReadByUserIdResponse{}
	read, err := s.gar.GetAnnouncementReadByUserId(uint(request.GroupId), uint(request.AnnouncementId), request.UserId)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrGetGroupAnnouncementReadFailed.Code()), err.Error())
	}
	resp.ID = uint32(read.ID)
	resp.AnnouncementId = uint32(read.AnnouncementId)
	resp.GroupId = uint32(read.GroupID)
	resp.ReadAt = uint64(read.ReadAt)
	resp.UserId = read.UserId
	return resp, nil
}
