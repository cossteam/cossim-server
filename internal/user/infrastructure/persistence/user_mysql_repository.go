package persistence

import (
	"context"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/user"
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
	"log"
)

type UserModel struct {
	ID           string `gorm:"type:varchar(64);primary_key;comment:用户id" json:"id"`
	CossID       string `gorm:"type:varchar(64);"`
	Email        string `gorm:"type:varchar(100);uniqueIndex;comment:邮箱" json:"email"`
	Tel          string `gorm:"type:varchar(50);comment:联系电话" json:"tel"`
	NickName     string `gorm:"comment:昵称" json:"nickname"`
	Avatar       string `gorm:"type:longtext;comment:头像" json:"avatar"`
	PublicKey    string `gorm:"comment:用户pgp公钥" json:"public_key,omitempty"`
	Password     string `gorm:"type:varchar(50);comment:登录密码" json:"password,omitempty"`
	LastIp       string `gorm:"type:varchar(20);comment:最后登录IP" json:"last_ip"`
	LineIp       string `gorm:"type:varchar(20);comment:最后在线IP（接口）" json:"line_ip"`
	CreatedIp    string `gorm:"type:varchar(20);comment:注册IP" json:"created_ip"`
	Signature    string `gorm:"type:varchar(255);comment:个性签名" json:"signature"`
	LineAt       int64  `gorm:"comment:最后在线时间（接口）" json:"line_at"`
	LastAt       int64  `gorm:"comment:最后登录时间" json:"last_at"`
	Status       uint   `gorm:"type:tinyint(4);default:0;comment:用户状态" json:"status"`
	EmailVerity  uint   `gorm:"type:tinyint(1);default:0;comment:邮箱是否已验证" json:"email_verity"`
	Bot          uint   `gorm:"type:tinyint(4);default:0;comment:是否机器人" json:"bot"`
	SecretBundle string `gorm:"type:longtext;comment:用户密钥" json:"secret_bundle,omitempty"`
	CreatedAt    int64  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt    int64  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt    int64  `gorm:"default:0;comment:删除时间" json:"deleted_at"`
}

func (m *UserModel) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	m.CreatedAt = now
	m.UpdatedAt = now
	cid, err := utils.GenCossID()
	if err != nil {
		return err
	}
	m.CossID = cid

	return nil
}

func (m *UserModel) BeforeUpdate(tx *gorm.DB) error {
	m.UpdatedAt = time.Now()
	return nil
}

func (m *UserModel) TableName() string {
	return "users"
}

func (m *UserModel) FromEntity(u *user.User) error {
	m.ID = u.ID
	m.CossID = u.CossID
	m.Email = u.Email
	m.Tel = u.Tel
	m.NickName = u.NickName
	m.Avatar = u.Avatar
	m.PublicKey = u.PublicKey
	m.Password = u.Password
	m.LastIp = u.LastIp
	m.LineIp = u.LineIp
	m.CreatedIp = u.CreatedIp
	m.Signature = u.Signature
	m.LineAt = u.LineAt
	m.LastAt = u.LastAt
	m.Status = uint(u.Status)
	m.EmailVerity = u.EmailVerity
	m.Bot = u.Bot
	m.SecretBundle = u.SecretBundle
	m.CreatedAt = u.CreatedAt
	m.UpdatedAt = u.UpdatedAt
	m.DeletedAt = u.DeletedAt
	return nil
}

func (m *UserModel) ToEntity() (*user.User, error) {
	return &user.User{
		ID:           m.ID,
		CossID:       m.CossID,
		Email:        m.Email,
		Tel:          m.Tel,
		NickName:     m.NickName,
		Avatar:       m.Avatar,
		PublicKey:    m.PublicKey,
		Password:     m.Password,
		LastIp:       m.LastIp,
		LineIp:       m.LineIp,
		CreatedIp:    m.CreatedIp,
		Signature:    m.Signature,
		LineAt:       m.LineAt,
		LastAt:       m.LastAt,
		Status:       user.UserStatus(m.Status),
		EmailVerity:  m.EmailVerity,
		Bot:          m.Bot,
		SecretBundle: m.SecretBundle,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
		DeletedAt:    m.DeletedAt,
	}, nil
}

var _ user.UserRepository = &MySQLUserRepository{}

func NewMySQLUserRepository(db *gorm.DB, cache cache.UserCache) *MySQLUserRepository {
	return &MySQLUserRepository{
		db:    db,
		cache: cache,
	}
}

type MySQLUserRepository struct {
	db    *gorm.DB
	cache cache.UserCache
}

func (r *MySQLUserRepository) GetUserInfoByEmail(ctx context.Context, email string) (*user.User, error) {
	var model UserModel

	if err := r.db.WithContext(ctx).Where("email = ? AND deleted_at = 0", email).First(&model).Error; err != nil {
		return nil, err
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (r *MySQLUserRepository) GetUserInfoByUid(ctx context.Context, id string) (*user.User, error) {
	var model UserModel

	if r.cache != nil {
		info, err := r.cache.GetUserInfo(ctx, id)
		if err == nil && info != nil {
			return info, nil
		}
	}

	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at = 0", id).First(&model).Error; err != nil {
		return nil, err
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, err
	}

	if r.cache != nil {
		if err := r.cache.SetUserInfo(ctx, entity.ID, entity, cache.UserExpireTime); err != nil {
			log.Println("cache set user info error:", utils.NewErrorWithStack(err.Error()))
		}
	}

	return entity, nil
}

func (r *MySQLUserRepository) GetUserInfoByCossID(ctx context.Context, cossId string) (*user.User, error) {
	var model UserModel

	if err := r.db.WithContext(ctx).Where("coss_id = ?", cossId).First(&model).Error; err != nil {
		return nil, err
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, err
	}

	if r.cache != nil {
		if err := r.cache.SetUserInfo(ctx, entity.ID, entity, cache.UserExpireTime); err != nil {
			log.Println("cache set user info error:", utils.NewErrorWithStack(err.Error()))
		}
	}

	return entity, nil
}

func (r *MySQLUserRepository) UpdateUser(ctx context.Context, user *user.User) (*user.User, error) {
	var model UserModel
	if err := model.FromEntity(user); err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Where("id = ?", user.ID).Updates(&model).Error; err != nil {
		return nil, err
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, err
	}

	if r.cache != nil {
		if err := r.cache.DeleteUsersInfo(ctx, []string{entity.ID}); err != nil {
			log.Println("cache set user info error:", utils.NewErrorWithStack(err.Error()))
		}
	}

	return entity, nil
}

func (r *MySQLUserRepository) InsertUser(ctx context.Context, user *user.User) (*user.User, error) {
	var model UserModel
	if err := model.FromEntity(user); err != nil {
		return nil, err
	}

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, err
	}

	if r.cache != nil {
		if err := r.cache.SetUserInfo(ctx, entity.ID, entity, cache.UserExpireTime); err != nil {
			log.Println("cache set user info error:", utils.NewErrorWithStack(err.Error()))
		}
	}

	return entity, nil
}

func (r *MySQLUserRepository) GetBatchGetUserInfoByIDs(ctx context.Context, userIds []string) ([]*user.User, error) {
	var models []UserModel

	if r.cache != nil {
		users, err := r.cache.GetUsersInfo(ctx, userIds)
		if err == nil && len(users) != 0 {
			return users, nil
		}
	}

	if err := r.db.WithContext(ctx).Where("id IN ?", userIds).Find(&models).Error; err != nil {
		return nil, err
	}

	entitys := make([]*user.User, 0, len(models))
	for _, model := range models {
		entity, err := model.ToEntity()
		if err != nil {
			return nil, err
		}
		entitys = append(entitys, entity)
	}

	if r.cache != nil {
		for _, v := range entitys {
			if err := r.cache.SetUserInfo(ctx, v.ID, v, cache.UserExpireTime); err != nil {
				return nil, err
			}
		}
	}

	return entitys, nil
}

func (r *MySQLUserRepository) SetUserPublicKey(ctx context.Context, userId, publicKey string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", userId).UpdateColumn("public_key", publicKey).Error; err != nil {
		return err
	}

	if r.cache != nil {
		if err := r.cache.DeleteUsersInfo(ctx, []string{userId}); err != nil {
			log.Println("cache set user public key error:", utils.NewErrorWithStack(err.Error()))
		}
	}

	return nil

}

func (r *MySQLUserRepository) GetUserPublicKey(ctx context.Context, userId string) (string, error) {
	var model UserModel

	if r.cache != nil {
		userInfo, err := r.cache.GetUserInfo(ctx, userId)
		if err == nil && userInfo != nil {
			return userInfo.PublicKey, nil
		}
	}

	if err := r.db.WithContext(ctx).Where("id = ?", userId).First(&model).Error; err != nil {
		return "", err
	}

	return model.PublicKey, nil
}

func (r *MySQLUserRepository) SetUserSecretBundle(ctx context.Context, userId, secretBundle string) error {
	if err := r.db.WithContext(ctx).Where("id = ?", userId).Update("secret_bundle", secretBundle).Error; err != nil {
		return err
	}

	if r.cache != nil {
		if err := r.cache.DeleteUsersInfo(ctx, []string{userId}); err != nil {
			log.Println("cache set user public key error:", utils.NewErrorWithStack(err.Error()))
		}
	}

	return nil
}

func (r *MySQLUserRepository) GetUserSecretBundle(ctx context.Context, userId string) (string, error) {
	var model UserModel

	if r.cache != nil {
		userInfo, err := r.cache.GetUserInfo(ctx, userId)
		if err == nil && userInfo != nil {
			return userInfo.SecretBundle, nil
		}
	}

	if err := r.db.WithContext(ctx).Where("id = ?", userId).Select("secret_bundle").First(&model).Error; err != nil {
		return "", err
	}
	return model.SecretBundle, nil
}

func (r *MySQLUserRepository) UpdateUserColumn(ctx context.Context, userId string, column string, value interface{}) error {
	if err := r.db.WithContext(ctx).Where("id = ?", userId).UpdateColumn(column, value).Error; err != nil {
		return err
	}

	if r.cache != nil {
		if err := r.cache.DeleteUsersInfo(ctx, []string{userId}); err != nil {
			log.Printf("无法删除用户信息缓存: %v", utils.NewErrorWithStack(err.Error()))
		}
	}

	return nil
}

func (r *MySQLUserRepository) InsertAndUpdateUser(ctx context.Context, user *user.User) error {
	var model UserModel

	if err := model.FromEntity(user); err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).Where(&UserModel{ID: model.ID}).Assign(UserModel{
		NickName: model.NickName,
		Password: model.Password,
		Email:    model.Email,
		Avatar:   model.Avatar}).FirstOrCreate(&model).Error; err != nil {
		return err
	}

	if r.cache != nil {
		if err := r.cache.DeleteUsersInfo(ctx, []string{model.ID}); err != nil {
			log.Printf("无法删除用户信息缓存: %v", utils.NewErrorWithStack(err.Error()))
		}
	}

	return nil
}

func (r *MySQLUserRepository) DeleteUser(ctx context.Context, userId string) error {
	if err := r.db.WithContext(ctx).Delete(&UserModel{ID: userId}).Error; err != nil {
		return err
	}

	if r.cache != nil {
		if err := r.cache.DeleteUsersInfo(ctx, []string{userId}); err != nil {
			log.Printf("无法删除用户信息缓存: %v", utils.NewErrorWithStack(err.Error()))
		}
	}

	return nil
}
