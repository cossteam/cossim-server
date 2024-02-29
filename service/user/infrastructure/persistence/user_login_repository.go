package persistence

import (
	"github.com/cossim/coss-server/service/user/domain/entity"
	"gorm.io/gorm"
)

// UserLoginRepo 需要实现UserLoginRepository接口
type UserLoginRepo struct {
	db *gorm.DB
}

func NewUserLoginRepo(db *gorm.DB) *UserLoginRepo {
	return &UserLoginRepo{db: db}
}

func (u UserLoginRepo) InsertUserLogin(user *entity.UserLogin) error {
	return u.db.Where(entity.UserLogin{UserId: user.UserId, DriverId: user.DriverId}).Assign(entity.UserLogin{LoginCount: user.LoginCount, DriverToken: user.DriverToken, Token: user.Token}).FirstOrCreate(&user).Error
}

func (u UserLoginRepo) GetUserLoginByToken(token string) (*entity.UserLogin, error) {
	var user *entity.UserLogin
	err := u.db.Where("token = ?", token).First(&user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u UserLoginRepo) GetUserLoginByDriverIdAndUserId(driverId, userId string) (*entity.UserLogin, error) {
	var user *entity.UserLogin
	err := u.db.Where("driver_id = ? AND user_id = ?", driverId, userId).First(&user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u UserLoginRepo) UpdateUserLoginTokenByDriverId(driverId string, token string, userId string) error {
	return u.db.Where("driver_id = ? AND user_id = ?", driverId, userId).Update("token", token).Error
}

func (u UserLoginRepo) GetUserDriverTokenByUserId(userId string) ([]string, error) {
	var driverTokens []string
	err := u.db.Where("user_id = ?", userId).Pluck("driver_token", &driverTokens).Error
	if err != nil {
		return nil, err
	}
	return driverTokens, err
}

func (u UserLoginRepo) GetUserByUserId(userId string) (*entity.UserLogin, error) {
	var user *entity.UserLogin
	err := u.db.Where("user_id = ?", userId).First(&user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}
