package persistence

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/cache"
	"github.com/cossim/coss-server/internal/relation/domain/relation"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
	"log"
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

func (m *UserRelationModel) FromEntity(u *relation.UserRelation) error {
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

func (m *UserRelationModel) ToEntity() *relation.UserRelation {
	return &relation.UserRelation{
		ID:                   m.ID,
		CreatedAt:            m.CreatedAt,
		UserID:               m.UserID,
		FriendID:             m.FriendID,
		Status:               relation.UserRelationStatus(m.Status),
		DialogId:             m.DialogId,
		Remark:               m.Remark,
		Label:                m.Label,
		SilentNotification:   m.SilentNotification,
		OpenBurnAfterReading: m.OpenBurnAfterReading,
	}
}

var _ relation.UserRepository = &MySQLRelationUserRepository{}

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

func (m *MySQLRelationUserRepository) RestoreFriendship(ctx context.Context, dialogID uint32, userId, friendId string) error {
	if err := m.db.WithContext(ctx).
		Model(&UserRelationModel{}).
		Where("dialog_id = ? AND user_id = ? AND friend_id = ?", dialogID, userId, friendId).
		Update("status", relation.UserStatusNormal).
		Update("deleted_at", 0).
		Error; err != nil {
		return err
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

func (m *MySQLRelationUserRepository) EstablishFriendship(ctx context.Context, dialogID uint32, senderID, receiverID string) error {
	models := []UserRelationModel{
		{
			UserID:   senderID,
			FriendID: receiverID,
			Status:   uint(relation.UserStatusNormal),
			DialogId: dialogID,
		},
		{
			UserID:   receiverID,
			FriendID: senderID,
			Status:   uint(relation.UserStatusNormal),
			DialogId: dialogID,
		},
	}

	if err := m.db.WithContext(ctx).Create(&models).Error; err != nil {
		return err
	}

	return nil
}

func (m *MySQLRelationUserRepository) UpdateStatus(ctx context.Context, id uint32, status relation.UserRelationStatus) error {
	var model UserRelationModel

	if err := m.db.WithContext(ctx).
		Model(&UserRelationModel{}).
		Where("id = ?", id).
		First(&model).
		Update("status", uint(status)).Error; err != nil {
		return err
	}

	if m.cache != nil {
		if err := m.cache.DeleteRelation(ctx, model.UserID, []string{model.FriendID}); err != nil {
			log.Printf("delete relation cache error: %v", err)
		}
		if err := m.cache.DeleteFriendList(ctx, model.UserID); err != nil {
			log.Printf("failed to delete cache friend list: %v", err)
		}
	}

	return nil
}

func (m *MySQLRelationUserRepository) Get(ctx context.Context, userId, friendId string) (*relation.UserRelation, error) {
	var model UserRelationModel

	if err := m.db.WithContext(ctx).
		Where("user_id = ? AND friend_id = ? AND deleted_at = 0", userId, friendId).
		First(&model).Error; err != nil {
		return nil, err
	}

	return model.ToEntity(), nil
}

func (m *MySQLRelationUserRepository) Create(ctx context.Context, ur *relation.UserRelation) (*relation.UserRelation, error) {
	var model UserRelationModel

	if err := model.FromEntity(ur); err != nil {
		return nil, err
	}

	if err := m.db.Create(&model).Error; err != nil {
		return nil, err
	}

	entity := model.ToEntity()

	return entity, nil
}

func (m *MySQLRelationUserRepository) Update(ctx context.Context, ur *relation.UserRelation) (*relation.UserRelation, error) {
	var model UserRelationModel

	if err := model.FromEntity(ur); err != nil {
		return nil, err
	}

	if err := m.db.Where("id = ?", ur.ID).Updates(&model).Error; err != nil {
		return nil, err
	}

	entity := model.ToEntity()

	if m.cache != nil {
		if err := m.cache.DeleteRelation(ctx, ur.UserID, []string{ur.FriendID}); err != nil {
			log.Printf("delete relation cache error: %v", err)
		}

		if err := m.cache.DeleteFriendList(ctx, ur.UserID, ur.FriendID); err != nil {
			log.Printf("failed to delete cache friend list: %v", err)
		}
	}

	return entity, nil
}

func (m *MySQLRelationUserRepository) Delete(ctx context.Context, userId, friendId string) error {
	if err := m.db.Model(&UserRelationModel{}).WithContext(ctx).
		Where("user_id = ? AND friend_id = ? AND deleted_at = 0", userId, friendId).
		Update("status", relation.UserStatusDeleted).
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

func (m *MySQLRelationUserRepository) Find(ctx context.Context, query *relation.UserQuery) ([]*relation.UserRelation, error) {
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

	var list []*relation.UserRelation
	for _, v := range models {
		list = append(list, v.ToEntity())
	}

	if m.cache != nil {
		for _, v := range list {
			if err := m.cache.SetRelation(ctx, v.UserID, v.FriendID, v, cache.RelationExpireTime); err != nil {
				log.Printf("failed to set relation cache: %v", err)
			}
			if err := m.cache.SetRelation(ctx, v.FriendID, v.UserID, v, cache.RelationExpireTime); err != nil {
				log.Printf("failed to set relation cache: %v", err)
			}
		}
	}

	return list, nil
}

func (m *MySQLRelationUserRepository) Blacklist(ctx context.Context, userId string) (*relation.Blacklist, error) {
	if m.cache != nil {
		cachedList, err := m.cache.GetBlacklist(ctx, userId)
		if err == nil && cachedList != nil {
			return cachedList, nil
		}
	}

	var relations []*UserRelationModel
	if err := m.db.WithContext(ctx).Where("user_id = ? AND status = ? AND deleted_at = 0", userId, relation.UserStatusBlocked).
		Find(&relations).Error; err != nil {
		return nil, err
	}

	blacklist := &relation.Blacklist{}
	for _, v := range relations {
		blacklist.List = append(blacklist.List, v.FriendID)
		blacklist.Total++
	}

	if m.cache != nil {
		if err := m.cache.SetBlacklist(ctx, userId, blacklist, cache.RelationExpireTime); err != nil {
			log.Printf("failed to set blacklist cache: %v", err)
		}
	}

	return blacklist, nil
}

func (m *MySQLRelationUserRepository) FriendRequestList(ctx context.Context, userId string) ([]*relation.Friend, error) {
	if m.cache != nil {
		r, err := m.cache.GetFriendList(ctx, userId)
		if err == nil && r != nil {
			return r, nil
		}
	}

	var relations []*UserRelationModel
	if err := m.db.WithContext(ctx).
		Where("user_id = ? AND status NOT IN ? AND deleted_at = 0", userId,
			[]relation.UserRelationStatus{relation.UserStatusBlocked, relation.UserStatusDeleted}).
		Find(&relations).Error; err != nil {
		return nil, err
	}

	var friends []*relation.Friend
	for _, v := range relations {
		friends = append(friends, &relation.Friend{
			UserId:                      v.FriendID,
			DialogId:                    v.DialogId,
			Remark:                      v.Remark,
			Status:                      relation.UserRelationStatus(v.Status),
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

	if m.cache != nil {
		if err := m.cache.DeleteRelation(ctx, userId, []string{friendId}); err != nil {
			log.Printf("delete relation cache failed: %v", err)
		}
		if err := m.cache.DeleteFriendList(ctx, userId); err != nil {
			log.Printf("delete friend request list cache failed: %v", err)
		}
	}

	return nil
}

func (m *MySQLRelationUserRepository) SetUserOpenBurnAfterReading(ctx context.Context, userId, friendId string, openBurnAfterReading bool, burnAfterReadingTimeOut int64) error {
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
		if err := m.cache.DeleteFriendList(ctx, userId); err != nil {
			log.Printf("delete friend request list cache failed: %v", err)
		}
	}

	return nil
}
