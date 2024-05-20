package service

import (
	"context"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/domain/service"
	"go.uber.org/zap"
)

type Service interface {
	CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	DeleteUser(ctx context.Context, id string) error
	UpdateUser(ctx context.Context, user *entity.User, partialUpdate bool) (*entity.User, error)
	UpdatePassword(ctx context.Context, id string) (string, error)
	GetUserinfo(ctx context.Context, username string) (*entity.User, error)
	GetUserList(ctx context.Context) ([]entity.User, error)
	GetUserDetail(ctx context.Context, id string) (*entity.User, error)
	UserLogin(ctx context.Context, ul *entity.LoginRequest) (*entity.LoginRequest, error)
}

var _ Service = &serviceImpl{}

type serviceImpl struct {
	ad  service.AuthDomain
	ud  service.UserDomain
	uld service.UserLoginDomain

	logger *zap.Logger
}

func (s *serviceImpl) UserLogin(ctx context.Context, ul *entity.LoginRequest) (*entity.LoginRequest, error) {
	//user, err := s.ud.GetUserWithOpts(
	//	ctx,
	//	entity.WithEmail(ul.Email),
	//	entity.WithPassword(ul.Password),
	//)
	//if err != nil {
	//	s.logger.Error("获取用户信息失败", zap.Error(err))
	//	return nil, fmt.Errorf("login failed: username '%s' and password do not match", ul.Email)
	//}
	//
	//// 登录是否受限，例如账户未激活、达到设备限制等
	//if err := s.uld.IsLoginRestricted(ctx, user.ID); err != nil {
	//	return nil, err
	//}
	//
	//token, err := s.ad.GenerateUserToken(ctx, &entity.AuthClaims{
	//	ID:    user.ID,
	//	Email:     user.Email,
	//	DriverID:  ul.DriverID,
	//	PublicKey: user.PublicKey,
	//})
	//if err != nil {
	//	s.logger.Error("生成用户token失败", zap.Error(err))
	//	return nil, err
	//}
	//
	//lastLoginTime, err := s.uld.LastLoginTime(ctx, user.ID)
	//if err != nil {
	//	s.logger.Error("获取用户最近一次登录时间失败", zap.Error(err))
	//	return nil, err
	//}
	//
	//isNewDevice, err := s.uld.IsNewDeviceLogin(ctx, user.ID, ul.DriverID)
	//if err != nil {
	//	return nil, err
	//}

	return nil, nil
}

func (s *serviceImpl) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	return s.ud.CreateUser(ctx, user)
}

func (s *serviceImpl) DeleteUser(ctx context.Context, id string) error {
	return s.ud.DeleteUser(ctx, id)
}

func (s *serviceImpl) UpdateUser(ctx context.Context, user *entity.User, partialUpdate bool) (*entity.User, error) {
	return s.ud.UpdateUser(ctx, user, partialUpdate)
}

func (s *serviceImpl) UpdatePassword(ctx context.Context, id string) (string, error) {
	//return s.ud.UpdatePassword(ctx, id)
	return "", nil
}

func (s *serviceImpl) GetUserinfo(ctx context.Context, username string) (*entity.User, error) {
	//return s.ud.GetUserinfo(ctx, username)
	return nil, nil
}

func (s *serviceImpl) GetUserList(ctx context.Context) ([]entity.User, error) {
	//TODO implement me
	panic("implement me")
}

func (s *serviceImpl) GetUserDetail(ctx context.Context, id string) (*entity.User, error) {
	//TODO implement me
	panic("implement me")
}
