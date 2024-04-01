package service

import (
	"context"
	"fmt"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	grpchandler "github.com/cossim/coss-server/internal/relation/interface/grpc"
	userv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/msg_queue"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Service struct
type Service struct {
	ac                  *pkgconfig.AppConfig
	logger              *zap.Logger
	sid                 string
	dtmGrpcServer       string
	relationServiceAddr string
	msgServiceAddr      string
	userServiceAddr     string
	groupServiceAddr    string
	cache               bool

	relationGroupService             relationgrpcv1.GroupRelationServiceServer
	relationUserService              relationgrpcv1.UserRelationServiceServer
	relationUserFriendRequestService relationgrpcv1.UserFriendRequestServiceServer
	relationGroupJoinRequestService  relationgrpcv1.GroupJoinRequestServiceServer
	relationGroupAnnouncementService relationgrpcv1.GroupAnnouncementServiceServer
	relationDialogService            relationgrpcv1.DialogServiceServer
	userService                      userv1.UserServiceClient
	groupService                     groupgrpcv1.GroupServiceClient
	pushService                      pushgrpcv1.PushServiceClient

	msgClient      msggrpcv1.MsgServiceClient
	rabbitMQClient *msg_queue.RabbitMQ
	redisClient    *cache.RedisClient
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
	s.relationGroupService = grpcService
	s.relationUserService = grpcService
	s.relationUserFriendRequestService = grpcService
	s.relationGroupJoinRequestService = grpcService
	s.relationGroupAnnouncementService = grpcService
	s.relationDialogService = grpcService
	return s
}

func (s *Service) HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) error {
	addr := conn.Target()
	switch serviceName {
	case "user_service":
		s.userServiceAddr = addr
		s.userService = userv1.NewUserServiceClient(conn)
	case "group_service":
		s.groupServiceAddr = addr
		s.groupService = groupgrpcv1.NewGroupServiceClient(conn)
	case "msg_service":
		s.msgServiceAddr = addr
		s.msgClient = msggrpcv1.NewMsgServiceClient(conn)
	case "relation_service":
		s.relationServiceAddr = addr
	case "push_service":
		s.pushService = pushgrpcv1.NewPushServiceClient(conn)
	default:
		return nil
	}
	s.logger.Info("gRPC client service initialized", zap.String("service", serviceName), zap.String("addr", addr))
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
