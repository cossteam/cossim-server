package persistence

import (
	"github.com/cossim/coss-server/internal/relation/domain/relation"
	"gorm.io/gorm"
)

type Repositories struct {
	Urr  relation.UserRepository
	Grr  relation.GroupRepository
	Ufqr relation.UserFriendRequestRepository
	Gjqr relation.GroupJoinRequestRepository
	GAr  relation.GroupAnnouncementRepository
	Dr   relation.DialogRepository
	Dur  relation.DialogUserRepository
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
