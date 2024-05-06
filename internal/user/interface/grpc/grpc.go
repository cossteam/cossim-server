package grpc

import (
	"context"
	api "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/repository"
	"github.com/cossim/coss-server/internal/user/infra/persistence"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	"github.com/cossim/coss-server/pkg/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"log"
	"strconv"
)

type UserServiceServer struct {
	ac  *pkgconfig.AppConfig
	ur  repository.UserRepository
	ulr repository.UserLoginRepository
	//AuthSrv   *authServiceServer
	secret    string
	userCache cache.UserCache
	stop      func() func(ctx context.Context) error
}

func (s *UserServiceServer) Init(cfg *pkgconfig.AppConfig) error {
	mysql, err := db.NewMySQL(cfg.MySQL.Address, strconv.Itoa(cfg.MySQL.Port), cfg.MySQL.Username, cfg.MySQL.Password, cfg.MySQL.Database, int64(cfg.Log.Level), cfg.MySQL.Opts)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}

	dbConn, err := mysql.GetConnection()
	if err != nil {
		return err
	}

	userCache, err := cache.NewUserCacheRedis(cfg.Redis.Addr(), cfg.Redis.Password, 0)
	if err != nil {
		return err
	}

	infra := persistence.NewRepositories(dbConn, userCache)
	if err = infra.Automigrate(); err != nil {
		return err
	}

	s.stop = func() func(ctx context.Context) error {
		return func(ctx context.Context) error {
			if err := userCache.DeleteAllCache(ctx); err != nil {
				log.Printf("failed to delete all cache: %v", err)
			}
			return userCache.Close()
		}
	}

	s.ur = infra.UR
	s.ulr = infra.ULR
	s.ac = cfg
	s.secret = cfg.SystemConfig.JwtSecret
	s.userCache = userCache
	//s.AuthSrv = &authServiceServer{
	//	secret:    "cfg.SystemConfig.JwtSecret",
	//	ur:        infra.UR,
	//	userCache: userCache,
	//}

	//fmt.Println("s.AuthSrv => ", s.AuthSrv)

	return nil
}

func (s *UserServiceServer) Name() string {
	//TODO implement me
	return "user_service"
}

func (s *UserServiceServer) Version() string { return version.FullVersion() }

func (s *UserServiceServer) Register(srv *grpc.Server) {
	api.RegisterUserServiceServer(srv, s)
	api.RegisterUserLoginServiceServer(srv, s)
	api.RegisterUserAuthServiceServer(srv, s)
}

func (s *UserServiceServer) RegisterHealth(srv *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
}

func (s *UserServiceServer) Stop(ctx context.Context) error {
	return s.stop()(ctx)
}

func (s *UserServiceServer) DiscoverServices(services map[string]*grpc.ClientConn) error { return nil }
