package repository

import "github.com/cossim/coss-server/internal/user/domain/entity"

type UserLoginRepository interface {
	InsertUserLogin(user *entity.UserLogin) error
	GetUserLoginByDriverIdAndUserId(driverId, userId string) (*entity.UserLogin, error)
	UpdateUserLoginTokenByDriverId(driverId string, token string, userId string) error
	GetUserLoginByToken(token string) (*entity.UserLogin, error)
	GetUserDriverTokenByUserId(userId string) ([]string, error)
	GetUserLoginById(id uint32) (*entity.UserLogin, error)
	DeleteUserLoginByID(id uint32) error
	GetUserByUserId(userId string) (*entity.UserLogin, error)
}
