package service

import (
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/cossim/coss-server/service/relation/domain/repository"
	"github.com/cossim/coss-server/service/relation/infrastructure/persistence"
	"github.com/rs/xid"
	"gorm.io/gorm"
	"log"
)

func NewService(repo *persistence.Repositories, db *gorm.DB, ac *pkgconfig.AppConfig) *Service {
	return &Service{
		urr:  repo.Urr,
		grr:  repo.Grr,
		dr:   repo.Dr,
		db:   db,
		ufqr: repo.Ufqr,
		gjqr: repo.Gjqr,
		ac:   ac,
		sid:  xid.New().String(),
	}
}

type Service struct {
	db   *gorm.DB
	urr  repository.UserRelationRepository
	grr  repository.GroupRelationRepository
	ufqr repository.UserFriendRequestRepository
	gjqr repository.GroupJoinRequestRepository
	dr   repository.DialogRepository
	v1.UnimplementedUserRelationServiceServer
	v1.UnimplementedGroupRelationServiceServer
	v1.UnimplementedDialogServiceServer
	v1.UnimplementedUserFriendRequestServiceServer
	v1.UnimplementedGroupJoinRequestServiceServer
	ac        *pkgconfig.AppConfig
	discovery discovery.Discovery
	sid       string
}

func (s *Service) Start(discover bool) {
	if !discover {
		return
	}
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

func (s *Service) Stop(discover bool) error {
	if !discover {
		return nil
	}
	if err := s.discovery.Cancel(s.sid); err != nil {
		log.Printf("Failed to cancel service registration: %v", err)
		return err
	}
	log.Printf("Service registration canceled ServiceName: %s  Addr: %s  ID: %s", s.ac.Register.Name, s.ac.GRPC.Addr(), s.sid)
	return nil
}

func (s *Service) Restart(discover bool) {
	s.Stop(discover)
	s.Start(discover)
}
