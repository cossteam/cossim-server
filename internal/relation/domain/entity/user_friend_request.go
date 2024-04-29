package entity

type FriendRequestList struct {
	List  []*FriendRequest
	Total int64
}

type FriendRequest struct {
	ID uint32
	// SenderId 发送方id
	SenderId string
	// ReceiverId 接收方id
	ReceiverId string
	Status     RequestStatus
	// Remark 请求备注
	Remark   string
	CreateAt uint64
	// OwnerID 所有者id
	OwnerID string
}

//type FriendRequestStatus int32
//
//const (
//	// FriendRequestPending 申请中
//	FriendRequestPending FriendRequestStatus = 0
//	// FriendRequestAccept 已同意
//	FriendRequestAccept FriendRequestStatus = 1
//	// FriendRequestReject 已拒绝
//	FriendRequestReject FriendRequestStatus = 2
//	// FriendRequestExpired 过期
//	FriendRequestExpired FriendRequestStatus = 3
//)

type Blacklist struct {
	List  []string
	Total int64
}
