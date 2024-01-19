package entity

type DialogUser struct {
	ID        uint   `gorm:"primaryKey;autoIncrement;comment:对话用户ID" json:"id"`
	DialogId  uint   `gorm:"default:0;comment:对话ID" json:"dialog_id"`
	UserId    string `gorm:"default:0;comment:会员ID" json:"user_id"`
	IsShow    int    `gorm:"default:1;comment:对话是否显示" json:"is_show"`
	TopAt     int64  `gorm:"comment:置顶时间" json:"top_at"`
	CreatedAt int64  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt int64  `gorm:"default:0;comment:删除时间" json:"deleted_at"`
}

type ShowSession uint

const (
	NotShow ShowSession = iota // 不显示对话
	IsShow                     // 显示对话
)
