package persistence

import (
	"github.com/cossim/coss-server/internal/msg/domain/entity"
	"github.com/cossim/coss-server/internal/msg/domain/repository"
	"gorm.io/gorm"
)

type Repositories struct {
	Umr  repository.UserMessageRepository
	Gmr  repository.GroupMessageRepository
	Gmrr repository.GroupMsgReadRepository
	db   *gorm.DB
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Umr:  NewUserMsgRepo(db),
		Gmr:  NewGroupMsgRepo(db),
		Gmrr: NewGroupMsgReadRepo(db),
		db:   db,
	}
}

func (s *Repositories) Automigrate() error {
	return s.db.AutoMigrate(&entity.GroupMessage{}, &entity.UserMessage{}, &entity.GroupMessageRead{})
}
