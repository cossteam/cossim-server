package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/cossim/coss-server/pkg/http/middleware"
	"github.com/cossim/coss-server/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/rs/xid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"sync"
	"time"
)

var (
	_ Registry = &HttpService{}
)

// New returns a new server with sane defaults.
func New(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:           handler,
		MaxHeaderBytes:    1 << 20,
		IdleTimeout:       90 * time.Second, // matches http.DefaultTransport keep-alive timeout
		ReadHeaderTimeout: 32 * time.Second,
	}
}

type HttpService struct {
	server     *http.Server
	handler    gin.IRouter
	svc        HTTPService
	registry   discovery.Registry
	ac         *config.AppConfig
	logger     logr.Logger
	sid        string
	addr       string
	healthAddr string
}

func NewHttpService(c *config.AppConfig, svc HTTPService, healthAddr string, logger logr.Logger) *HttpService {
	s := &HttpService{
		logger:     logger.WithValues("kind", "http server", "name", c.HTTP.Name),
		ac:         c,
		svc:        svc,
		addr:       c.HTTP.Addr(),
		sid:        xid.New().String(),
		healthAddr: c.HTTP.Address + healthAddr,
	}

	handler := gin.New()
	handler.Use(middleware.GinLogger(log.NewLogger(c.Log.Format, int8(c.Log.Level), true)))
	s.handler = handler
	s.server = &http.Server{
		Handler:           handler,
		Addr:              s.ac.HTTP.Addr(),
		MaxHeaderBytes:    1 << 20,
		IdleTimeout:       90 * time.Second, // matches http.DefaultTransport keep-alive timeout
		ReadHeaderTimeout: 32 * time.Second,
	}
	return s
}

func (s *HttpService) Start(ctx context.Context) error {
	if err := s.svc.Init(s.ac); err != nil {
		return err
	}

	s.svc.RegisterRoute(s.handler)

	if s.ac.Register.Register || s.ac.Register.Discover {
		d, err := discovery.NewConsulRegistry(s.ac.Register.Addr())
		if err != nil {
			return err
		}
		s.registry = d
	}

	if s.ac.Register.Register {
		if err := s.RegisterHTTP(s.ac.HTTP.Name, s.ac.HTTP.Addr(), s.sid); err != nil {
			return err
		}
	}

	if s.ac.Register.Discover {
		go func() {
			if err := s.Discover(); err != nil {
				s.logger.Error(err, "发现服务失败")
			}
		}()
	}

	serverShutdown := make(chan struct{})
	go func() {
		<-ctx.Done()
		s.logger.Info("shutting down httpServer", "addr", s.addr)
		s.cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			s.logger.Error(err, "error shutting down httpServer")
		}
		close(serverShutdown)
	}()
	s.logger.Info("starting httpServer", "addr", s.ac.HTTP.Addr())
	if err := s.server.ListenAndServe(); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error(err, fmt.Sprintf("启动 [%s] http服务失败：%v", s.ac.HTTP.Name, err))
			return fmt.Errorf("启动 [%s] http服务失败：%v", s.ac.HTTP.Name, err)
		}
		return nil
	}

	<-serverShutdown
	return nil
}

func (s *HttpService) Stop(_ context.Context) error {
	return nil
}

func (s *HttpService) RegisterGRPC(serviceName, addr string, serviceID string) error {
	//TODO implement me
	panic("implement me")
}

func (s *HttpService) RegisterHTTP(serviceName, addr string, serviceID string) error {
	if err := s.registry.RegisterHTTP(serviceName, addr, s.sid, s.Health()); err != nil {
		return err
	}
	s.logger.Info("Service register success", "service", serviceName, "addr", addr, "sid", serviceID)
	return nil
}

func (s *HttpService) Discover() error {
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

func (s *HttpService) cancel() {
	if s.registry != nil {
		if err := s.registry.Cancel(s.sid); err != nil {
			s.logger.Error(err, "Service unregister failed", "service", s.ac.HTTP.Name, "addr", s.ac.HTTP.Addr(), "id", s.sid)
		}
		s.logger.Info("Service unregister success", "service", s.ac.HTTP.Name, "addr", s.ac.HTTP.Addr(), "id", s.sid)
	}
}

func (s *HttpService) Health() string {
	return "http://" + s.healthAddr
}
