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

// RequestStatus 邀请状态 根据发送者和接受者和状态判断
// Pending 表示申请记录待接受者同意或拒绝申请
// Accepted 对于发送者表示对方已通过申请 对于接受者已通过对方的申请
// Rejected 对于发送者表示对方已拒绝申请 对于接受者已拒绝对方的申请
// Invitation 表示该申请记录是通过邀请发送的 你邀请别人或别人邀请你
type RequestStatus uint

const (
	Pending    RequestStatus = iota // 待处理
	Accepted                        // 已接受
	Rejected                        // 已拒绝
	Invitation                      // 邀请
)
