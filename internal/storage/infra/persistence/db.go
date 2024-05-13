package persistence

import (
	"github.com/cossim/coss-server/internal/storage/domain/repository"
	"github.com/cossim/coss-server/internal/storage/infra/persistence/po"
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
	return s.db.AutoMigrate(&po.File{})
}
