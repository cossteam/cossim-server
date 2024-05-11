package service

import (
	"context"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/domain/repository"
	"github.com/google/uuid"
)

type UserDomain interface {
	CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	DeleteUser(ctx context.Context, id string) error
	UpdateUser(ctx context.Context, user *entity.User, partialUpdate bool) (*entity.User, error)
	UpdatePassword(ctx context.Context, userID, password string) (string, error)
	GetUser(ctx context.Context, id string) (*entity.User, error)
	GetUsers(ctx context.Context) ([]entity.User, error)
	GetUserWithOpts(ctx context.Context, opts ...entity.UserOpt) (*entity.User, error)

	UserRegister(ctx context.Context, ur *entity.UserRegister) (string, error)
	ActivateUser(ctx context.Context, userID ...string) error
}

// userDomain 实现了 UserDomain 接口
type userDomain struct {
	ur repository.UserRepository
}

func (d *userDomain) UserRegister(ctx context.Context, ur *entity.UserRegister) (string, error) {
	uid := uuid.New().String()
	_, err := d.ur.InsertUser(ctx, &entity.User{
		ID:        uid,
		Email:     ur.Email,
		Password:  ur.Password,
		NickName:  ur.NickName,
		Avatar:    ur.Avatar,
		PublicKey: ur.PublicKey,
		Status:    entity.UserStatusNormal,
	})
	if err != nil {
		return uid, err
	}
	return uid, nil
}

func (d *userDomain) ActivateUser(ctx context.Context, userID ...string) error {
	emailVerity := true
	return d.ur.UpdateUserStatus(ctx, &entity.UpdateUserStatus{EmailVerity: &emailVerity}, userID...)
}

// NewUserDomain 返回 UserDomain 接口的实例
func NewUserDomain(ur repository.UserRepository) UserDomain {
	return &userDomain{ur: ur}
}

// CreateUser 创建用户
func (d *userDomain) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	return d.ur.InsertUser(ctx, user)
}

// DeleteUser 删除用户
func (d *userDomain) DeleteUser(ctx context.Context, id string) error {
	return d.ur.DeleteUser(ctx, id)
}

// UpdateUser 更新用户信息
func (d *userDomain) UpdateUser(ctx context.Context, user *entity.User, partialUpdate bool) (*entity.User, error) {
	if partialUpdate {
		// 如果是部分更新，调用 UserRepository 的 UpdateUser 方法
		return d.ur.UpdateUser(ctx, user)
	}
	// 否则，调用 UserRepository 的 InsertAndUpdateUser 方法
	return nil, d.ur.InsertAndUpdateUser(ctx, user)
}

// UpdatePassword 更新用户密码
func (d *userDomain) UpdatePassword(ctx context.Context, userID, password string) (string, error) {
	return d.ur.UpdatePassword(ctx, userID, password)
}

// GetUser 获取用户信息
func (d *userDomain) GetUser(ctx context.Context, id string) (*entity.User, error) {
	return d.ur.GetUserInfoByUid(ctx, id)
}

// GetUsers 获取所有用户信息
func (d *userDomain) GetUsers(ctx context.Context) ([]entity.User, error) {
	// 实现 GetUsers 的逻辑
	return nil, nil
}

// GetUserWithOpts 根据选项获取用户信息
func (d *userDomain) GetUserWithOpts(ctx context.Context, opts ...entity.UserOpt) (*entity.User, error) {
	user := new(entity.User)

	// Apply the options to the user object
	entity.UserOpts(opts).Apply(user)

	return d.ur.GetWithOptions(ctx, user)
}
