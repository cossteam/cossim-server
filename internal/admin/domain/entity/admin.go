package entity

import (
	"github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

type Admin struct {
	BaseModel
	UserId string      `gorm:"comment:管理员id" json:"user_id"`
	Role   Role        `gorm:"comment:管理员角色" json:"role"`
	Status AdminStatus `gorm:"comment:管理员状态" json:"status"`
}

type BaseModel struct {
	ID        uint  `gorm:"primaryKey;autoIncrement;" json:"id"`
	CreatedAt int64 `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt int64 `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt int64 `gorm:"default:0;comment:删除时间" json:"deleted_at"`
}

func (bm *BaseModel) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	bm.CreatedAt = now
	bm.UpdatedAt = now
	return nil
}

func (bm *BaseModel) BeforeUpdate(tx *gorm.DB) error {
	bm.UpdatedAt = time.Now()
	return nil
}

type Role uint

const (
	SuperAdminRole Role = 1 //超级管理员
	AdminRole      Role = 2 //管理员
)

type AdminStatus uint

const (
	NormalStatus   AdminStatus = 1
	DisabledStatus AdminStatus = 2
)
