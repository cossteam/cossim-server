package po

type GroupRelation struct {
	BaseModel
	GroupID            uint32   `gorm:"comment:群组ID" json:"group_id"`
	Identity           uint8    `gorm:"comment:身份 (0=普通用户, 1=管理员, 2=群主)"`
	EntryMethod        uint8    `gorm:"comment:入群方式"`
	JoinedAt           int64    `gorm:"comment:加入时间"`
	MuteEndTime        int64    `gorm:"comment:禁言结束时间"`
	UserID             string   `gorm:"type:varchar(64);comment:用户ID"`
	Inviter            string   `gorm:"type:varchar(64);comment:邀请人id"`
	Remark             string   `gorm:"type:varchar(255);comment:添加群聊备注"`
	Label              []string `gorm:"type:varchar(255);comment:标签"`
	SilentNotification bool     `gorm:"comment:是否开启静默通知"`
	PrivacyMode        bool     `gorm:"comment:隐私模式"`
}

func (m *GroupRelation) TableName() string {
	return "group_relations"
}
