package service

import (
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	"github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/cossim/coss-server/service/relation/domain/repository"
	"github.com/cossim/coss-server/service/relation/infrastructure/persistence"
)

func NewService(c *config.AppConfig) *Service {
	dbConn, err := db.NewMySQLFromDSN(c.MySQL.DSN).GetConnection()
	if err != nil {
		panic(err)
	}

	return &Service{
		urr: persistence.NewUserRelationRepo(dbConn),
		grr: persistence.NewGroupRelationRepo(dbConn),
	}
}

type Service struct {
	urr repository.UserRelationRepository
	grr repository.GroupRelationRepository
	v1.UnimplementedUserRelationServiceServer
	v1.UnimplementedGroupRelationServiceServer
}
