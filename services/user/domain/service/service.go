package service

import (
	"fmt"
	api "github.com/cossim/coss-server/services/user/api/v1"
	"github.com/cossim/coss-server/services/user/domain/entity"
	"github.com/cossim/coss-server/services/user/domain/repository"
	"github.com/cossim/coss-server/services/user/utils"
	"gorm.io/gorm"
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
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("用户不存在或密码错误")
		}
		return nil, err
	}
	if user.Password != request.Password {
		return nil, fmt.Errorf("用户不存在或密码错误")
	}
	if user.Status == entity.UserStatusLock {
		return nil, fmt.Errorf("用户暂时被锁定,请先解锁")
	}
	return user, nil
}

func (g UserService) Register(request *api.UserRegisterRequest) (*entity.User, error) {
	//参数校验
	_, err := g.ur.GetUserInfoByEmail(request.Email)
	if err == nil {
		return nil, fmt.Errorf("邮箱已被注册")
	}
	user, err := g.ur.InsertUser(&entity.User{
		Email:    request.Email,
		Password: request.Password,
		NickName: request.NickName,
		Avatar:   request.Avatar,
		Status:   entity.UserStatusLock,
		ID:       utils.GenUUid(),
	})
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
