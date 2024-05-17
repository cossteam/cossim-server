package persistence

import (
	"github.com/cossim/coss-server/internal/admin/domain/repository"
	"github.com/cossim/coss-server/internal/admin/infra/persistence/po"
	"gorm.io/gorm"
)

type Repositories struct {
	Ar repository.AdminRepository
	db *gorm.DB
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Ar: NewAdminRepo(db),
		db: db,
	}
}

func (s *Repositories) Automigrate() error {
	return s.db.AutoMigrate(&po.Admin{})
}
