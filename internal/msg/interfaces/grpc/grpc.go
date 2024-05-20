package grpc

import (
	"context"
	api "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	"github.com/cossim/coss-server/internal/msg/domain/service"
	"github.com/cossim/coss-server/internal/msg/infra/persistence"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	"github.com/cossim/coss-server/pkg/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"gorm.io/gorm"
	"strconv"
)

var _ api.MsgServiceServer = &Handler{}

type Handler struct {
	db   *gorm.DB
	ac   *pkgconfig.AppConfig
	gmd  service.GroupMsgDomain
	umd  service.UserMsgDomain
	gmrd service.GroupMsgReadDomain
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

	repo := persistence.NewRepositories(dbConn)

	s.umd = service.NewUserMsgDomain(dbConn, cfg, repo)
	s.gmd = service.NewGroupMsgDomain(dbConn, cfg, repo)
	s.gmrd = service.NewGroupMsgReadDomain(dbConn, cfg, repo)

	s.db = dbConn
	s.ac = cfg
	//s.cache = msgCache
	//s.cacheEnable = cfg.Cache.Enable
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
