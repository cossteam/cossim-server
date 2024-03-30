package entity

type GroupJoinRequest struct {
	BaseModel
	GroupID     uint
	InviterTime int64  `gorm:"comment:邀请时间"`
	UserID      string `gorm:"comment:被邀请人id"`
	Inviter     string `gorm:"comment:邀请人ID"`
	Remark      string `gorm:"comment:邀请备注"`
	OwnerID     string `gorm:"comment:所有者id"`
	Status      RequestStatus
}

type RequestStatus uint

const (
	Pending RequestStatus = iota
	Accepted
	Rejected
	Invitation
)
