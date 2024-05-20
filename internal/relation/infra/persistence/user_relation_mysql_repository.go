package persistence

import (
	"context"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/internal/relation/cache"
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
	"log"
	"reflect"
)

// UserRelationModel 对应 relation.UserRelation
type UserRelationModel struct {
	BaseModel
	Status                  uint     `gorm:"comment:好友关系状态 (0=拉黑 1=正常 2=删除)"`
	UserID                  string   `gorm:"type:varchar(64);comment:用户ID"`
	FriendID                string   `gorm:"type:varchar(64);comment:好友ID"`
	DialogId                uint32   `gorm:"comment:对话ID"`
	Remark                  string   `gorm:"type:varchar(255);comment:备注"`
	Label                   []string `gorm:"type:varchar(255);comment:标签"`
	SilentNotification      bool     `gorm:"comment:是否开启静默通知"`
	OpenBurnAfterReading    bool     `gorm:"comment:是否开启阅后即焚消息"`
	BurnAfterReadingTimeOut int64    `gorm:"default:10;comment:阅后即焚时间"`
}

func (m *UserRelationModel) TableName() string {
	return "user_relations"
}

func (m *UserRelationModel) FromEntity(u *entity.UserRelation) error {
	m.ID = u.ID
	m.UserID = u.UserID
	m.FriendID = u.FriendID
	m.Status = uint(u.Status)
	m.DialogId = u.DialogId
	m.Remark = u.Remark
	m.Label = u.Label
	m.SilentNotification = u.SilentNotification
	m.OpenBurnAfterReading = u.OpenBurnAfterReading
	m.BurnAfterReadingTimeOut = u.BurnAfterReadingTimeOut
	return nil
}

func (m *UserRelationModel) ToEntity() *entity.UserRelation {
	return &entity.UserRelation{
		ID:                      m.ID,
		CreatedAt:               m.CreatedAt,
		UserID:                  m.UserID,
		FriendID:                m.FriendID,
		Status:                  entity.UserRelationStatus(m.Status),
		DialogId:                m.DialogId,
		Remark:                  m.Remark,
		Label:                   m.Label,
		SilentNotification:      m.SilentNotification,
		OpenBurnAfterReading:    m.OpenBurnAfterReading,
		BurnAfterReadingTimeOut: m.BurnAfterReadingTimeOut,
	}
}

var _ repository.UserRelationRepository = &MySQLRelationUserRepository{}

func NewMySQLRelationUserRepository(db *gorm.DB, cache cache.RelationUserCache) *MySQLRelationUserRepository {
	return &MySQLRelationUserRepository{
		db:    db,
		cache: cache,
	}
}

type MySQLRelationUserRepository struct {
	db    *gorm.DB
	cache cache.RelationUserCache
}

func (m *MySQLRelationUserRepository) DeleteRollback(ctx context.Context, id uint32) error {
	var model UserRelationModel

	if err := m.db.WithContext(ctx).
		Model(&UserRelationModel{}).
		Where("id = ?", id).
		First(&model).
		Update("status", uint(entity.UserStatusNormal)).
		Update("deleted_at", 0).
		Error; err != nil {
		return err
	}

	return nil
}

func (m *MySQLRelationUserRepository) RestoreFriendship(ctx context.Context, dialogID uint32, userId, friendId string) error {
	if err := m.db.WithContext(ctx).
		Model(&UserRelationModel{}).
		Where("dialog_id = ? AND user_id = ? AND friend_id = ?", dialogID, userId, friendId).
		Update("status", entity.UserStatusNormal).
		Update("deleted_at", 0).
		Error; err != nil {
		return err
	}

	if m.cache != nil {
		if err := m.cache.DeleteRelation(ctx, userId, []string{friendId}); err != nil {
			log.Printf("delete relation cache error: %v", err)
		}
		if err := m.cache.DeleteFriendList(ctx, userId, friendId); err != nil {
			log.Printf("failed to delete cache friend list: %v", err)
		}
	}

	return nil
}

func (m *MySQLRelationUserRepository) EstablishFriendship(ctx context.Context, dialogID uint32, senderID, receiverID string) error {
	models := []UserRelationModel{
		{
			UserID:   senderID,
			FriendID: receiverID,
			Status:   uint(entity.UserStatusNormal),
			DialogId: dialogID,
		},
		{
			UserID:   receiverID,
			FriendID: senderID,
			Status:   uint(entity.UserStatusNormal),
			DialogId: dialogID,
		},
	}

	if err := m.db.WithContext(ctx).Create(&models).Error; err != nil {
		return err
	}

	if m.cache != nil {
		if err := m.cache.DeleteRelation(ctx, senderID, []string{receiverID}); err != nil {
			log.Printf("delete relation cache error: %v", err)
		}
		if err := m.cache.DeleteRelation(ctx, receiverID, []string{senderID}); err != nil {
			log.Printf("delete relation cache error: %v", err)
		}
		if err := m.cache.DeleteFriendList(ctx, senderID, receiverID); err != nil {
			log.Printf("failed to delete cache friend list: %v", err)
		}
	}

	return nil
}

func (m *MySQLRelationUserRepository) UpdateStatus(ctx context.Context, id uint32, status entity.UserRelationStatus) error {
	var model UserRelationModel
	var deletedAt int64 = 0

	if status == entity.UserStatusDeleted {
		deletedAt = ptime.Now()
	}

	if err := m.db.WithContext(ctx).
		Model(&UserRelationModel{}).
		Where("id = ?", id).
		First(&model).
		Update("status", uint(status)).
		Update("deleted_at", deletedAt).
		Error; err != nil {
		return err
	}

	if m.cache != nil {
		if err := m.cache.DeleteRelation(ctx, model.UserID, []string{model.FriendID}); err != nil {
			log.Printf("delete relation cache error: %v", err)
		}
		if err := m.cache.DeleteFriendList(ctx, model.UserID, model.FriendID); err != nil {
			log.Printf("failed to delete cache friend list: %v", err)
		}
		if err := m.cache.DeleteBlacklist(ctx, model.UserID, model.FriendID); err != nil {
			log.Printf("failed to delete cache friend list: %v", err)
		}
	}

	return nil
}

func (m *MySQLRelationUserRepository) Get(ctx context.Context, userId, friendId string) (*entity.UserRelation, error) {
	var model UserRelationModel

	if m.cache != nil {
		e, err := m.cache.GetRelation(ctx, userId, friendId)
		if err == nil && e != nil {
			return e, nil
		}
	}

	if err := m.db.WithContext(ctx).
		Where("user_id = ? AND friend_id = ? AND deleted_at = 0", userId, friendId).
		First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, code.NotFound
		}
		return nil, err
	}

	fmt.Println("model.Status", model.Status)

	e := model.ToEntity()

	go func() {
		if m.cache != nil {
			if err := m.cache.SetRelation(context.Background(), userId, friendId, e, cache.RelationExpireTime); err != nil {
				log.Printf("cache.SetRelation failed err:%v", err)
			}
		}
	}()

	return e, nil
}

func (m *MySQLRelationUserRepository) Create(ctx context.Context, ur *entity.UserRelation) (*entity.UserRelation, error) {
	var model UserRelationModel

	if err := model.FromEntity(ur); err != nil {
		return nil, err
	}

	if err := m.db.Create(&model).Error; err != nil {
		return nil, err
	}

	e := model.ToEntity()

	return e, nil
}

func (m *MySQLRelationUserRepository) Update(ctx context.Context, ur *entity.UserRelation) (*entity.UserRelation, error) {
	var model UserRelationModel

	if err := model.FromEntity(ur); err != nil {
		return nil, err
	}

	userValue := reflect.ValueOf(ur).Elem()
	for i := 0; i < userValue.NumField(); i++ {
		fieldName := userValue.Type().Field(i).Name
		fieldValue := userValue.Field(i)

		if !fieldValue.IsZero() {
			result := m.db.Model(&model).Where("id = ?", model.ID).Update(fieldName, fieldValue.Interface())
			if result.Error != nil {
				return nil, result.Error
			}
		}
	}

	//entityUser := converter.UserPOToEntity(poUser)

	//return entityUser, nil

	//if err := m.db.Where("id = ?", ur.ID).Updates(&model).Error; err != nil {
	//	return nil, err
	//}

	e := model.ToEntity()

	if m.cache != nil {
		if err := m.cache.DeleteRelation(ctx, ur.UserID, []string{ur.FriendID}); err != nil {
			log.Printf("delete relation cache error: %v", err)
		}

		if err := m.cache.DeleteFriendList(ctx, ur.UserID, ur.FriendID); err != nil {
			log.Printf("failed to delete cache friend list: %v", err)
		}
	}

	return e, nil
}

func (m *MySQLRelationUserRepository) Delete(ctx context.Context, userId, friendId string) error {
	if err := m.db.Model(&UserRelationModel{}).WithContext(ctx).
		Where("user_id = ? AND friend_id = ? AND deleted_at = 0", userId, friendId).
		Update("status", entity.UserStatusDeleted).
		Update("deleted_at", ptime.Now()).Error; err != nil {
	}

	if m.cache != nil {
		if err := m.cache.DeleteRelation(ctx, userId, []string{friendId}); err != nil {
			log.Printf("delete relation cache error: %v", err)
		}
		if err := m.cache.DeleteFriendList(ctx, userId); err != nil {
			log.Printf("failed to delete cache friend list: %v", err)
		}
	}

	return nil
}

func (m *MySQLRelationUserRepository) Find(ctx context.Context, query *repository.UserQuery) ([]*entity.UserRelation, error) {
	var models []*UserRelationModel

	// 构建查询条件
	db := m.db.WithContext(ctx).Model(&UserRelationModel{})
	if query.UserId != "" {
		db = db.Where("user_id = ?", query.UserId)
	}
	if query.FriendId != nil && len(query.FriendId) > 0 {
		db = db.Where("friend_id IN (?)", query.FriendId)
	}

	if query.Status != nil {
		db = db.Where("status = ?", query.Status)
	}

	if !query.Force {
		db = db.Where("deleted_at = 0")
	}

	// 执行查询
	if err := db.Find(&models).Error; err != nil {
		return nil, err
	}

	var list []*entity.UserRelation
	for _, v := range models {
		list = append(list, v.ToEntity())
	}

	if m.cache != nil {
		for _, v := range list {
			if err := m.cache.SetRelation(ctx, v.UserID, v.FriendID, v, cache.RelationExpireTime); err != nil {
				log.Printf("failed to set relation cache: %v", err)
			}
		}
	}

	return list, nil
}

func (m *MySQLRelationUserRepository) Blacklist(ctx context.Context, opts *entity.BlacklistOptions) (*entity.Blacklist, error) {
	if m.cache != nil {
		cachedList, err := m.cache.GetBlacklist(ctx, opts.UserID)
		if err == nil && cachedList != nil {
			return cachedList, nil
		}
	}

	var relations []*UserRelationModel
	if err := m.db.WithContext(ctx).Where("user_id = ? AND status = ? AND deleted_at = 0", opts.UserID, entity.UserStatusBlocked).
		Find(&relations).Error; err != nil {
		return nil, err
	}

	blacklist := &entity.Blacklist{
		Page: int32(opts.PageNum),
	}
	for _, v := range relations {
		blacklist.List = append(blacklist.List, &entity.Black{
			UserID:    v.FriendID,
			CreatedAt: v.CreatedAt,
		})
		blacklist.Total++
	}

	if m.cache != nil {
		if err := m.cache.SetBlacklist(ctx, opts.UserID, blacklist, cache.RelationExpireTime); err != nil {
			log.Printf("failed to set blacklist cache: %v", err)
		}
	}

	return blacklist, nil
}

func (m *MySQLRelationUserRepository) ListFriend(ctx context.Context, userId string) ([]*entity.Friend, error) {
	if m.cache != nil {
		r, err := m.cache.GetFriendList(ctx, userId)
		fmt.Println("ListFriend cache => ", r)
		if err == nil && r != nil {
			return r, nil
		}
	}

	var relations []*UserRelationModel
	if err := m.db.WithContext(ctx).
		Where("user_id = ? AND status NOT IN ? AND friend_id NOT IN ? AND deleted_at = 0",
			userId,
			[]entity.UserRelationStatus{entity.UserStatusBlocked, entity.UserStatusDeleted},
			constants.SystemUserList,
		).
		Find(&relations).Error; err != nil {
		return nil, err
	}

	var friends []*entity.Friend
	for _, v := range relations {
		friends = append(friends, &entity.Friend{
			UserID:                      v.FriendID,
			DialogID:                    v.DialogId,
			Remark:                      v.Remark,
			Status:                      entity.UserRelationStatus(v.Status),
			OpenBurnAfterReading:        v.OpenBurnAfterReading,
			IsSilent:                    v.SilentNotification,
			OpenBurnAfterReadingTimeOut: v.BurnAfterReadingTimeOut,
		})
	}

	if m.cache != nil {
		if err := m.cache.SetFriendList(ctx, userId, friends, cache.RelationExpireTime); err != nil {
			log.Printf("failed to set friend list cache: %v", err)
		}
	}

	return friends, nil
}

func (m *MySQLRelationUserRepository) UpdateFields(ctx context.Context, id uint32, fields map[string]interface{}) error {
	if err := m.db.Model(&UserRelationModel{}).WithContext(ctx).
		Where("id = ?", id).
		Unscoped().
		Updates(fields).Error; err != nil {
		return err
	}

	return nil
}

func (m *MySQLRelationUserRepository) SetUserFriendSilentNotification(ctx context.Context, userId, friendId string, silentNotification bool) error {
	if err := m.db.Model(&UserRelationModel{}).WithContext(ctx).
		Where("user_id = ? AND friend_id = ? AND deleted_at = 0", userId, friendId).
		Update("silent_notification", silentNotification).Error; err != nil {
		return err
	}

	go func() {
		if m.cache != nil {
			if err := m.cache.DeleteRelation(context.Background(), userId, []string{friendId}); err != nil {
				log.Printf("delete relation cache failed: %v", err)
			}
			if err := m.cache.DeleteFriendList(context.Background(), userId); err != nil {
				log.Printf("delete friend request list cache failed: %v", err)
			}
		}
	}()

	return nil
}

func (m *MySQLRelationUserRepository) SetUserOpenBurnAfterReading(ctx context.Context, userId, friendId string, openBurnAfterReading bool, burnAfterReadingTimeOut uint32) error {
	if err := m.db.Model(&UserRelationModel{}).WithContext(ctx).
		Where("user_id = ? AND friend_id = ? AND deleted_at = 0", userId, friendId).
		Updates(map[string]interface{}{
			"open_burn_after_reading":     openBurnAfterReading,
			"burn_after_reading_time_out": burnAfterReadingTimeOut,
		}).Error; err != nil {
		return err
	}

	if m.cache != nil {
		if err := m.cache.DeleteRelation(ctx, userId, []string{friendId}); err != nil {
			log.Printf("delete relation cache failed: %v", err)
		}
		if err := m.cache.DeleteRelation(ctx, friendId, []string{userId}); err != nil {
			log.Printf("delete relation cache failed: %v", err)
		}
		if err := m.cache.DeleteFriendList(ctx, userId); err != nil {
			log.Printf("delete friend request list cache failed: %v", err)
		}
	}

	return nil
}

func (m *MySQLRelationUserRepository) SetFriendRemark(ctx context.Context, userId, friendId string, remark string) error {
	if err := m.db.Model(&UserRelationModel{}).WithContext(ctx).
		Where("user_id = ? AND friend_id = ? AND deleted_at = 0", userId, friendId).
		Update("remark", remark).Error; err != nil {
	}

	if m.cache != nil {
		if err := m.cache.DeleteRelation(ctx, userId, []string{friendId}); err != nil {
			log.Printf("delete relation cache failed: %v", err)
		}
		if err := m.cache.DeleteRelation(ctx, friendId, []string{userId}); err != nil {
			log.Printf("delete relation cache failed: %v", err)
		}
		if err := m.cache.DeleteFriendList(ctx, userId); err != nil {
			log.Printf("delete friend request list cache failed: %v", err)
		}
	}

	return nil
}
