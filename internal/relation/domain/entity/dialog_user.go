package entity

type DialogUser struct {
	BaseModel
	DialogId uint   `gorm:"default:0;comment:对话ID" json:"dialog_id"`
	UserId   string `gorm:"default:0;comment:会员ID" json:"user_id"`
	IsShow   int    `gorm:"default:1;comment:对话是否显示" json:"is_show"`
	TopAt    int64  `gorm:"comment:置顶时间" json:"top_at"`
}

type ShowSession uint

const (
	NotShow ShowSession = iota // 不显示对话
	IsShow                     // 显示对话
)
