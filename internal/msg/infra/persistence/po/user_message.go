package po

type UserMessage struct {
	BaseModel
	Type               uint   `gorm:";comment:消息类型" json:"type"`
	DialogId           uint   `gorm:"default:0;comment:对话ID" json:"dialog_id"`
	IsRead             uint   `gorm:"default:0;comment:是否已读" json:"is_read"`
	ReplyId            uint   `gorm:"default:0;comment:回复ID" json:"reply_id"`
	ReadAt             int64  `gorm:"comment:阅读时间" json:"read_at"`
	ReceiveID          string `gorm:"default:0;comment:接收用户id" json:"receive_id"`
	SendID             string `gorm:"default:0;comment:发送用户id" json:"send_id"`
	Content            string `gorm:"longtext;comment:详细消息" json:"content"`
	IsLabel            uint   `gorm:"default:0;comment:是否标注" json:"is_label"`
	IsBurnAfterReading uint   `gorm:"default:0;comment:是否阅后即焚消息" json:"is_burn_after_reading"`
	ReplyEmoji         string `gorm:"comment:回复时使用的 Emoji" json:"reply_emoji"`
}
