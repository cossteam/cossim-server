package service

import (
	storagev1 "github.com/cossim/coss-server/internal/storage/api/grpc/v1"
	grpchandler "github.com/cossim/coss-server/internal/storage/interface/grpc"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/storage"
	"github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Service struct {
	ac     *pkgconfig.AppConfig
	logger *zap.Logger

	userService    usergrpcv1.UserServiceClient
	sp             storage.StorageProvider
	storageService storagev1.StorageServiceServer

	sid            string
	downloadURL    string
	gatewayAddress string
	gatewayPort    string
	dis            bool
	cache          bool
}

func New(ac *pkgconfig.AppConfig, grpcService *grpchandler.Handler) *Service {
	logger := setupLogger(ac)
	svc := &Service{
		downloadURL: "/api/v1/storage/files/download",
		sp:          setMinIOProvider(ac),
		logger:      logger,
		ac:          ac,
		sid:         xid.New().String(),
		dis:         false,
	}
	svc.setLoadSystem()
	svc.storageService = grpcService
	return svc
}

func setMinIOProvider(ac *pkgconfig.AppConfig) storage.StorageProvider {
	var err error
	sp, err := minio.NewMinIOStorage(ac.OSS.Addr(), ac.OSS.AccessKey, ac.OSS.SecretKey, ac.OSS.SSL)
	if err != nil {
		panic(err)
	}

	return sp
}

func (s *Service) HandlerGrpcClient(serviceName string, addr string) error {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}

	switch serviceName {
	case "user_service":
		s.userService = usergrpcv1.NewUserServiceClient(conn)
		s.logger.Info("gRPC client for user service initialized", zap.String("service", "user"), zap.String("addr", conn.Target()))
	}

	return nil
}

func setupLogger(c *pkgconfig.AppConfig) *zap.Logger {
	return plog.NewDefaultLogger("group_bff", int8(c.Log.Level))
}

func (s *Service) setLoadSystem() {
	env := s.ac.SystemConfig.Environment
	if env == "" {
		env = "dev"
	}

	switch env {
	case "prod":
		gatewayAdd := s.ac.SystemConfig.GatewayAddress
		if gatewayAdd == "" {
			gatewayAdd = "43.229.28.107"
		}

		s.gatewayAddress = gatewayAdd

		gatewayPo := s.ac.SystemConfig.GatewayPort
		if gatewayPo == "" {
			gatewayPo = "8080"
		}
		s.gatewayPort = gatewayPo
	default:
		gatewayAdd := s.ac.SystemConfig.GatewayAddressDev
		if gatewayAdd == "" {
			gatewayAdd = "127.0.0.1"
		}

		s.gatewayAddress = gatewayAdd

		gatewayPo := s.ac.SystemConfig.GatewayPortDev
		if gatewayPo == "" {
			gatewayPo = "8080"
		}
		s.gatewayPort = gatewayPo
	}

	if !s.ac.SystemConfig.Ssl {
		s.gatewayAddress = s.gatewayAddress + ":" + s.gatewayPort
	}
}
