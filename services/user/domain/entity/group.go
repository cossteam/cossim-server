package entity

import "gorm.io/gorm"

type Group struct {
	gorm.Model
	Type            GroupType   `gorm:"default:1;comment:群聊类型" json:"type"`
	Status          GroupStatus `gorm:"comment:群聊状态" json:"status"`
	MaxMembersLimit int         `gorm:"comment:群聊人数限制" json:"max_members_limit"`
	CreatorID       string      `gorm:"varchar(128);comment:创建者id" json:"creator_id"`
	Name            string      `gorm:"comment:群聊名称" json:"name"`
	Avatar          string      `gorm:"default:'';comment:头像（群）" json:"avatar"`
}

type GroupType uint

const (
	TypePublic  GroupType = iota + 1 // 公开群
	TypePrivate                      // 私密群
)

type GroupStatus uint

const (
	StatusActive   GroupStatus = iota + 1 // 活跃状态
	StatusInactive                        // 不活跃状态
	StatusClosed                          // 已关闭状态
)
