package grpc

import (
	"context"
	api "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/infra/persistence"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/version"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type RelationServiceServer struct {
	ac                  *pkgconfig.AppConfig
	UserServiceServer   *userServiceServer
	GroupServiceServer  *groupServiceServer
	DialogServiceServer *dialogServiceServer

	stop func() func(ctx context.Context) error
}

func (s *RelationServiceServer) Init(cfg *pkgconfig.AppConfig) error {
	infra := persistence.NewRepositories(cfg)
	if err := infra.Automigrate(); err != nil {
		return err
	}

	s.stop = func() func(ctx context.Context) error {
		return func(ctx context.Context) error {
			return infra.Close()
		}
	}

	s.ac = cfg
	s.UserServiceServer = &userServiceServer{
		repos: infra,
	}
	s.GroupServiceServer = &groupServiceServer{
		repos: infra,
	}
	s.DialogServiceServer = &dialogServiceServer{
		repos: infra,
	}
	return nil
}

func (s *RelationServiceServer) Name() string {
	//TODO implement me
	return "relation_service"
}

func (s *RelationServiceServer) Version() string { return version.FullVersion() }

func (s *RelationServiceServer) Register(srv *grpc.Server) {
	api.RegisterUserRelationServiceServer(srv, s.UserServiceServer)
	api.RegisterGroupRelationServiceServer(srv, s.GroupServiceServer)
	api.RegisterDialogServiceServer(srv, s.DialogServiceServer)
}

func (s *RelationServiceServer) RegisterHealth(srv *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
}

func (s *RelationServiceServer) Stop(ctx context.Context) error {
	//return s.UserServiceServer.cache.DeleteAllCache(ctx)
	return s.stop()(ctx)
}

func (s *RelationServiceServer) DiscoverServices(services map[string]*grpc.ClientConn) error {
	//TODO implement me
	return nil
}
