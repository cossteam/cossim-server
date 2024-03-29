package server

import (
	"context"
	"errors"
	"fmt"
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
	"time"
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
	server := New(handler)
	server.Addr = s.ac.HTTP.Addr()
	s.server = server
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
		if err := s.register(); err != nil {
			return err
		}
		go s.watchRegistry(ctx)
	}

	if s.ac.Register.Discover {
		go s.discover()
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

func (s *HttpService) register() error {
	serviceName := s.ac.HTTP.Name
	addr := s.ac.HTTP.Addr()
	serviceID := s.sid
	if err := s.registry.RegisterHTTP(serviceName, addr, serviceID, s.Health()); err != nil {
		s.logger.Error(err, "Service register failed", "service", serviceName, "addr", addr, "sid", serviceID)
		return err
	}
	s.logger.Info("Service register success", "service", serviceName, "addr", addr, "sid", serviceID)
	return nil
}

func (s *HttpService) discover() {
	// 定时器，每隔5秒执行一次服务发现
	ticker := time.NewTicker(5 * time.Second)
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

// watchRegistration 监听注册状态
func (s *HttpService) watchRegistry(ctx context.Context) {
	s.registry.KeepAlive(s.ac.HTTP.Name, s.sid, &discovery.RegisterOption{
		HealthCheckCallbackFn: func(b bool) {
			if !b {
				s.register()
			}
		},
	})
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
