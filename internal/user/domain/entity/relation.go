package entity

type Relation struct {
	DialogID                    uint32
	Status                      UserRelationStatus
	OpenBurnAfterReading        bool
	SilentNotification          bool
	OpenBurnAfterReadingTimeOut int64
	Remark                      string
}
