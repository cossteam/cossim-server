package persistence

import (
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"gorm.io/gorm"
)

type Repositories struct {
	Urr  repository.UserRelationRepository
	Grr  repository.GroupRelationRepository
	Ufqr repository.UserFriendRequestRepository
	Gjqr repository.GroupJoinRequestRepository
	GAr  repository.GroupAnnouncementRepository
	Dr   repository.DialogRepository
	db   *gorm.DB
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Urr:  NewUserRelationRepo(db),
		Grr:  NewGroupRelationRepo(db),
		Dr:   NewDialogRepo(db),
		Gjqr: NewGroupJoinRequestRepo(db),
		GAr:  NewGroupAnnouncementRepository(db),
		Ufqr: NewUserFriendRequestRepo(db),
		db:   db,
	}
}

func (s *Repositories) Automigrate() error {
	return s.db.AutoMigrate(&entity.GroupRelation{}, &entity.UserRelation{}, &entity.Dialog{}, &entity.DialogUser{}, &entity.UserFriendRequest{}, &entity.GroupJoinRequest{}, &entity.GroupAnnouncement{}, &entity.GroupAnnouncementRead{})
}
