package service

import (
	"context"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/service/group/api/v1"
	"github.com/cossim/coss-server/service/group/domain/entity"
	"github.com/cossim/coss-server/service/group/domain/repository"
	"github.com/cossim/coss-server/service/group/infrastructure/persistence"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewService(repo *persistence.Repositories) *Service {
	return &Service{gr: repo.Gr}
}

type Service struct {
	gr repository.GroupRepository
	v1.UnimplementedGroupServiceServer
}

func (s *Service) GetGroupInfoByGid(ctx context.Context, request *v1.GetGroupInfoRequest) (*v1.Group, error) {
	resp := &v1.Group{}

	group, err := s.gr.GetGroupInfoByGid(uint(request.GetGid()))
	if err != nil {
		return resp, status.Error(codes.Code(code.GroupErrGetGroupInfoByGidFailed.Code()), err.Error())
	}

	// 将领域模型转换为 gRPC API 模型
	resp = &v1.Group{
		Id:              uint32(group.ID),
		Type:            int32(uint32(group.Type)),
		Status:          int32(uint32(group.Status)),
		MaxMembersLimit: int32(group.MaxMembersLimit),
		CreatorId:       group.CreatorID,
		Name:            group.Name,
		Avatar:          group.Avatar,
	}
	return resp, nil
}

func (s *Service) GetBatchGroupInfoByIDs(ctx context.Context, request *v1.GetBatchGroupInfoRequest) (*v1.GetBatchGroupInfoResponse, error) {
	resp := &v1.GetBatchGroupInfoResponse{}

	if len(request.GetGroupIds()) == 0 {
		return resp, nil
	}

	//将uint32转成uint
	groupIds := make([]uint, len(request.GetGroupIds()))
	for i, id := range request.GetGroupIds() {
		groupIds[i] = uint(id)
	}

	groups, err := s.gr.GetBatchGetGroupInfoByIDs(groupIds)
	if err != nil {
		return resp, status.Error(codes.Code(code.GroupErrGetBatchGroupInfoByIDsFailed.Code()), err.Error())
	}

	// 将领域模型转换为 gRPC API 模型
	var groupAPIs []*v1.Group
	for _, group := range groups {
		groupAPI := &v1.Group{
			Id:              uint32(group.ID),
			Type:            int32(uint32(group.Type)),
			Status:          int32(uint32(group.Status)),
			MaxMembersLimit: int32(group.MaxMembersLimit),
			CreatorId:       group.CreatorID,
			Name:            group.Name,
			Avatar:          group.Avatar,
		}
		groupAPIs = append(groupAPIs, groupAPI)
	}

	resp.Groups = groupAPIs
	return resp, nil
}

func (s *Service) UpdateGroup(ctx context.Context, request *v1.UpdateGroupRequest) (*v1.Group, error) {
	resp := &v1.Group{}

	group := &entity.Group{
		Type:            entity.GroupType(request.GetGroup().Type),
		Status:          entity.GroupStatus(request.GetGroup().Status),
		MaxMembersLimit: int(request.GetGroup().MaxMembersLimit),
		CreatorID:       request.GetGroup().CreatorId,
		Name:            request.Group.Name,
		Avatar:          request.GetGroup().Avatar,
	}

	updatedGroup, err := s.gr.UpdateGroup(group)
	if err != nil {
		return resp, status.Error(codes.Code(code.GroupErrUpdateGroupFailed.Code()), err.Error())
	}

	// 将领域模型转换为 gRPC API 模型
	resp = &v1.Group{
		Id:              uint32(updatedGroup.ID),
		Type:            int32(uint32(updatedGroup.Type)),
		Status:          int32(uint32(updatedGroup.Status)),
		MaxMembersLimit: int32(updatedGroup.MaxMembersLimit),
		CreatorId:       updatedGroup.CreatorID,
		Name:            updatedGroup.Name,
		Avatar:          updatedGroup.Avatar,
	}
	return resp, nil
}

func (s *Service) InsertGroup(ctx context.Context, request *v1.InsertGroupRequest) (*v1.Group, error) {
	resp := &v1.Group{}

	group := &entity.Group{
		Type:            entity.GroupType(request.GetGroup().Type),
		Status:          entity.GroupStatus(request.GetGroup().Status),
		MaxMembersLimit: int(request.GetGroup().MaxMembersLimit),
		CreatorID:       request.GetGroup().CreatorId,
		Name:            request.Group.Name,
		Avatar:          request.GetGroup().Avatar,
	}

	createdGroup, err := s.gr.InsertGroup(group)
	if err != nil {
		return resp, status.Error(codes.Code(code.GroupErrInsertGroupFailed.Code()), err.Error())
	}

	// 将领域模型转换为 gRPC API 模型
	resp = &v1.Group{
		Id:              uint32(createdGroup.ID),
		Type:            int32(uint32(createdGroup.Type)),
		Status:          int32(uint32(createdGroup.Status)),
		MaxMembersLimit: int32(createdGroup.MaxMembersLimit),
		CreatorId:       createdGroup.CreatorID,
		Name:            createdGroup.Name,
		Avatar:          createdGroup.Avatar,
	}
	return resp, nil
}

func (s *Service) DeleteGroup(ctx context.Context, request *v1.DeleteGroupRequest) (*v1.EmptyResponse, error) {
	resp := &v1.EmptyResponse{}

	if err := s.gr.DeleteGroup(uint(request.GetGid())); err != nil {
		return resp, status.Error(codes.Code(code.GroupErrDeleteGroupFailed.Code()), err.Error())
	}

	return resp, nil
}
