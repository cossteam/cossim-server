package service

import (
	"context"
	"fmt"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	"github.com/cossim/coss-server/pkg/version"
	"github.com/cossim/coss-server/service/msg/api/v1"
	api "github.com/cossim/coss-server/service/msg/api/v1"
	"github.com/cossim/coss-server/service/msg/domain/repository"
	"github.com/cossim/coss-server/service/msg/infrastructure/persistence"
	"github.com/rs/xid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"
)

type Service struct {
	db   *gorm.DB
	ac   *pkgconfig.AppConfig
	sid  string
	mr   repository.MsgRepository
	gmrr repository.GroupMsgReadRepository
	v1.UnimplementedMsgServiceServer
	v1.UnimplementedGroupMessageServiceServer
}

func (s *Service) Init(cfg *pkgconfig.AppConfig) error {
	fmt.Println("cfg.MySQL.DSN => ", cfg.MySQL.DSN)
	dbConn, err := db.NewMySQLFromDSN(cfg.MySQL.DSN).GetConnection()
	if err != nil {
		return err
	}

	infra := persistence.NewRepositories(dbConn)
	if err = infra.Automigrate(); err != nil {
		return err
	}

	s.mr = infra.Mr
	s.gmrr = infra.Gmrr
	s.db = dbConn
	s.ac = cfg
	s.sid = xid.New().String()
	return nil
}

func (s *Service) Name() string {
	//TODO implement me
	return "msg_service"
}

func (s *Service) Version() string { return version.FullVersion() }

func (s *Service) Register(srv *grpc.Server) {
	api.RegisterMsgServiceServer(srv, s)
	api.RegisterGroupMessageServiceServer(srv, s)
}

func (s *Service) RegisterHealth(srv *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
}

func (s *Service) Stop(ctx context.Context) error {
	return nil
}

func (s *Service) DiscoverServices(services map[string]*grpc.ClientConn) error {
	return nil
}
