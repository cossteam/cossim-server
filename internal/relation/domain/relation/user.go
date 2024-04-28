package relation

type RequestStatus uint

const (
	Pending    RequestStatus = iota // 等待中
	Accepted                        // 已通过
	Rejected                        // 已拒绝
	Invitation                      // 邀请中
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
