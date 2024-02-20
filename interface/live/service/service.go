package service

import (
	"fmt"
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

	ac *pkgconfig.AppConfig

	livekitServer string
	liveApiKey    string
	liveApiSecret string
	liveTimeout   time.Duration

	logger    *zap.Logger
	sid       string
	discovery discovery.Registry

	lock sync.Mutex
}

func New(ac *pkgconfig.AppConfig) *Service {
	return &Service{
		liveTimeout:   ac.Livekit.Timeout,
		liveApiKey:    ac.Livekit.ApiKey,
		liveApiSecret: ac.Livekit.ApiSecret,
		livekitServer: ac.Livekit.Url,
		roomService:   lksdk.NewRoomServiceClient(ac.Livekit.Addr(), ac.Livekit.ApiKey, ac.Livekit.ApiSecret),
		redisClient: redis.NewClient(&redis.Options{
			Addr:     ac.Redis.Addr(),
			Password: ac.Redis.Password, // no password set
			DB:       0,                 // use default DB
		}),
		mqClient: setRabbitMQProvider(ac),
		ac:       ac,
		sid:      xid.New().String(),
		logger:   plog.NewDevLogger("live_user_bff"),
	}
}

func (s *Service) HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) error {
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

func setRabbitMQProvider(ac *pkgconfig.AppConfig) *msg_queue.RabbitMQ {
	rmq, err := msg_queue.NewRabbitMQ(fmt.Sprintf("amqp://%s:%s@%s", ac.MessageQueue.Username, ac.MessageQueue.Password, ac.MessageQueue.Addr()))
	if err != nil {
		panic(err)
	}
	return rmq
}
