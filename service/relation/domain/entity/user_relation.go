package entity

type UserRelation struct {
	ID                 uint               `gorm:"primaryKey;autoIncrement;comment:好友关系ID" json:"id"`
	Status             UserRelationStatus `gorm:"comment:好友关系状态" json:"status"`
	UserID             string             `gorm:"type:varchar(64);comment:用户ID" json:"user_id"`
	FriendID           string             `gorm:"type:varchar(64);comment:好友ID" json:"friend_id"`
	DialogId           uint               `gorm:"comment:对话ID" json:"dialog_id"`
	Remark             string             `gorm:"type:varchar(255);comment:备注" json:"remark"`
	Label              []string           `gorm:"type:varchar(255);comment:标签" json:"label"`
	SilentNotification bool               `gorm:"comment:静默通知" json:"silent_notification"`
	CreatedAt          int64              `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt          int64              `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt          int64              `gorm:"default:null;comment:删除时间" json:"deleted_at"`
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
