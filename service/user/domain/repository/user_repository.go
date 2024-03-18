package repository

import "github.com/cossim/coss-server/service/user/domain/entity"

type UserRepository interface {
	GetUserInfoByEmail(email string) (*entity.User, error)
	GetUserInfoByUid(id string) (*entity.User, error)
	GetUserInfoByCossID(cossId string) (*entity.User, error)
	UpdateUser(user *entity.User) (*entity.User, error)
	InsertUser(user *entity.User) (*entity.User, error)
	GetBatchGetUserInfoByIDs(userIds []string) ([]*entity.User, error)
	SetUserPublicKey(userId, publicKey string) error
	GetUserPublicKey(userId string) (string, error)
	SetUserSecretBundle(userId, secretBundle string) error
	GetUserSecretBundle(userId string) (string, error)
	UpdateUserColumn(userId string, column string, value interface{}) error
	InsertAndUpdateUser(user *entity.User) error
	DeleteUser(userId string) error
}
