package po

type UserFriendRequest struct {
	BaseModel
	SenderID   string `gorm:"comment:发送者id"`
	ReceiverID string `gorm:"comment:接收者id"`
	Remark     string `gorm:"comment:添加备注"`
	OwnerID    string `gorm:"comment:所有者id"`
	Status     uint8  `gorm:"comment:申请记录状态 (0=申请中 1=已同意 2=已拒绝)"`
	ExpiredAt  int64  `gorm:"comment:过期时间"`
}

func (m *UserFriendRequest) TableName() string {
	return "user_friend_requests"
}
