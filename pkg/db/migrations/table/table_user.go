package table

import (
	"github.com/cossim/coss-server/service/user/domain/entity"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func (d InitDatabase) AddTableUser() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202401031401",
		Migrate: func(tx *gorm.DB) error {
			// 执行迁移操作，例如创建表
			return tx.AutoMigrate(&entity.User{})
		},
	}
}

func (d InitDatabase) AddTableGroup() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202401031400",
		Migrate: func(tx *gorm.DB) error {
			// 执行迁移操作，例如创建表
			return tx.AutoMigrate(&entity.Group{})
		},
	}
}
