package po

import (
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

type File struct {
	ID        string `gorm:"type:char(64);primary_key;comment:文件id"`
	Name      string `gorm:"type:varchar(50);comment:文件名"`
	Owner     string `gorm:"type:char(64);comment:所属者id"`
	Content   string `gorm:"type:text;comment:文件内容"`
	Path      string `gorm:"type:text;comment:文件路径"`
	Type      uint   `gorm:"comment:文件类型"`
	Status    uint   `gorm:"comment:文件状态"`
	Provider  string `gorm:"default:MinIO;comment:文件供应商"`
	Share     bool   `gorm:"comment:是否共享"`
	Size      uint64 `gorm:"comment:文件大小"`
	CreatedAt int64  `gorm:"autoCreateTime;comment:创建时间"`
	UpdatedAt int64  `gorm:"autoUpdateTime;comment:更新时间"`
	DeletedAt int64  `gorm:"default:0;comment:删除时间"`
}

func (bm *File) BeforeCreate(tx *gorm.DB) error {
	now := ptime.Now()
	bm.CreatedAt = now
	bm.UpdatedAt = now
	return nil
}

func (bm *File) BeforeUpdate(tx *gorm.DB) error {
	bm.UpdatedAt = ptime.Now()
	return nil
}

func (bm *File) TableName() string {
	return "files"
}
