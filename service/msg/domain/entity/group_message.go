package entity

type GroupMessage struct {
	BaseModel
	DialogId  uint            `gorm:"default:0;comment:对话ID" json:"dialog_id"`
	GroupID   uint            `gorm:"comment:群聊id" json:"group_id"`
	Type      UserMessageType `gorm:"comment:消息类型" json:"type"`
	ReplyId   uint            `gorm:"default:0;comment:回复ID" json:"reply_id"`
	ReadCount int             `gorm:"default:0;comment:已读数量" json:"read_count"`
	UID       string          `gorm:"comment:用户ID" json:"uid"`
	Content   string          `gorm:"longtext;comment:详细消息" json:"content"`
}

type BaseModel struct {
	ID        uint  `gorm:"primaryKey;autoIncrement;" json:"id"`
	CreatedAt int64 `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt int64 `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt int64 `gorm:"default:0;comment:删除时间" json:"deleted_at"`
}
