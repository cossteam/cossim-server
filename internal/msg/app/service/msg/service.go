package msg

import (
	"context"
	groupApi "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	"github.com/cossim/coss-server/internal/msg/domain/service"
	"github.com/cossim/coss-server/internal/msg/infra/persistence"
	pushv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type Service interface {
	UserService
	GroupService
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
	logger                *zap.Logger
	relationGroupService  relationgrpcv1.GroupRelationServiceClient
	groupService          groupApi.GroupServiceClient

	ud   service.UserMsgDomain
	gmd  service.GroupMsgDomain
	gmrd service.GroupMsgReadDomain
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
	s.ud = service.NewUserMsgDomain(db, cfg, repo)
	s.gmd = service.NewGroupMsgDomain(db, cfg, repo)
	s.gmrd = service.NewGroupMsgReadDomain(db, cfg, repo)
	return nil
}

func (s *ServiceImpl) HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) error {
	addr := conn.Target()
	switch serviceName {
	case "user_service":
		s.userService = usergrpcv1.NewUserServiceClient(conn)
	case "relation_service":
		s.relationUserService = relationgrpcv1.NewUserRelationServiceClient(conn)
		s.relationGroupService = relationgrpcv1.NewGroupRelationServiceClient(conn)
		s.relationDialogService = relationgrpcv1.NewDialogServiceClient(conn)
	case "group_service":
		s.groupService = groupApi.NewGroupServiceClient(conn)
	case "push_service":
		s.pushService = pushv1.NewPushServiceClient(conn)
	default:
		return nil
	}
	s.logger.Info("gRPC client service initialized", zap.String("service", serviceName), zap.String("addr", addr))
	return nil
}
