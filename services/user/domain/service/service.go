package service

import (
	"fmt"
	"im/services/user/api/v1"
	"im/services/user/domain/entity"
	"im/services/user/domain/repository"
	"im/services/user/utils"
)

type UserService struct {
	ur repository.UserRepository
}

func NewUserService(ur repository.UserRepository) *UserService {
	return &UserService{
		ur: ur,
	}
}

func (g UserService) Login(request *api.UserLoginRequest) (*entity.User, error) {
	user, err := g.ur.GetUserInfoByEmail(request.Email)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (g UserService) Register(request *api.UserRegisterRequest) (*entity.User, error) {

	user, err := g.ur.InsertUser(&entity.User{
		Email:    request.Email,
		Password: request.Password,
		NickName: request.NickName,
		Avatar:   request.Avatar,
		ID:       utils.GenUUid(),
	})
	fmt.Println("user =>", user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (g UserService) UpdateUserInfo(user *entity.User) (*entity.User, error) {
	user, err := g.ur.UpdateUser(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (g UserService) GetUserInfoByEmail(email string) (*entity.User, error) {
	user, err := g.ur.GetUserInfoByEmail(email)
	if err != nil {
		return nil, err
	}
	return user, nil
}
