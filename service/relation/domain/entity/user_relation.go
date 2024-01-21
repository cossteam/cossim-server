package entity

type UserRelation struct {
	BaseModel
	Status             UserRelationStatus `gorm:"comment:好友关系状态 (0=申请中 1=待通过 2=已添加 3=已拒绝 4=拉黑 5=删除)" json:"status"`
	UserID             string             `gorm:"type:varchar(64);comment:用户ID" json:"user_id"`
	FriendID           string             `gorm:"type:varchar(64);comment:好友ID" json:"friend_id"`
	DialogId           uint               `gorm:"comment:对话ID" json:"dialog_id"`
	Remark             string             `gorm:"type:varchar(255);comment:备注" json:"remark"`
	Label              []string           `gorm:"type:varchar(255);comment:标签" json:"label"`
	SilentNotification bool               `gorm:"comment:静默通知" json:"silent_notification"`
}

type UserRelationStatus uint

const (
	UserStatusApplying UserRelationStatus = iota // 申请中
	UserStatusPending                            // 待通过
	UserStatusAdded                              // 已添加
	UserStatusRejected                           // 已拒绝
	UserStatusBlocked                            // 拉黑
	UserStatusDeleted                            // 删除
)
