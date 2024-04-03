package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/go-logr/logr"
	"github.com/rs/xid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"sync"
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

	// 存储服务和服务地址的映射关系
	serviceMap  map[string]string
	serviceLock sync.Mutex
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
	backoffSettings := backoff.NewExponentialBackOff()
	backoffSettings.InitialInterval = 1 * time.Second
	backoffSettings.MaxElapsedTime = 0 // 无限期重试

	//clients := make(map[string]*grpc.ClientConn)
	//var mu sync.Mutex // 用于对 clients 的并发访问进行保护
	var wg sync.WaitGroup

	// 控制并发数的信号量
	sem := make(chan struct{}, 10) // 限制并发数为 10

	for serviceName, c := range s.ac.Discovers {
		if c.Direct {
			conn, err := grpc.Dial(c.Addr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return err
			}
			//mu.Lock()
			//clients[c.Name] = conn
			//mu.Unlock()
			// 在每次成功发现服务后调用 DiscoverServices
			client := make(map[string]*grpc.ClientConn)
			client[c.Name] = conn
			if err := s.svc.DiscoverServices(client); err != nil {
				s.logger.Error(err, "Failed to set up gRPC client for service", "service", c.Name)
			}
			continue
		}
		sem <- struct{}{} // 获取信号量，限制并发数
		wg.Add(1)
		go func(serviceName string, c config.ServiceConfig) {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量

			retryFunc := func() error {
				addr, err := s.registry.Discover(c.Name)
				if err != nil {
					s.logger.Error(err, "Service discover failed", "service", c.Name)
					return err
				}
				s.logger.Info("Service discover success", "service", c.Name, "addr", addr)
				conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					return err
				}
				//mu.Lock()
				//clients[c.Name] = conn
				//mu.Unlock()
				client := make(map[string]*grpc.ClientConn)
				client[c.Name] = conn
				// 在每次成功发现服务后调用 DiscoverServices
				if err := s.svc.DiscoverServices(client); err != nil {
					s.logger.Error(err, "Failed to set up gRPC client for service", "service", c.Name)
				}
				return nil
			}
			if err := backoff.Retry(retryFunc, backoffSettings); err != nil {
				s.logger.Error(err, "Failed to initialize gRPC client for service after retries")
				return
			}
		}(serviceName, c)
	}
	wg.Wait()
	return nil // 异步调用 DiscoverServices，无需等待所有服务都发现
}

func (s *GrpcService) discover(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	serviceMap := make(map[string]*svcT)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, c := range s.ac.Discovers {
				var addr string
				var err error
				if c.Direct {
					addr = c.Addr()
				} else {
					addr, err = s.registry.Discover(c.Name)
					if err != nil {
						s.logger.Error(err, "Failed to discover gRPC server", "service", c.Name)
						continue
					}
				}

				if svc, ok := serviceMap[c.Name]; ok {
					if svc.addr == addr {
						//fmt.Printf("原地址 => %s 发现后地址 => %s 服务发现无变化\n", svc.addr, addr)
						continue
					}
					svc.addr = addr
					svc.f = true
				} else {
					serviceMap[c.Name] = &svcT{
						addr: addr,
						f:    true,
					}
				}
			}

			allFalse := true
			for _, v := range serviceMap {
				if v.f {
					allFalse = false
					break
				}
			}

			if allFalse {
				continue
			}

			ss := make(map[string]*grpc.ClientConn)
			for k, v := range serviceMap {
				if !v.f {
					continue
				}
				v.f = false
				conn, err := grpc.Dial(v.addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					s.logger.Error(err, "Failed to connect to gRPC server", "service", k)
					continue
				}
				ss[k] = conn
			}
			if len(ss) > 0 {
				if err := s.svc.DiscoverServices(ss); err != nil {
					s.logger.Error(err, "Failed to set up gRPC clients")
				}
			}
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
		s.sid = xid.New().String()
		if err := s.register(); err != nil {
			return err
		}
		s.logger.Info("Service register success", "service", s.ac.GRPC.Name, "addr", s.ac.GRPC.Addr(), "id", s.sid)
		go s.watchRegistry(ctx)
	}

	if s.ac.Register.Discover {
		go s.discover(ctx)
		//go s.Discover()
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
				s.logger.Info("Service health check failed, re-registering", "service", s.ac.HTTP.Name, "addr", s.ac.HTTP.Addr(), "id", s.sid)
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
	if s.registry != nil && s.sid != "" {
		if err := s.registry.Cancel(s.sid); err != nil {
			s.logger.Error(err, "Service unregister failed", "service", s.ac.GRPC.Name, "addr", s.ac.GRPC.Addr(), "id", s.sid)
		}
		s.logger.Info("Service unregister success", "service", s.ac.GRPC.Name, "addr", s.ac.GRPC.Addr(), "id", s.sid)
	}
}
