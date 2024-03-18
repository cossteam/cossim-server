package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/pkg/code"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	pkgtime "github.com/cossim/coss-server/pkg/utils/time"
	"github.com/cossim/coss-server/pkg/version"
	api "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/cossim/coss-server/service/user/domain/entity"
	"github.com/cossim/coss-server/service/user/domain/repository"
	"github.com/cossim/coss-server/service/user/infrastructure/persistence"
	"github.com/cossim/coss-server/service/user/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"strconv"
)

type Service struct {
	ac  *pkgconfig.AppConfig
	ur  repository.UserRepository
	ulr repository.UserLoginRepository
	api.UnimplementedUserServiceServer
	api.UnimplementedUserLoginServiceServer
}

func (s *Service) Init(cfg *pkgconfig.AppConfig) error {
	mysql, err := db.NewMySQL(cfg.MySQL.Address, strconv.Itoa(cfg.MySQL.Port), cfg.MySQL.Username, cfg.MySQL.Password, cfg.MySQL.Database, int64(cfg.Log.Level), cfg.MySQL.Opts)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}

	dbConn, err := mysql.GetConnection()
	if err != nil {
		return err
	}

	infra := persistence.NewRepositories(dbConn)
	if err = infra.Automigrate(); err != nil {
		return err
	}

	s.ur = infra.UR
	s.ulr = infra.ULR
	s.ac = cfg
	return nil
}

func (s *Service) Name() string {
	//TODO implement me
	return "user_service"
}

func (s *Service) Version() string { return version.FullVersion() }

func (s *Service) Register(srv *grpc.Server) {
	api.RegisterUserServiceServer(srv, s)
	api.RegisterUserLoginServiceServer(srv, s)
}

func (s *Service) RegisterHealth(srv *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
}

func (s *Service) Stop(ctx context.Context) error { return nil }

func (s *Service) DiscoverServices(services map[string]*grpc.ClientConn) error { return nil }

// 用户登录
func (s *Service) UserLogin(ctx context.Context, request *api.UserLoginRequest) (*api.UserLoginResponse, error) {
	resp := &api.UserLoginResponse{}
	userInfo := &entity.User{}
	userInfo, err := s.ur.GetUserInfoByEmail(request.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userInfo, err = s.ur.GetUserInfoByCossID(request.Email)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return resp, status.Error(codes.Code(code.UserErrNotExistOrPassword.Code()), code.UserErrNotExistOrPassword.Message())
				}
				return resp, status.Error(codes.Code(code.UserErrLoginFailed.Code()), err.Error())
			}

			//return resp, status.Error(codes.Code(code.UserErrNotExistOrPassword.Code()), err.Error())
		}
	}

	if userInfo.Password != request.Password {
		return resp, status.Error(codes.Code(code.UserErrNotExistOrPassword.Code()), code.UserErrNotExistOrPassword.Message())
	}
	//修改登录时间
	userInfo.LastAt = pkgtime.Now()
	_, err = s.ur.UpdateUser(userInfo)
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
		}, nil
	default:
		return resp, status.Error(codes.Code(code.UserErrStatusException.Code()), code.UserErrStatusException.Message())
	}
}

// 用户注册
func (s *Service) UserRegister(ctx context.Context, request *api.UserRegisterRequest) (*api.UserRegisterResponse, error) {
	resp := &api.UserRegisterResponse{}
	//添加用户
	_, err := s.ur.GetUserInfoByEmail(request.Email)
	if err == nil {
		//return resp, status.Error(codes.Code(code.UserErrEmailAlreadyRegistered.Code()), code.UserErrEmailAlreadyRegistered.Message())
		return resp, status.Error(codes.Aborted, code.UserErrEmailAlreadyRegistered.Message())
	}
	userInfo, err := s.ur.InsertUser(&entity.User{
		Email:     request.Email,
		Password:  request.Password,
		NickName:  request.NickName,
		Avatar:    request.Avatar,
		PublicKey: request.PublicKey,
		//Action:   entity.UserStatusLock,
		Status: entity.UserStatusNormal,
		ID:     utils.GenUUid(),
	})
	if err != nil {
		//return resp, status.Error(codes.Code(code.UserErrRegistrationFailed.Code()), err.Error())
		return resp, status.Error(codes.Aborted, err.Error())
	}
	resp.UserId = userInfo.ID
	return resp, nil
}

func (s *Service) UserInfo(ctx context.Context, request *api.UserInfoRequest) (*api.UserInfoResponse, error) {
	resp := &api.UserInfoResponse{}
	userInfo, err := s.ur.GetUserInfoByUid(request.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			//return resp, status.Error(codes.Code(code.UserErrNotExistOrPassword.Code()), err.Error())
			return resp, status.Error(codes.Aborted, err.Error())
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

func (s *Service) GetBatchUserInfo(ctx context.Context, request *api.GetBatchUserInfoRequest) (*api.GetBatchUserInfoResponse, error) {
	resp := &api.GetBatchUserInfoResponse{}
	users, err := s.ur.GetBatchGetUserInfoByIDs(request.UserIds)
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

func (s *Service) GetUserInfoByEmail(ctx context.Context, request *api.GetUserInfoByEmailRequest) (*api.UserInfoResponse, error) {
	resp := &api.UserInfoResponse{}
	userInfo := &entity.User{}
	userInfo, err := s.ur.GetUserInfoByEmail(request.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			userInfo, err = s.ur.GetUserInfoByCossID(request.Email)
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
	}
	return resp, nil
}

func (s *Service) GetUserPublicKey(ctx context.Context, in *api.UserRequest) (*api.GetUserPublicKeyResponse, error) {
	key, err := s.ur.GetUserPublicKey(in.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &api.GetUserPublicKeyResponse{}, status.Error(codes.Code(code.UserErrPublicKeyNotExist.Code()), err.Error())
		}
		return &api.GetUserPublicKeyResponse{}, status.Error(codes.Code(code.UserErrGetUserPublicKeyFailed.Code()), err.Error())
	}
	return &api.GetUserPublicKeyResponse{PublicKey: key}, nil
}

func (s *Service) SetUserPublicKey(ctx context.Context, in *api.SetPublicKeyRequest) (*api.UserResponse, error) {
	if err := s.ur.SetUserPublicKey(in.UserId, in.PublicKey); err != nil {
		return &api.UserResponse{}, status.Error(codes.Code(code.UserErrSaveUserPublicKeyFailed.Code()), err.Error())
	}
	return &api.UserResponse{UserId: in.UserId}, nil
}

func (s *Service) ModifyUserInfo(ctx context.Context, in *api.User) (*api.UserResponse, error) {
	resp := &api.UserResponse{}
	user, err := s.ur.UpdateUser(&entity.User{
		ID:        in.UserId,
		NickName:  in.NickName,
		Avatar:    in.Avatar,
		CossID:    in.CossId,
		Signature: in.Signature,
		Status:    entity.UserStatus(in.Status),
		Tel:       in.Tel,
	})
	if err != nil {
		return resp, err
	}
	resp = &api.UserResponse{UserId: user.ID}
	return resp, nil
}

func (s *Service) ModifyUserPassword(ctx context.Context, in *api.ModifyUserPasswordRequest) (*api.UserResponse, error) {
	resp := &api.UserResponse{}
	user, err := s.ur.UpdateUser(&entity.User{
		ID:       in.UserId,
		Password: in.Password,
	})
	if err != nil {
		return resp, err
	}
	resp = &api.UserResponse{UserId: user.ID}
	return resp, nil
}

func (s *Service) GetUserPasswordByUserId(ctx context.Context, in *api.UserRequest) (*api.GetUserPasswordByUserIdResponse, error) {
	resp := &api.GetUserPasswordByUserIdResponse{}
	userInfo, err := s.ur.GetUserInfoByUid(in.UserId)
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

func (s *Service) SetUserSecretBundle(ctx context.Context, in *api.SetUserSecretBundleRequest) (*api.SetUserSecretBundleResponse, error) {
	var resp = &api.SetUserSecretBundleResponse{}
	if err := s.ur.SetUserSecretBundle(in.UserId, in.SecretBundle); err != nil {
		return resp, status.Error(codes.Code(code.UserErrSetUserSecretBundleFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) GetUserSecretBundle(ctx context.Context, in *api.GetUserSecretBundleRequest) (*api.GetUserSecretBundleResponse, error) {
	var resp = &api.GetUserSecretBundleResponse{}
	secretBundle, err := s.ur.GetUserSecretBundle(in.UserId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.UserErrNotExist.Code()), err.Error())
		}
		return resp, status.Error(codes.Code(code.UserErrGetUserSecretBundleFailed.Code()), err.Error())
	}
	resp.SecretBundle = secretBundle
	resp.UserId = in.UserId
	return resp, nil
}

func (s *Service) ActivateUser(ctx context.Context, in *api.UserRequest) (*api.UserResponse, error) {
	var resp = &api.UserResponse{UserId: in.UserId}
	if err := s.ur.UpdateUserColumn(in.UserId, "email_verity", entity.Activated); err != nil {
		return resp, status.Error(codes.Code(code.UserErrActivateUserFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) CreateUser(ctx context.Context, in *api.CreateUserRequest) (*api.CreateUserResponse, error) {
	resp := &api.CreateUserResponse{}
	if err := s.ur.InsertAndUpdateUser(&entity.User{
		NickName:  in.NickName,
		Email:     in.Email,
		Password:  in.Password,
		Avatar:    in.Avatar,
		Status:    entity.UserStatus(in.Status),
		ID:        in.UserId,
		PublicKey: in.PublicKey,
		Bot:       uint(in.IsBot),
	}); err != nil {
		return resp, status.Error(codes.Code(code.UserErrCreateUserFailed.Code()), err.Error())
	}
	resp.UserId = in.UserId
	return resp, nil
}

func (s *Service) CreateUserRollback(ctx context.Context, in *api.CreateUserRollbackRequest) (*api.CreateUserRollbackResponse, error) {
	resp := &api.CreateUserRollbackResponse{}
	if err := s.ur.DeleteUser(in.UserId); err != nil {
		return resp, status.Error(codes.Code(code.UserErrCreateUserRollbackFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) GetUserInfoByCossId(ctx context.Context, in *api.GetUserInfoByCossIdlRequest) (*api.UserInfoResponse, error) {
	resp := &api.UserInfoResponse{}
	if userInfo, err := s.ur.GetUserInfoByCossID(in.CossId); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return resp, status.Error(codes.Code(code.UserErrNotExist.Code()), err.Error())
		}
		return resp, status.Error(codes.Code(code.UserErrNotExist.Code()), err.Error())
	} else {
		resp = &api.UserInfoResponse{
			UserId: userInfo.ID,
		}
	}
	return resp, nil
}
