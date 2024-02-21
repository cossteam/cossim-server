package persistence

import (
	"github.com/cossim/coss-server/service/user/domain/entity"
	"github.com/cossim/coss-server/service/user/domain/repository"
	"gorm.io/gorm"
)

type Repositories struct {
	UR  repository.UserRepository
	ULR repository.UserLoginRepository
	db  *gorm.DB
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		UR:  NewUserRepo(db),
		ULR: NewUserLoginRepo(db),
		db:  db,
	}
}

func (s *Repositories) Automigrate() error {
	return s.db.AutoMigrate(&entity.User{}, &entity.UserLogin{})
}
