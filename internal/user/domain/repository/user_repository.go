package repository

import (
	"context"
	"github.com/cossim/coss-server/internal/user/domain/entity"
)

type UserRepository interface {
	//GetUserInfoByEmail(ctx context.Context, email string) (*entity.User, error)
	//GetUserInfoByUid(ctx context.Context, id string) (*entity.User, error)
	//GetUserInfoByCossID(ctx context.Context, cossId string) (*entity.User, error)
	//InsertUser(ctx context.Context, user *entity.User) (*entity.User, error)
	//GetBatchGetUserInfoByIDs(ctx context.Context, userIds []string) ([]*entity.User, error)
	//SetUserPublicKey(ctx context.Context, userId, publicKey string) error
	//GetUserPublicKey(ctx context.Context, userId string) (string, error)
	//SetUserSecretBundle(ctx context.Context, userId, secretBundle string) error
	//GetUserSecretBundle(ctx context.Context, userId string) (string, error)
	//UpdateUserColumn(ctx context.Context, userId string, column string, value interface{}) error
	//InsertAndUpdateUser(ctx context.Context, user *entity.User) error
	//
	//UpdateUserInfo(ctx context.Context, user *entity.UpdateUser) error
	//// UpdateUserStatus 更新用户状态
	//UpdateUserStatus(ctx context.Context, Param *entity.UpdateUserStatus, userId ...string) error
	//// UpdatePassword 更新用户密码
	//UpdatePassword(ctx context.Context, userID, password string) (string, error)

	SaveUser(ctx context.Context, user *entity.User) (*entity.User, error)
	DeleteUser(ctx context.Context, id string) error
	UpdateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	UpdatesUser(ctx context.Context, user *entity.User) (*entity.User, error)
	GetUser(ctx context.Context, userID string) (*entity.User, error)
	GetWithOptions(ctx context.Context, query *entity.User) (*entity.User, error)
	ListUser(ctx context.Context, query *entity.ListUserOptions) ([]*entity.User, error)
}
