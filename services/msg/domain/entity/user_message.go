package entity

import (
	"gorm.io/gorm"
)

type UserMessage struct {
	gorm.Model
	Type      UserMessageType `gorm:"default:'';comment:消息类型" json:"type"`
	IsRead    uint            `gorm:"default:0;comment:是否已读" json:"is_read"`
	ReplyId   int             `gorm:"default:0;comment:回复ID" json:"reply_id"`
	ReadAt    int64           `gorm:"comment:阅读时间" json:"read_at"`
	ReceiveID string          `gorm:"default:0;comment:接收用户id" json:"receive_id"`
	SendID    string          `gorm:"default:0;comment:发送用户id" json:"send_id"`
	Content   string          `gorm:"longtext;comment:详细消息" json:"content"`
}

type UserMessageType uint

const (
	MessageTypeText      UserMessageType = iota + 1 // 文本消息
	MessageTypeVoice                                // 语音消息
	MessageTypeImage                                // 图片消息
	MessageTypeFile                                 // 文件消息
	MessageTypeVideo                                // 视频消息
	MessageTypeEmoji                                // Emoji表情
	MessageTypeSticker                              // 表情包
	MessageTypeVoiceCall                            // 语音通话
	MessageTypeVideoCall                            // 视频通话
)
