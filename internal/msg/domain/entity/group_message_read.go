package entity

type GroupMessageRead struct {
	BaseModel
	MsgId    uint   `gorm:"comment:消息ID" json:"msg_id"`
	DialogId uint   `gorm:"default:0;comment:对话ID" json:"dialog_id"`
	GroupID  uint   `gorm:"comment:群聊id" json:"group_id"`
	ReadAt   int64  `gorm:"comment:已读时间" json:"read_at"`
	UserId   string `gorm:"comment:用户ID" json:"user_id"`
}
