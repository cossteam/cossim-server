package service

import (
	"github.com/cossim/coss-server/service/msg/api/v1"
	"github.com/cossim/coss-server/service/msg/domain/repository"
	"github.com/cossim/coss-server/service/msg/infrastructure/persistence"
	"gorm.io/gorm"
)

func NewService(repo *persistence.Repositories, db *gorm.DB) *Service {
	return &Service{
		mr: repo.Mr,
		db: db,
	}
}

type Service struct {
	mr repository.MsgRepository
	db *gorm.DB
	v1.UnimplementedMsgServiceServer
}
