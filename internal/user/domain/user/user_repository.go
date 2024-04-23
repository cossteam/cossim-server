package user

import "context"

type UserLoginRepository interface {
	InsertUserLogin(ctx context.Context, user *UserLogin) error
	GetUserLoginByDriverIdAndUserId(ctx context.Context, driverId, userId string) (*UserLogin, error)
	UpdateUserLoginTokenByDriverId(ctx context.Context, driverId string, token string, userId string) error
	GetUserLoginByToken(ctx context.Context, token string) (*UserLogin, error)
	GetUserDriverTokenByUserId(ctx context.Context, userId string) ([]string, error)
	GetUserByUserId(ctx context.Context, userId string) (*UserLogin, error)
	DeleteUserLoginByID(ctx context.Context, id uint32) error
}

type UserRepository interface {
	GetUserInfoByEmail(ctx context.Context, email string) (*User, error)
	GetUserInfoByUid(ctx context.Context, id string) (*User, error)
	GetUserInfoByCossID(ctx context.Context, cossId string) (*User, error)
	UpdateUser(ctx context.Context, user *User) (*User, error)
	InsertUser(ctx context.Context, user *User) (*User, error)
	GetBatchGetUserInfoByIDs(ctx context.Context, userIds []string) ([]*User, error)
	SetUserPublicKey(ctx context.Context, userId, publicKey string) error
	GetUserPublicKey(ctx context.Context, userId string) (string, error)
	SetUserSecretBundle(ctx context.Context, userId, secretBundle string) error
	GetUserSecretBundle(ctx context.Context, userId string) (string, error)
	UpdateUserColumn(ctx context.Context, userId string, column string, value interface{}) error
	InsertAndUpdateUser(ctx context.Context, user *User) error
	DeleteUser(ctx context.Context, userId string) error
}
