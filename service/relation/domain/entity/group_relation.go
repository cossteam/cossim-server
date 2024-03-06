package entity

type GroupRelation struct {
	BaseModel
	GroupID                 uint                     `gorm:"comment:群组ID" json:"group_id"`
	Identity                GroupIdentity            `gorm:"comment:身份 (0=普通用户, 1=管理员, 2=群主)" json:"identity"`
	EntryMethod             EntryMethod              `gorm:"comment:入群方式" json:"entry_method"`
	JoinedAt                int64                    `gorm:"comment:加入时间" json:"joined_at"`
	MuteEndTime             int64                    `gorm:"comment:禁言结束时间" json:"mute_end_time"`
	UserID                  string                   `gorm:"type:varchar(64);comment:用户ID" json:"user_id"`
	GroupNickname           string                   `gorm:"comment:群昵称" json:"group_nickname"`
	Inviter                 string                   `gorm:"type:varchar(64);comment:邀请人id" json:"inviter"`
	Remark                  string                   `gorm:"type:varchar(255);comment:添加群聊备注" json:"remark"`
	Label                   []string                 `gorm:"type:varchar(255);comment:标签" json:"label"`
	SilentNotification      SilentNotification       `gorm:"comment:是否开启静默通知" json:"silent_notification"`
	PrivacyMode             bool                     `gorm:"comment:隐私模式" json:"privacy_mode"`
	OpenBurnAfterReading    OpenBurnAfterReadingType `gorm:"default:0;comment:是否开启阅后即焚消息" json:"open_burn_after_reading"`
	BurnAfterReadingTimeOut int64                    `gorm:"comment:阅后即焚时间" json:"burn_after_reading_time_out"`
}

type GroupIdentity uint

const (
	IdentityUser  GroupIdentity = iota // 普通用户
	IdentityAdmin                      // 管理员
	IdentityOwner                      // 群主
)

type EntryMethod uint

const (
	EntryInvitation EntryMethod = iota // 邀请
	EntrySearch                        // 搜索
)

type OpenBurnAfterReadingType uint

const (
	CloseBurnAfterReading OpenBurnAfterReadingType = iota //关闭阅后即焚
	OpenBurnAfterReading                                  //开启阅后即焚消息
)
