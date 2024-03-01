package service

import (
	"context"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	"github.com/cossim/coss-server/pkg/version"
	"github.com/cossim/coss-server/service/relation/api/v1"
	api "github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/cossim/coss-server/service/relation/domain/repository"
	"github.com/cossim/coss-server/service/relation/infrastructure/persistence"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"
)

type Service struct {
	ac   *pkgconfig.AppConfig
	db   *gorm.DB
	urr  repository.UserRelationRepository
	grr  repository.GroupRelationRepository
	ufqr repository.UserFriendRequestRepository
	gar  repository.GroupAnnouncementRepository
	gjqr repository.GroupJoinRequestRepository
	garr repository.GroupAnnouncementReadRepository
	dr   repository.DialogRepository
	v1.UnimplementedUserRelationServiceServer
	v1.UnimplementedGroupRelationServiceServer
	v1.UnimplementedDialogServiceServer
	v1.UnimplementedUserFriendRequestServiceServer
	v1.UnimplementedGroupJoinRequestServiceServer
	v1.UnimplementedGroupAnnouncementServiceServer
	v1.UnimplementedGroupAnnouncementReadServiceServer
}

func (s *Service) Init(cfg *pkgconfig.AppConfig) error {
	dbConn, err := db.NewMySQLFromDSN(cfg.MySQL.DSN).GetConnection()
	if err != nil {
		return err
	}

	infra := persistence.NewRepositories(dbConn)
	if err = infra.Automigrate(); err != nil {
		return err
	}

	s.ac = cfg
	s.db = dbConn
	s.urr = infra.Urr
	s.grr = infra.Grr
	s.ufqr = infra.Ufqr
	s.gar = infra.GAr
	s.gjqr = infra.Gjqr
	s.dr = infra.Dr
	s.garr = infra.Garr
	return nil
}

func (s *Service) Name() string {
	//TODO implement me
	return "relation_service"
}

func (s *Service) Version() string { return version.FullVersion() }

func (s *Service) Register(srv *grpc.Server) {
	api.RegisterUserRelationServiceServer(srv, s)
	api.RegisterGroupRelationServiceServer(srv, s)
	api.RegisterDialogServiceServer(srv, s)
	api.RegisterUserFriendRequestServiceServer(srv, s)
	api.RegisterGroupJoinRequestServiceServer(srv, s)
	api.RegisterGroupAnnouncementServiceServer(srv, s)
	api.RegisterGroupAnnouncementReadServiceServer(srv, s)
}

func (s *Service) RegisterHealth(srv *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
}

func (s *Service) Stop(ctx context.Context) error {
	//TODO implement me
	return nil
}

func (s *Service) DiscoverServices(services map[string]*grpc.ClientConn) error {
	//TODO implement me
	return nil
}
