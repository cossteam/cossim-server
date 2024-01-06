package entity

import "gorm.io/gorm"

type GroupMessage struct {
	gorm.Model
	GroupID   uint            `gorm:"comment:群聊id" json:"group_id"`
	Type      UserMessageType `gorm:"comment:消息类型" json:"type"`
	ReplyId   uint            `gorm:"default:0;comment:回复ID" json:"reply_id"`
	ReadCount int             `gorm:"default:0;comment:已读数量" json:"read_count"`
	UID       string          `gorm:"comment:用户ID" json:"uid"`
	Content   string          `gorm:"longtext;comment:详细消息" json:"content"`
}
