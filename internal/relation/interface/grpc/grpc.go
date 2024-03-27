package grpc

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	api "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/domain/repository"
	"github.com/cossim/coss-server/internal/relation/infrastructure/persistence"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	"github.com/cossim/coss-server/pkg/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"
	"strconv"
)

type Handler struct {
	ac   *pkgconfig.AppConfig
	db   *gorm.DB
	urr  repository.UserRelationRepository
	grr  repository.GroupRelationRepository
	ufqr repository.UserFriendRequestRepository
	gar  repository.GroupAnnouncementRepository
	gjqr repository.GroupJoinRequestRepository
	dr   repository.DialogRepository
	v1.UnimplementedUserRelationServiceServer
	v1.UnimplementedGroupRelationServiceServer
	v1.UnimplementedDialogServiceServer
	v1.UnimplementedUserFriendRequestServiceServer
	v1.UnimplementedGroupJoinRequestServiceServer
	v1.UnimplementedGroupAnnouncementServiceServer
}

func (s *Handler) Init(cfg *pkgconfig.AppConfig) error {
	mysql, err := db.NewMySQL(cfg.MySQL.Address, strconv.Itoa(cfg.MySQL.Port), cfg.MySQL.Username, cfg.MySQL.Password, cfg.MySQL.Database, int64(cfg.Log.Level), cfg.MySQL.Opts)
	if err != nil {
		return err
	}

	dbConn, err := mysql.GetConnection()
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
	return nil
}

func (s *Handler) Name() string {
	//TODO implement me
	return "relation_service"
}

func (s *Handler) Version() string { return version.FullVersion() }

func (s *Handler) Register(srv *grpc.Server) {
	api.RegisterUserRelationServiceServer(srv, s)
	api.RegisterGroupRelationServiceServer(srv, s)
	api.RegisterDialogServiceServer(srv, s)
	api.RegisterUserFriendRequestServiceServer(srv, s)
	api.RegisterGroupJoinRequestServiceServer(srv, s)
	api.RegisterGroupAnnouncementServiceServer(srv, s)
}

func (s *Handler) RegisterHealth(srv *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
}

func (s *Handler) Stop(ctx context.Context) error {
	//TODO implement me
	return nil
}

func (s *Handler) DiscoverServices(services map[string]*grpc.ClientConn) error {
	//TODO implement me
	return nil
}
