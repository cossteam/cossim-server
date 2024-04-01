package service

import (
	"fmt"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	user "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
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
	ac                  *pkgconfig.AppConfig
	logger              *zap.Logger
	livekitServer       string
	liveApiKey          string
	liveApiSecret       string
	sid                 string
	userServiceAddr     string
	relationServiceAddr string
	groupServiceAddr    string
	msgServiceAddr      string
	liveTimeout         time.Duration
	lock                sync.Mutex
	cache               bool

	userService           user.UserServiceClient
	relationUserService   relationgrpcv1.UserRelationServiceClient
	relationGroupService  relationgrpcv1.GroupRelationServiceClient
	groupService          groupgrpcv1.GroupServiceClient
	msgService            msggrpcv1.MsgServiceClient
	relationDialogService relationgrpcv1.DialogServiceClient
	pushService           pushgrpcv1.PushServiceClient
	roomService           *lksdk.RoomServiceClient
	mqClient              *msg_queue.RabbitMQ
	redisClient           *cache.RedisClient
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
	addr := conn.Target()
	switch serviceName {
	case "user_service":
		s.userServiceAddr = addr
		s.userService = user.NewUserServiceClient(conn)
	case "relation_service":
		s.relationServiceAddr = addr
		s.relationUserService = relationgrpcv1.NewUserRelationServiceClient(conn)
		s.relationGroupService = relationgrpcv1.NewGroupRelationServiceClient(conn)
	case "group_service":
		s.groupServiceAddr = addr
		s.groupService = groupgrpcv1.NewGroupServiceClient(conn)
	case "msg_service":
		s.msgServiceAddr = addr
		s.msgService = msggrpcv1.NewMsgServiceClient(conn)
	case "push_service":
		s.pushService = pushgrpcv1.NewPushServiceClient(conn)
	default:
		return nil
	}
	s.logger.Info("gRPC client service initialized", zap.String("service", serviceName), zap.String("addr", addr))
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
