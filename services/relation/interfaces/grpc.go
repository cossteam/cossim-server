package interfaces

import (
	"context"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	api "github.com/cossim/coss-server/services/relation/api/v1"
	"github.com/cossim/coss-server/services/relation/domain/service"
	"github.com/cossim/coss-server/services/relation/infrastructure/persistence"
)

func NewGrpcHandler(c *config.AppConfig) *GrpcHandler {
	dbConn, err := db.NewMySQLFromDSN(c.MySQL.DSN).GetConnection()
	if err != nil {
		panic(err)
	}

	return &GrpcHandler{
		svc: service.NewUserService(persistence.NewUserRelationRepo(dbConn)),
	}
}

type GrpcHandler struct {
	svc *service.UserService
	api.UnimplementedRelationServiceServer
}

func (g *GrpcHandler) AddFriend(ctx context.Context, request *api.AddFriendRequest) (*api.AddFriendResponse, error) {
	resp := &api.AddFriendResponse{}
	if _, err := g.svc.AddFriend(ctx, request.GetUserId(), request.GetFriendId()); err != nil {
		return nil, err
	}
	return resp, nil
}

func (g *GrpcHandler) ConfirmFriend(ctx context.Context, request *api.ConfirmFriendRequest) (*api.ConfirmFriendResponse, error) {
	resp := &api.ConfirmFriendResponse{}
	if _, err := g.svc.ConfirmFriend(ctx, request.GetUserId(), request.GetFriendId()); err != nil {
		return nil, err
	}
	return resp, nil
}

func (g *GrpcHandler) DeleteFriend(ctx context.Context, request *api.DeleteFriendRequest) (*api.DeleteFriendResponse, error) {
	resp := &api.DeleteFriendResponse{}
	if _, err := g.svc.DeleteFriend(ctx, request.GetUserId(), request.GetFriendId()); err != nil {
		return nil, err
	}
	return resp, nil
}

func (g *GrpcHandler) AddBlacklist(ctx context.Context, request *api.AddBlacklistRequest) (*api.AddBlacklistResponse, error) {
	resp := &api.AddBlacklistResponse{}
	if _, err := g.svc.AddBlacklist(ctx, request.GetUserId(), request.GetFriendId()); err != nil {
		return nil, err
	}
	return resp, nil
}

func (g *GrpcHandler) DeleteBlacklist(ctx context.Context, request *api.DeleteBlacklistRequest) (*api.DeleteBlacklistResponse, error) {
	resp := &api.DeleteBlacklistResponse{}
	if _, err := g.svc.DeleteBlacklist(ctx, request.GetUserId(), request.GetFriendId()); err != nil {
		return nil, err
	}
	return resp, nil
}

func (g *GrpcHandler) GetFriendList(ctx context.Context, request *api.GetFriendListRequest) (*api.GetFriendListResponse, error) {
	resp := &api.GetFriendListResponse{} // 初始化 resp
	friends, err := g.svc.GetFriendList(ctx, request.GetUserId())
	if err != nil {
		return nil, err
	}

	for _, friend := range friends {
		resp.FriendList = append(resp.FriendList, &api.Friend{UserId: friend.FriendID})
	}

	return resp, nil
}

func (g *GrpcHandler) GetBlacklist(ctx context.Context, request *api.GetBlacklistRequest) (*api.GetBlacklistResponse, error) {
	resp := &api.GetBlacklistResponse{} // 初始化 resp
	blacks, err := g.svc.GetBlacklist(ctx, request.GetUserId())
	if err != nil {
		return nil, err
	}

	for _, black := range blacks {
		resp.Blacklist = append(resp.Blacklist, &api.Blacklist{UserId: black.UserID})
	}

	return resp, nil
}
