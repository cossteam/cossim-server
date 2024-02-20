package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
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
		logger:     logger.WithValues("kind", "http server", "name", c.Register.Name),
		ac:         c,
		svc:        svc,
		addr:       c.HTTP.Addr(),
		sid:        xid.New().String(),
		healthAddr: c.HTTP.Address + healthAddr,
	}

	handler := gin.Default()
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
		if err := s.RegisterHTTP(s.ac.Register.Name, s.ac.HTTP.Addr(), s.sid); err != nil {
			return err
		}
	}

	go func() {
		if s.ac.Register.Discover {
			if err := s.Discover(); err != nil {
				s.logger.Error(err, "发现服务失败")
			}
		}
	}()

	serverShutdown := make(chan struct{})
	go func() {
		<-ctx.Done()
		s.logger.Info("shutting down httpServer")

		if err := s.server.Shutdown(ctx); err != nil {
			s.logger.Error(err, "error shutting down httpServer")
		}
		close(serverShutdown)
	}()

	s.logger.Info(fmt.Sprintf("%s http service start", s.ac.Register.Name), "addr", s.ac.HTTP.Addr())
	if err := s.server.ListenAndServe(); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			s.logger.Info(fmt.Sprintf("%s http service stop", s.ac.Register.Name), "addr", s.addr)
			return nil
		}
		s.logger.Error(err, fmt.Sprintf("启动 [%s] http服务失败：%v", s.ac.Register.Name, err))
		return fmt.Errorf("启动 [%s] http服务失败：%v", s.ac.Register.Name, err)
	}

	<-serverShutdown
	return nil
}

//func (s *HttpService) Stop(_ context.Context) error {
//	s.logger.Info(fmt.Sprintf("%s http service try stop", s.ac.RegisterGRPC.Name), "addr", s.addr)
//	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
//	defer cancel()
//	if err := s.server.Shutdown(ctx); err != nil {
//		s.logger.Info(fmt.Sprintf("%s http service try stop failed", s.ac.RegisterGRPC.Name), "msg", err.Error(), "addr", s.addr)
//	}
//	return nil
//}

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

	clients := make(map[string]*grpc.ClientConn)
	var mu sync.Mutex // 用于对 clients 的并发访问进行保护
	var wg sync.WaitGroup

	// 控制并发数的信号量
	sem := make(chan struct{}, 10) // 限制并发数为 10

	for serviceName, c := range s.ac.Discovers {
		if c.Direct {
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
					s.logger.Error(err, "Service discovery failed", "service", c.Name)
					return err
				}
				s.logger.Info("Service discovery success", "service", c.Name, "addr", addr)
				conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					return err
				}
				mu.Lock()
				clients[c.Name] = conn
				mu.Unlock()
				return nil
			}
			if err := backoff.Retry(retryFunc, backoffSettings); err != nil {
				s.logger.Error(err, "Failed to initialize gRPC client for service after retries")
				return
			}
		}(serviceName, c)
	}
	wg.Wait()
	return s.svc.DiscoverServices(clients)
}

func (s *HttpService) Health() string {
	return "http://" + s.healthAddr
}
