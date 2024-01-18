package service

import (
	"github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/cossim/coss-server/service/relation/domain/repository"
	"github.com/cossim/coss-server/service/relation/infrastructure/persistence"
	"gorm.io/gorm"
)

func NewService(repo *persistence.Repositories, db *gorm.DB) *Service {
	return &Service{
		urr: repo.Urr,
		grr: repo.Grr,
		dr:  repo.Dr,
		db:  db,
	}
}

type Service struct {
	db  *gorm.DB
	urr repository.UserRelationRepository
	grr repository.GroupRelationRepository
	dr  repository.DialogRepository
	v1.UnimplementedUserRelationServiceServer
	v1.UnimplementedGroupRelationServiceServer
	v1.UnimplementedDialogServiceServer
}
