package persistence

import (
	"gorm.io/gorm"
	"im/services/user/domain/entity"
)

// UserRepo 需要实现UserRepository接口
type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

var UserR *UserRepo

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
	return nil, nil
}

// 添加用户
func (ur *UserRepo) InsertUser(user *entity.User) (*entity.User, error) {
	if err := ur.db.Save(user).Error; err != nil {
		return nil, err
	}
	return nil, nil
}
