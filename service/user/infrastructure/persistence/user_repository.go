package persistence

import (
	"github.com/cossim/coss-server/service/user/domain/entity"
	"gorm.io/gorm"
)

// UserRepo 需要实现UserRepository接口
type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// 根据邮箱获取用户信息
func (ur *UserRepo) GetUserInfoByEmail(email string) (*entity.User, error) {
	var user entity.User
	if err := ur.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// 根据uid获取用户信息
func (ur *UserRepo) GetUserInfoByUid(id string) (*entity.User, error) {
	var user entity.User
	if err := ur.db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// 修改用户信息
func (ur *UserRepo) UpdateUser(user *entity.User) (*entity.User, error) {
	if err := ur.db.Save(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// 添加用户
func (ur *UserRepo) InsertUser(user *entity.User) (*entity.User, error) {
	if err := ur.db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (ur *UserRepo) GetBatchGetUserInfoByIDs(userIds []string) ([]*entity.User, error) {
	var users []*entity.User

	if err := ur.db.Where("id IN ?", userIds).Find(&users).Error; err != nil {
		return nil, err
	}

	return users, nil
}

// GetUserPublicKey 获取用户公钥
func (ur *UserRepo) GetUserPublicKey(userId string) (string, error) {
	var user entity.User
	if err := ur.db.Where("id = ?", userId).First(&user).Error; err != nil {
		return "", err
	}
	return user.PublicKey, nil
}

// SetUserPublicKey 设置用户公钥
func (ur *UserRepo) SetUserPublicKey(userId, publicKey string) error {
	return ur.db.Model(entity.User{}).Where("id = ?", userId).Update("public_key", publicKey).Error
}
