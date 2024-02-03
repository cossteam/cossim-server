package service

import (
	"fmt"
	"github.com/cossim/coss-server/interface/group/config"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/msg_queue"
	groupgrpcv1 "github.com/cossim/coss-server/service/group/api/v1"
	dialoggrpcv1 "github.com/cossim/coss-server/service/relation/api/v1/dialog"
	grouprelationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1/group_relation"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1/user_relation"
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
	relationDialogClient dialoggrpcv1.DialogServiceClient
	relationGroupClient  grouprelationgrpcv1.GroupRelationServiceClient
	relationUserClient   relationgrpcv1.UserRelationServiceClient
	userClient           usergrpcv1.UserServiceClient
	rabbitMQClient       *msg_queue.RabbitMQ

	logger    *zap.Logger
	sid       string
	discovery discovery.Discovery
	conf      *pkgconfig.AppConfig

	dtmGrpcServer      string
	relationGrpcServer string
	dialogGrpcServer   string
	groupGrpcServer    string
}

func New() *Service {
	logger := setupLogger()
	rabbitMQClient := setRabbitMQProvider(config.Conf)

	return &Service{
		rabbitMQClient: rabbitMQClient,
		logger:         logger,
		conf:           config.Conf,
		sid:            xid.New().String(),
		dtmGrpcServer:  config.Conf.Dtm.Addr(),
	}
}

func (s *Service) Start(discover bool) {
	if discover {
		d, err := discovery.NewConsulRegistry(s.conf.Register.Addr())
		if err != nil {
			panic(err)
		}
		s.discovery = d
		if err = s.discovery.RegisterHTTP(s.conf.Register.Name, s.conf.HTTP.Addr(), s.sid); err != nil {
			panic(err)
		}
		s.logger.Info("Service registration successful", zap.String("service", s.conf.Register.Name), zap.String("addr", s.conf.HTTP.Addr()), zap.String("sid", s.sid))
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

func (s *Service) handlerGrpcClient(serviceName string, addr string) error {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	switch serviceName {
	case "user_relation":
		s.userClient = usergrpcv1.NewUserServiceClient(conn)
		s.logger.Info("gRPC client for user_relation service initialized", zap.String("service", "user_relation"), zap.String("addr", conn.Target()))
	case "relation":
		s.relationGrpcServer = addr
		s.dialogGrpcServer = addr
		s.relationUserClient = relationgrpcv1.NewUserRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "userRelation"), zap.String("addr", conn.Target()))

		s.relationGroupClient = grouprelationgrpcv1.NewGroupRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "groupRelation"), zap.String("addr", conn.Target()))

		s.relationDialogClient = dialoggrpcv1.NewDialogServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "dialogRelation"), zap.String("addr", conn.Target()))
	case "group_relation":
		s.groupGrpcServer = addr
		s.groupClient = groupgrpcv1.NewGroupServiceClient(conn)
		s.logger.Info("gRPC client for group_relation service initialized", zap.String("service", "group_relation"), zap.String("addr", conn.Target()))
	}

	return nil
}

func setupLogger() *zap.Logger {
	return plog.NewDevLogger("group_bff")
}

func setRabbitMQProvider(c *pkgconfig.AppConfig) *msg_queue.RabbitMQ {
	rmq, err := msg_queue.NewRabbitMQ(fmt.Sprintf("amqp://%s:%s@%s", c.MessageQueue.Username, c.MessageQueue.Password, c.MessageQueue.Addr()))
	if err != nil {
		panic(err)
	}
	return rmq
}
