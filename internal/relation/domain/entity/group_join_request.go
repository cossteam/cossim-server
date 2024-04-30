package entity

type GroupJoinRequestList struct {
	List  []*GroupJoinRequest
	Total int64
}

type GroupJoinRequest struct {
	ID        uint32
	GroupID   uint32
	CreatedAt int64
	InviterAt int64
	ExpiredAt int64
	UserID    string
	Inviter   string
	Remark    string
	OwnerID   string
	Status    RequestStatus
}
