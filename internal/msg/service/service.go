package service

import (
	"context"
	"fmt"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	grpcHandler "github.com/cossim/coss-server/internal/msg/interface/grpc"
	pushv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/encryption"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Service struct
type Service struct {
	ac                  *pkgconfig.AppConfig
	dtmGrpcServer       string
	relationServiceAddr string
	userServiceAddr     string
	groupServiceAddr    string
	msgServiceAddr      string
	redisClient         *cache.RedisClient
	logger              *zap.Logger
	sid                 string
	cache               bool
	enc                 encryption.Encryptor

	relationUserService   relationgrpcv1.UserRelationServiceClient
	relationGroupService  relationgrpcv1.GroupRelationServiceClient
	relationDialogService relationgrpcv1.DialogServiceClient
	userService           usergrpcv1.UserServiceClient
	userLoginService      usergrpcv1.UserLoginServiceClient
	groupService          groupgrpcv1.GroupServiceClient
	msgService            msggrpcv1.MsgServiceServer
	msgGroupService       msggrpcv1.GroupMessageServiceServer
	pushService           pushv1.PushServiceClient
	//msgClient            *grpcHandler.Handler
}

func New(ac *pkgconfig.AppConfig, handler *grpcHandler.Handler) *Service {
	s := &Service{
		ac:            ac,
		logger:        plog.NewDefaultLogger("msg_bff", int8(ac.Log.Level)),
		sid:           xid.New().String(),
		redisClient:   setupRedis(ac),
		dtmGrpcServer: ac.Dtm.Addr(),
		//rabbitMQClient: mqClient,
		//pool:     make(map[string]map[string][]*client),
		//mqClient: mqClient,
		//pool:     make(map[string]map[string][]*client),
	}
	s.msgService = handler
	s.msgGroupService = handler
	s.cache = s.setCacheConfig()
	s.setupEncryption(ac)
	return s
}

func (s *Service) Stop(ctx context.Context) error {
	return nil
}

func (s *Service) HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) error {
	addr := conn.Target()
	switch serviceName {
	case "user_service":
		s.userServiceAddr = addr
		s.userService = usergrpcv1.NewUserServiceClient(conn)
		s.userLoginService = usergrpcv1.NewUserLoginServiceClient(conn)
	case "relation_service":
		s.relationServiceAddr = addr
		s.relationUserService = relationgrpcv1.NewUserRelationServiceClient(conn)
		s.relationGroupService = relationgrpcv1.NewGroupRelationServiceClient(conn)
		s.relationDialogService = relationgrpcv1.NewDialogServiceClient(conn)
	case "group_service":
		s.groupServiceAddr = addr
		s.groupService = groupgrpcv1.NewGroupServiceClient(conn)
	case "push_service":
		s.pushService = pushv1.NewPushServiceClient(conn)
		fmt.Println("push_service", s.pushService)
	default:
		return nil
	}
	s.logger.Info("gRPC client service initialized", zap.String("service", serviceName), zap.String("addr", addr))
	return nil
}

func setupRedis(ac *pkgconfig.AppConfig) *cache.RedisClient {
	return cache.NewRedisClient(ac.Redis.Addr(), ac.Redis.Password)
}

func (s *Service) setupEncryption(ac *pkgconfig.AppConfig) {
	enc2 := encryption.NewEncryptor([]byte(ac.Encryption.Passphrase), ac.Encryption.Name, ac.Encryption.Email, ac.Encryption.RsaBits, ac.Encryption.Enable)

	err := enc2.ReadKeyPair()
	if err != nil {
		return
	}

	s.enc = enc2
}

func (s *Service) setCacheConfig() bool {
	if s.redisClient == nil && s.ac.Cache.Enable {
		panic("redis is nil")
	}
	return s.ac.Cache.Enable
}
