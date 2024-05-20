package storage

import (
	"context"
	"github.com/cossim/coss-server/internal/storage/domain/service"
	"github.com/cossim/coss-server/internal/storage/infra/persistence"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/storage"
	"github.com/cossim/coss-server/pkg/storage/minio"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type Service interface {
	StorageService
	Init(db *gorm.DB, cfg *pkgconfig.AppConfig) error
	HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) error
	Stop(ctx context.Context) error
}

type ServiceImpl struct {
	logger      *zap.Logger
	userService usergrpcv1.UserServiceClient
	sd          service.StorageDomain
	sp          storage.StorageProvider
	ac          *pkgconfig.AppConfig

	downloadURL    string
	gatewayAddress string
	gatewayPort    string
}

func (s *ServiceImpl) Stop(ctx context.Context) error {
	return nil
}

func NewService(logger *zap.Logger) Service {
	return &ServiceImpl{
		logger: logger,
	}
}

func (s *ServiceImpl) Init(db *gorm.DB, cfg *pkgconfig.AppConfig) error {
	s.ac = cfg
	repo := persistence.NewRepositories(db)
	err := repo.Automigrate()
	if err != nil {
		return err
	}

	s.sd = service.NewStorageDomain(db, cfg, repo)
	s.sp = setMinIOProvider(cfg)
	s.downloadURL = "/api/v1/storage/files/download"
	s.setLoadSystem()

	return nil
}

func (s *ServiceImpl) HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) error {
	addr := conn.Target()
	switch serviceName {
	case "user_service":
		s.userService = usergrpcv1.NewUserServiceClient(conn)
	default:
		return nil
	}
	s.logger.Info("gRPC client service initialized", zap.String("service", serviceName), zap.String("addr", addr))
	return nil
}

func (s *ServiceImpl) setLoadSystem() {
	s.gatewayAddress = s.ac.SystemConfig.GatewayAddress
	s.gatewayPort = s.ac.SystemConfig.GatewayPort
	if !s.ac.SystemConfig.Ssl {
		s.gatewayAddress = s.gatewayAddress + ":" + s.gatewayPort
	}
}

func setMinIOProvider(ac *pkgconfig.AppConfig) storage.StorageProvider {
	var err error
	sp, err := minio.NewMinIOStorage(ac.OSS.Addr(), ac.OSS.AccessKey, ac.OSS.SecretKey, ac.OSS.SSL)
	if err != nil {
		panic(err)
	}

	return sp
}
