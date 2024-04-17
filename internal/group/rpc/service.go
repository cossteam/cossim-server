package rpc

import (
	"context"
	"github.com/cossim/coss-server/internal/group/adapters"
	api "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	"github.com/cossim/coss-server/internal/group/domain/group"
	"github.com/cossim/coss-server/pkg/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	"github.com/cossim/coss-server/pkg/manager/server"
	"github.com/cossim/coss-server/pkg/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"log"
	"strconv"
)

var _ server.GRPCService = &GroupServiceServer{}

const (
	ServiceName = "group_service"
)

type GroupServiceServer struct {
	repo        group.Repository
	cache       cache.GroupCache
	cacheEnable bool
}

func (s *GroupServiceServer) Init(cfg *pkgconfig.AppConfig) error {
	mysql, err := db.NewMySQL(cfg.MySQL.Address, strconv.Itoa(cfg.MySQL.Port), cfg.MySQL.Username, cfg.MySQL.Password, cfg.MySQL.Database, int64(cfg.Log.Level), cfg.MySQL.Opts)
	if err != nil {
		return err
	}
	dbConn, err := mysql.GetConnection()
	if err != nil {
		return err
	}

	repo := adapters.NewMySQLGroupRepository(dbConn)
	if err = repo.Automigrate(); err != nil {
		return err
	}
	s.repo = repo

	groupCache, err := cache.NewGroupCacheRedis(cfg.Redis.Addr(), cfg.Redis.Password, 0)
	if err != nil {
		return err
	}
	s.cache = groupCache
	s.cacheEnable = cfg.Cache.Enable
	return nil
}

func (s *GroupServiceServer) Name() string {
	return ServiceName
}

func (s *GroupServiceServer) Version() string {
	return version.FullVersion()
}

func (s *GroupServiceServer) Register(srv *grpc.Server) {
	api.RegisterGroupServiceServer(srv, s)
}

func (s *GroupServiceServer) RegisterHealth(srv *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
}

func (s *GroupServiceServer) Stop(ctx context.Context) error {
	if s.cacheEnable && s.cache != nil {
		if err := s.cache.DeleteAllCache(ctx); err != nil {
			log.Printf("delete all group cache error: %v", err)
		}
		return s.cache.Close()
	}
	return nil
}

func (s *GroupServiceServer) DiscoverServices(services map[string]*grpc.ClientConn) error {
	return nil
}
