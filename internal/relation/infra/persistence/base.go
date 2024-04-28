package persistence

import (
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint32 `gorm:"primaryKey;autoIncrement;"`
	CreatedAt int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt int64  `gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt int64  `gorm:"default:0;comment:删除时间"`
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
