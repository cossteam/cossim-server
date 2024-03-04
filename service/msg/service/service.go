package service

import (
	"context"
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
	"strconv"
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
	mysql, err := db.NewMySQL(cfg.MySQL.Address, strconv.Itoa(cfg.MySQL.Port), cfg.MySQL.Username, cfg.MySQL.Password, cfg.MySQL.Database, int64(cfg.Log.Level), map[string]string{"allowNativePasswords": "true", "timeout": "800ms", "readTimeout": "200ms", "writeTimeout": "800ms", "parseTime": "true", "loc": "Local", "charset": "utf8mb4"})
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
