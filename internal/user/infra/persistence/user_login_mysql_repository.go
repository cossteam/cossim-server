package persistence

import (
	"context"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/domain/repository"
	"github.com/cossim/coss-server/internal/user/infra/persistence/converter"
	"github.com/cossim/coss-server/internal/user/infra/persistence/po"
	"github.com/cossim/coss-server/pkg/utils"
	"gorm.io/gorm"
	"log"
)

var _ repository.UserLoginRepository = &MySQLUserLoginRepository{}

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

func (r *MySQLUserLoginRepository) GetWithFields(ctx context.Context, fields map[string]interface{}) (*entity.UserLogin, error) {
	var model po.UserLogin

	query := r.db.WithContext(ctx)

	for field, value := range fields {
		query = query.Where(field, value)
	}

	if err := query.First(&model).Error; err != nil {
		//if errors.Is(err, gorm.ErrRecordNotFound) {
		//	return nil, nil
		//}
		return nil, err
	}

	userLogin := converter.UserLoginPOToEntity(&model)
	return userLogin, nil
}

func (r *MySQLUserLoginRepository) InsertUserLogin(ctx context.Context, user *entity.UserLogin) error {
	model := converter.UserLoginEntityToPO(user)

	if err := r.db.WithContext(ctx).Where(po.UserLogin{
		UserId:   model.UserId,
		DriverId: model.DriverId,
		LastAt:   model.LastAt,
	}).
		Assign(po.UserLogin{
			DriverId:    model.DriverId,
			LoginCount:  model.LoginCount,
			DriverToken: model.DriverToken,
			Token:       model.Token,
			LastAt:      model.LastAt,
		}).
		FirstOrCreate(model).Error; err != nil {
		return err
	}

	user.ID = model.ID

	return nil
}

func (r *MySQLUserLoginRepository) GetUserLoginByDriverIdAndUserId(ctx context.Context, driverId, userId string) (*entity.UserLogin, error) {
	var model po.UserLogin

	if err := r.db.WithContext(ctx).Where("driver_id = ? AND user_id = ?", driverId, userId).Order("created_at DESC").First(&model).Error; err != nil {
		return nil, err
	}

	e := converter.UserLoginPOToEntity(&model)

	//if r.cache != nil {
	//	if err := r.cache.SetUserLoginInfo(ctx, entity.ID, int(entity.LoginCount), entity, cache.UserExpireTime); err != nil {
	//		log.Printf("cache set user login info error: %v", utils.NewErrorWithStack(err.Error()))
	//	}
	//}

	return e, nil
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

func (r *MySQLUserLoginRepository) GetUserLoginByToken(ctx context.Context, token string) (*entity.UserLogin, error) {
	var model po.UserLogin

	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&model).Error; err != nil {
		return nil, err
	}

	e := converter.UserLoginPOToEntity(&model)

	return e, nil
}

func (r *MySQLUserLoginRepository) GetUserDriverTokenByUserId(ctx context.Context, userId string) ([]string, error) {
	var driverTokens []string
	err := r.db.WithContext(ctx).Where("user_id = ?", userId).Pluck("driver_token", &driverTokens).Error
	if err != nil {
		return nil, err
	}
	return driverTokens, err
}

func (r *MySQLUserLoginRepository) GetUserByUserId(ctx context.Context, userId string) (*entity.UserLogin, error) {
	var model po.UserLogin

	if err := r.db.WithContext(ctx).Where("user_id = ?", userId).Order("created_at DESC").First(&model).Error; err != nil {
		return nil, err
	}

	e := converter.UserLoginPOToEntity(&model)

	return e, nil
}

func (r *MySQLUserLoginRepository) GetUserLoginById(ctx context.Context, id uint32) (*entity.UserLogin, error) {
	var model po.UserLogin

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return nil, err
	}

	e := converter.UserLoginPOToEntity(&model)

	return e, nil
}

func (r *MySQLUserLoginRepository) DeleteUserLoginByID(ctx context.Context, id uint32) error {
	var model po.UserLogin
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error; err != nil {
		return err
	}

	if err := r.db.WithContext(ctx).Where("id = ?", id).Delete(&po.UserLogin{}).Error; err != nil {
		return err
	}

	if r.cache != nil {
		if err := r.cache.DeleteUserLoginInfo(ctx, model.UserId, model.DriverId); err != nil {
			log.Printf("cache del user login info error: %v", utils.NewErrorWithStack(err.Error()))
		}
	}

	return nil
}
