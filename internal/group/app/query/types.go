package query

type GroupInfo struct {
	Id              uint32
	Avatar          string
	Name            string
	Type            uint
	Status          int
	MaxMembersLimit int32
	CreatorId       string
	DialogId        uint32
	SilenceTime     int64
	JoinApprove     bool
	Encrypt         bool
	Preferences     *Preferences
}

type Group struct {
	Id              uint32
	Type            uint8
	Status          int
	MaxMembersLimit int
	Member          int
	Avatar          string
	Name            string
	CreatorID       string
}

type Preferences struct {
	EntryMethod        uint
	JoinedAt           int64
	MuteEndTime        int64
	SilentNotification bool
	Inviter            string
	Remark             string
	Identity           uint
}
