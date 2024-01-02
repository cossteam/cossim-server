package service

import (
	"context"
	"im/services/user/domain/entity"
	"im/services/user/domain/repository"
)

type UserService struct {
	ur repository.UserRepository
}

func (g UserService) Login(ctx context.Context, name string, ss string) (*entity.User, error) {
	//TODO implement me
	panic("implement me")
}

func (g UserService) Register(ctx context.Context) (*entity.User, error) {
	//TODO implement me
	panic("implement me")
}

func (g UserService) Logout(ctx context.Context) (*entity.User, error) {
	//TODO implement me
	panic("implement me")
}

func (g UserService) mustEmbedUnimplementedUserServiceServer() {
	//TODO implement me
	panic("implement me")
}
