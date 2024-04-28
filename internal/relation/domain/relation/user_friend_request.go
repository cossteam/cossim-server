package relation

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
	Status     FriendRequestStatus
	// Remark 请求备注
	Remark   string
	CreateAt uint64
	// OwnerID 所有者id
	OwnerID string
}

type FriendRequestStatus int32

const (
	// FriendRequestStatus_FriendRequestStatus_PENDING 申请中
	FriendRequestStatus_FriendRequestStatus_PENDING FriendRequestStatus = 0
	// FriendRequestStatus_FriendRequestStatus_ACCEPT 已同意
	FriendRequestStatus_FriendRequestStatus_ACCEPT FriendRequestStatus = 1
	// FriendRequestStatus_FriendRequestStatus_REJECT 已拒绝
	FriendRequestStatus_FriendRequestStatus_REJECT FriendRequestStatus = 2
)

type Blacklist struct {
	List  []string
	Total int64
}
