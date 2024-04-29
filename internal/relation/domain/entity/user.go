package entity

type RequestStatus uint8

// RequestStatus 邀请状态 根据发送者和接受者和状态判断
// Pending 表示申请记录待接受者同意或拒绝申请
// Accepted 对于发送者表示对方已通过申请 对于接受者已通过对方的申请
// Rejected 对于发送者表示对方已拒绝申请 对于接受者已拒绝对方的申请
// Invitation 表示该申请记录是通过邀请发送的 你邀请别人或别人邀请你
// Expired 表示该申请记录已经过期
const (
	Pending    RequestStatus = iota // 等待中
	Accepted                        // 已通过
	Rejected                        // 已拒绝
	Invitation                      // 邀请中
	Expired
)

type UserFriendRequest struct {
	ID         uint32
	CreatedAt  int64
	SenderID   string
	ReceiverID string
	Remark     string
	OwnerID    string
	Status     RequestStatus
}

type UserFriendRequestList struct {
	List  []*UserFriendRequest
	Total int64
}

type UserRelation struct {
	ID                   uint32
	CreatedAt            int64
	Status               UserRelationStatus
	UserID               string
	FriendID             string
	DialogId             uint32
	Remark               string
	Label                []string
	SilentNotification   bool
	OpenBurnAfterReading bool
	// BurnAfterReadingTimeOut 阅后即焚 单位秒
	BurnAfterReadingTimeOut int64
}

type UserRelationStatus uint

const (
	UserStatusBlocked UserRelationStatus = iota // 拉黑
	UserStatusNormal                            // 正常
	UserStatusDeleted                           // 删除
)

type Friend struct {
	UserId                      string
	DialogId                    uint32
	Remark                      string
	Status                      UserRelationStatus
	OpenBurnAfterReading        bool
	IsSilent                    bool
	OpenBurnAfterReadingTimeOut int64
}
