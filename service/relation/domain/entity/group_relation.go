package entity

type GroupRelation struct {
	ID                 uint                `gorm:"primaryKey;autoIncrement;comment:群组关系ID" json:"id"`
	GroupID            uint                `gorm:"comment:群组ID" json:"group_id"`
	Status             GroupRelationStatus `gorm:"comment:状态（0:申请中 1:已加入 2:被拒绝 3:被封禁）" json:"status"`
	Identity           GroupIdentity       `gorm:"comment:身份（比如管理员、普通用户）" json:"identity"`
	EntryMethod        EntryMethod         `gorm:"comment:入群方式" json:"entry_method"`
	JoinedAt           int64               `gorm:"comment:加入时间" json:"joined_at"`
	MuteEndTime        int64               `gorm:"comment:禁言结束时间" json:"mute_end_time"`
	UserID             string              `gorm:"type:varchar(64);comment:用户ID" json:"user_id"`
	GroupNickname      string              `gorm:"comment:群昵称" json:"group_nickname"`
	Inviter            string              `gorm:"type:varchar(64);comment:邀请人id" json:"inviter"`
	Remark             string              `gorm:"type:varchar(255);comment:添加群聊备注" json:"remark"`
	Label              []string            `gorm:"type:varchar(255);comment:标签" json:"label"`
	SilentNotification bool                `gorm:"comment:静默通知" json:"silent_notification"`
	PrivacyMode        bool                `gorm:"comment:隐私模式" json:"privacy_mode"`
	CreatedAt          int64               `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt          int64               `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt          int64               `gorm:"default:0;comment:删除时间" json:"deleted_at"`
}

type GroupRelationStatus uint

const (
	GroupStatusApplying GroupRelationStatus = iota // 申请中
	GroupStatusJoined                              // 已加入
	GroupStatusReject                              // 被拒绝
	GroupStatusBlocked                             // 被封禁
)

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
