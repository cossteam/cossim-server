package persistence

import (
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/repository"
	"github.com/cossim/coss-server/internal/user/infra/persistence/po"
	"gorm.io/gorm"
)

type Repositories struct {
	UR  repository.UserRepository
	ULR repository.UserLoginRepository
	db  *gorm.DB
}

func NewRepositories(db *gorm.DB, cache cache.UserCache) *Repositories {
	return &Repositories{
		UR:  NewMySQLUserRepository(db, cache),
		ULR: NewMySQLUserLoginRepository(db, cache),
		db:  db,
	}
}

func (s *Repositories) Automigrate() error {
	return s.db.AutoMigrate(&po.User{}, &po.UserLogin{})
}
