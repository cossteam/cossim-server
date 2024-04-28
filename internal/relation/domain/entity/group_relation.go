package entity

type GroupRelation struct {
	BaseModel
	GroupID                 uint
	Identity                GroupIdentity
	EntryMethod             EntryMethod
	JoinedAt                int64
	MuteEndTime             int64
	UserID                  string
	Inviter                 string
	Remark                  string
	Label                   []string
	SilentNotification      SilentNotification
	PrivacyMode             bool                     `gorm:"comment:隐私模式" json:"privacy_mode"`
	OpenBurnAfterReading    OpenBurnAfterReadingType `gorm:"default:0;comment:是否开启阅后即焚消息" json:"open_burn_after_reading"`
	BurnAfterReadingTimeOut int64                    `gorm:"default:10;comment:阅后即焚时间" json:"burn_after_reading_time_out"`
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
