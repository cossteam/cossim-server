package interfaces

import (
	"context"
	"im/services/user/api"
	"im/services/user/domain/service"
)

type GrpcHandler struct {
	svc service.UserService
}

func (g GrpcHandler) UserService(ctx context.Context, request *api.UserRequest) (*api.UserDetailResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (g GrpcHandler) UserRegister(ctx context.Context, request *api.UserRequest) (*api.UserCommonResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (g GrpcHandler) UserLogout(ctx context.Context, request *api.UserRequest) (*api.UserCommonResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (g GrpcHandler) mustEmbedUnimplementedUserServiceServer() {
	//TODO implement me
	panic("implement me")
}
