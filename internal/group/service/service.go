package service

//import (
//	"fmt"
//	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
//	mgrpc "github.com/cossim/coss-server/internal/group/interfaces/grpc"
//	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
//	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
//	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
//	"github.com/cossim/coss-server/pkg/cache"
//	pkgconfig "github.com/cossim/coss-server/pkg/config"
//	plog "github.com/cossim/coss-server/pkg/log"
//	"github.com/cossim/coss-server/pkg/msg_queue"
//	"github.com/cossim/coss-server/pkg/storage"
//	"github.com/cossim/coss-server/pkg/storage/minio"
//	"github.com/rs/xid"
//	"go.uber.org/zap"
//	"google.golang.org/grpc"
//)
//
//// Service struct
//type Service struct {
//	ac                  *pkgconfig.AppConfig
//	logger              *zap.Logger
//	sid                 string
//	dtmGrpcServer       string
//	relationServiceAddr string
//	dialogServiceAddr   string
//	groupServiceAddr    string
//	userServiceAddr     string
//	downloadURL         string
//	gatewayAddress      string
//	gatewayPort         string
//	cache               bool
//
//	groupService groupgrpcv1.GroupServiceServer
//	//groupClient           *mgrpc.GroupServiceServer
//	relationDialogService relationgrpcv1.DialogServiceClient
//	relationGroupService  relationgrpcv1.GroupRelationServiceClient
//	relationUserService   relationgrpcv1.UserRelationServiceClient
//	userService           usergrpcv1.UserServiceClient
//	pushService           pushgrpcv1.PushServiceClient
//
//	rabbitMQClient *msg_queue.RabbitMQ
//	redisClient    *cache.RedisClient
//	sp             storage.StorageProvider
//}
//
//func New(ac *pkgconfig.AppConfig, grpcService *mgrpc.GroupServiceServer) *Service {
//	logger := setupLogger(ac)
//	rabbitMQClient := setRabbitMQProvider(ac)
//	svc := &Service{
//		rabbitMQClient: rabbitMQClient,
//		redisClient:    setupRedis(ac),
//		downloadURL:    "/api/v1/storage/files/download",
//		sp:             setMinIOProvider(ac),
//		logger:         logger,
//		ac:             ac,
//		sid:            xid.New().String(),
//		dtmGrpcServer:  ac.Dtm.Addr(),
//	}
//	svc.cache = svc.setCacheConfig()
//	svc.setLoadSystem()
//	svc.groupService = grpcService
//	return svc
//}
//
//func setupRedis(ac *pkgconfig.AppConfig) *cache.RedisClient {
//	return cache.NewRedisClient(ac.Redis.Addr(), ac.Redis.Password)
//}
//
//func setMinIOProvider(ac *pkgconfig.AppConfig) storage.StorageProvider {
//	var err error
//	sp, err := minio.NewMinIOStorage(ac.OSS.Addr(), ac.OSS.AccessKey, ac.OSS.SecretKey, ac.OSS.SSL)
//	if err != nil {
//		panic(err)
//	}
//
//	return sp
//}
//
//func (s *Service) HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) error {
//	addr := conn.Target()
//	switch serviceName {
//	case "user_service":
//		s.userServiceAddr = addr
//		s.userService = usergrpcv1.NewUserServiceClient(conn)
//	case "relation_service":
//		s.relationServiceAddr = addr
//		s.dialogServiceAddr = addr
//		s.relationUserService = relationgrpcv1.NewUserRelationServiceClient(conn)
//		s.relationGroupService = relationgrpcv1.NewGroupRelationServiceClient(conn)
//		s.relationDialogService = relationgrpcv1.NewDialogServiceClient(conn)
//	case "push_service":
//		s.pushService = pushgrpcv1.NewPushServiceClient(conn)
//		s.logger.Info("gRPC client for push service initialized", zap.String("service", "push"), zap.String("addr", conn.Target()))
//
//	default:
//		return nil
//	}
//	s.logger.Info("gRPC client service initialized", zap.String("service", serviceName), zap.String("addr", addr))
//	return nil
//}
//
//func setupLogger(c *pkgconfig.AppConfig) *zap.Logger {
//	return plog.NewDefaultLogger("group_bff", int8(c.Log.Level))
//}
//
//func setRabbitMQProvider(c *pkgconfig.AppConfig) *msg_queue.RabbitMQ {
//	rmq, err := msg_queue.NewRabbitMQ(fmt.Sprintf("amqp://%s:%s@%s", c.MessageQueue.Username, c.MessageQueue.Password, c.MessageQueue.Addr()))
//	if err != nil {
//		panic(err)
//	}
//	return rmq
//}
//
//func (s *Service) setCacheConfig() bool {
//	if s.redisClient == nil && s.ac.Cache.Enable {
//		panic("redis is nil")
//	}
//	return s.ac.Cache.Enable
//}
//
//func (s *Service) setLoadSystem() {
//	env := s.ac.SystemConfig.Environment
//	if env == "" {
//		env = "dev"
//	}
//
//	switch env {
//	case "prod":
//		gatewayAdd := s.ac.SystemConfig.GatewayAddress
//		if gatewayAdd == "" {
//			gatewayAdd = "43.229.28.107"
//		}
//
//		s.gatewayAddress = gatewayAdd
//
//		gatewayPo := s.ac.SystemConfig.GatewayPort
//		if gatewayPo == "" {
//			gatewayPo = "8080"
//		}
//		s.gatewayPort = gatewayPo
//	default:
//		gatewayAdd := s.ac.SystemConfig.GatewayAddressDev
//		if gatewayAdd == "" {
//			gatewayAdd = "127.0.0.1"
//		}
//
//		s.gatewayAddress = gatewayAdd
//
//		gatewayPo := s.ac.SystemConfig.GatewayPortDev
//		if gatewayPo == "" {
//			gatewayPo = "8080"
//		}
//		s.gatewayPort = gatewayPo
//	}
//	if !s.ac.SystemConfig.Ssl {
//		s.gatewayAddress = s.gatewayAddress + ":" + s.gatewayPort
//	}
//}
