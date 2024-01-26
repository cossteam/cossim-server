package persistence

import (
	"github.com/cossim/coss-server/service/relation/domain/entity"
	"github.com/cossim/coss-server/service/relation/domain/repository"
	"gorm.io/gorm"
)

type Repositories struct {
	Urr  repository.UserRelationRepository
	Grr  repository.GroupRelationRepository
	Ufqr repository.UserFriendRequestRepository
	Dr   repository.DialogRepository
	db   *gorm.DB
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Urr:  NewUserRelationRepo(db),
		Grr:  NewGroupRelationRepo(db),
		Dr:   NewDialogRepo(db),
		Ufqr: NewUserFriendRequestRepo(db),
		db:   db,
	}
}

func (s *Repositories) Automigrate() error {
	return s.db.AutoMigrate(&entity.GroupRelation{}, &entity.UserRelation{}, &entity.Dialog{}, &entity.DialogUser{}, &entity.UserFriendRequest{}, &entity.GroupJoinRequest{})
}
