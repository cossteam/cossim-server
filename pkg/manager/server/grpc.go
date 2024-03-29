package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/go-logr/logr"
	"github.com/rs/xid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"time"
)

type ServiceInfo struct {
	ServiceName string
	Addr        string
}

type Option func(*GrpcService)

func WithLogger(logger logr.Logger) Option {
	return func(s *GrpcService) {
		s.logger = logger
	}
}

func WithDiscovery(discovery discovery.Registry) Option {
	return func(s *GrpcService) {
		s.registry = discovery
	}
}

func WithGrpcDiscoverFunc(disFunc GrpcDiscoverFunc) Option {
	return func(s *GrpcService) {
		s.disFunc = disFunc
	}
}

func NewGRPCService(c *config.AppConfig, svc GRPCService, logger logr.Logger, opts ...Option) *GrpcService {
	d, err := discovery.NewConsulRegistry(c.Register.Addr())
	if err != nil {
		panic(err)
	}
	s := &GrpcService{
		server:   grpc.NewServer(),
		logger:   logger.WithValues("kind", "grpc server", "name", c.GRPC.Name),
		ac:       c,
		sid:      xid.New().String(),
		registry: d,
		svc:      svc,
	}
	for _, opt := range opts {
		opt(s)
	}

	return s
}

type GrpcDiscoverFunc func(serviceName, addr string) error

type GrpcService struct {
	server   *grpc.Server
	logger   logr.Logger
	ac       *config.AppConfig
	sid      string
	registry discovery.Registry
	disFunc  GrpcDiscoverFunc
	svc      GRPCService
}

func (s *GrpcService) RegisterGRPC(serviceName, addr string, serviceID string) error {
	//TODO implement me
	panic("implement me")
}

func (s *GrpcService) RegisterHTTP(serviceName, addr string, serviceID string) error {
	//TODO implement me
	panic("implement me")
}

func (s *GrpcService) discover() {
	// 定时器，每隔5秒执行一次服务发现
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	clients := make(map[string]*grpc.ClientConn)

	for {
		select {
		case <-ticker.C:
			for _, c := range s.ac.Discovers {
				var conn *grpc.ClientConn
				var err error
				if c.Direct {
					conn, err = grpc.Dial(c.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
				} else {
					addr, err := s.registry.Discover(c.Name)
					if err == nil {
						conn, err = grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
					}
				}
				if err != nil {
					s.logger.Error(err, "Failed to connect to gRPC server", "service", c.Name)
					continue
				}
				clients[c.Name] = conn
			}
			if err := s.svc.DiscoverServices(clients); err != nil {
				s.logger.Error(err, "Failed to set up gRPC clients")
			}
			clients = make(map[string]*grpc.ClientConn)
		}
	}
}

func (s *GrpcService) Start(ctx context.Context) error {
	if err := s.svc.Init(s.ac); err != nil {
		return err
	}

	// 注册grpc服务
	s.svc.Register(s.server)

	if s.ac.Register.Register || s.ac.Register.Discover {
		d, err := discovery.NewConsulRegistry(s.ac.Register.Addr())
		if err != nil {
			return err
		}
		s.registry = d
	}

	if s.ac.Register.Register {
		// 注册服务到注册中心
		if err := s.register(); err != nil {
			return err
		}
		s.logger.Info("Service register success", "service", s.ac.GRPC.Name, "addr", s.ac.GRPC.Addr(), "id", s.sid)
		go s.watchRegistry(ctx)
	}

	if s.ac.Register.Discover {
		go s.discover()
	}

	lisAddr := fmt.Sprintf("%s", s.ac.GRPC.Addr())
	lis, err := net.Listen("tcp", lisAddr)
	if err != nil {
		return err
	}

	serverShutdown := make(chan struct{})
	go func() {
		<-ctx.Done()
		s.logger.Info("Shutting down grpcServer", "addr", lisAddr)
		s.cancel()
		s.server.GracefulStop()
		close(serverShutdown)
	}()

	s.logger.Info("Starting  grpcServer", "addr", lisAddr)
	if err := s.server.Serve(lis); err != nil {
		if !errors.Is(err, grpc.ErrServerStopped) {
			s.logger.Error(err, "failed to start grpcServer")
		}
	}

	<-serverShutdown
	return nil
}

// watchRegistration 监听注册状态
func (s *GrpcService) watchRegistry(ctx context.Context) {
	s.registry.KeepAlive(s.ac.GRPC.Name, s.sid, &discovery.RegisterOption{
		HealthCheckCallbackFn: func(b bool) {
			if !b {
				s.register()
			}
		},
	})
}

func (s *GrpcService) register() error {
	// 注册到注册中心要实现健康检查
	s.svc.RegisterHealth(s.server)
	return s.registry.RegisterGRPC(s.ac.GRPC.Name, s.ac.GRPC.Addr(), s.sid)
}

func (s *GrpcService) cancel() {
	if s.registry != nil {
		if err := s.registry.Cancel(s.sid); err != nil {
			s.logger.Error(err, "Service unregister failed", "service", s.ac.GRPC.Name, "addr", s.ac.GRPC.Addr(), "id", s.sid)
		}
		s.logger.Info("Service unregister success", "service", s.ac.GRPC.Name, "addr", s.ac.GRPC.Addr(), "id", s.sid)
	}
}
