package query

type GroupInfo struct {
	Id              uint32       `json:"id"`
	Avatar          string       `json:"avatar"`
	Name            string       `json:"name"`
	Type            uint32       `json:"type"`
	Status          int          `json:"status"`
	MaxMembersLimit int32        `json:"max_members_limit"`
	CreatorId       string       `json:"creator_id"`
	DialogId        uint32       `json:"dialog_id"`
	Preferences     *Preferences `json:"preferences,omitempty"`
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
