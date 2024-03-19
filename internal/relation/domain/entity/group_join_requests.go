package entity

type GroupJoinRequest struct {
	BaseModel
	UserID  string        `json:"user_id" gorm:"column:user_id"`
	GroupID uint          `json:"group_id" gorm:"column:group_id"`
	Status  RequestStatus `json:"status" gorm:"column:status"`
	//邀请人
	Inviter string `json:"inviter" gorm:"column:inviter"`
	//邀请时间
	InviterTime int64  `json:"inviter_time" gorm:"column:inviter_time"`
	Remark      string `json:"remark" gorm:"column:remark"`
}

type RequestStatus uint

const (
	Pending RequestStatus = iota
	Accepted
	Rejected
	Invitation
)
