package po

import (
	"github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

type GroupMessage struct {
	BaseModel
	DialogId           uint     `gorm:"default:0;comment:对话ID" json:"dialog_id"`
	GroupID            uint     `gorm:"comment:群聊id" json:"group_id"`
	Type               uint     `gorm:"comment:消息类型" json:"type"`
	ReplyId            uint     `gorm:"default:0;comment:回复ID" json:"reply_id"`
	ReadCount          int      `gorm:"default:0;comment:已读数量" json:"read_count"`
	UserID             string   `gorm:"comment:用户ID" json:"user_id"`
	Content            string   `gorm:"longtext;comment:详细消息" json:"content"`
	IsLabel            uint     `gorm:"default:0;comment:是否标注" json:"is_label"`
	ReplyEmoji         string   `gorm:"comment:回复时使用的 Emoji" json:"reply_emoji"`
	AtAllUser          uint     `gorm:"default:0;comment:是否at全体用户" json:"at_all_users"`
	AtUsers            []string `gorm:"serializer:json;comment:at的用户" json:"at_users"`
	IsBurnAfterReading bool     `gorm:"default:0;comment:是否阅后即焚消息" json:"is_burn_after_reading"`
}

type BaseModel struct {
	ID        uint  `gorm:"primaryKey;autoIncrement;" json:"id"`
	CreatedAt int64 `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt int64 `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt int64 `gorm:"default:0;comment:删除时间" json:"deleted_at"`
}

func (bm *BaseModel) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	bm.CreatedAt = now
	bm.UpdatedAt = now
	return nil
}

func (bm *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	bm.UpdatedAt = time.Now()
	return nil
}
