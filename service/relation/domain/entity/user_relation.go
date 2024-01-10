package entity

type UserRelation struct {
	ID                 uint           `gorm:"primaryKey;autoIncrement;comment:好友关系ID" json:"id"`
	Status             RelationStatus `gorm:"comment:好友关系状态" json:"status"`
	UserID             string         `gorm:"type:varchar(64);comment:用户ID" json:"user_id"`
	FriendID           string         `gorm:"type:varchar(64);comment:好友ID" json:"friend_id"`
	Remark             string         `gorm:"type:varchar(255);comment:备注" json:"remark"`
	Label              []string       `gorm:"type:varchar(255);comment:标签" json:"label"`
	SilentNotification bool           `gorm:"comment:静默通知" json:"silent_notification"`
	CreatedAt          int64          `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt          int64          `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt          int64          `gorm:"comment:删除时间" json:"deleted_at"`
}

type RelationStatus uint

const (
	RelationStatusPending  RelationStatus = iota + 1 // 申请中
	RelationStatusAdded                              // 已添加
	RelationStatusRejected                           // 已拒绝
	RelationStatusBlocked                            // 拉黑
	RelationStatusDeleted                            // 删除
)
