package entity

type UserRelation struct {
	BaseModel
	Status                  UserRelationStatus       `gorm:"comment:好友关系状态 (0=拉黑 1=正常 3=删除 )" json:"status"`
	UserID                  string                   `gorm:"type:varchar(64);comment:用户ID" json:"user_id"`
	FriendID                string                   `gorm:"type:varchar(64);comment:好友ID" json:"friend_id"`
	DialogId                uint                     `gorm:"comment:对话ID" json:"dialog_id"`
	Remark                  string                   `gorm:"type:varchar(255);comment:备注" json:"remark"`
	Label                   []string                 `gorm:"type:varchar(255);comment:标签" json:"label"`
	SilentNotification      SilentNotification       `gorm:"comment:是否开启静默通知" json:"silent_notification"`
	OpenBurnAfterReading    OpenBurnAfterReadingType `gorm:"default:0;comment:是否开启阅后即焚消息" json:"open_burn_after_reading"`
	BurnAfterReadingTimeOut int64                    `gorm:"comment:阅后即焚时间" json:"burn_after_reading_time_out"`
}

type SilentNotification uint

const (
	NotSilentNotification SilentNotification = iota //不开启静默通知
	IsSilentNotification                            //开启静默通知
)

type UserRelationStatus uint

const (
	//正常关系
	UserStatusBlocked UserRelationStatus = iota //拉黑
	UserStatusNormal                            //正常
	UserStatusDeleted                           // 删除
)
