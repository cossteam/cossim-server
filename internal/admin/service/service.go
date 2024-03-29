package service

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/admin/infrastructure/persistence"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	user "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/msg_queue"
	"github.com/cossim/coss-server/pkg/storage"
	"github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"strconv"
)

type Service struct {
	ac             *pkgconfig.AppConfig
	logger         *zap.Logger
	repo           *persistence.Repositories
	userClient     user.UserServiceClient
	relationClient relationgrpcv1.UserRelationServiceClient
	msgClient      msggrpcv1.MsgServiceClient
	rabbitMQClient *msg_queue.RabbitMQ
	sp             storage.StorageProvider

	dtmGrpcServer       string
	relationServiceAddr string
	userServiceAddr     string
	msgServiceAddr      string
	gatewayAddress      string
	gatewayPort         string
	sid                 string
	downloadURL         string
}

func New(ac *pkgconfig.AppConfig) (s *Service) {
	s = &Service{
		ac:             ac,
		logger:         plog.NewDefaultLogger("admin_bff", int8(ac.Log.Level)),
		sid:            xid.New().String(),
		rabbitMQClient: setRabbitMQProvider(ac),
		dtmGrpcServer:  ac.Dtm.Addr(),
		sp:             setMinIOProvider(ac),
		downloadURL:    "/api/v1/storage/files/download",
	}
	s.setLoadSystem()
	s.setupDBConn()
	return s
}

func (s *Service) HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) error {
	addr := conn.Target()
	switch serviceName {
	case "user_service":
		s.userServiceAddr = addr
		s.userClient = user.NewUserServiceClient(conn)
		err := s.InitAdmin()
		if err != nil {
			return nil
		}
	case "relation_service":
		s.relationServiceAddr = addr
		s.relationClient = relationgrpcv1.NewUserRelationServiceClient(conn)
	case "msg_service":
		s.msgServiceAddr = addr
		s.msgClient = msggrpcv1.NewMsgServiceClient(conn)
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

func (s *Service) Stop(ctx context.Context) error {
	return nil
}

func (s *Service) setupDBConn() {
	mysql, err := db.NewMySQL(s.ac.MySQL.Address, strconv.Itoa(s.ac.MySQL.Port), s.ac.MySQL.Username, s.ac.MySQL.Password, s.ac.MySQL.Database, int64(s.ac.Log.Level), s.ac.MySQL.Opts)
	if err != nil {
		panic(err)
	}

	dbConn, err := mysql.GetConnection()
	if err != nil {
		panic(err)
	}

	infra := persistence.NewRepositories(dbConn)
	if err = infra.Automigrate(); err != nil {
		panic(err)
	}
	s.repo = infra
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

func setMinIOProvider(ac *pkgconfig.AppConfig) storage.StorageProvider {
	var err error
	//sp, err := minio.NewMinIOStorage(ac.OSS["minio"].Addr(), ac.OSS["minio"].AccessKey, ac.OSS["minio"].SecretKey, ac.OSS["minio"].SSL)
	sp, err := minio.NewMinIOStorage(ac.OSS.Addr(), ac.OSS.AccessKey, ac.OSS.SecretKey, ac.OSS.SSL)
	if err != nil {
		panic(err)
	}

	return sp
}
