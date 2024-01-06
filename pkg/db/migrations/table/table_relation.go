package table

import (
	"github.com/cossim/coss-server/service/relation/domain/entity"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func (d InitDatabase) AddTableUserRelation() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202401031301",
		Migrate: func(tx *gorm.DB) error {
			// 执行迁移操作，例如创建表
			return tx.AutoMigrate(&entity.UserRelation{})
		},
	}
}
func (d InitDatabase) AddTableUserGroup() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "202401031301",
		Migrate: func(tx *gorm.DB) error {
			// 执行迁移操作，例如创建表
			return tx.AutoMigrate(&entity.UserGroup{})
		},
	}
}
