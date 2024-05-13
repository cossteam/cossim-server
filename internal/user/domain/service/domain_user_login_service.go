package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/domain/repository"
	"github.com/cossim/coss-server/pkg/code"
	"gorm.io/gorm"
)

type UserLoginDomain interface {
	Get(ctx context.Context, userID string) (*entity.UserLogin, error)
	List(ctx context.Context, userID string) ([]*entity.UserLogin, error)
	Create(ctx context.Context, user *entity.UserLogin) error
	Delete(ctx context.Context, id uint32) error

	GetByUserIDAndDriverID(ctx context.Context, userID, driverID string) (*entity.UserLogin, error)

	DeleteByUserIDAndDriverID(ctx context.Context, userID, driverID string) error

	// LastLoginTime 获取用户上次登录时间
	LastLoginTime(ctx context.Context, userID string) (int64, error)
	// IsLoginRestricted 判断是否被限制登录
	IsLoginRestricted(ctx context.Context, userID string) error
	// IsNewDeviceLogin 判断是否是新设备登录
	IsNewDeviceLogin(ctx context.Context, userID, deviceID string) (bool, error)
}

// TODO UserLoginDomain的IsLoginRestricted、LastLoginTime后期可以删除、使用聚合根、值对象解决

var _ UserLoginDomain = &userLoginDomain{}

type userLoginDomain struct {
	ur        repository.UserRepository
	ulr       repository.UserLoginRepository
	userCache cache.UserCache

	multipleDeviceLimit    bool
	multipleDeviceLimitMax int
}

func NewUserLoginDomain(ur repository.UserRepository, ulr repository.UserLoginRepository, userCache cache.UserCache, multipleDeviceLimit bool, multipleDeviceLimitMax int) UserLoginDomain {
	return &userLoginDomain{ur: ur, ulr: ulr, userCache: userCache, multipleDeviceLimit: multipleDeviceLimit, multipleDeviceLimitMax: multipleDeviceLimitMax}
}

func (d *userLoginDomain) DeleteByUserIDAndDriverID(ctx context.Context, userID, driverID string) error {
	r, err := d.ulr.GetUserLoginByDriverIdAndUserId(ctx, driverID, userID)
	if err != nil {
		return err
	}
	return d.ulr.DeleteUserLoginByID(ctx, uint32(r.ID))
}

func (d *userLoginDomain) GetByUserIDAndDriverID(ctx context.Context, userID, driverID string) (*entity.UserLogin, error) {
	return d.ulr.GetUserLoginByDriverIdAndUserId(ctx, driverID, userID)
}

func (d *userLoginDomain) Delete(ctx context.Context, uint32 uint32) error {
	if err := d.ulr.DeleteUserLoginByID(ctx, uint32); err != nil {
		return err
	}

	return nil
}

func (d *userLoginDomain) Create(ctx context.Context, ul *entity.UserLogin) error {
	if err := d.ulr.InsertUserLogin(ctx, ul); err != nil {
		return err
	}

	return nil
}

func (d *userLoginDomain) List(ctx context.Context, userID string) ([]*entity.UserLogin, error) {
	infos, err := d.userCache.GetUserLoginInfos(ctx, userID)
	if err != nil {
		return nil, err
	}

	return infos, nil
}

func (d *userLoginDomain) IsNewDeviceLogin(ctx context.Context, userID, deviceID string) (bool, error) {
	info, err := d.ulr.GetUserLoginByDriverIdAndUserId(ctx, deviceID, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return true, nil
		}
		return false, err
	}

	return info.DriverID == "", nil
}

func (d *userLoginDomain) LastLoginTime(ctx context.Context, userID string) (int64, error) {
	info, err := d.ulr.GetUserByUserId(ctx, userID)
	if err != nil {
		return 0, err
	}

	return info.LastAt, nil
}

func (d *userLoginDomain) Get(ctx context.Context, userID string) (*entity.UserLogin, error) {
	info, err := d.ulr.GetUserByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	return info, nil
}

func (d *userLoginDomain) IsLoginRestricted(ctx context.Context, userID string) error {
	user, err := d.ur.GetUser(ctx, userID)
	if err != nil {
		return err
	}

	if user.Status != entity.UserStatusNormal {
		return code.UserErrStatusException.CustomMessage(fmt.Sprintf("无法登录，用户状态为：%s", user.Status.String()))
	}

	infos, err := d.userCache.GetUserLoginInfos(ctx, userID)
	if err != nil {
		return err
	}

	if d.multipleDeviceLimit && len(infos) >= d.multipleDeviceLimitMax {
		return code.MyCustomErrorCode.CustomMessage("登录设备超出限制")
	}

	return nil
}
