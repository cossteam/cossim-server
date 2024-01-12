package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/service/user/domain/entity"
	"github.com/cossim/coss-server/service/user/domain/repository"
	"github.com/cossim/coss-server/service/user/utils"
	codes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"time"

	api "github.com/cossim/coss-server/service/user/api/v1"
)

func NewService(ur repository.UserRepository) *Service {
	return &Service{
		ur: ur,
	}
}

type Service struct {
	ur repository.UserRepository
	api.UnimplementedUserServiceServer
}

// 用户登录
func (g *Service) UserLogin(ctx context.Context, request *api.UserLoginRequest) (*api.UserLoginResponse, error) {
	userInfo, err := g.ur.GetUserInfoByEmail(request.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.Code(code.UserErrNotExistOrPassword.Code()), err.Error())
		}
		return nil, status.Error(codes.Code(code.UserErrLoginFailed.Code()), err.Error())
	}

	if userInfo.Password != request.Password {
		return nil, status.Error(codes.Code(code.UserErrNotExistOrPassword.Code()), code.UserErrNotExistOrPassword.Message())
	}
	//修改登录时间
	userInfo.LastAt = time.Now().Unix()
	_, err = g.ur.UpdateUser(userInfo)
	if err != nil {
		return nil, status.Error(codes.Code(code.UserErrLoginFailed.Code()), err.Error())
	}

	switch userInfo.Status {
	case entity.UserStatusLock:
		return nil, status.Error(codes.Code(code.UserErrLocked.Code()), code.UserErrLocked.Message())
	case entity.UserStatusDeleted:
		return nil, status.Error(codes.Code(code.UserErrDeleted.Code()), code.UserErrDeleted.Message())
	case entity.UserStatusDisable:
		return nil, status.Error(codes.Code(code.UserErrDisabled.Code()), code.UserErrDisabled.Message())
	case entity.UserStatusNormal:
		return &api.UserLoginResponse{
			UserId:   userInfo.ID,
			NickName: userInfo.NickName,
			Email:    userInfo.Email,
			Tel:      userInfo.Tel,
			Avatar:   userInfo.Avatar,
		}, nil
	default:
		return nil, status.Error(codes.Code(code.UserErrStatusException.Code()), code.UserErrStatusException.Message())
	}
}

// 用户注册
func (g *Service) UserRegister(ctx context.Context, request *api.UserRegisterRequest) (*api.UserRegisterResponse, error) {
	//添加用户
	_, err := g.ur.GetUserInfoByEmail(request.Email)
	if err == nil {
		return nil, status.Error(codes.Code(code.UserErrEmailAlreadyRegistered.Code()), err.Error())
	}
	userInfo, err := g.ur.InsertUser(&entity.User{
		Email:    request.Email,
		Password: request.Password,
		NickName: request.NickName,
		Avatar:   request.Avatar,
		//Status:   entity.UserStatusLock,
		Status: entity.UserStatusNormal,
		ID:     utils.GenUUid(),
	})
	if err != nil {
		return nil, status.Error(codes.Code(code.UserErrRegistrationFailed.Code()), err.Error())
	}

	return &api.UserRegisterResponse{
		UserId: userInfo.ID,
	}, nil
}

func (g *Service) UserInfo(ctx context.Context, request *api.UserInfoRequest) (*api.UserInfoResponse, error) {
	userInfo, err := g.ur.GetUserInfoByUid(request.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.Code(code.UserErrNotExistOrPassword.Code()), err.Error())
		}
		return nil, status.Error(codes.Code(code.UserErrGetUserInfoFailed.Code()), err.Error())
	}
	return &api.UserInfoResponse{
		UserId:    userInfo.ID,
		NickName:  userInfo.NickName,
		Email:     userInfo.Email,
		Tel:       userInfo.Tel,
		Avatar:    userInfo.Avatar,
		Signature: userInfo.Signature,
		Status:    api.UserStatus(userInfo.Status),
	}, nil
}

func (g *Service) GetBatchUserInfo(ctx context.Context, request *api.GetBatchUserInfoRequest) (*api.GetBatchUserInfoResponse, error) {
	resp := &api.GetBatchUserInfoResponse{}
	users, err := g.ur.GetBatchGetUserInfoByIDs(request.UserIds)
	if err != nil {
		fmt.Printf("无法获取用户列表信息: %v\n", err)
		return nil, status.Error(codes.Code(code.UserErrUnableToGetUserListInfo.Code()), err.Error())
	}
	for _, user := range users {
		resp.Users = append(resp.Users, &api.UserInfoResponse{
			UserId:    user.ID,
			NickName:  user.NickName,
			Email:     user.Email,
			Tel:       user.Tel,
			Avatar:    user.Avatar,
			Signature: user.Signature,
			Status:    api.UserStatus(user.Status),
		})
	}

	return resp, nil
}

func (g *Service) GetUserInfoByEmail(ctx context.Context, request *api.GetUserInfoByEmailRequest) (*api.UserInfoResponse, error) {
	userInfo, err := g.ur.GetUserInfoByEmail(request.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.Code(code.UserErrNotExistOrPassword.Code()), err.Error())
		}
		return nil, status.Error(codes.Code(code.UserErrGetUserInfoFailed.Code()), err.Error())
	}
	return &api.UserInfoResponse{
		UserId:    userInfo.ID,
		NickName:  userInfo.NickName,
		Email:     userInfo.Email,
		Tel:       userInfo.Tel,
		Avatar:    userInfo.Avatar,
		Signature: userInfo.Signature,
		Status:    api.UserStatus(userInfo.Status),
	}, nil
}

func (g *Service) GetUserPublicKey(ctx context.Context, in *api.UserRequest) (*api.GetUserPublicKeyResponse, error) {
	key, err := g.ur.GetUserPublicKey(in.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.Code(code.UserErrPublicKeyNotExist.Code()), err.Error())
		}
		return &api.GetUserPublicKeyResponse{}, status.Error(codes.Code(code.UserErrGetUserPublicKeyFailed.Code()), err.Error())
	}
	return &api.GetUserPublicKeyResponse{PublicKey: key}, nil
}

func (g *Service) SetUserPublicKey(ctx context.Context, in *api.SetPublicKeyRequest) (*api.UserResponse, error) {
	if err := g.ur.SetUserPublicKey(in.UserId, in.PublicKey); err != nil {
		return nil, status.Error(codes.Code(code.UserErrSaveUserPublicKeyFailed.Code()), err.Error())
	}
	return &api.UserResponse{UserId: in.UserId}, nil
}

//
//func (u *UserService) UpdateUserInfo(user *entity.User) (*entity.User, error) {
//	user, err := u.ur.UpdateUser(user)
//	if err != nil {
//		return nil, err
//	}
//	return user, nil
//}

//func (u *UserService) GetUserInfo(userId string) (*entity.User, error) {
//	user, err := u.ur.GetUserInfoByUid(userId)
//	if err != nil {
//		if errors.Is(err, gorm.ErrRecordNotFound) {
//			return nil, fmt.Errorf("未找到用户")
//		}
//		return nil, fmt.Errorf("获取用户信息失败: %w", err)
//	}
//
//	return user, nil
//}
