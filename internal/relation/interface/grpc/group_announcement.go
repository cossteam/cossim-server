package grpc

import (
	"context"
	v1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/internal/relation/infra/persistence"
	"github.com/cossim/coss-server/pkg/code"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var _ v1.GroupAnnouncementServiceServer = &groupAnnouncementServer{}

type groupAnnouncementServer struct {
	repos *persistence.Repositories
}

func (s *groupAnnouncementServer) CreateGroupAnnouncement(ctx context.Context, request *v1.CreateGroupAnnouncementRequest) (*v1.CreateGroupAnnouncementResponse, error) {
	announcement := &v1.CreateGroupAnnouncementResponse{}

	// TODO 改用 Validation 验证
	if request.GroupId == 0 || request.UserId == "" || request.Title == "" || request.Content == "" {
		return nil, code.InvalidParameter
	}

	// TODO 验证用户是否是群主

	ra, err := s.repos.GroupAnnouncementRepo.Create(ctx, &entity.GroupAnnouncement{
		GroupID: request.GroupId,
		Title:   request.Title,
		Content: request.Content,
		UserID:  request.UserId,
	})
	if err != nil {
		return nil, err
	}
	announcement.ID = ra.ID
	return announcement, nil
}

func (s *groupAnnouncementServer) GetGroupAnnouncementList(ctx context.Context, request *v1.GetGroupAnnouncementListRequest) (*v1.GetGroupAnnouncementListResponse, error) {
	resp := &v1.GetGroupAnnouncementListResponse{}
	//announcements, err := s.gar.GetGroupAnnouncementList(request.GroupId)
	//if err != nil {
	//	return resp, status.Error(codes.Code(code.RelationGroupErrGetGroupAnnouncementListFailed.Code()), err.Error())
	//}

	announcements, err := s.repos.GroupAnnouncementRepo.Find(ctx, &repository.GroupAnnouncementQuery{
		GroupID: []uint32{request.GroupId},
	})
	if err != nil {
		return nil, err
	}

	for _, announcement := range announcements {
		resp.AnnouncementList = append(resp.AnnouncementList, &v1.GroupAnnouncementInfo{
			ID:        announcement.ID,
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

func (s *groupAnnouncementServer) GetGroupAnnouncement(ctx context.Context, request *v1.GetGroupAnnouncementRequest) (*v1.GetGroupAnnouncementResponse, error) {
	resp := &v1.GetGroupAnnouncementResponse{}
	//announcement, err := s.gar.GetGroupAnnouncement(request.ID)
	//if err != nil {
	//	return resp, status.Error(codes.Code(code.RelationGroupErrGetGroupAnnouncementFailed.Code()), err.Error())
	//}

	announcement, err := s.repos.GroupAnnouncementRepo.Get(ctx, request.ID)
	if err != nil {
		return nil, err
	}

	resp.AnnouncementInfo = &v1.GroupAnnouncementInfo{
		ID:        announcement.ID,
		Content:   announcement.Content,
		GroupId:   announcement.GroupID,
		Title:     announcement.Title,
		UserId:    announcement.UserID,
		CreatedAt: announcement.CreatedAt,
		UpdatedAt: announcement.UpdatedAt,
	}
	return resp, nil
}

func (s *groupAnnouncementServer) UpdateGroupAnnouncement(ctx context.Context, request *v1.UpdateGroupAnnouncementRequest) (*v1.UpdateGroupAnnouncementResponse, error) {
	resp := &v1.UpdateGroupAnnouncementResponse{}

	if err := s.repos.GroupAnnouncementRepo.Update(ctx, &entity.UpdateGroupAnnouncement{
		ID:      request.ID,
		Title:   request.Title,
		Content: request.Content,
	}); err != nil {
		return nil, err
	}

	resp.ID = request.ID
	return resp, nil
}

func (s *groupAnnouncementServer) DeleteGroupAnnouncement(ctx context.Context, request *v1.DeleteGroupAnnouncementRequest) (*v1.DeleteGroupAnnouncementResponse, error) {
	resp := &v1.DeleteGroupAnnouncementResponse{}

	if err := s.repos.GroupAnnouncementRepo.Delete(ctx, request.ID); err != nil {
		return nil, err
	}

	resp.ID = request.ID
	return resp, nil
}

func (s *groupAnnouncementServer) MarkAnnouncementAsRead(ctx context.Context, request *v1.MarkAnnouncementAsReadRequest) (*v1.MarkAnnouncementAsReadResponse, error) {
	resp := &v1.MarkAnnouncementAsReadResponse{}

	if err := s.repos.GroupAnnouncementRepo.MarkAsRead(ctx, request.GroupId, request.AnnouncementId, request.UserIds); err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *groupAnnouncementServer) GetReadUsers(ctx context.Context, request *v1.GetReadUsersRequest) (*v1.GetReadUsersResponse, error) {
	resp := &v1.GetReadUsersResponse{}
	list, err := s.repos.GroupAnnouncementRepo.GetReadUsers(ctx, request.GroupId, request.AnnouncementId)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrGetGroupAnnouncementReadUsersFailed.Code()), err.Error())
	}
	if len(list) > 0 {
		var reads []*v1.AnnouncementRead
		for _, v := range list {
			reads = append(reads, &v1.AnnouncementRead{
				UserId:         v.UserId,
				ReadAt:         uint64(v.ReadAt),
				GroupId:        v.GroupID,
				AnnouncementId: v.AnnouncementId,
				ID:             v.ID,
			})
		}
		resp.AnnouncementReadUsers = reads
	}
	return resp, nil
}

func (s *groupAnnouncementServer) GetAnnouncementReadByUserId(ctx context.Context, request *v1.GetAnnouncementReadByUserIdRequest) (*v1.GetAnnouncementReadByUserIdResponse, error) {
	resp := &v1.GetAnnouncementReadByUserIdResponse{}

	read, err := s.repos.GroupAnnouncementRepo.GetReadByUserId(ctx, request.GroupId, request.AnnouncementId, request.UserId)
	if err != nil {
		return nil, err
	}

	resp.ID = read.ID
	resp.AnnouncementId = read.AnnouncementId
	resp.GroupId = read.GroupID
	resp.ReadAt = uint64(read.ReadAt)
	resp.UserId = read.UserId
	return resp, nil
}
