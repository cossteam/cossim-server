package po

// UserRelation 对应 relation.UserRelation
type UserRelation struct {
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

func (m *UserRelation) TableName() string {
	return "user_relations"
}
