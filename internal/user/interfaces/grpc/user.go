package grpc

import (
	"context"
	"errors"
	"fmt"
	api "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/pkg/code"
	pkgtime "github.com/cossim/coss-server/pkg/utils/time"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

// 用户登录
func (s *UserServiceServer) UserLogin(ctx context.Context, request *api.UserLoginRequest) (*api.UserLoginResponse, error) {
	resp := &api.UserLoginResponse{}
	userInfo := &entity.User{}
	userInfo, err := s.ur.GetWithOptions(ctx, &entity.User{Email: request.Email})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userInfo, err = s.ur.GetWithOptions(ctx, &entity.User{CossID: userInfo.Email})
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return resp, status.Error(codes.Code(code.UserErrNotExistOrPassword.Code()), code.UserErrNotExistOrPassword.Message())
				}
				return resp, status.Error(codes.Code(code.UserErrLoginFailed.Code()), err.Error())
			}

		}
	}

	if userInfo.Password != request.Password {
		return resp, status.Error(codes.Code(code.UserErrNotExistOrPassword.Code()), code.UserErrNotExistOrPassword.Message())
	}
	//修改登录时间
	userInfo.LastAt = pkgtime.Now()
	_, err = s.ur.UpdateUser(ctx, userInfo)
	if err != nil {
		return resp, status.Error(codes.Code(code.UserErrLoginFailed.Code()), err.Error())
	}
	switch userInfo.Status {
	case entity.UserStatusLock:
		return resp, status.Error(codes.Code(code.UserErrLocked.Code()), code.UserErrLocked.Message())
	case entity.UserStatusDeleted:
		return resp, status.Error(codes.Code(code.UserErrDeleted.Code()), code.UserErrDeleted.Message())
	case entity.UserStatusDisable:
		return resp, status.Error(codes.Code(code.UserErrDisabled.Code()), code.UserErrDisabled.Message())
	case entity.UserStatusNormal:
		return &api.UserLoginResponse{
			UserId:    userInfo.ID,
			NickName:  userInfo.NickName,
			Email:     userInfo.Email,
			Tel:       userInfo.Tel,
			Avatar:    userInfo.Avatar,
			Signature: userInfo.Signature,
			CossId:    userInfo.CossID,
			PublicKey: userInfo.PublicKey,
		}, nil
	default:
		return resp, status.Error(codes.Code(code.UserErrStatusException.Code()), code.UserErrStatusException.Message())
	}
}

// 用户注册
func (s *UserServiceServer) UserRegister(ctx context.Context, request *api.UserRegisterRequest) (*api.UserRegisterResponse, error) {
	resp := &api.UserRegisterResponse{}
	//添加用户
	_, err := s.ur.GetWithOptions(ctx, &entity.User{Email: request.Email})
	if err == nil {
		return resp, status.Error(codes.Aborted, code.UserErrEmailAlreadyRegistered.Message())
	}

	var stats = entity.UserStatusNormal
	if s.ac.Email.Enable {
		stats = entity.UserStatusLock
	}

	userInfo, err := s.ur.SaveUser(ctx, &entity.User{
		ID:        uuid.New().String(),
		Email:     request.Email,
		Password:  request.Password,
		NickName:  request.NickName,
		Avatar:    request.Avatar,
		PublicKey: request.PublicKey,
		Status:    stats,
	})
	if err != nil {
		return resp, status.Error(codes.Aborted, err.Error())
	}
	resp.UserId = userInfo.ID
	return resp, nil
}

func (s *UserServiceServer) UserInfo(ctx context.Context, request *api.UserInfoRequest) (*api.UserInfoResponse, error) {
	resp := &api.UserInfoResponse{}

	userInfo, err := s.ur.GetWithOptions(ctx, &entity.User{
		ID: request.UserId,
	})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.UserErrNotExist.Code()), err.Error())
		}
		return resp, status.Error(codes.Code(code.UserErrGetUserInfoFailed.Code()), err.Error())
	}
	resp = &api.UserInfoResponse{
		UserId:    userInfo.ID,
		NickName:  userInfo.NickName,
		Email:     userInfo.Email,
		Tel:       userInfo.Tel,
		Avatar:    userInfo.Avatar,
		Signature: userInfo.Signature,
		Status:    api.UserStatus(userInfo.Status),
		CossId:    userInfo.CossID,
	}

	return resp, nil
}

func (s *UserServiceServer) GetBatchUserInfo(ctx context.Context, request *api.GetBatchUserInfoRequest) (*api.GetBatchUserInfoResponse, error) {
	resp := &api.GetBatchUserInfoResponse{}
	if len(request.UserIds) == 0 {
		return resp, nil
	}

	users, err := s.ur.ListUser(ctx, &entity.ListUserOptions{
		UserID: request.UserIds,
	})
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
			CossId:    user.CossID,
		})
	}

	return resp, nil
}

func (s *UserServiceServer) GetUserInfoByEmail(ctx context.Context, request *api.GetUserInfoByEmailRequest) (*api.UserInfoResponse, error) {
	resp := &api.UserInfoResponse{}
	userInfo, err := s.ur.GetWithOptions(ctx, &entity.User{Email: request.Email})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userInfo, err = s.ur.GetWithOptions(ctx, &entity.User{CossID: request.Email})
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return resp, status.Error(codes.Code(code.UserErrNotExistOrPassword.Code()), err.Error())
				}
				return resp, status.Error(codes.Code(code.UserErrGetUserInfoFailed.Code()), err.Error())
			}
		}
	}

	resp = &api.UserInfoResponse{
		UserId:    userInfo.ID,
		NickName:  userInfo.NickName,
		Email:     userInfo.Email,
		Tel:       userInfo.Tel,
		Avatar:    userInfo.Avatar,
		Signature: userInfo.Signature,
		Status:    api.UserStatus(userInfo.Status),
		CossId:    userInfo.CossID,
		PublicKey: userInfo.PublicKey,
	}
	return resp, nil
}

func (s *UserServiceServer) GetUserPublicKey(ctx context.Context, request *api.UserRequest) (*api.GetUserPublicKeyResponse, error) {
	user, err := s.ur.GetUser(ctx, request.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &api.GetUserPublicKeyResponse{}, status.Error(codes.Code(code.UserErrPublicKeyNotExist.Code()), err.Error())
		}
		return &api.GetUserPublicKeyResponse{}, status.Error(codes.Code(code.UserErrGetUserPublicKeyFailed.Code()), err.Error())
	}
	return &api.GetUserPublicKeyResponse{PublicKey: user.PublicKey}, nil
}

func (s *UserServiceServer) SetUserPublicKey(ctx context.Context, request *api.SetPublicKeyRequest) (*api.UserResponse, error) {
	_, err := s.ur.UpdateUser(ctx, &entity.User{ID: request.UserId, PublicKey: request.PublicKey})
	if err != nil {
		return nil, err
	}
	//if err := s.ur.SetUserPublicKey(ctx, request.UserId, request.PublicKey); err != nil {
	//	return &api.UserResponse{}, status.Error(codes.Code(code.UserErrSaveUserPublicKeyFailed.Code()), err.Error())
	//}
	return &api.UserResponse{UserId: request.UserId}, nil
}

func (s *UserServiceServer) ModifyUserInfo(ctx context.Context, request *api.User) (*api.UserResponse, error) {
	resp := &api.UserResponse{}
	user, err := s.ur.UpdateUser(ctx, &entity.User{
		ID:        request.UserId,
		NickName:  request.NickName,
		Avatar:    request.Avatar,
		CossID:    request.CossId,
		Signature: request.Signature,
		Status:    entity.UserStatus(request.Status),
		Tel:       request.Tel,
	})
	if err != nil {
		return resp, err
	}
	resp = &api.UserResponse{UserId: user.ID}

	return resp, nil
}

func (s *UserServiceServer) ModifyUserPassword(ctx context.Context, request *api.ModifyUserPasswordRequest) (*api.UserResponse, error) {
	resp := &api.UserResponse{}
	user, err := s.ur.UpdateUser(ctx, &entity.User{
		ID:       request.UserId,
		Password: request.Password,
	})
	if err != nil {
		return resp, err
	}
	resp.UserId = user.ID
	return resp, nil
}

func (s *UserServiceServer) GetUserPasswordByUserId(ctx context.Context, request *api.UserRequest) (*api.GetUserPasswordByUserIdResponse, error) {
	resp := &api.GetUserPasswordByUserIdResponse{}
	userInfo, err := s.ur.GetUser(ctx, request.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.UserErrNotExistOrPassword.Code()), err.Error())
		}
		return resp, status.Error(codes.Code(code.UserErrGetUserInfoFailed.Code()), err.Error())
	}
	resp = &api.GetUserPasswordByUserIdResponse{
		Password: userInfo.Password,
		UserId:   userInfo.ID,
	}
	return resp, nil
}

func (s *UserServiceServer) SetUserSecretBundle(ctx context.Context, request *api.SetUserSecretBundleRequest) (*api.SetUserSecretBundleResponse, error) {
	var resp = &api.SetUserSecretBundleResponse{}
	if _, err := s.ur.UpdateUser(ctx, &entity.User{
		ID:           request.UserId,
		SecretBundle: request.SecretBundle,
	}); err != nil {
		return resp, status.Error(codes.Code(code.UserErrSetUserSecretBundleFailed.Code()), err.Error())
	}
	//if err := s.ur.SetUserSecretBundle(ctx, request.UserId, request.SecretBundle); err != nil {
	//	return resp, status.Error(codes.Code(code.UserErrSetUserSecretBundleFailed.Code()), err.Error())
	//}
	return resp, nil
}

func (s *UserServiceServer) GetUserSecretBundle(ctx context.Context, request *api.GetUserSecretBundleRequest) (*api.GetUserSecretBundleResponse, error) {
	var resp = &api.GetUserSecretBundleResponse{}
	user, err := s.ur.GetUser(ctx, request.UserId)
	if err != nil {
		return nil, err
	}
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.UserErrNotExist.Code()), err.Error())
		}
		return resp, status.Error(codes.Code(code.UserErrGetUserSecretBundleFailed.Code()), err.Error())
	}
	resp.SecretBundle = user.SecretBundle
	resp.UserId = user.ID
	return resp, nil
}

func (s *UserServiceServer) ActivateUser(ctx context.Context, request *api.UserRequest) (*api.UserResponse, error) {
	var resp = &api.UserResponse{UserId: request.UserId}
	if _, err := s.ur.UpdateUser(ctx, &entity.User{
		ID:          request.UserId,
		EmailVerity: true,
	}); err != nil {
		return resp, status.Error(codes.Code(code.UserErrActivateUserFailed.Code()), err.Error())
	}
	//if err := s.ur.UpdateUserColumn(ctx, request.UserId, "email_verity", entity.Activated); err != nil {
	//	return resp, status.Error(codes.Code(code.UserErrActivateUserFailed.Code()), err.Error())
	//}
	return resp, nil
}

func (s *UserServiceServer) CreateUser(ctx context.Context, request *api.CreateUserRequest) (*api.CreateUserResponse, error) {
	resp := &api.CreateUserResponse{}
	//if err := s.ur.InsertAndUpdateUser(ctx, &entity.User{
	//	NickName:  request.NickName,
	//	Email:     request.Email,
	//	Password:  request.Password,
	//	Avatar:    request.Avatar,
	//	Status:    entity.UserStatus(request.Status),
	//	ID:        request.UserId,
	//	PublicKey: request.PublicKey,
	//	Bot:       uint(request.IsBot),
	//}); err != nil {
	//	return resp, status.Error(codes.Code(code.UserErrCreateUserFailed.Code()), err.Error())
	//}
	//resp.UserId = request.UserId
	return resp, nil
}

func (s *UserServiceServer) CreateUserRollback(ctx context.Context, request *api.CreateUserRollbackRequest) (*api.CreateUserRollbackResponse, error) {
	resp := &api.CreateUserRollbackResponse{}
	if err := s.ur.DeleteUser(ctx, request.UserId); err != nil {
		return resp, status.Error(codes.Code(code.UserErrCreateUserRollbackFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *UserServiceServer) GetUserInfoByCossId(ctx context.Context, request *api.GetUserInfoByCossIdlRequest) (*api.UserInfoResponse, error) {
	resp := &api.UserInfoResponse{}
	if userInfo, err := s.ur.GetWithOptions(ctx, &entity.User{CossID: request.CossId}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.UserErrNotExist.Code()), err.Error())
		}
		return resp, status.Error(codes.Code(code.UserErrGetUserInfoFailed.Code()), err.Error())
	} else {
		resp = &api.UserInfoResponse{
			UserId: userInfo.ID,
		}
	}
	return resp, nil
}
