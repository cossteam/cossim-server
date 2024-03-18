package grpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/pkg/code"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	"github.com/cossim/coss-server/pkg/version"
	"google.golang.org/grpc/health/grpc_health_v1"

	//"github.com/cossim/coss-interface/internal/group/api/grpc/v1"
	api "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	"github.com/cossim/coss-server/internal/group/domain/entity"
	"github.com/cossim/coss-server/internal/group/domain/repository"
	"github.com/cossim/coss-server/internal/group/infrastructure/persistence"
	"github.com/rs/xid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"strconv"
)

type Handler struct {
	gr  repository.GroupRepository
	ac  *pkgconfig.AppConfig
	sid string
	api.UnimplementedGroupServiceServer
}

func (s *Handler) Register(srv *grpc.Server) {
	api.RegisterGroupServiceServer(srv, s)
}

func (s *Handler) RegisterHealth(srv *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
}

func (s *Handler) Init(cfg *pkgconfig.AppConfig) error {
	mysql, err := db.NewMySQL(cfg.MySQL.Address, strconv.Itoa(cfg.MySQL.Port), cfg.MySQL.Username, cfg.MySQL.Password, cfg.MySQL.Database, int64(cfg.Log.Level), cfg.MySQL.Opts)
	if err != nil {
		return err
	}

	dbConn, err := mysql.GetConnection()
	if err != nil {
		return err
	}
	infra := persistence.NewRepositories(dbConn)
	if err = infra.Automigrate(); err != nil {
		return err
	}
	fmt.Println("初始化db")
	s.gr = infra.Gr
	s.ac = cfg
	s.sid = xid.New().String()
	return nil
}

func (s *Handler) Name() string {
	return "group_service"
}

func (s *Handler) Version() string {
	return version.FullVersion()
}

func (s *Handler) Stop(ctx context.Context) error { return nil }

func (s *Handler) DiscoverServices(services map[string]*grpc.ClientConn) error { return nil }

func (s *Handler) GetGroupInfoByGid(ctx context.Context, request *api.GetGroupInfoRequest) (*api.Group, error) {
	resp := &api.Group{}

	group, err := s.gr.GetGroupInfoByGid(uint(request.GetGid()))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.GroupErrGroupNotFound.Code()), err.Error())
		}
		return resp, status.Error(codes.Code(code.GroupErrGetGroupInfoByGidFailed.Code()), err.Error())
	}

	// 将领域模型转换为 gRPC API 模型
	resp = &api.Group{
		Id:              uint32(group.ID),
		Type:            api.GroupType(group.Type),
		Status:          api.GroupStatus(group.Status),
		MaxMembersLimit: int32(group.MaxMembersLimit),
		CreatorId:       group.CreatorID,
		Name:            group.Name,
		Avatar:          group.Avatar,
	}
	return resp, nil
}

func (s *Handler) GetBatchGroupInfoByIDs(ctx context.Context, request *api.GetBatchGroupInfoRequest) (*api.GetBatchGroupInfoResponse, error) {
	resp := &api.GetBatchGroupInfoResponse{}

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
	var groupAPIs []*api.Group
	for _, group := range groups {
		groupAPI := &api.Group{
			Id:              uint32(group.ID),
			Type:            api.GroupType(group.Type),
			Status:          api.GroupStatus(group.Status),
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

func (s *Handler) UpdateGroup(ctx context.Context, request *api.UpdateGroupRequest) (*api.Group, error) {
	resp := &api.Group{}

	group := &entity.Group{
		BaseModel: entity.BaseModel{
			ID: uint(request.Group.Id),
		},
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
	resp = &api.Group{
		Id:              uint32(updatedGroup.ID),
		Type:            api.GroupType(group.Type),
		Status:          api.GroupStatus(group.Status),
		MaxMembersLimit: int32(updatedGroup.MaxMembersLimit),
		CreatorId:       updatedGroup.CreatorID,
		Name:            updatedGroup.Name,
		Avatar:          updatedGroup.Avatar,
	}
	return resp, nil
}

func (s *Handler) CreateGroup(ctx context.Context, request *api.CreateGroupRequest) (*api.Group, error) {
	resp := &api.Group{}

	group := &entity.Group{
		Type:            entity.GroupType(request.GetGroup().Type),
		Status:          entity.GroupStatus(request.GetGroup().Status),
		MaxMembersLimit: int(request.GetGroup().MaxMembersLimit),
		CreatorID:       request.GetGroup().CreatorId,
		Name:            request.Group.Name,
		Avatar:          request.GetGroup().Avatar,
	}
	fmt.Println("gr ->", s.gr)

	createdGroup, err := s.gr.InsertGroup(group)
	if err != nil {
		return resp, status.Error(codes.Code(code.GroupErrInsertGroupFailed.Code()), err.Error())
	}

	// 将领域模型转换为 gRPC API 模型
	resp = &api.Group{
		Id:              uint32(createdGroup.ID),
		Type:            api.GroupType(group.Type),
		Status:          api.GroupStatus(group.Status),
		MaxMembersLimit: int32(createdGroup.MaxMembersLimit),
		CreatorId:       createdGroup.CreatorID,
		Name:            createdGroup.Name,
		Avatar:          createdGroup.Avatar,
	}
	return resp, nil
}

func (s *Handler) CreateGroupRevert(ctx context.Context, request *api.CreateGroupRequest) (*api.Group, error) {
	resp := &api.Group{}
	if err := s.gr.DeleteGroup(uint(request.Group.Id)); err != nil {
		return resp, status.Error(codes.Code(code.GroupErrDeleteGroupFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Handler) DeleteGroup(ctx context.Context, request *api.DeleteGroupRequest) (*api.EmptyResponse, error) {
	resp := &api.EmptyResponse{}

	//return resp, status.Error(codes.Aborted, errors.New("测试回滚").Error())

	if err := s.gr.DeleteGroup(uint(request.GetGid())); err != nil {
		return resp, status.Error(codes.Aborted, err.Error())
	}
	return resp, nil
}

func (s *Handler) DeleteGroupRevert(ctx context.Context, request *api.DeleteGroupRequest) (*api.EmptyResponse, error) {
	resp := &api.EmptyResponse{}
	if err := s.gr.UpdateGroupByGroupID(uint(request.GetGid()), map[string]interface{}{
		"deleted_at": 0,
	}); err != nil {
		return resp, status.Error(codes.Code(code.GroupErrDeleteGroupFailed.Code()), err.Error())
	}
	return resp, nil
}
