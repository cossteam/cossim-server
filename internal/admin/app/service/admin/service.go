package admin

import (
	"context"
	"github.com/cossim/coss-server/internal/admin/domain/service"
	"github.com/cossim/coss-server/internal/admin/infra/persistence"
	groupApi "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	pushv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/storage"
	"github.com/cossim/coss-server/pkg/storage/minio"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type Service interface {
	AdminService
	Init(db *gorm.DB, cfg *pkgconfig.AppConfig) error
	HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) error
	Stop(ctx context.Context) error
}

type ServiceImpl struct {
	relationUserService   relationgrpcv1.UserRelationServiceClient
	relationDialogService relationgrpcv1.DialogServiceClient
	userService           usergrpcv1.UserServiceClient
	pushService           pushv1.PushServiceClient
	dtmGrpcServer         string
	userServiceAddr       string
	logger                *zap.Logger
	relationGroupService  relationgrpcv1.GroupRelationServiceClient
	groupService          groupApi.GroupServiceClient
	msgService            msggrpcv1.MsgServiceClient
	ad                    service.AdminDomain
	gatewayPort           string
	gatewayAddress        string
	downloadURL           string
	sp                    storage.StorageProvider
	ac                    *pkgconfig.AppConfig
}

func (s *ServiceImpl) Stop(ctx context.Context) error {
	return nil
}

func NewService(dtmGrpcServer string, logger *zap.Logger) Service {
	return &ServiceImpl{
		dtmGrpcServer: dtmGrpcServer,
		logger:        logger,
	}
}

func (s *ServiceImpl) Init(db *gorm.DB, cfg *pkgconfig.AppConfig) error {
	repo := persistence.NewRepositories(db)
	err := repo.Automigrate()
	if err != nil {
		return err
	}

	s.ad = service.NewAdminDomain(db, cfg, repo)
	s.ac = cfg
	s.downloadURL = constants.DownLoadAddress
	s.setLoadSystem()
	s.sp = setMinIOProvider(cfg)
	return nil
}

func (s *ServiceImpl) HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) error {
	addr := conn.Target()
	switch serviceName {
	case "user_service":
		s.userServiceAddr = addr
		s.userService = usergrpcv1.NewUserServiceClient(conn)
		err := s.InitAdmin()
		if err != nil {
			return err
		}
	case "relation_service":
		s.relationUserService = relationgrpcv1.NewUserRelationServiceClient(conn)
		s.relationDialogService = relationgrpcv1.NewDialogServiceClient(conn)
	case "msg_service":
		s.msgService = msggrpcv1.NewMsgServiceClient(conn)
	case "push_service":
		s.pushService = pushv1.NewPushServiceClient(conn)
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
	//sp, err := minio.NewMinIOStorage(ac.OSS["minio"].Addr(), ac.OSS["minio"].AccessKey, ac.OSS["minio"].SecretKey, ac.OSS["minio"].SSL)
	sp, err := minio.NewMinIOStorage(ac.OSS.Addr(), ac.OSS.AccessKey, ac.OSS.SecretKey, ac.OSS.SSL)
	if err != nil {
		panic(err)
	}

	return sp
}
