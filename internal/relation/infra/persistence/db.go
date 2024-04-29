package persistence

import (
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"gorm.io/gorm"
)

type Repositories struct {
	Urr  repository.UserRepository
	Grr  repository.GroupRepository
	Ufqr repository.UserFriendRequestRepository
	Gjqr repository.GroupJoinRequestRepository
	GAr  entity.GroupAnnouncementRepository
	Dr   repository.DialogRepository
	Dur  repository.DialogUserRepository
	db   *gorm.DB
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Urr:  NewMySQLRelationUserRepository(db, nil),
		Grr:  NewMySQLRelationGroupRepository(db, nil),
		Dr:   NewMySQLMySQLDialogRepository(db, nil),
		Dur:  NewMySQLDialogUserRepository(db, nil),
		Gjqr: NewMySQLGroupJoinRequestRepository(db, nil),
		GAr:  NewMySQLRelationGroupAnnouncementRepository(db, nil),
		Ufqr: NewMySQLUserFriendRequestRepository(db, nil),
		db:   db,
	}
}

func (s *Repositories) Automigrate() error {
	return s.db.AutoMigrate(
		&GroupRelationModel{},
		&UserRelationModel{},
		&DialogModel{},
		&DialogUserModel{},
		&UserFriendRequestModel{},
		&GroupJoinRequestModel{},
		&GroupAnnouncementModel{},
		&GroupAnnouncementReadModel{},
	)
}
