package service

import (
	"context"
	"fmt"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	pushv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	userv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/internal/user/cache"
	grpchandler "github.com/cossim/coss-server/internal/user/interfaces/grpc"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/email"
	"github.com/cossim/coss-server/pkg/email/smtp"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/storage"
	"github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

// Service struct
type Service struct {
	ac     *pkgconfig.AppConfig
	logger *zap.Logger

	userService      userv1.UserServiceServer
	userLoginService userv1.UserLoginServiceServer
	relationService  relationgrpcv1.UserRelationServiceClient
	dialogService    relationgrpcv1.DialogServiceClient
	pushService      pushv1.PushServiceClient
	msgService       msggrpcv1.MsgServiceClient
	authService      userv1.UserAuthServiceServer

	storageService storage.StorageProvider
	//redisClient    *cache.RedisClient
	smtpClient email.EmailProvider
	userCache  cache.UserCache

	dtmGrpcServer       string
	relationServiceAddr string
	userGrpcServer      string
	sid                 string
	downloadURL         string
	gatewayAddress      string
	gatewayPort         string
	tokenExpiration     time.Duration
	cache               bool
	jwtSecret           string
	stop                func() func(ctx context.Context) error
}

func New(ac *pkgconfig.AppConfig, grpcService *grpchandler.UserServiceServer) (s *Service) {
	s = &Service{
		ac:              ac,
		logger:          plog.NewDefaultLogger("user_bff", int8(ac.Log.Level)),
		sid:             xid.New().String(),
		tokenExpiration: 60 * 60 * 24 * 3 * time.Second,
		storageService:  setMinIOProvider(ac),
		dtmGrpcServer:   ac.Dtm.Addr(),
		downloadURL:     "/api/v1/storage/files/download",
		smtpClient:      setupSmtpProvider(ac),
		jwtSecret:       ac.SystemConfig.JwtSecret,
	}
	userCache, err := cache.NewUserCacheRedis(s.ac.Redis.Addr(), s.ac.Redis.Password, 0)
	if err != nil {
		panic(err)
	}

	s.stop = func() func(ctx context.Context) error {
		return func(ctx context.Context) error {
			if err := userCache.DeleteAllCache(ctx); err != nil {
				s.logger.Warn("delete all cache error", zap.Error(err))
			}
			return userCache.Close()
		}
	}

	s.setLoadSystem()
	s.setupRedisClient()
	s.userCache = userCache
	s.cache = s.ac.Cache.Enable
	s.userService = grpcService
	s.userLoginService = grpcService
	s.authService = grpcService
	return s
}

func (s *Service) HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) error {
	addr := conn.Target()
	switch serviceName {
	case "relation_service":
		s.relationServiceAddr = addr
		s.relationService = relationgrpcv1.NewUserRelationServiceClient(conn)
		s.dialogService = relationgrpcv1.NewDialogServiceClient(conn)
	case "push_service":
		s.pushService = pushv1.NewPushServiceClient(conn)
	case "msg_service":
		s.msgService = msggrpcv1.NewMsgServiceClient(conn)
	default:
		return nil
	}
	s.logger.Info("gRPC client service initialized", zap.String("service", serviceName), zap.String("addr", addr))
	return nil
}

func (s *Service) setupRedisClient() {
	//s.redisClient = cache.NewRedisClient(s.ac.Redis.Addr(), s.ac.Redis.Password)
}

func setMinIOProvider(ac *pkgconfig.AppConfig) storage.StorageProvider {
	var err error
	fmt.Println("ac.OSS => ", ac.OSS)
	sp, err := minio.NewMinIOStorage(ac.OSS.Addr(), ac.OSS.AccessKey, ac.OSS.SecretKey, ac.OSS.SSL)
	if err != nil {
		panic(err)
	}

	return sp
}

func setupSmtpProvider(ac *pkgconfig.AppConfig) email.EmailProvider {
	smtpStorage, err := smtp.NewSmtpStorage(ac.Email.SmtpServer, ac.Email.Port, ac.Email.Username, ac.Email.Password)
	if err != nil {
		panic(err)
	}
	return smtpStorage
}

func (s *Service) setLoadSystem() {
	s.gatewayAddress = s.ac.SystemConfig.GatewayAddress
	s.gatewayPort = s.ac.SystemConfig.GatewayPort
	if !s.ac.SystemConfig.Ssl {
		s.gatewayAddress = s.gatewayAddress + ":" + s.gatewayPort
	}
}

func (s *Service) Stop(ctx context.Context) error {
	return nil
}
