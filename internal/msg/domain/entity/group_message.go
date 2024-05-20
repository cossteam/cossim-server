package entity

import (
	v1 "github.com/cossim/coss-server/internal/msg/api/http/v1"
)

type GroupMessage struct {
	BaseModel
	DialogID           uint
	GroupID            uint
	Type               UserMessageType
	ReplyId            uint
	ReadCount          int
	UserID             string
	Content            string
	IsLabel            uint
	ReplyEmoji         string
	AtAllUser          AtAllUserType
	AtUsers            []string
	IsBurnAfterReading bool
}

type AtAllUserType uint

const (
	NotAtAllUser = iota
	AtAllUser
)

type BaseModel struct {
	ID        uint  `gorm:"primaryKey;autoIncrement;" json:"id"`
	CreatedAt int64 `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt int64 `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt int64 `gorm:"default:0;comment:删除时间" json:"deleted_at"`
}

func (gm *GroupMessage) ToMessage() *v1.Message {
	return &v1.Message{
		AtAllUser:          gm.AtAllUser == AtAllUser,
		AtUsers:            gm.AtUsers,
		Content:            gm.Content,
		GroupId:            int(gm.GroupID),
		IsBurnAfterReading: gm.IsBurnAfterReading,
		IsLabel:            gm.IsLabel == uint(IsLabel),
		IsRead:             gm.ReadCount > 0, // 根据 ReadCount 判断是否已读
		MsgId:              int(gm.ID),
		MsgType:            int(gm.Type),
		ReplyId:            int(gm.ReplyId),
		SendAt:             int(gm.CreatedAt),
		SenderId:           gm.UserID, // 或者根据实际情况选择其他字段
		DialogId:           int(gm.DialogID),
	}
}
