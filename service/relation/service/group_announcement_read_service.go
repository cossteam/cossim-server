package service

import (
	"context"
	"github.com/cossim/coss-server/pkg/code"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) MarkAnnouncementAsRead(ctx context.Context, request *relationgrpcv1.MarkAnnouncementAsReadRequest) (*relationgrpcv1.MarkAnnouncementAsReadResponse, error) {
	resp := &relationgrpcv1.MarkAnnouncementAsReadResponse{}
	err := s.garr.MarkAnnouncementAsRead(uint(request.GroupId), uint(request.AnnouncementId), request.UserIds)
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrGroupAnnouncementReadFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) GetReadUsers(ctx context.Context, request *relationgrpcv1.GetReadUsersRequest) (*relationgrpcv1.GetReadUsersResponse, error) {
	resp := &relationgrpcv1.GetReadUsersResponse{}
	list, err := s.garr.GetReadUsers(uint(request.GroupId), uint(request.AnnouncementId))
	if err != nil {
		return resp, status.Error(codes.Code(code.RelationErrGetGroupAnnouncementReadUsersFailed.Code()), err.Error())
	}
	if len(list) > 0 {
		var reads []*relationgrpcv1.AnnouncementRead
		for _, v := range list {
			reads = append(reads, &relationgrpcv1.AnnouncementRead{
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

func (s *Service) GetAnnouncementReadByUserId(ctx context.Context, request *relationgrpcv1.GetAnnouncementReadByUserIdRequest) (*relationgrpcv1.GetAnnouncementReadByUserIdResponse, error) {
	resp := &relationgrpcv1.GetAnnouncementReadByUserIdResponse{}
	read, err := s.garr.GetAnnouncementReadByUserId(uint(request.GroupId), uint(request.AnnouncementId), request.UserId)
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
