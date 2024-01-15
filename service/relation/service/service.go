package service

import (
	"github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/cossim/coss-server/service/relation/domain/repository"
	"github.com/cossim/coss-server/service/relation/infrastructure/persistence"
)

func NewService(repo *persistence.Repositories) *Service {
	return &Service{
		urr: repo.Urr,
		grr: repo.Grr,
		dr:  repo.Dr,
	}
}

type Service struct {
	urr repository.UserRelationRepository
	grr repository.GroupRelationRepository
	dr  repository.DialogRepository
	v1.UnimplementedUserRelationServiceServer
	v1.UnimplementedGroupRelationServiceServer
	v1.UnimplementedDialogServiceServer
}
