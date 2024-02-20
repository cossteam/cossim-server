package server

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/cossim/coss-server/pkg/log"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"net"
	"sync"
	"time"
)

type ServiceInfo struct {
	ServiceName string
	Addr        string
}

type Option func(*GrpcService)

func WithLogger(logger *zap.Logger) Option {
	return func(s *GrpcService) {
		s.logger = logger
	}
}

func WithDiscovery(discovery discovery.Registry) Option {
	return func(s *GrpcService) {
		s.discovery = discovery
	}
}

func WithGrpcDiscoverFunc(disFunc GrpcDiscoverFunc) Option {
	return func(s *GrpcService) {
		s.disFunc = disFunc
	}
}

func NewGRPCService(c *config.AppConfig, svc GRPCService, opts ...Option) *GrpcService {
	d, err := discovery.NewConsulRegistry(c.Register.Addr())
	if err != nil {
		panic(err)
	}
	s := &GrpcService{
		server: grpc.NewServer(),
		logger: log.NewDevLogger("manager grpc service"),
		cfg:    c,

		sid:       xid.New().String(),
		discovery: d,

		grpcSvc: svc,
	}
	for _, opt := range opts {
		opt(s)
	}

	return s
}

type GrpcDiscoverFunc func(serviceName, addr string) error

var (
	_ Registry = &GrpcService{}
)

type GrpcService struct {
	server *grpc.Server
	logger *zap.Logger
	cfg    *config.AppConfig

	sid       string
	discovery discovery.Registry
	disFunc   GrpcDiscoverFunc

	grpcSvc GRPCService
}

func (s *GrpcService) RegisterGRPC(serviceName, addr string, serviceID string) error {
	//TODO implement me
	panic("implement me")
}

func (s *GrpcService) RegisterHTTP(serviceName, addr string, serviceID string) error {
	//TODO implement me
	panic("implement me")
}

func (s *GrpcService) Discover() error {
	//TODO implement me
	panic("implement me")
}

func (s *GrpcService) Start(ctx context.Context) error {
	if err := s.grpcSvc.Init(s.cfg); err != nil {
		return err
	}
	// 进行服务发现注册
	if err := s.discover(); err != nil {
		return err
	}

	lisAddr := fmt.Sprintf("%s", s.cfg.GRPC.Addr())
	var err error
	// 启动 gRPC 服务器
	lis, err := net.Listen("tcp", lisAddr)
	if err != nil {
		panic(err)
	}

	s.grpcSvc.Registry(s.server)
	// 注册服务开启健康检查
	grpc_health_v1.RegisterHealthServer(s.server, health.NewServer())
	s.logger.Info("GRPC service start", zap.String("service", s.cfg.Register.Name), zap.String("addr", lisAddr))

	return s.server.Serve(lis)
}

func (s *GrpcService) discover() error {
	if s.cfg.Register.Register {
		if err := s.register(); err != nil {
			return err
		}
		s.logger.Info("Service registration successful", zap.String("service", s.cfg.Register.Name), zap.String("addr", s.cfg.GRPC.Addr()), zap.String("id", s.sid))
	}

	var wg sync.WaitGroup
	//ch := make(chan ServiceInfo, len(s.ac.Discovers))

	ss := make([]ServiceInfo, 0, len(s.cfg.Discovers))

	for serviceName, c := range s.cfg.Discovers {
		wg.Add(1)
		go func(serviceName string, c config.ServiceConfig) {
			defer wg.Done()
			for {
				if c.Direct {
					addr := c.Addr()
					s.logger.Info("Service direct successful", zap.String("service", s.cfg.Register.Name), zap.String("addr", addr))
					ss = append(ss, ServiceInfo{ServiceName: serviceName, Addr: addr})
					//ch <- ServiceInfo{ServiceName: serviceName, Addr: addr}
					break
				}
				addr, err := s.discovery.Discover(c.Name)
				if err != nil {
					s.logger.Info("Service discovery failed", zap.String("service", c.Name), zap.Error(err))
					time.Sleep(15 * time.Second)
					continue
				}
				s.logger.Info("Service discovery successful", zap.String("service", s.cfg.Register.Name), zap.String("addr", addr))
				ss = append(ss, ServiceInfo{ServiceName: serviceName, Addr: addr})
				//ch <- ServiceInfo{ServiceName: serviceName, Addr: addr}
				break
			}
		}(serviceName, c)
	}

	//go func() {
	wg.Wait()
	//close(ch)
	//}()

	//for info := range ch {
	//	if err := s.disFunc(info.ServiceName, info.Addr); err != nil {
	//		s.logger.Warn("Failed to initialize gRPC client for", zap.String("service", info.ServiceName), zap.String("msg", err.Error()))
	//	}
	//}

	return nil
}

func (s *GrpcService) register() error {
	return s.discovery.RegisterGRPC(s.cfg.Register.Name, s.cfg.GRPC.Addr(), s.sid)
}
