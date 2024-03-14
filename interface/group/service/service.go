package service

import (
	"fmt"
	"github.com/cossim/coss-server/pkg/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/msg_queue"
	"github.com/cossim/coss-server/pkg/storage"
	"github.com/cossim/coss-server/pkg/storage/minio"
	groupgrpcv1 "github.com/cossim/coss-server/service/group/api/v1"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	usergrpcv1 "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"sync"
	"time"
)

// Service struct
type Service struct {
	groupClient          groupgrpcv1.GroupServiceClient
	relationDialogClient relationgrpcv1.DialogServiceClient
	relationGroupClient  relationgrpcv1.GroupRelationServiceClient
	relationUserClient   relationgrpcv1.UserRelationServiceClient
	userClient           usergrpcv1.UserServiceClient
	rabbitMQClient       *msg_queue.RabbitMQ
	redisClient          *cache.RedisClient
	sp                   storage.StorageProvider

	logger    *zap.Logger
	sid       string
	discovery discovery.Registry
	conf      *pkgconfig.AppConfig

	dtmGrpcServer      string
	relationGrpcServer string
	dialogGrpcServer   string
	groupGrpcServer    string
	downloadURL        string
	gatewayAddress     string
	gatewayPort        string

	dis   bool
	cache bool
}

func New(ac *pkgconfig.AppConfig) *Service {
	logger := setupLogger(ac)
	rabbitMQClient := setRabbitMQProvider(ac)
	svc := &Service{
		rabbitMQClient: rabbitMQClient,
		redisClient:    setupRedis(ac),
		downloadURL:    "/api/v1/storage/files/download",
		sp:             setMinIOProvider(ac),
		logger:         logger,
		conf:           ac,
		sid:            xid.New().String(),
		dtmGrpcServer:  ac.Dtm.Addr(),
		dis:            false,
	}
	svc.cache = svc.setCacheConfig()
	svc.setLoadSystem()
	return svc
}

func (s *Service) Start() error {
	if s.dis {
		d, err := discovery.NewConsulRegistry(s.conf.Register.Addr())
		if err != nil {
			return err
		}
		s.discovery = d
		if err = s.discovery.RegisterHTTP(s.conf.Register.Name, s.conf.HTTP.Addr(), s.sid, ""); err != nil {
			panic(err)
		}
		s.logger.Info("Service registration successful", zap.String("service", s.conf.Register.Name), zap.String("addr", s.conf.HTTP.Addr()), zap.String("sid", s.sid))
		go s.discover()
	} else {
		s.direct()
	}
	return nil
}

func setupRedis(ac *pkgconfig.AppConfig) *cache.RedisClient {
	return cache.NewRedisClient(ac.Redis.Addr(), ac.Redis.Password)
}

func setMinIOProvider(ac *pkgconfig.AppConfig) storage.StorageProvider {
	var err error
	sp, err := minio.NewMinIOStorage(ac.OSS.Addr(), ac.OSS.AccessKey, ac.OSS.SecretKey, ac.OSS.SSL)
	if err != nil {
		panic(err)
	}

	return sp
}

func (s *Service) Stop() error {
	if !s.dis {
		return nil
	}
	if err := s.discovery.Cancel(s.sid); err != nil {
		s.logger.Error("Failed to cancel service registration: %v", zap.Error(err))
		return err
	}
	s.logger.Info("Service registration canceled", zap.String("service", s.conf.Register.Name), zap.String("addr", s.conf.GRPC.Addr()), zap.String("sid", s.sid))
	return nil
}

func (s *Service) discover() {
	var wg sync.WaitGroup
	type serviceInfo struct {
		ServiceName string
		Addr        string
	}
	ch := make(chan serviceInfo)

	for serviceName, c := range s.conf.Discovers {
		wg.Add(1)
		go func(serviceName string, c pkgconfig.ServiceConfig) {
			defer wg.Done()
			for {
				addr, err := s.discovery.Discover(c.Name)
				if err != nil {
					s.logger.Info("Service discovery failed", zap.String("service", c.Name))
					time.Sleep(15 * time.Second)
					continue
				}
				s.logger.Info("Service discovery successful", zap.String("service", s.conf.Register.Name), zap.String("addr", addr))

				ch <- serviceInfo{ServiceName: serviceName, Addr: addr}
				break
			}
		}(serviceName, c)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for info := range ch {
		if err := s.handlerGrpcClient(info.ServiceName, info.Addr); err != nil {
			s.logger.Info("Failed to initialize gRPC client for service", zap.String("service", info.ServiceName), zap.String("addr", info.Addr))
		}
	}
}

func (s *Service) direct() {
	for serviceName, _ := range s.conf.Discovers {
		if err := s.handlerGrpcClient(serviceName, s.conf.Discovers[serviceName].Addr()); err != nil {
			panic(err)
		}
	}
}

func (s *Service) HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) error {
	switch serviceName {
	case "user_service":
		s.userClient = usergrpcv1.NewUserServiceClient(conn)
		s.logger.Info("gRPC client for user service initialized", zap.String("service", "user"), zap.String("addr", conn.Target()))
	case "relation_service":
		s.relationGrpcServer = conn.Target()
		s.dialogGrpcServer = conn.Target()
		s.relationUserClient = relationgrpcv1.NewUserRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "userRelation"), zap.String("addr", conn.Target()))

		s.relationGroupClient = relationgrpcv1.NewGroupRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "groupRelation"), zap.String("addr", conn.Target()))

		s.relationDialogClient = relationgrpcv1.NewDialogServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "dialogRelation"), zap.String("addr", conn.Target()))
	case "group_service":
		s.groupGrpcServer = conn.Target()
		s.groupClient = groupgrpcv1.NewGroupServiceClient(conn)
		s.logger.Info("gRPC client for group service initialized", zap.String("service", "group"), zap.String("addr", conn.Target()))
	}

	return nil
}

func (s *Service) handlerGrpcClient(serviceName string, addr string) error {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	switch serviceName {
	case "user_service":
		s.userClient = usergrpcv1.NewUserServiceClient(conn)
		s.logger.Info("gRPC client for user service initialized", zap.String("service", "user"), zap.String("addr", conn.Target()))
	case "relation_service":
		s.relationGrpcServer = addr
		s.dialogGrpcServer = addr
		s.relationUserClient = relationgrpcv1.NewUserRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "userRelation"), zap.String("addr", conn.Target()))

		s.relationGroupClient = relationgrpcv1.NewGroupRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "groupRelation"), zap.String("addr", conn.Target()))

		s.relationDialogClient = relationgrpcv1.NewDialogServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "dialogRelation"), zap.String("addr", conn.Target()))
	case "group_service":
		s.groupGrpcServer = addr
		s.groupClient = groupgrpcv1.NewGroupServiceClient(conn)
		s.logger.Info("gRPC client for group service initialized", zap.String("service", "group"), zap.String("addr", conn.Target()))
	}

	return nil
}

func setupLogger(c *pkgconfig.AppConfig) *zap.Logger {
	return plog.NewDefaultLogger("group_bff", int8(c.Log.Level))
}

func setRabbitMQProvider(c *pkgconfig.AppConfig) *msg_queue.RabbitMQ {
	rmq, err := msg_queue.NewRabbitMQ(fmt.Sprintf("amqp://%s:%s@%s", c.MessageQueue.Username, c.MessageQueue.Password, c.MessageQueue.Addr()))
	if err != nil {
		panic(err)
	}
	return rmq
}

func (s *Service) setCacheConfig() bool {
	if s.redisClient == nil && s.conf.Cache.Enable {
		panic("redis is nil")
	}
	return s.conf.Cache.Enable
}

func (s *Service) setLoadSystem() {
	env := s.conf.SystemConfig.Environment
	if env == "" {
		env = "dev"
	}

	switch env {
	case "prod":
		gatewayAdd := s.conf.SystemConfig.GatewayAddress
		if gatewayAdd == "" {
			gatewayAdd = "43.229.28.107"
		}

		s.gatewayAddress = gatewayAdd

		gatewayPo := s.conf.SystemConfig.GatewayPort
		if gatewayPo == "" {
			gatewayPo = "8080"
		}
		s.gatewayPort = gatewayPo
	default:
		gatewayAdd := s.conf.SystemConfig.GatewayAddressDev
		if gatewayAdd == "" {
			gatewayAdd = "127.0.0.1"
		}

		s.gatewayAddress = gatewayAdd

		gatewayPo := s.conf.SystemConfig.GatewayPortDev
		if gatewayPo == "" {
			gatewayPo = "8080"
		}
		s.gatewayPort = gatewayPo
	}
	if !s.conf.SystemConfig.Ssl {
		s.gatewayAddress = s.gatewayAddress + ":" + s.gatewayPort
	}
}
