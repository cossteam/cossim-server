package persistence

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/domain/repository"
	"github.com/cossim/coss-server/internal/user/infra/persistence/converter"
	"github.com/cossim/coss-server/internal/user/infra/persistence/po"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils"
	"gorm.io/gorm"
	"log"
	"reflect"
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

func (r *MySQLUserRepository) GetWithOptions(ctx context.Context, query *entity.User) (*entity.User, error) {
	var model po.User

	if err := r.db.WithContext(ctx).Where(query).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.NotFound
		}
		return nil, err
	}

	return converter.UserPOToEntity(&model), nil
}

func (r *MySQLUserRepository) UpdateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	poUser := converter.UserEntityToPO(user)

	userValue := reflect.ValueOf(user).Elem()
	for i := 0; i < userValue.NumField(); i++ {
		fieldName := userValue.Type().Field(i).Name
		fieldValue := userValue.Field(i)

		if !fieldValue.IsZero() {
			if err := r.db.Model(&poUser).Where("id = ?", poUser.ID).Update(fieldName, fieldValue.Interface()).Error; err != nil {
				return nil, err
			}
		}
	}

	entityUser := converter.UserPOToEntity(poUser)

	return entityUser, nil
}

func (r *MySQLUserRepository) SaveUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	poUser := converter.UserEntityToPO(user)
	result := r.db.WithContext(ctx).
		Create(poUser)
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, errors.New("unable to save user")
	}

	entityUser := converter.UserPOToEntity(poUser)

	return entityUser, nil
}

func (r *MySQLUserRepository) DeleteUser(ctx context.Context, id string) error {
	if err := r.db.WithContext(ctx).Delete(&po.User{ID: id}).Error; err != nil {
		return err
	}

	if r.cache != nil {
		if err := r.cache.DeleteUsersInfo(ctx, []string{id}); err != nil {
			log.Printf("无法删除用户信息缓存: %v", utils.NewErrorWithStack(err.Error()))
		}
	}

	return nil
}

func (r *MySQLUserRepository) UpdatesUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	poUser := converter.UserEntityToPO(user)
	result := r.db.Model(poUser).Where("id = ?", poUser.ID).Updates(*poUser)
	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, errors.New("unable to update user")
	}

	entityUser := converter.UserPOToEntity(poUser)

	return entityUser, nil
}

func (r *MySQLUserRepository) GetUser(ctx context.Context, id string) (*entity.User, error) {
	user := &po.User{}

	if err := r.db.WithContext(ctx).
		Where(&po.User{
			ID:        id,
			DeletedAt: 0,
		}).
		First(user).Error; err != nil {
		return nil, err
	}

	entityUser := converter.UserPOToEntity(user)

	return entityUser, nil
}

func (r *MySQLUserRepository) ListUser(ctx context.Context, query *entity.ListUserOptions) ([]*entity.User, error) {
	var model []*po.User

	db := r.db.WithContext(ctx)

	if len(query.UserID) > 0 {
		db = db.Where("id IN (?)", query.UserID)
	}

	if err := db.Where(&po.User{
		DeletedAt: 0,
	}).Find(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.NotFound
		}
		return nil, err
	}

	var entityUser = make([]*entity.User, 0)
	for _, v := range model {
		entityUser = append(entityUser, converter.UserPOToEntity(v))
	}

	return entityUser, nil
}

//func (r *MySQLUserRepository) UpdateUserInfo(ctx context.Context, user *entity.UpdateUser) error {
//	if user == nil {
//		return code.InvalidParameter.CustomMessage("内容不能为空")
//	}
//
//	updateData := make(map[string]interface{})
//
//	if user.Email != nil {
//		updateData["email"] = *user.Email
//	}
//	if user.Tel != nil {
//		updateData["tel"] = *user.Tel
//	}
//	if user.NickName != nil {
//		updateData["nick_name"] = *user.NickName
//	}
//	if user.Avatar != nil {
//		updateData["avatar"] = *user.Avatar
//	}
//	if user.PublicKey != nil {
//		updateData["public_key"] = *user.PublicKey
//	}
//	if user.Password != nil {
//		updateData["password"] = *user.Password
//	}
//
//	if err := r.db.WithContext(ctx).Model(&po.User{}).Updates(updateData).Error; err != nil {
//		return err
//	}
//
//	return nil
//}

//func (r *MySQLUserRepository) UpdateUserStatus(ctx context.Context, Param *entity.UpdateUserStatus, userIDs ...string) error {
//	for _, userID := range userIDs {
//		user := &po.User{}
//		if err := r.db.WithContext(ctx).Where("id = ?", userID).First(user).Error; err != nil {
//			return err
//		}
//
//		if Param.Status != nil {
//			user.Status = uint(*Param.Status)
//		}
//
//		if Param.EmailVerity != nil {
//			user.EmailVerity = *Param.EmailVerity
//		}
//
//		if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
//			return err
//		}
//	}
//
//	return nil
//}

//func (r *MySQLUserRepository) UpdatePassword(ctx context.Context, userID, password string) (string, error) {
//	var newPassword = password
//	if password == "" {
//		return "", code.InvalidParameter.CustomMessage("password can not be empty")
//	}
//
//	user := &po.User{}
//	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(user).Error; err != nil {
//		return "", err
//	}
//
//	user.Password = newPassword
//
//	if err := r.db.WithContext(ctx).Save(user).Error; err != nil {
//		return "", err
//	}
//
//	return newPassword, nil
//}

//func (r *MySQLUserRepository) GetUserInfoByEmail(ctx context.Context, email string) (*entity.User, error) {
//	var model po.User
//
//	if err := r.db.WithContext(ctx).Where("email = ? AND deleted_at = 0", email).First(&model).Error; err != nil {
//		return nil, err
//	}
//
//	entity := converter.UserPOToEntity(&model)
//
//	return entity, nil
//}

//func (r *MySQLUserRepository) GetUserInfoByUid(ctx context.Context, id string) (*entity.User, error) {
//	var model po.User
//
//	if r.cache != nil {
//		info, err := r.cache.GetUserInfo(ctx, id)
//		if err == nil && info != nil {
//			return info, nil
//		}
//	}
//
//	if err := r.db.WithContext(ctx).Where("id = ? AND deleted_at = 0", id).First(&model).Error; err != nil {
//		return nil, err
//	}
//
//	e := converter.UserPOToEntity(&model)
//
//	if r.cache != nil {
//		if err := r.cache.SetUserInfo(ctx, e.ID, e, cache.UserExpireTime); err != nil {
//			log.Println("cache set user info error:", utils.NewErrorWithStack(err.Error()))
//		}
//	}
//
//	return e, nil
//}

//
//func (r *MySQLUserRepository) GetUserInfoByCossID(ctx context.Context, cossId string) (*entity.User, error) {
//	var model po.User
//
//	if err := r.db.WithContext(ctx).Where("coss_id = ?", cossId).First(&model).Error; err != nil {
//		return nil, err
//	}
//
//	entity := converter.UserPOToEntity(&model)
//
//	if r.cache != nil {
//		if err := r.cache.SetUserInfo(ctx, entity.ID, entity, cache.UserExpireTime); err != nil {
//			log.Println("cache set user info error:", utils.NewErrorWithStack(err.Error()))
//		}
//	}
//
//	return entity, nil
//}

//func (r *MySQLUserRepository) InsertUser(ctx context.Context, user *entity.User) (*entity.User, error) {
//	model := converter.UserEntityToPO(user)
//
//	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
//		return nil, err
//	}
//
//	entity := converter.UserPOToEntity(model)
//
//	if r.cache != nil {
//		if err := r.cache.SetUserInfo(ctx, entity.ID, entity, cache.UserExpireTime); err != nil {
//			log.Println("cache set user info error:", utils.NewErrorWithStack(err.Error()))
//		}
//	}
//
//	return entity, nil
//}
//
//func (r *MySQLUserRepository) GetBatchGetUserInfoByIDs(ctx context.Context, userIds []string) ([]*entity.User, error) {
//	var models []*po.User
//
//	if len(userIds) == 0 {
//		return nil, nil
//	}
//
//	if r.cache != nil {
//		users, err := r.cache.GetUsersInfo(ctx, userIds)
//		if err == nil && len(users) != 0 {
//			return users, nil
//		}
//	}
//
//	if err := r.db.WithContext(ctx).Where("id IN (?)", userIds).Find(&models).Error; err != nil {
//		return nil, err
//	}
//
//	entitys := make([]*entity.User, 0, len(models))
//	for _, model := range models {
//		entity := converter.UserPOToEntity(model)
//		entitys = append(entitys, entity)
//	}
//
//	if r.cache != nil {
//		for _, v := range entitys {
//			if err := r.cache.SetUserInfo(ctx, v.ID, v, cache.UserExpireTime); err != nil {
//				return nil, err
//			}
//		}
//	}
//
//	return entitys, nil
//}
//
//func (r *MySQLUserRepository) SetUserPublicKey(ctx context.Context, userId, publicKey string) error {
//	if err := r.db.WithContext(ctx).Where("id = ?", userId).UpdateColumn("public_key", publicKey).Error; err != nil {
//		return err
//	}
//
//	if r.cache != nil {
//		if err := r.cache.DeleteUsersInfo(ctx, []string{userId}); err != nil {
//			log.Println("cache set user public key error:", utils.NewErrorWithStack(err.Error()))
//		}
//	}
//
//	return nil
//
//}
//
//func (r *MySQLUserRepository) GetUserPublicKey(ctx context.Context, userId string) (string, error) {
//	var model po.User
//
//	if r.cache != nil {
//		userInfo, err := r.cache.GetUserInfo(ctx, userId)
//		if err == nil && userInfo != nil {
//			return userInfo.PublicKey, nil
//		}
//	}
//
//	if err := r.db.WithContext(ctx).Where("id = ?", userId).First(&model).Error; err != nil {
//		return "", err
//	}
//
//	return model.PublicKey, nil
//}
//
//func (r *MySQLUserRepository) SetUserSecretBundle(ctx context.Context, userId, secretBundle string) error {
//	if err := r.db.WithContext(ctx).Model(&po.User{}).
//		Where("id = ?", userId).Update("secret_bundle", secretBundle).Error; err != nil {
//		return err
//	}
//
//	if r.cache != nil {
//		if err := r.cache.DeleteUsersInfo(ctx, []string{userId}); err != nil {
//			log.Println("cache set user public key error:", utils.NewErrorWithStack(err.Error()))
//		}
//	}
//
//	return nil
//}
//
//func (r *MySQLUserRepository) GetUserSecretBundle(ctx context.Context, userId string) (string, error) {
//	var model po.User
//
//	if r.cache != nil {
//		userInfo, err := r.cache.GetUserInfo(ctx, userId)
//		if err == nil && userInfo != nil {
//			return userInfo.SecretBundle, nil
//		}
//	}
//
//	if err := r.db.WithContext(ctx).Where("id = ?", userId).Select("secret_bundle").First(&model).Error; err != nil {
//		return "", err
//	}
//	return model.SecretBundle, nil
//}
//
//func (r *MySQLUserRepository) UpdateUserColumn(ctx context.Context, userId string, column string, value interface{}) error {
//	if err := r.db.WithContext(ctx).
//		Model(&po.User{}).
//		Where("id = ?", userId).UpdateColumn(column, value).Error; err != nil {
//		return err
//	}
//
//	if r.cache != nil {
//		if err := r.cache.DeleteUsersInfo(ctx, []string{userId}); err != nil {
//			log.Printf("无法删除用户信息缓存: %v", utils.NewErrorWithStack(err.Error()))
//		}
//	}
//
//	return nil
//}
//
//func (r *MySQLUserRepository) InsertAndUpdateUser(ctx context.Context, user *entity.User) error {
//	model := converter.UserEntityToPO(user)
//
//	if err := r.db.WithContext(ctx).Where(&po.User{ID: model.ID}).Assign(po.User{
//		NickName: model.NickName,
//		Password: model.Password,
//		Email:    model.Email,
//		Avatar:   model.Avatar}).FirstOrCreate(&model).Error; err != nil {
//		return err
//	}
//
//	if r.cache != nil {
//		if err := r.cache.DeleteUsersInfo(ctx, []string{model.ID}); err != nil {
//			log.Printf("无法删除用户信息缓存: %v", utils.NewErrorWithStack(err.Error()))
//		}
//	}
//
//	return nil
//}
