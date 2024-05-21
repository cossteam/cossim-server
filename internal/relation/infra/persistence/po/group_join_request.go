package po

type GroupJoinRequest struct {
	BaseModel
	GroupID   uint32 `gorm:"comment:群聊id"`
	InviterAt int64  `gorm:"comment:邀请时间"`
	ExpiredAt int64  `gorm:"comment:过期时间"`
	UserID    string `gorm:"comment:被邀请人id"`
	Inviter   string `gorm:"comment:邀请人ID"`
	Remark    string `gorm:"comment:邀请备注"`
	OwnerID   string `gorm:"comment:所有者id"`
	Status    uint8  `gorm:"comment:申请记录状态 (0=待处理 1=已接受 2=已拒绝 3=邀请)"`
}

func (m *GroupJoinRequest) TableName() string {
	return "group_join_requests"
}
