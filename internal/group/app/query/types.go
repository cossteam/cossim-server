package query

type GroupInfo struct {
	Id              uint32
	Avatar          string
	Name            string
	Type            uint32
	Status          int
	MaxMembersLimit int32
	CreatorId       string
	DialogId        uint32
	SilenceTime     int64
	JoinApprove     bool
	Encrypt         bool
	Preferences     *Preferences
}

type Preferences struct {
	EntryMethod          uint
	JoinedAt             int64
	MuteEndTime          int64
	SilentNotification   uint
	Inviter              string
	Remark               string
	OpenBurnAfterReading uint
	Identity             uint
}
