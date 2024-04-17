package adapters

import (
	"context"
	"github.com/cossim/coss-server/internal/group/app/command"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
)

//var _ command.UserService = &UserGrpc{}

func NewUserGrpc(client usergrpcv1.UserServiceClient) *UserGrpc {
	return &UserGrpc{client: client}
}

type UserGrpc struct {
	client usergrpcv1.UserServiceClient
}

func (s *UserGrpc) GetUserInfo(ctx context.Context, userID string) (*command.User, error) {
	resp, err := s.client.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: userID})
	if err != nil {
		return nil, err
	}

	return &command.User{
		ID:       resp.UserId,
		NickName: resp.NickName,
	}, nil
}
