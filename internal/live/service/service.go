package service

import (
	"fmt"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	user "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/msg_queue"
	lksdk "github.com/livekit/server-sdk-go"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"sync"
	"time"
)

const (
	liveUserPrefix  = "live.User."
	liveGroupPrefix = "live.Group."
	liveRoomPrefix  = "live.Room."
)

type Service struct {
	userClient           user.UserServiceClient
	relUserClient        relationgrpcv1.UserRelationServiceClient
	relGroupClient       relationgrpcv1.GroupRelationServiceClient
	groupClient          groupgrpcv1.GroupServiceClient
	msgClient            msggrpcv1.MsgServiceClient
	relationDialogClient relationgrpcv1.DialogServiceClient
	roomService          *lksdk.RoomServiceClient
	mqClient             *msg_queue.RabbitMQ
	redisClient          *cache.RedisClient

	ac *pkgconfig.AppConfig

	livekitServer string
	liveApiKey    string
	liveApiSecret string
	liveTimeout   time.Duration

	logger    *zap.Logger
	sid       string
	discovery discovery.Registry

	cache bool

	lock sync.Mutex
}

func New(ac *pkgconfig.AppConfig) *Service {
	s := &Service{
		liveTimeout:   ac.Livekit.Timeout,
		liveApiKey:    ac.Livekit.ApiKey,
		liveApiSecret: ac.Livekit.ApiSecret,
		livekitServer: ac.Livekit.Url,
		roomService:   lksdk.NewRoomServiceClient(ac.Livekit.Addr(), ac.Livekit.ApiKey, ac.Livekit.ApiSecret),
		redisClient:   setupRedisClient(ac),
		mqClient:      setRabbitMQProvider(ac),
		ac:            ac,
		sid:           xid.New().String(),
		logger:        plog.NewDefaultLogger("live_user_bff", int8(ac.Log.Level)),
	}
	s.cache = s.setCacheConfig()
	return s
}

func (s *Service) HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) error {
	switch serviceName {
	case "user_service":
		s.userClient = user.NewUserServiceClient(conn)
		s.logger.Info("gRPC client for user service initialized", zap.String("service", "user"), zap.String("addr", conn.Target()))
	case "relation_service":
		s.relUserClient = relationgrpcv1.NewUserRelationServiceClient(conn)
		s.relGroupClient = relationgrpcv1.NewGroupRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "userRelation"), zap.String("addr", conn.Target()))
	case "group_service":
		s.groupClient = groupgrpcv1.NewGroupServiceClient(conn)
		s.logger.Info("gRPC client for group service initialized", zap.String("service", "group"), zap.String("addr", conn.Target()))
	case "msg_service":
		s.msgClient = msggrpcv1.NewMsgServiceClient(conn)
		s.logger.Info("gRPC client for group service initialized", zap.String("service", "msg"), zap.String("addr", conn.Target()))
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

func setupRedisClient(cfg *pkgconfig.AppConfig) *cache.RedisClient {
	return cache.NewRedisClient(cfg.Redis.Addr(), cfg.Redis.Password)
}

func (s *Service) setCacheConfig() bool {
	if s.redisClient == nil && s.ac.Cache.Enable {
		return false
	}
	return s.ac.Cache.Enable
}