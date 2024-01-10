package persistence

import (
	"github.com/cossim/coss-server/service/relation/domain/entity"
	"github.com/cossim/coss-server/service/relation/domain/repository"
	"gorm.io/gorm"
)

type Repositories struct {
	Urr repository.UserRelationRepository
	Grr repository.GroupRelationRepository
	db  *gorm.DB
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Urr: NewUserRelationRepo(db),
		Grr: NewGroupRelationRepo(db),
		db:  db,
	}
}

func (s *Repositories) Automigrate() error {
	return s.db.AutoMigrate(&entity.UserGroup{}, &entity.UserRelation{})
}
