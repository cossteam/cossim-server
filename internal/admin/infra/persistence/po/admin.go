package po

import (
	"github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

type Admin struct {
	ID        uint   `gorm:"primaryKey;autoIncrement;" json:"id"`
	UserId    string `gorm:"comment:管理员id" json:"user_id"`
	Role      uint   `gorm:"comment:管理员角色" json:"role"`
	Status    uint   `gorm:"comment:管理员状态" json:"status"`
	CreatedAt int64  `gorm:"autoCreateTime;comment:创建时间" json:"created_at"`
	UpdatedAt int64  `gorm:"autoUpdateTime;comment:更新时间" json:"updated_at"`
	DeletedAt int64  `gorm:"default:0;comment:删除时间" json:"deleted_at"`
}

func (bm *Admin) BeforeCreate(tx *gorm.DB) error {
	now := time.Now()
	bm.CreatedAt = now
	bm.UpdatedAt = now
	return nil
}

func (bm *Admin) BeforeUpdate(tx *gorm.DB) error {
	bm.UpdatedAt = time.Now()
	return nil
}
