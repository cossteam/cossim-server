package service

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/push/connect"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/db"
	"github.com/cossim/coss-server/pkg/encryption"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/msg_queue"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"strconv"
)

type Service struct {
	logger          *zap.Logger
	rabbitMQClient  *msg_queue.RabbitMQ
	relationService relationgrpcv1.UserRelationServiceClient
	redisClient     *cache.RedisClient
	ac              *pkgconfig.AppConfig
	enc             encryption.Encryptor
	Buckets         map[constants.DriverType]*connect.Bucket
	db              *gorm.DB
}

var wsRid int64 = 0 //全局客户端id

func New(ac *pkgconfig.AppConfig) *Service {
	mqClient, err := msg_queue.NewRabbitMQ(fmt.Sprintf("amqp://%s:%s@%s", ac.MessageQueue.Username, ac.MessageQueue.Password, ac.MessageQueue.Addr()))
	if err != nil {
		panic(err)
	}

	mysql, err := db.NewMySQL(ac.MySQL.Address, strconv.Itoa(ac.MySQL.Port), ac.MySQL.Username, ac.MySQL.Password, ac.MySQL.Database, int64(ac.Log.Level), ac.MySQL.Opts)
	if err != nil {
		panic(err)
	}

	dbConn, err := mysql.GetConnection()
	if err != nil {
		panic(err)
	}

	s := &Service{
		ac:             ac,
		logger:         plog.NewDefaultLogger("msg_bff", int8(ac.Log.Level)),
		rabbitMQClient: mqClient,
		redisClient:    setupRedis(ac),
		Buckets:        make(map[constants.DriverType]*connect.Bucket),
		db:             dbConn,
	}

	s.setupEncryption(ac)
	for _, driverType := range constants.GetDriverTypeList() {
		s.Buckets[driverType] = connect.NewBucket()
	}

	return s
}

func (s *Service) Init(ac *pkgconfig.AppConfig) {
	*s = *New(ac)
}

func (s *Service) setupEncryption(ac *pkgconfig.AppConfig) {
	s.enc = encryption.NewEncryptor(
		[]byte(ac.Encryption.Passphrase),
		ac.Encryption.Name,
		ac.Encryption.Email,
		ac.Encryption.RsaBits,
		ac.Encryption.Enable,
	)
}
func setupRedis(ac *pkgconfig.AppConfig) *cache.RedisClient {
	return cache.NewRedisClient(ac.Redis.Addr(), ac.Redis.Password)
}

func (s *Service) Stop(ctx context.Context) error {
	s.Buckets = make(map[constants.DriverType]*connect.Bucket)
	s.rabbitMQClient.Close()
	return nil
}

func (s *Service) HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) error {
	addr := conn.Target()
	switch serviceName {
	case "relation_service":
		s.relationService = relationgrpcv1.NewUserRelationServiceClient(conn)
	default:
		return nil
	}
	s.logger.Info("gRPC client service initialized", zap.String("service", serviceName), zap.String("addr", addr))
	return nil
}
