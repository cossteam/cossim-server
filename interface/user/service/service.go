package service

import (
	"fmt"
	"github.com/cossim/coss-server/interface/user/config"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/msg_queue"
	"github.com/cossim/coss-server/pkg/storage"
	"github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/cossim/coss-server/pkg/utils/os"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	user "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"sync"
	"time"
)

// Service struct
type Service struct {
	conf   *pkgconfig.AppConfig
	logger *zap.Logger

	discovery       discovery.Discovery
	userClient      user.UserServiceClient
	relClient       relationgrpcv1.UserRelationServiceClient
	sp              storage.StorageProvider
	redisClient     *redis.Client
	rabbitMQClient  *msg_queue.RabbitMQ
	sid             string
	tokenExpiration time.Duration
	appPath         string
	downloadURL     string
	gatewayAddress  string
	gatewayPort     string
}

func New() (s *Service) {

	s = &Service{
		conf:            config.Conf,
		sid:             xid.New().String(),
		tokenExpiration: 60 * 60 * 24 * 3 * time.Second,
		rabbitMQClient:  setRabbitMQProvider(),
		//appPath:         path,
		sp:          setMinIOProvider(),
		downloadURL: "/api/v1/storage/files/download",
	}
	s.logger = setupLogger()
	s.setLoadSystem()
	s.setupRedis()
	return s
}

func (s *Service) Start(discover bool) {
	//gate := s.conf.Discovers["gateway"]
	s.gatewayAddress = s.gatewayAddress + ":" + s.gatewayPort
	if discover {
		d, err := discovery.NewConsulRegistry(s.conf.Register.Addr())
		if err != nil {
			panic(err)
		}
		s.discovery = d
		if err = s.discovery.RegisterHTTP(s.conf.Register.Name, s.conf.HTTP.Addr(), s.sid); err != nil {
			panic(err)
		}
		s.logger.Info("Service register success", zap.String("name", s.conf.Register.Name), zap.String("addr", s.conf.HTTP.Addr()), zap.String("id", s.sid))
		go s.discover()
	} else {
		s.direct()
	}
}

func (s *Service) Stop(discover bool) error {
	if !discover {
		return nil
	}
	if err := s.discovery.Cancel(s.sid); err != nil {
		log.Printf("Failed to cancel service registration: %v", err)
		return err
	}
	log.Printf("Service registration canceled ServiceName: %s  Addr: %s  ID: %s", s.conf.Register.Name, s.conf.GRPC.Addr(), s.sid)
	return nil
}

func Restart(discover bool) *Service {
	s := New()
	s.logger.Info("Service restart")
	s.Start(discover)
	return s
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
					s.logger.Info("Service discovery failed", zap.String("service", c.Name), zap.Error(err))
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
			log.Printf("Failed to initialize gRPC client for service: %s, Error: %v\n", info.ServiceName, err)
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

func (s *Service) handlerGrpcClient(serviceName string, addr string) error {
	switch serviceName {
	case "user":
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		s.userClient = user.NewUserServiceClient(conn)
		s.logger.Info("gRPC client for user service initialized", zap.String("addr", conn.Target()))
	case "relation":
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		s.relClient = relationgrpcv1.NewUserRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("addr", conn.Target()))
	}

	return nil
}

func setupLogger() *zap.Logger {
	return plog.NewDefaultLogger("user_bff")
}

func (s *Service) setupRedis() {
	s.redisClient = redis.NewClient(&redis.Options{
		Addr:     s.conf.Redis.Addr(),
		Password: s.conf.Redis.Password, // no password set
		DB:       0,                     // use default DB
		//Protocol: cfg,
	})
}

func (s *Service) Ping() {
}

func setRabbitMQProvider() *msg_queue.RabbitMQ {
	rmq, err := msg_queue.NewRabbitMQ(fmt.Sprintf("amqp://%s:%s@%s", config.Conf.MessageQueue.Username, config.Conf.MessageQueue.Password, config.Conf.MessageQueue.Addr()))
	if err != nil {
		panic(err)
	}
	return rmq
}

func setMinIOProvider() storage.StorageProvider {
	var err error
	sp, err := minio.NewMinIOStorage(config.Conf.OSS["minio"].Addr(), config.Conf.OSS["minio"].AccessKey, config.Conf.OSS["minio"].SecretKey, config.Conf.OSS["minio"].SSL)
	if err != nil {
		panic(err)
	}

	return sp
}

func (s *Service) setLoadSystem() {

	env := config.Conf.SystemConfig.Environment
	if env == "" {
		env = "dev"
	}

	switch env {
	case "prod":
		path := config.Conf.SystemConfig.AvatarFilePath
		if path == "" {
			path = "/.catch/"
		}
		s.appPath = path

		gatewayAdd := config.Conf.SystemConfig.GatewayAddress
		if gatewayAdd == "" {
			gatewayAdd = "43.229.28.107"
		}

		s.gatewayAddress = gatewayAdd

		gatewayPo := config.Conf.SystemConfig.GatewayPort
		if gatewayPo == "" {
			gatewayPo = "8080"
		}
		s.gatewayPort = gatewayPo
	default:
		path := config.Conf.SystemConfig.AvatarFilePathDev
		if path == "" {
			npath, err := os.GetPackagePath()
			if err != nil {
				panic(err)
			}
			path = npath + "deploy/docker/config/common/"
		}
		s.appPath = path

		gatewayAdd := config.Conf.SystemConfig.GatewayAddressDev
		if gatewayAdd == "" {
			gatewayAdd = "127.0.0.1"
		}

		s.gatewayAddress = gatewayAdd

		gatewayPo := config.Conf.SystemConfig.GatewayPortDev
		if gatewayPo == "" {
			gatewayPo = "8080"
		}
		s.gatewayPort = gatewayPo
	}

}
