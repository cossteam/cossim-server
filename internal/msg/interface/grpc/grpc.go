package grpc

import (
	"context"
	"github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	api "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	"github.com/cossim/coss-server/internal/msg/domain/repository"
	"github.com/cossim/coss-server/internal/msg/infrastructure/persistence"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	"github.com/cossim/coss-server/pkg/version"
	"github.com/rs/xid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"
	"strconv"
)

type Handler struct {
	db       *gorm.DB
	ac       *pkgconfig.AppConfig
	sid      string
	mr       repository.MsgRepository
	gmrr     repository.GroupMsgReadRepository
	stopFunc func()
	v1.UnimplementedMsgServiceServer
	v1.UnimplementedGroupMessageServiceServer
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
	s.stopFunc = func() {
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

func (s *Handler) Name() string {
	//TODO implement me
	return "msg_service"
}

func (s *Handler) Version() string { return version.FullVersion() }

func (s *Handler) Register(srv *grpc.Server) {
	api.RegisterMsgServiceServer(srv, s)
	api.RegisterGroupMessageServiceServer(srv, s)
}

func (s *Handler) RegisterHealth(srv *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
}

func (s *Handler) Stop(ctx context.Context) error {
	return nil
}

func (s *Handler) DiscoverServices(services map[string]*grpc.ClientConn) error {
	return nil
}
