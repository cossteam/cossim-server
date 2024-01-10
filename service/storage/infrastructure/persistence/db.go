package persistence

import (
	"github.com/cossim/coss-server/service/storage/domain/entity"
	"github.com/cossim/coss-server/service/storage/domain/repository"
	"gorm.io/gorm"
)

type Repositories struct {
	FR repository.FileRepository
	db *gorm.DB
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		FR: NewFileRepo(db),
		db: db,
	}
}

func (s *Repositories) Automigrate() error {
	return s.db.AutoMigrate(&entity.File{})
}
