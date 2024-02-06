package service

import (
	"fmt"
	"github.com/cossim/coss-server/interface/live/config"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/msg_queue"
	groupgrpcv1 "github.com/cossim/coss-server/service/group/api/v1"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	user "github.com/cossim/coss-server/service/user/api/v1"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"sync"
	"time"
)

const (
	liveUserPrefix  = "liveUser:"
	liveGroupPrefix = "liveGroup:"
)

type Service struct {
	userClient     user.UserServiceClient
	relUserClient  relationgrpcv1.UserRelationServiceClient
	relGroupClient relationgrpcv1.GroupRelationServiceClient
	groupClient    groupgrpcv1.GroupServiceClient
	roomService    *lksdk.RoomServiceClient
	mqClient       *msg_queue.RabbitMQ
	redisClient    *redis.Client

	livekitServer string
	liveApiKey    string
	liveApiSecret string
	liveTimeout   time.Duration

	logger    *zap.Logger
	sid       string
	discovery discovery.Discovery

	lock sync.Mutex
}

func New() *Service {
	return &Service{
		liveTimeout:   config.Conf.Livekit.Timeout,
		liveApiKey:    config.Conf.Livekit.ApiKey,
		liveApiSecret: config.Conf.Livekit.ApiSecret,
		livekitServer: config.Conf.Livekit.Url,
		roomService:   lksdk.NewRoomServiceClient(config.Conf.Livekit.Addr(), config.Conf.Livekit.ApiKey, config.Conf.Livekit.ApiSecret),

		redisClient: redis.NewClient(&redis.Options{
			Addr:     config.Conf.Redis.Addr(),
			Password: config.Conf.Redis.Password, // no password set
			DB:       0,                          // use default DB
		}),
		mqClient: setRabbitMQProvider(),
		// mqClient: msg_queue.NewRabbitMQ()

		sid: xid.New().String(),

		logger: plog.NewDevLogger("live_user_bff"),
	}
}

func (s *Service) Start(discover bool) {
	if discover {
		d, err := discovery.NewConsulRegistry(config.Conf.Register.Addr())
		if err != nil {
			panic(err)
		}
		s.discovery = d
		if err = s.discovery.RegisterHTTP(config.Conf.Register.Name, config.Conf.HTTP.Addr(), s.sid); err != nil {
			panic(err)
		}
		s.logger.Info("Service register success", zap.String("name", config.Conf.Register.Name), zap.String("addr", config.Conf.HTTP.Addr()), zap.String("id", s.sid))
		go s.discover()
	} else {
		s.direct()
	}
}

func Restart(discover bool) *Service {
	s := New()
	s.logger.Info("Service restart")
	s.Start(discover)
	return s
}

func (s *Service) Stop(discover bool) error {
	if !discover {
		return nil
	}
	if err := s.discovery.Cancel(s.sid); err != nil {
		log.Printf("Failed to cancel service registration: %v", err)
		return err
	}
	log.Printf("Service registration canceled ServiceName: %s  Addr: %s  ID: %s", config.Conf.Register.Name, config.Conf.GRPC.Addr(), s.sid)
	return nil
}

func (s *Service) discover() {
	var wg sync.WaitGroup
	type serviceInfo struct {
		ServiceName string
		Addr        string
	}
	ch := make(chan serviceInfo)

	for serviceName, c := range config.Conf.Discovers {
		if c.Direct {
			continue
		}
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
				s.logger.Info("Service discovery successful", zap.String("service", config.Conf.Register.Name), zap.String("addr", addr))
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
	for serviceName, _ := range config.Conf.Discovers {
		if err := s.handlerGrpcClient(serviceName, config.Conf.Discovers[serviceName].Addr()); err != nil {
			panic(err)
		}
	}
}

func (s *Service) handlerGrpcClient(serviceName string, addr string) error {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	switch serviceName {
	case "user":
		s.userClient = user.NewUserServiceClient(conn)
		s.logger.Info("gRPC client for user service initialized", zap.String("service", "user"), zap.String("addr", conn.Target()))
	case "relation":
		s.relUserClient = relationgrpcv1.NewUserRelationServiceClient(conn)
		s.relGroupClient = relationgrpcv1.NewGroupRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "userRelation"), zap.String("addr", conn.Target()))
	case "group":
		s.groupClient = groupgrpcv1.NewGroupServiceClient(conn)
		s.logger.Info("gRPC client for group service initialized", zap.String("service", "group"), zap.String("addr", conn.Target()))
	}

	return nil
}

func setRabbitMQProvider() *msg_queue.RabbitMQ {
	rmq, err := msg_queue.NewRabbitMQ(fmt.Sprintf("amqp://%s:%s@%s", config.Conf.MessageQueue.Username, config.Conf.MessageQueue.Password, config.Conf.MessageQueue.Addr()))
	if err != nil {
		panic(err)
	}
	return rmq
}
