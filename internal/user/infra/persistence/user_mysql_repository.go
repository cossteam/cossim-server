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

var _ repository.UserRepository = &MySQLUserRepository{}

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

func (r *MySQLUserRepository) GetUserInfoByEmail(ctx context.Context, email string) (*entity.User, error) {
	var model po.User

	if err := r.db.WithContext(ctx).Where("email = ? AND deleted_at = 0", email).First(&model).Error; err != nil {
		return nil, err
	}

	entity := converter.UserPOToEntity(&model)

	return entity, nil
}

func (r *MySQLUserRepository) GetUserInfoByUid(ctx context.Context, id string) (*entity.User, error) {
	var model po.User

	if r.cache != nil {
		info, err := r.cache.GetUserInfo(ctx, id)
		if err == nil && info != nil {
			return info, nil
		}
	}

	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at = 0", id).First(&model).Error; err != nil {
		return nil, err
	}

	entity := converter.UserPOToEntity(&model)

	if r.cache != nil {
		if err := r.cache.SetUserInfo(ctx, entity.ID, entity, cache.UserExpireTime); err != nil {
			log.Println("cache set user info error:", utils.NewErrorWithStack(err.Error()))
		}
	}

	return entity, nil
}

func (r *MySQLUserRepository) GetUserInfoByCossID(ctx context.Context, cossId string) (*entity.User, error) {
	var model po.User

	if err := r.db.WithContext(ctx).Where("coss_id = ?", cossId).First(&model).Error; err != nil {
		return nil, err
	}

	entity := converter.UserPOToEntity(&model)

	if r.cache != nil {
		if err := r.cache.SetUserInfo(ctx, entity.ID, entity, cache.UserExpireTime); err != nil {
			log.Println("cache set user info error:", utils.NewErrorWithStack(err.Error()))
		}
	}

	return entity, nil
}

func (r *MySQLUserRepository) UpdateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	model := converter.UserEntityToPO(user)

	if err := r.db.WithContext(ctx).Where("id = ?", user.ID).Updates(&model).Error; err != nil {
		return nil, err
	}

	entity := converter.UserPOToEntity(model)

	if r.cache != nil {
		if err := r.cache.DeleteUsersInfo(ctx, []string{entity.ID}); err != nil {
			log.Println("cache set user info error:", utils.NewErrorWithStack(err.Error()))
		}
	}

	return entity, nil
}

func (r *MySQLUserRepository) InsertUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	model := converter.UserEntityToPO(user)

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return nil, err
	}

	entity := converter.UserPOToEntity(model)

	if r.cache != nil {
		if err := r.cache.SetUserInfo(ctx, entity.ID, entity, cache.UserExpireTime); err != nil {
			log.Println("cache set user info error:", utils.NewErrorWithStack(err.Error()))
		}
	}

	return entity, nil
}

func (r *MySQLUserRepository) GetBatchGetUserInfoByIDs(ctx context.Context, userIds []string) ([]*entity.User, error) {
	var models []*po.User

	if r.cache != nil {
		users, err := r.cache.GetUsersInfo(ctx, userIds)
		if err == nil && len(users) != 0 {
			return users, nil
		}
	}

	if err := r.db.WithContext(ctx).Where("id IN ?", userIds).Find(models).Error; err != nil {
		return nil, err
	}

	entitys := make([]*entity.User, 0, len(models))
	for _, model := range models {
		entity := converter.UserPOToEntity(model)
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
	var model po.User

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
	var model po.User

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

func (r *MySQLUserRepository) InsertAndUpdateUser(ctx context.Context, user *entity.User) error {
	model := converter.UserEntityToPO(user)

	if err := r.db.WithContext(ctx).Where(&po.User{ID: model.ID}).Assign(po.User{
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
	if err := r.db.WithContext(ctx).Delete(&po.User{ID: userId}).Error; err != nil {
		return err
	}

	if r.cache != nil {
		if err := r.cache.DeleteUsersInfo(ctx, []string{userId}); err != nil {
			log.Printf("无法删除用户信息缓存: %v", utils.NewErrorWithStack(err.Error()))
		}
	}

	return nil
}
