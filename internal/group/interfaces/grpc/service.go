package grpc

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/group/adapters"
	api "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	"github.com/cossim/coss-server/internal/group/cache"
	"github.com/cossim/coss-server/internal/group/domain/repository"
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
	repo repository.Repository
	stop func() func(ctx context.Context) error
}

func (s *GroupServiceServer) Init(cfg *pkgconfig.AppConfig) error {
	fmt.Println("sb")
	mysql, err := db.NewMySQL(cfg.MySQL.Address, strconv.Itoa(cfg.MySQL.Port), cfg.MySQL.Username, cfg.MySQL.Password, cfg.MySQL.Database, int64(cfg.Log.Level), cfg.MySQL.Opts)
	if err != nil {
		log.Printf("init mysql error: %v\n", err)
		return err
	}
	dbConn, err := mysql.GetConnection()
	if err != nil {
		log.Printf("get mysql connection error: %v\n", err)
		return err
	}

	gcache, err := cache.NewGroupCacheRedis(cfg.Redis.Addr(), cfg.Redis.Password, 0)
	if err != nil {
		log.Printf("init group cache error: %v\n", err)
		return err
	}

	s.stop = func() func(ctx context.Context) error {
		return func(ctx context.Context) error {
			if err := gcache.DeleteAllCache(ctx); err != nil {
				log.Printf("delete all group cache error: %v\n", err)
			}
			return gcache.Close()
		}
	}

	repo := adapters.NewMySQLGroupRepository(dbConn, gcache)
	if err = repo.Automigrate(); err != nil {
		log.Printf("automigrate error: %v\n", err)
		return err
	}
	s.repo = repo
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
	return s.stop()(ctx)
}

func (s *GroupServiceServer) DiscoverServices(services map[string]*grpc.ClientConn) error {
	return nil
}
