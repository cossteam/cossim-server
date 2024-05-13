package service

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/domain/repository"
	"github.com/google/uuid"
	"strings"
)

type UserDomain interface {
	CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	DeleteUser(ctx context.Context, id string) error
	UpdateUser(ctx context.Context, user *entity.User, partialUpdate bool) (*entity.User, error)
	GetUser(ctx context.Context, id string) (*entity.User, error)
	GetUsers(ctx context.Context) ([]entity.User, error)
	GetUserWithOpts(ctx context.Context, opts ...entity.UserOpt) (*entity.User, error)

	UpdatePassword(ctx context.Context, userID, password string) (string, error)
	UpdateBundle(ctx context.Context, userID, bundle string) error
	UserRegister(ctx context.Context, ur *entity.UserRegister) (string, error)
	ActivateUser(ctx context.Context, userID ...string) error
	SetUserPublicKey(ctx context.Context, id string, publickey string) error
}

// userDomain 实现了 UserDomain 接口
type userDomain struct {
	ur repository.UserRepository
	//ur repository.UserRepositoryBase
}

func (d *userDomain) SetUserPublicKey(ctx context.Context, userID string, publickey string) error {
	_, err := d.ur.UpdateUser(ctx, &entity.User{ID: userID, PublicKey: publickey})
	if err != nil {
		return err
	}

	return nil
	//return d.ur.UpdateUserColumn(ctx, userID, "public_key", publickey)
}

func (d *userDomain) UpdateBundle(ctx context.Context, userID, bundle string) error {
	_, err := d.ur.UpdateUser(ctx, &entity.User{ID: userID, SecretBundle: bundle})
	if err != nil {
		return err
	}

	return nil
	//return d.ur.UpdateUserColumn(ctx, userID, "secret_bundle", bundle)
}

func (d *userDomain) UserRegister(ctx context.Context, ur *entity.UserRegister) (string, error) {
	uid := uuid.New().String()
	e := &entity.User{
		ID:        uid,
		Email:     ur.Email,
		Password:  ur.Password,
		NickName:  ur.NickName,
		Avatar:    ur.Avatar,
		PublicKey: ur.PublicKey,
		Status:    entity.UserStatusNormal,
	}
	_, err := d.ur.SaveUser(ctx, e)
	if err != nil {
		return "", err
	}

	return uid, nil
}

func (d *userDomain) ActivateUser(ctx context.Context, userID ...string) error {
	for _, id := range userID {
		_, err := d.ur.UpdateUser(ctx, &entity.User{
			ID:          id,
			EmailVerity: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// NewUserDomain 返回 UserDomain 接口的实例
func NewUserDomain(ur repository.UserRepository) UserDomain {
	return &userDomain{ur: ur}
}

// CreateUser 创建用户
func (d *userDomain) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	u, err := d.ur.SaveUser(ctx, user)
	if err != nil {
		if strings.Contains(err.Error(), "Duplicate entry") {
			return nil, fmt.Errorf("email ‘%s’ already exists", user.Email)
		}
		return nil, err
	}
	return u, nil
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
	return d.ur.UpdatesUser(ctx, user)
}

// UpdatePassword 更新用户密码
func (d *userDomain) UpdatePassword(ctx context.Context, userID, password string) (string, error) {
	user, err := d.ur.UpdateUser(ctx, &entity.User{ID: userID, Password: password})
	if err != nil {
		return "", err
	}
	return user.Password, nil
}

// GetUser 获取用户信息
func (d *userDomain) GetUser(ctx context.Context, id string) (*entity.User, error) {
	return d.ur.GetUser(ctx, id)
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
