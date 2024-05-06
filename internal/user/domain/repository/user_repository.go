package repository

import (
	"context"
	"github.com/cossim/coss-server/internal/user/domain/entity"
)

type UserRepository interface {
	GetUserInfoByEmail(ctx context.Context, email string) (*entity.User, error)
	GetUserInfoByUid(ctx context.Context, id string) (*entity.User, error)
	GetUserInfoByCossID(ctx context.Context, cossId string) (*entity.User, error)
	UpdateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	InsertUser(ctx context.Context, user *entity.User) (*entity.User, error)
	GetBatchGetUserInfoByIDs(ctx context.Context, userIds []string) ([]*entity.User, error)
	SetUserPublicKey(ctx context.Context, userId, publicKey string) error
	GetUserPublicKey(ctx context.Context, userId string) (string, error)
	SetUserSecretBundle(ctx context.Context, userId, secretBundle string) error
	GetUserSecretBundle(ctx context.Context, userId string) (string, error)
	UpdateUserColumn(ctx context.Context, userId string, column string, value interface{}) error
	InsertAndUpdateUser(ctx context.Context, user *entity.User) error
	DeleteUser(ctx context.Context, userId string) error
}
