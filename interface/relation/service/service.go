package service

import (
	"context"
	"fmt"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/msg_queue"
	groupgrpcv1 "github.com/cossim/coss-server/service/group/api/v1"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	user "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Service struct
type Service struct {
	dialogClient            relationgrpcv1.DialogServiceClient
	groupRelationClient     relationgrpcv1.GroupRelationServiceClient
	userRelationClient      relationgrpcv1.UserRelationServiceClient
	userFriendRequestClient relationgrpcv1.UserFriendRequestServiceClient
	groupAnnouncementClient relationgrpcv1.GroupAnnouncementServiceClient
	groupJoinRequestClient  relationgrpcv1.GroupJoinRequestServiceClient
	userClient              user.UserServiceClient
	groupClient             groupgrpcv1.GroupServiceClient
	rabbitMQClient          *msg_queue.RabbitMQ

	logger    *zap.Logger
	sid       string
	discovery discovery.Registry
	ac        *pkgconfig.AppConfig

	dtmGrpcServer      string
	relationGrpcServer string
	dialogGrpcServer   string
}

func New(ac *pkgconfig.AppConfig) *Service {
	return &Service{
		rabbitMQClient: setRabbitMQProvider(ac),
		logger:         plog.NewDevLogger("relation_bff"),
		ac:             ac,
		sid:            xid.New().String(),
		dtmGrpcServer:  ac.Dtm.Addr(),
	}
}

func (s *Service) HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) error {
	switch serviceName {
	case "user_service":
		s.userClient = user.NewUserServiceClient(conn)
		s.logger.Info("gRPC client for user service initialized", zap.String("service", "user"), zap.String("addr", conn.Target()))
	case "relation_service":
		s.relationGrpcServer = conn.Target()
		s.dialogGrpcServer = conn.Target()
		s.userRelationClient = relationgrpcv1.NewUserRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "userRelation"), zap.String("addr", conn.Target()))

		s.userFriendRequestClient = relationgrpcv1.NewUserFriendRequestServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "userFriendRequestRelation"), zap.String("addr", conn.Target()))

		s.groupJoinRequestClient = relationgrpcv1.NewGroupJoinRequestServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "groupJoinRequestRelation"), zap.String("addr", conn.Target()))

		s.groupRelationClient = relationgrpcv1.NewGroupRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "groupRelation"), zap.String("addr", conn.Target()))

		s.groupAnnouncementClient = relationgrpcv1.NewGroupAnnouncementServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "groupAnnouncementRelation"), zap.String("addr", conn.Target()))

		s.dialogClient = relationgrpcv1.NewDialogServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "dialogRelation"), zap.String("addr", conn.Target()))
	case "group_service":
		s.groupClient = groupgrpcv1.NewGroupServiceClient(conn)
		s.logger.Info("gRPC client for group service initialized", zap.String("service", "group"), zap.String("addr", conn.Target()))
	}

	return nil
}

func (s *Service) Stop(ctx context.Context) error {
	return nil
}

func setRabbitMQProvider(ac *pkgconfig.AppConfig) *msg_queue.RabbitMQ {
	rmq, err := msg_queue.NewRabbitMQ(fmt.Sprintf("amqp://%s:%s@%s", ac.MessageQueue.Username, ac.MessageQueue.Password, ac.MessageQueue.Addr()))
	if err != nil {
		panic(err)
	}
	return rmq
}
