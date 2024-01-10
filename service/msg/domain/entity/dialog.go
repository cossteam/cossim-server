package entity

import "gorm.io/gorm"

type Dialog struct {
	gorm.Model
	OwnerId string     `gorm:"comment:用户id" json:"owner_id"`
	Type    DialogType `gorm:"comment:对话类型" json:"type"`
	GroupId uint       `gorm:"comment:群组id" json:"group_id"`
	//LastAt  int64  `gorm:"comment:最后发送消息时间" json:"last_at"`
}

type DialogType uint32

const (
	UserDialog = iota
	GroupDialog
)
