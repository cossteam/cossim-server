package persistence

import (
	"github.com/cossim/coss-server/internal/user/domain/entity"
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
	if err := ur.db.Where("email = ? AND deleted_at = 0", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// 根据uid获取用户信息
func (ur *UserRepo) GetUserInfoByUid(id string) (*entity.User, error) {
	var user entity.User
	if err := ur.db.Where("id = ? AND deleted_at = 0", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// 修改用户信息
func (ur *UserRepo) UpdateUser(user *entity.User) (*entity.User, error) {
	if err := ur.db.Model(&entity.User{}).Where("id = ?", user.ID).Updates(user).Error; err != nil {
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
	return ur.db.Model(&entity.User{}).Where("id = ?", userId).UpdateColumn("public_key", publicKey).Error
}

// 设置用户密钥包
func (ur *UserRepo) SetUserSecretBundle(userId, secretBundle string) error {
	return ur.db.Model(&entity.User{}).Where("id = ?", userId).Update("secret_bundle", secretBundle).Error
}

// 获取用户密钥包
func (ur *UserRepo) GetUserSecretBundle(userId string) (string, error) {
	var user entity.User
	err := ur.db.Model(&entity.User{}).Where("id = ?", userId).Select("secret_bundle").First(&user).Error
	if err != nil {
		return "", err
	}
	return user.SecretBundle, nil
}

func (ur *UserRepo) UpdateUserColumn(userId string, column string, value interface{}) error {
	if err := ur.db.Model(&entity.User{}).Where("id = ?", userId).UpdateColumn(column, value).Error; err != nil {
		return err
	}
	return nil
}

func (ur *UserRepo) InsertAndUpdateUser(user *entity.User) error {
	return ur.db.Where(entity.User{ID: user.ID}).Assign(entity.User{NickName: user.NickName, Password: user.Password, Email: user.Email, Avatar: user.Avatar}).FirstOrCreate(&user).Error
}

func (ur *UserRepo) DeleteUser(userId string) error {
	return ur.db.Delete(&entity.User{ID: userId}).Error
}

func (ur *UserRepo) GetUserInfoByCossID(cossId string) (*entity.User, error) {
	var user entity.User
	return &user, ur.db.Where("coss_id = ?", cossId).First(&user).Error
}
