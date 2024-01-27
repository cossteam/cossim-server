package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/pkg/code"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/cossim/coss-server/service/group/api/v1"
	"github.com/cossim/coss-server/service/group/domain/entity"
	"github.com/cossim/coss-server/service/group/domain/repository"
	"github.com/cossim/coss-server/service/group/infrastructure/persistence"
	"github.com/rs/xid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"log"
)

func NewService(repo *persistence.Repositories, ac pkgconfig.AppConfig) *Service {
	return &Service{
		gr:  repo.Gr,
		ac:  ac,
		sid: xid.New().String(),
	}
}

type Service struct {
	gr repository.GroupRepository
	v1.UnimplementedGroupServiceServer

	ac        pkgconfig.AppConfig
	discovery discovery.Discovery
	sid       string
}

func (s *Service) Start(discover bool) {
	if !discover {
		return
	}
	d, err := discovery.NewConsulRegistry(s.ac.Register.Addr())
	if err != nil {
		panic(err)
	}
	s.discovery = d
	if err = s.discovery.Register(s.ac.Register.Name, s.ac.GRPC.Addr(), s.sid); err != nil {
		panic(err)
	}
	log.Printf("Service registration successful ServiceName: %s  Addr: %s  ID: %s", s.ac.Register.Name, s.ac.GRPC.Addr(), s.sid)
}

func (s *Service) Close(discover bool) error {
	if !discover {
		return nil
	}
	if err := s.discovery.Cancel(s.sid); err != nil {
		log.Printf("Failed to cancel service registration: %v", err)
		return err
	}
	log.Printf("Service registration canceled ServiceName: %s  Addr: %s  ID: %s", s.ac.Register.Name, s.ac.GRPC.Addr(), s.sid)
	return nil
}

func (s *Service) GetGroupInfoByGid(ctx context.Context, request *v1.GetGroupInfoRequest) (*v1.Group, error) {
	resp := &v1.Group{}

	group, err := s.gr.GetGroupInfoByGid(uint(request.GetGid()))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.GroupErrGroupNotFound.Code()), err.Error())
		}
		return resp, status.Error(codes.Code(code.GroupErrGetGroupInfoByGidFailed.Code()), err.Error())
	}

	// 将领域模型转换为 gRPC API 模型
	resp = &v1.Group{
		Id:              uint32(group.ID),
		Type:            v1.GroupType(group.Type),
		Status:          v1.GroupStatus(group.Status),
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
			Type:            v1.GroupType(group.Type),
			Status:          v1.GroupStatus(group.Status),
			MaxMembersLimit: int32(group.MaxMembersLimit),
			CreatorId:       group.CreatorID,
			Name:            group.Name,
			Avatar:          group.Avatar,
		}
		groupAPIs = append(groupAPIs, groupAPI)
	}

	resp.Groups = groupAPIs

	fmt.Println("len => ", len(resp.Groups))
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
		Type:            v1.GroupType(group.Type),
		Status:          v1.GroupStatus(group.Status),
		MaxMembersLimit: int32(updatedGroup.MaxMembersLimit),
		CreatorId:       updatedGroup.CreatorID,
		Name:            updatedGroup.Name,
		Avatar:          updatedGroup.Avatar,
	}
	return resp, nil
}

func (s *Service) CreateGroup(ctx context.Context, request *v1.CreateGroupRequest) (*v1.Group, error) {
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
		Type:            v1.GroupType(group.Type),
		Status:          v1.GroupStatus(group.Status),
		MaxMembersLimit: int32(createdGroup.MaxMembersLimit),
		CreatorId:       createdGroup.CreatorID,
		Name:            createdGroup.Name,
		Avatar:          createdGroup.Avatar,
	}
	return resp, nil
}

func (s *Service) CreateGroupRevert(ctx context.Context, request *v1.CreateGroupRequest) (*v1.Group, error) {
	resp := &v1.Group{}
	if err := s.gr.DeleteGroup(uint(request.Group.Id)); err != nil {
		return resp, status.Error(codes.Code(code.GroupErrDeleteGroupFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) DeleteGroup(ctx context.Context, request *v1.DeleteGroupRequest) (*v1.EmptyResponse, error) {
	resp := &v1.EmptyResponse{}

	//return resp, status.Error(codes.Aborted, errors.New("测试回滚").Error())

	if err := s.gr.DeleteGroup(uint(request.GetGid())); err != nil {
		return resp, status.Error(codes.Aborted, err.Error())
	}
	return resp, nil
}

func (s *Service) DeleteGroupRevert(ctx context.Context, request *v1.DeleteGroupRequest) (*v1.EmptyResponse, error) {
	resp := &v1.EmptyResponse{}
	if err := s.gr.UpdateGroupByGroupID(uint(request.GetGid()), map[string]interface{}{
		"deleted_at": 0,
	}); err != nil {
		return resp, status.Error(codes.Code(code.GroupErrDeleteGroupFailed.Code()), err.Error())
	}
	return resp, nil
}
