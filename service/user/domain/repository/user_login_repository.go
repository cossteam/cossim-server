package repository

import "github.com/cossim/coss-server/service/user/domain/entity"

type UserLoginRepository interface {
	InsertUserLogin(user *entity.UserLogin) error
	GetUserLoginByDriverIdAndUserId(driverId, userId string) (*entity.UserLogin, error)
	UpdateUserLoginTokenByDriverId(driverId string, token string, userId string) error
	GetUserLoginByToken(token string) (*entity.UserLogin, error)
	GetUserDriverTokenByUserId(userId string) ([]string, error)

	GetUserByUserId(userId string) (*entity.UserLogin, error)
}
