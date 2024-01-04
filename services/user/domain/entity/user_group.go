package entity

type UserGroup struct {
	GroupID            uint            `gorm:"comment:群组ID" json:"group_id"`
	Status             UserGroupStatus `gorm:"comment:状态（比如已加入、申请中等）" json:"status"`
	Identity           GroupIdentity   `gorm:"comment:身份（比如管理员、普通用户）" json:"identity"`
	EntryMethod        EntryMethod     `gorm:"comment:入群方式" json:"entry_method"`
	JoinedAt           int64           `gorm:"comment:加入时间" json:"joined_at"`
	MuteEndTime        int64           `gorm:"comment:禁言结束时间" json:"mute_end_time"`
	TopAt              int64           `gorm:"comment:置顶时间" json:"top_at"`
	UID                string          `gorm:"type:varchar(64);comment:用户ID" json:"uid"`
	GroupNickname      string          `gorm:"comment:群昵称" json:"group_nickname"`
	Inviter            string          `gorm:"type:varchar(64);comment:邀请人id" json:"inviter"`
	Label              []string        `gorm:"type:varchar(255);comment:标签" json:"label"`
	IsTop              bool            `gorm:"comment:是否置顶" json:"is_top"`
	SilentNotification bool            `gorm:"comment:静默通知" json:"silent_notification"`
	PrivacyMode        bool            `gorm:"comment:隐私模式" json:"privacy_mode"`
}

type UserGroupStatus uint

const (
	StatusJoined   UserGroupStatus = iota + 1 // 已加入
	StatusApplying                            // 申请中
	StatusBlocked                             // 被封禁
	StatusReject                              // 被拒绝
)

type GroupIdentity uint

const (
	IdentityAdmin GroupIdentity = iota + 1 // 管理员
	IdentityUser                           // 普通用户
)

type EntryMethod uint

const (
	EntryInvitation EntryMethod = iota + 1 // 邀请
	EntrySearch                            // 搜索
)
