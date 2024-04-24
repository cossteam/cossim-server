package persistence

import (
	"context"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/user"
	"github.com/cossim/coss-server/pkg/utils"
	"gorm.io/gorm"
	"log"
)

type UserLoginModel struct {
	BaseModel
	UserId      string `gorm:"type:varchar(64);comment:用户id" json:"user_id"`
	LoginCount  uint   `gorm:"default:0;comment:登录次数" json:"login_count"`
	LastAt      int64  `gorm:"comment:最后登录时间" json:"last_at"`
	Token       string `gorm:"type:longtext;comment:登录token" json:"token"`
	DriverId    string `gorm:"type:longtext;comment:登录设备id" json:"driver_id"`
	DriverToken string `gorm:"type:varchar(255);comment:登录设备token" json:"driver_token"`
	Platform    string `gorm:"type:varchar(50);comment:手机厂商" json:"platform"`
}

type BaseModel struct {
	ID        uint  `gorm:"primaryKey;autoIncrement;" json:"id"`
	CreatedAt int64 `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt int64 `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt int64 `gorm:"default:0;comment:删除时间" json:"deleted_at"`
}

func (m *UserLoginModel) TableName() string {
	return "user_logins"
}

func (m *UserLoginModel) FromEntity(ul *user.UserLogin) error {
	m.UserId = ul.UserId
	m.LoginCount = ul.LoginCount
	m.LastAt = ul.LastAt
	m.Token = ul.Token
	m.DriverId = ul.DriverId
	m.DriverToken = ul.DriverToken
	m.Platform = ul.Platform
	m.BaseModel = BaseModel{
		ID:        ul.ID,
		CreatedAt: ul.CreatedAt,
		UpdatedAt: ul.UpdatedAt,
		DeletedAt: ul.DeletedAt,
	}
	return nil
}

func (m *UserLoginModel) ToEntity() (*user.UserLogin, error) {
	ul := &user.UserLogin{
		UserId:      m.UserId,
		LoginCount:  m.LoginCount,
		LastAt:      m.LastAt,
		Token:       m.Token,
		DriverId:    m.DriverId,
		DriverToken: m.DriverToken,
		Platform:    m.Platform,
		ID:          m.ID,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
		DeletedAt:   m.DeletedAt,
	}
	return ul, nil
}

var _ user.UserLoginRepository = &MySQLUserLoginRepository{}

func NewMySQLUserLoginRepository(db *gorm.DB, cache cache.UserCache) *MySQLUserLoginRepository {
	return &MySQLUserLoginRepository{
		db:    db,
		cache: cache,
	}
}

type MySQLUserLoginRepository struct {
	db    *gorm.DB
	cache cache.UserCache
}

func (r *MySQLUserLoginRepository) InsertUserLogin(ctx context.Context, user *user.UserLogin) error {
	var model UserLoginModel

	if err := model.FromEntity(user); err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).Where(UserLoginModel{
		UserId:   model.UserId,
		DriverId: model.DriverId,
		LastAt:   model.LastAt,
	}).
		Assign(UserLoginModel{
			LoginCount:  model.LoginCount,
			DriverToken: model.DriverToken,
			Token:       model.Token,
			LastAt:      model.LastAt,
		}).
		FirstOrCreate(&model).Error; err != nil {
		return err
	}

	user.ID = model.ID

	return nil
}

func (r *MySQLUserLoginRepository) GetUserLoginByDriverIdAndUserId(ctx context.Context, driverId, userId string) (*user.UserLogin, error) {
	var model UserLoginModel

	if err := r.db.WithContext(ctx).Where("driver_id = ? AND user_id = ?", driverId, userId).Order("created_at DESC").First(&model).Error; err != nil {
		return nil, err
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, err
	}

	if r.cache != nil {
		if err := r.cache.SetUserLoginInfo(ctx, entity.UserId, int(entity.LoginCount), entity, cache.UserExpireTime); err != nil {
			log.Printf("cache set user login info error: %v", utils.NewErrorWithStack(err.Error()))
		}
	}

	return entity, nil
}

func (r *MySQLUserLoginRepository) UpdateUserLoginTokenByDriverId(ctx context.Context, driverId string, token string, userId string) error {
	if err := r.db.WithContext(ctx).Where("driver_id = ? AND user_id = ?", driverId, userId).Update("token", token).Error; err != nil {
		return err
	}

	//if r.cache != nil {
	//	if err := r.cache.DeleteUserLoginInfo(ctx, userId); err != nil {
	//		log.Printf("cache del user login info error: %v", utils.NewErrorWithStack(err.Error()))
	//	}
	//}

	return nil
}

func (r *MySQLUserLoginRepository) GetUserLoginByToken(ctx context.Context, token string) (*user.UserLogin, error) {
	var model UserLoginModel

	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&model).Error; err != nil {
		return nil, err
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (r *MySQLUserLoginRepository) GetUserDriverTokenByUserId(ctx context.Context, userId string) ([]string, error) {
	var driverTokens []string
	err := r.db.WithContext(ctx).Where("user_id = ?", userId).Pluck("driver_token", &driverTokens).Error
	if err != nil {
		return nil, err
	}
	return driverTokens, err
}

func (r *MySQLUserLoginRepository) GetUserByUserId(ctx context.Context, userId string) (*user.UserLogin, error) {
	var model UserLoginModel

	if err := r.db.WithContext(ctx).Where("user_id = ?", userId).Order("created_at DESC").First(&model).Error; err != nil {
		return nil, err
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (r *MySQLUserLoginRepository) GetUserLoginById(ctx context.Context, id uint32) (*user.UserLogin, error) {
	var model UserLoginModel

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, err
	}

	entity, err := model.ToEntity()
	if err != nil {
		return nil, err
	}

	return entity, nil
}

func (r *MySQLUserLoginRepository) DeleteUserLoginByID(ctx context.Context, id uint32) error {
	var model UserLoginModel
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&UserLoginModel{}).Error; err != nil {
		return err
	}

	if r.cache != nil {
		if err := r.cache.DeleteUserLoginInfo(ctx, model.UserId, int(model.LoginCount)); err != nil {
			log.Printf("cache del user login info error: %v", utils.NewErrorWithStack(err.Error()))
		}
	}

	return nil
}
