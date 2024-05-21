package po

type Dialog struct {
	BaseModel
	OwnerId string `gorm:"comment:用户id"`
	Type    uint8  `gorm:"comment:对话类型"`
	GroupId uint32 `gorm:"comment:群组id"`
}

func (m *Dialog) TableName() string {
	return "dialogs"
}

type DialogUser struct {
	BaseModel
	DialogId uint32 `gorm:"default:0;comment:对话ID"`
	UserId   string `gorm:"default:0;comment:会员ID"`
	IsShow   bool   `gorm:"default:false;comment:对话是否显示"`
	TopAt    int64  `gorm:"comment:置顶时间"`
}

func (m *DialogUser) TableName() string {
	return "dialog_users"
}
