package entity

type UserGroup struct {
	ID                 uint            `gorm:"primaryKey;autoIncrement;comment:群组关系ID" json:"id"`
	GroupID            uint            `gorm:"comment:群组ID" json:"group_id"`
	Status             UserGroupStatus `gorm:"comment:状态（比如已加入、申请中等）" json:"status"`
	Identity           GroupIdentity   `gorm:"comment:身份（比如管理员、普通用户）" json:"identity"`
	EntryMethod        EntryMethod     `gorm:"comment:入群方式" json:"entry_method"`
	JoinedAt           int64           `gorm:"comment:加入时间" json:"joined_at"`
	MuteEndTime        int64           `gorm:"comment:禁言结束时间" json:"mute_end_time"`
	UID                string          `gorm:"type:varchar(64);comment:用户ID" json:"uid"`
	GroupNickname      string          `gorm:"comment:群昵称" json:"group_nickname"`
	Inviter            string          `gorm:"type:varchar(64);comment:邀请人id" json:"inviter"`
	Remark             string          `gorm:"type:varchar(255);comment:添加好友备注" json:"remark"`
	Label              []string        `gorm:"type:varchar(255);comment:标签" json:"label"`
	SilentNotification bool            `gorm:"comment:静默通知" json:"silent_notification"`
	PrivacyMode        bool            `gorm:"comment:隐私模式" json:"privacy_mode"`
	CreatedAt          int64           `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt          int64           `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt          int64           `gorm:"comment:删除时间" json:"deleted_at"`
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
