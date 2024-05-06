package repository

import (
	"context"
	"github.com/cossim/coss-server/internal/user/domain/entity"
)

type UserLoginRepository interface {
	InsertUserLogin(ctx context.Context, user *entity.UserLogin) error
	GetUserLoginByDriverIdAndUserId(ctx context.Context, driverId, userId string) (*entity.UserLogin, error)
	UpdateUserLoginTokenByDriverId(ctx context.Context, driverId string, token string, userId string) error
	GetUserLoginByToken(ctx context.Context, token string) (*entity.UserLogin, error)
	GetUserDriverTokenByUserId(ctx context.Context, userId string) ([]string, error)
	GetUserByUserId(ctx context.Context, userId string) (*entity.UserLogin, error)
	DeleteUserLoginByID(ctx context.Context, id uint32) error
}
