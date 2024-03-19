package service

import (
	"context"
	"fmt"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	grpchandler "github.com/cossim/coss-server/internal/relation/interface/grpc"
	user "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/msg_queue"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Service struct
type Service struct {
	//dialogClient                relationgrpcv1.DialogServiceClient
	msgClient msggrpcv1.MsgServiceClient
	//groupRelationClient         relationgrpcv1.GroupRelationServiceClient
	//userRelationClient
	//userFriendRequestClient     relationgrpcv1.UserFriendRequestServiceClient
	//groupAnnouncementClient     relationgrpcv1.GroupAnnouncementServiceClient
	//groupAnnouncementReadClient relationgrpcv1.GroupAnnouncementReadServiceClient
	//groupJoinRequestClient      relationgrpcv1.GroupJoinRequestServiceClient
	svc            *grpchandler.Handler
	userClient     user.UserServiceClient
	groupClient    groupgrpcv1.GroupServiceClient
	rabbitMQClient *msg_queue.RabbitMQ
	redisClient    *cache.RedisClient

	logger    *zap.Logger
	sid       string
	discovery discovery.Registry
	ac        *pkgconfig.AppConfig

	dtmGrpcServer      string
	relationGrpcServer string
	msgGrpcServer      string
	dialogGrpcServer   string
	cache              bool
}

func New(ac *pkgconfig.AppConfig, grpcService *grpchandler.Handler) *Service {
	s := &Service{
		rabbitMQClient: setRabbitMQProvider(ac),
		redisClient:    setupRedis(ac),
		logger:         plog.NewDefaultLogger("relation_bff", int8(ac.Log.Level)),
		ac:             ac,
		sid:            xid.New().String(),
		dtmGrpcServer:  ac.Dtm.Addr(),
	}
	s.cache = s.setCacheConfig()
	s.svc = grpcService
	return s
}

func (s *Service) HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) error {
	switch serviceName {
	case "user_service":
		s.userClient = user.NewUserServiceClient(conn)
		s.logger.Info("gRPC client for user service initialized", zap.String("service", "user"), zap.String("addr", conn.Target()))
	case "group_service":
		s.groupClient = groupgrpcv1.NewGroupServiceClient(conn)
		s.logger.Info("gRPC client for group service initialized", zap.String("service", "group"), zap.String("addr", conn.Target()))
	case "msg_service":
		s.msgGrpcServer = conn.Target()
		s.msgClient = msggrpcv1.NewMsgServiceClient(conn)
		s.logger.Info("gRPC client for msg service initialized", zap.String("service", "msg"), zap.String("addr", conn.Target()))
	}

	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	return nil
}

func setRabbitMQProvider(ac *pkgconfig.AppConfig) *msg_queue.RabbitMQ {
	rmq, err := msg_queue.NewRabbitMQ(fmt.Sprintf("amqp://%s:%s@%s", ac.MessageQueue.Username, ac.MessageQueue.Password, ac.MessageQueue.Addr()))
	if err != nil {
		panic(err)
	}
	return rmq
}

func setupRedis(ac *pkgconfig.AppConfig) *cache.RedisClient {
	return cache.NewRedisClient(ac.Redis.Addr(), ac.Redis.Password)
}

func (s *Service) setCacheConfig() bool {
	if s.redisClient == nil && s.ac.Cache.Enable {
		panic("redis is nil")
	}
	return s.ac.Cache.Enable
}
