package po

import (
	"fmt"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint32 `gorm:"primaryKey;autoIncrement;"`
	CreatedAt int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt int64  `gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt int64  `gorm:"default:0;comment:删除时间"`
}

type Group struct {
	BaseModel
	Type            uint   `gorm:"default:0;comment:群聊类型(0=私密群, 1=公开群)"`
	Status          uint   `gorm:"comment:群聊状态(0=正常状态, 1=锁定状态, 2=删除状态)"`
	MaxMembersLimit int    `gorm:"comment:群聊人数限制"`
	CreatorID       string `gorm:"type:varchar(64);comment:创建者id"`
	Name            string `gorm:"comment:群聊名称"`
	Avatar          string `gorm:"default:'';comment:头像（群）"`
	SilenceTime     int64  `gorm:"comment:全员禁言结束时间"`
	JoinApprove     bool   `gorm:"default:false;comment:是否开启入群验证"`
	Encrypt         bool   `gorm:"default:false;comment:是否开启群聊加密，只有当群聊类型为私密群时，该字段才有效"`
}

func (bm *BaseModel) TableName() string {
	fmt.Println("table name")
	return "groups"
}

func (bm *BaseModel) BeforeCreate(tx *gorm.DB) error {
	now := ptime.Now()
	bm.CreatedAt = now
	bm.UpdatedAt = now
	return nil
}

func (bm *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	bm.UpdatedAt = ptime.Now()
	return nil
}
