package rpc

import (
	"context"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"google.golang.org/grpc"
)

type User struct {
	ID        string
	CossID    string
	Nickname  string
	Email     string
	Avatar    string
	Tel       string
	Signature string
	Status    uint
}

type UserService interface {
	GetUserInfo(ctx context.Context, userID string) (*User, error)
	GetUsersInfo(ctx context.Context, userIDs []string) (map[string]*User, error)
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

func (s *UserGrpc) GetUsersInfo(ctx context.Context, userIDs []string) (map[string]*User, error) {
	req := &usergrpcv1.GetBatchUserInfoRequest{UserIds: userIDs}
	resp, err := s.client.GetBatchUserInfo(ctx, req)
	if err != nil {
		return nil, err
	}

	users := make(map[string]*User, len(resp.Users))
	for _, userInfo := range resp.Users {
		users[userInfo.UserId] = &User{
			ID:        userInfo.UserId,
			CossID:    userInfo.CossId,
			Nickname:  userInfo.NickName,
			Email:     userInfo.Email,
			Avatar:    userInfo.Avatar,
			Tel:       userInfo.Tel,
			Signature: userInfo.Signature,
			Status:    uint(userInfo.Status),
		}
	}

	return users, nil
}

func (s *UserGrpc) GetUserInfo(ctx context.Context, userID string) (*User, error) {
	r, err := s.client.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: userID})
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        r.UserId,
		CossID:    r.CossId,
		Nickname:  r.NickName,
		Email:     r.Email,
		Avatar:    r.Avatar,
		Tel:       r.Tel,
		Signature: r.Signature,
		Status:    uint(r.Status),
	}, nil
}
