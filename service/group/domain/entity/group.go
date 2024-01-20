package entity

import "gorm.io/gorm"

type Group struct {
	gorm.Model
	Type            GroupType   `gorm:"default:0;comment:群聊类型(0=私密群, 1=公开群)" json:"type"`
	Status          GroupStatus `gorm:"comment:群聊状态(0=正常状态, 1=锁定状态, 2=删除状态)" json:"status"`
	MaxMembersLimit int         `gorm:"comment:群聊人数限制" json:"max_members_limit"`
	CreatorID       string      `gorm:"type:varchar(64);comment:创建者id" json:"creator_id"`
	Name            string      `gorm:"comment:群聊名称" json:"name"`
	Avatar          string      `gorm:"default:'';comment:头像（群）" json:"avatar"`
}

type GroupType uint

const (
	TypePublic  GroupType = iota // 公开群
	TypePrivate                  // 私密群
)

type GroupStatus uint

const (
	GroupStatusNormal  GroupStatus = iota // 正常状态
	GroupStatusLocked                     // 锁定状态
	GroupStatusDeleted                    // 删除状态
)
