package service

import (
	"errors"
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

func (u *UserService) Login(request *api.UserLoginRequest) (*entity.User, error) {
	user, err := u.ur.GetUserInfoByEmail(request.Email)
	if err == gorm.ErrRecordNotFound || user.Password != request.Password {
		return nil, fmt.Errorf("用户不存在或密码错误")
	}
	if err != nil {
		return nil, err
	}

	switch user.Status {
	case entity.UserStatusLock:
		return nil, fmt.Errorf("用户暂时被锁定，请先解锁")
	case entity.UserStatusDeleted:
		return nil, fmt.Errorf("用户已被删除")
	case entity.UserStatusDisable:
		return nil, fmt.Errorf("用户已被禁用")
	case entity.UserStatusNormal:
		return user, nil
	default:
		return nil, fmt.Errorf("用户状态异常")
	}
}

func (u *UserService) Register(request *api.UserRegisterRequest) (*entity.User, error) {
	//参数校验
	_, err := u.ur.GetUserInfoByEmail(request.Email)
	if err == nil {
		return nil, fmt.Errorf("邮箱已被注册")
	}
	user, err := u.ur.InsertUser(&entity.User{
		Email:    request.Email,
		Password: request.Password,
		NickName: request.NickName,
		Avatar:   request.Avatar,
		//Status:   entity.UserStatusLock,
		Status: entity.UserStatusNormal,
		ID:     utils.GenUUid(),
	})
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserService) UpdateUserInfo(user *entity.User) (*entity.User, error) {
	user, err := u.ur.UpdateUser(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserService) GetUserInfo(userId string) (*entity.User, error) {
	user, err := u.ur.GetUserInfoByUid(userId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("未找到用户")
		}
		return nil, fmt.Errorf("获取用户信息失败: %w", err)
	}

	return user, nil
}

func (u *UserService) GetBatchGetUserInfo(userIds []string) ([]*entity.User, error) {
	users, err := u.ur.GetBatchGetUserInfoByIDs(userIds)
	if err != nil {
		fmt.Printf("无法获取用户列表信息: %v\n", err)
		return nil, fmt.Errorf("无法获取用户列表信息")
	}
	return users, nil
}
