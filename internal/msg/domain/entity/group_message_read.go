package entity

type GroupMessageRead struct {
	BaseModel
	MsgID    uint
	DialogID uint
	GroupID  uint
	ReadAt   int64
	UserID   string
}
