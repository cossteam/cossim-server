package service

import (
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/cossim/coss-server/service/msg/api/v1"
	"github.com/cossim/coss-server/service/msg/domain/repository"
	"github.com/cossim/coss-server/service/msg/infrastructure/persistence"
	"github.com/rs/xid"
	"gorm.io/gorm"
	"log"
)

func NewService(ac *pkgconfig.AppConfig, repo *persistence.Repositories, db *gorm.DB) *Service {
	return &Service{
		mr:   repo.Mr,
		db:   db,
		gmrr: repo.Gmrr,
		ac:   ac,
		sid:  xid.New().String(),
	}
}

type Service struct {
	mr repository.MsgRepository
	db *gorm.DB
	v1.UnimplementedMsgServiceServer
	v1.UnimplementedGroupMessageServiceServer
	gmrr      repository.GroupMsgReadRepository
	ac        *pkgconfig.AppConfig
	discovery discovery.Registry
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
	if err = s.discovery.RegisterGRPC(s.ac.Register.Name, s.ac.GRPC.Addr(), s.sid); err != nil {
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
