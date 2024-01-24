package service

import (
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/cossim/coss-server/service/relation/domain/repository"
	"github.com/cossim/coss-server/service/relation/infrastructure/persistence"
	"github.com/rs/xid"
	"gorm.io/gorm"
	"log"
)

func NewService(repo *persistence.Repositories, db *gorm.DB, ac config.AppConfig) *Service {
	return &Service{
		urr: repo.Urr,
		grr: repo.Grr,
		dr:  repo.Dr,
		db:  db,
		ac:  ac,
		sid: xid.New().String(),
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

	ac        config.AppConfig
	discovery discovery.Discovery
	sid       string
}

func (s *Service) Start() {
	d, err := discovery.NewConsulRegistry(s.ac.Register.Addr())
	if err != nil {
		panic(err)
	}
	s.discovery = d
	if err = s.discovery.Register(s.ac.Register.Name, s.ac.GRPC.Addr(), s.sid); err != nil {
		panic(err)
	}
	log.Printf("Service registration successful ServiceName: %s  Addr: %s  ID: %s", s.ac.Register.Name, s.ac.GRPC.Addr(), s.sid)
}

func (s *Service) Close() error {
	if err := s.discovery.Cancel(s.sid); err != nil {
		log.Printf("Failed to cancel service registration: %v", err)
		return err
	}
	log.Printf("Service registration canceled ServiceName: %s  Addr: %s  ID: %s", s.ac.Register.Name, s.ac.GRPC.Addr(), s.sid)
	return nil
}
