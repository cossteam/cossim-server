package entity

type GroupRelationList struct {
	List  []*GroupRelation
	Total int64
}

type GroupRelation struct {
	ID                      uint32
	CreatedAt               int64
	GroupID                 uint32
	Identity                GroupIdentity
	EntryMethod             EntryMethod
	JoinedAt                int64
	MuteEndTime             int64
	UserID                  string
	Inviter                 string
	Remark                  string
	Label                   []string
	SilentNotification      bool
	PrivacyMode             bool
	OpenBurnAfterReading    bool
	BurnAfterReadingTimeOut int64
}

type GroupIdentity uint8

const (
	IdentityUser  GroupIdentity = iota // 普通用户
	IdentityAdmin                      // 管理员
	IdentityOwner                      // 群主
)

type EntryMethod uint

const (
	EntryInvitation EntryMethod = iota // 邀请
	EntrySearch                        // 搜索
)

type CreateGroupRelation struct {
	GroupID     uint32
	UserID      string
	Identity    GroupIdentity
	EntryMethod EntryMethod
	Inviter     string
	JoinedAt    int64
}
