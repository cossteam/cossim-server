package interfaces

import (
	"context"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	api "github.com/cossim/coss-server/service/group/api/v1"
	"github.com/cossim/coss-server/service/group/domain/entity"
	"github.com/cossim/coss-server/service/group/domain/service"
	"github.com/cossim/coss-server/service/group/infrastructure/persistence"
)

type GrpcHandler struct {
	svc *service.GroupService
	api.UnimplementedGroupServiceServer
}

func NewGrpcHandler(c *config.AppConfig) *GrpcHandler {
	dbConn, err := db.NewMySQLFromDSN(c.MySQL.DSN).GetConnection()
	if err != nil {
		panic(err)
	}

	return &GrpcHandler{
		svc: service.NewGroupService(persistence.NewGroupRepo(dbConn)),
	}
}

func (g GrpcHandler) GetGroupInfoByGid(ctx context.Context, in *api.GetGroupInfoRequest) (*api.Group, error) {
	group, err := g.svc.GetGroupInfoByGid(uint(in.Gid))
	if err != nil {
		return nil, err
	}

	// 将领域模型转换为 gRPC API 模型
	groupAPI := &api.Group{
		Id:              uint32(group.ID),
		Type:            int32(uint32(group.Type)),
		Status:          int32(uint32(group.Status)),
		MaxMembersLimit: int32(group.MaxMembersLimit),
		CreatorId:       group.CreatorID,
		Name:            group.Name,
		Avatar:          group.Avatar,
	}

	return groupAPI, nil
}

func (g GrpcHandler) GetBatchGroupInfoByIDs(ctx context.Context, in *api.GetBatchGroupInfoRequest) (*api.GetBatchGroupInfoResponse, error) {
	if len(in.GroupIds) == 0 {
		return &api.GetBatchGroupInfoResponse{}, nil
	}
	//将uint32转成uint
	groupIds := make([]uint, len(in.GroupIds))
	for i, id := range in.GroupIds {
		groupIds[i] = uint(id)
	}
	groups, err := g.svc.GetBatchGetGroupInfoByIDs(groupIds)
	if err != nil {
		return nil, err
	}

	// 将领域模型转换为 gRPC API 模型
	var groupAPIs []*api.Group
	for _, group := range groups {
		groupAPI := &api.Group{
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

	response := &api.GetBatchGroupInfoResponse{
		Groups: groupAPIs,
	}

	return response, nil
}

func (g GrpcHandler) UpdateGroup(ctx context.Context, in *api.UpdateGroupRequest) (*api.Group, error) {
	group := &entity.Group{
		Type:            entity.GroupType(in.Group.Type),
		Status:          entity.GroupStatus(in.Group.Status),
		MaxMembersLimit: int(in.Group.MaxMembersLimit),
		CreatorID:       in.Group.CreatorId,
		Name:            in.Group.Name,
		Avatar:          in.Group.Avatar,
	}

	updatedGroup, err := g.svc.UpdateGroup(group)
	if err != nil {
		return nil, err
	}

	// 将领域模型转换为 gRPC API 模型
	updatedGroupAPI := &api.Group{
		Id:              uint32(updatedGroup.ID),
		Type:            int32(uint32(updatedGroup.Type)),
		Status:          int32(uint32(updatedGroup.Status)),
		MaxMembersLimit: int32(updatedGroup.MaxMembersLimit),
		CreatorId:       updatedGroup.CreatorID,
		Name:            updatedGroup.Name,
		Avatar:          updatedGroup.Avatar,
	}

	return updatedGroupAPI, nil
}

func (g GrpcHandler) InsertGroup(ctx context.Context, in *api.InsertGroupRequest) (*api.Group, error) {
	group := &entity.Group{
		Type:            entity.GroupType(in.Group.Id),
		Status:          entity.GroupStatus(in.Group.Status),
		MaxMembersLimit: int(in.Group.MaxMembersLimit),
		CreatorID:       in.Group.CreatorId,
		Name:            in.Group.Name,
		Avatar:          in.Group.Avatar,
	}

	createdGroup, err := g.svc.InsertGroup(group)
	if err != nil {
		return nil, err
	}

	// 将领域模型转换为 gRPC API 模型
	createdGroupAPI := &api.Group{
		Id:              uint32(createdGroup.ID),
		Type:            int32(uint32(createdGroup.Type)),
		Status:          int32(uint32(createdGroup.Status)),
		MaxMembersLimit: int32(createdGroup.MaxMembersLimit),
		CreatorId:       createdGroup.CreatorID,
		Name:            createdGroup.Name,
		Avatar:          createdGroup.Avatar,
	}

	return createdGroupAPI, nil
}

func (g GrpcHandler) DeleteGroup(ctx context.Context, in *api.DeleteGroupRequest) (*api.EmptyResponse, error) {
	err := g.svc.DeleteGroup(uint(in.Gid))
	if err != nil {
		return nil, err
	}

	return &api.EmptyResponse{}, nil
}
