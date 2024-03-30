package entity

type UserFriendRequest struct {
	BaseModel
	SenderID   string `gorm:"comment:发送者id"`
	ReceiverID string `gorm:"comment:接收者id"`
	Remark     string `gorm:"comment:添加备注"`
	OwnerID    string `gorm:"comment:所有者id"`
	Status     RequestStatus
}
