package rpc

import (
	"context"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"google.golang.org/grpc"
)

type User struct {
	ID       string
	NickName string
}

type UserService interface {
	GetUserInfo(ctx context.Context, userID string) (*entity.UserInfo, error)
}

func NewUserGrpc(addr string) (UserService, error) {
	var grpcOptions = []grpc.DialOption{grpc.WithInsecure()}
	grpcOptions = append(grpcOptions, grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`))
	conn, err := grpc.Dial(
		addr,
		grpcOptions...,
	)
	if err != nil {
		return nil, err
	}

	return &UserGrpc{client: usergrpcv1.NewUserServiceClient(conn)}, nil
}

func NewUserGrpcWithClient(client usergrpcv1.UserServiceClient) UserService {
	return &UserGrpc{client: client}
}

type UserGrpc struct {
	client usergrpcv1.UserServiceClient
}

func (s *UserGrpc) GetUserInfo(ctx context.Context, userID string) (*entity.UserInfo, error) {
	r, err := s.client.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: userID})
	if err != nil {
		return nil, err
	}

	return &entity.UserInfo{
		UserID:    r.UserId,
		CossID:    r.CossId,
		Nickname:  r.NickName,
		Email:     r.Email,
		Tel:       r.Tel,
		Avatar:    r.Avatar,
		Signature: r.Signature,
		Status:    entity.UserStatus(r.Status),
	}, nil
}
