package entity

type Dialog struct {
	ID        uint32
	CreatedAt int64
	OwnerId   string
	Type      DialogType
	GroupId   uint32
}

type DialogType uint8

const (
	UserDialog = iota
	GroupDialog
)

type DialogUser struct {
	ID        uint32
	CreatedAt int64
	DialogId  uint32
	UserId    string
	IsShow    bool
	TopAt     int64
}
