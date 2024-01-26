package service

import (
	"fmt"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/cossim/coss-server/pkg/msg_queue"
	groupgrpcv1 "github.com/cossim/coss-server/service/group/api/v1"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	user "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"sync"
	"time"
)

// Service struct
type Service struct {
	dialogClient            relationgrpcv1.DialogServiceClient
	groupRelationClient     relationgrpcv1.GroupRelationServiceClient
	userRelationClient      relationgrpcv1.UserRelationServiceClient
	userFriendRequestClient relationgrpcv1.UserFriendRequestServiceClient
	userClient              user.UserServiceClient
	groupClient             groupgrpcv1.GroupServiceClient
	rabbitMQClient          *msg_queue.RabbitMQ

	logger    *zap.Logger
	sid       string
	discovery discovery.Discovery
	conf      *pkgconfig.AppConfig

	dtmGrpcServer      string
	relationGrpcServer string
	dialogGrpcServer   string
}

func New(c *pkgconfig.AppConfig) *Service {
	logger := setupLogger(c)
	rabbitMQClient := setRabbitMQProvider(c)

	return &Service{
		//dialogClient:        dialogClient,
		//groupRelationClient: groupRelationClient,
		//userRelationClient:  userRelationClient,
		//userClient:          userClient,
		//groupClient:         groupClient,

		rabbitMQClient: rabbitMQClient,
		logger:         logger,
		conf:           c,

		sid: xid.New().String(),

		dtmGrpcServer: c.Dtm.Addr(),
		//relationGrpcServer: c.Discovers["relation"].Addr(),
		//dialogGrpcServer:   c.Discovers["relation"].Addr(),
	}
}

func (s *Service) Start(discover bool) {
	if discover {
		d, err := discovery.NewConsulRegistry(s.conf.Register.Addr())
		if err != nil {
			panic(err)
		}
		s.discovery = d
		if err = s.discovery.RegisterHTTP(s.conf.Register.Name, s.conf.HTTP.Addr(), s.sid); err != nil {
			panic(err)
		}
		log.Printf("Service registration successful ServiceName: %s  Addr: %s  ID: %s", s.conf.Register.Name, s.conf.HTTP.Addr(), s.sid)
		go s.discover()
	} else {
		s.direct()
	}
}

func (s *Service) discover() {
	var wg sync.WaitGroup
	type serviceInfo struct {
		ServiceName string
		Addr        string
	}
	ch := make(chan serviceInfo)

	for serviceName, c := range s.conf.Discovers {
		wg.Add(1)
		go func(serviceName string, c pkgconfig.ServiceConfig) {
			defer wg.Done()
			for {
				addr, err := s.discovery.Discover(c.Name)
				if err != nil {
					log.Printf("Service discovery failed ServiceName: %s %v\n", c.Name, err)
					time.Sleep(5 * time.Second)
					continue
				}
				log.Printf("Service discovery successful ServiceName: %s  Addr: %s\n", c.Name, addr)

				ch <- serviceInfo{ServiceName: serviceName, Addr: addr}
				break
			}
		}(serviceName, c)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for info := range ch {
		if err := s.handlerGrpcClient(info.ServiceName, info.Addr); err != nil {
			log.Printf("Failed to initialize gRPC client for service: %s, Error: %v\n", info.ServiceName, err)
		}
	}
}

func (s *Service) direct() {
	for serviceName, _ := range s.conf.Discovers {
		if err := s.handlerGrpcClient(serviceName, s.conf.Discovers[serviceName].Addr()); err != nil {
			panic(err)
		}
	}
}

func (s *Service) handlerGrpcClient(serviceName string, addr string) error {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	switch serviceName {
	case "user":
		s.userClient = user.NewUserServiceClient(conn)
		s.logger.Info("gRPC client for user service initialized", zap.String("service", "user"), zap.String("addr", conn.Target()))
	case "relation":
		s.relationGrpcServer = addr
		s.dialogGrpcServer = addr

		s.userRelationClient = relationgrpcv1.NewUserRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "userRelation"), zap.String("addr", conn.Target()))

		s.userFriendRequestClient = relationgrpcv1.NewUserFriendRequestServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "userFriendRequestRelation"), zap.String("addr", conn.Target()))

		s.groupRelationClient = relationgrpcv1.NewGroupRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "groupRelation"), zap.String("addr", conn.Target()))

		s.dialogClient = relationgrpcv1.NewDialogServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "dialogRelation"), zap.String("addr", conn.Target()))
	case "group":
		s.groupClient = groupgrpcv1.NewGroupServiceClient(conn)
		s.logger.Info("gRPC client for group service initialized", zap.String("service", "group"), zap.String("addr", conn.Target()))
	}

	return nil
}

func (s *Service) Close(discover bool) error {
	if !discover {
		return nil
	}
	if err := s.discovery.Cancel(s.sid); err != nil {
		log.Printf("Failed to cancel service registration: %v", err)
		return err
	}
	log.Printf("Service registration canceled ServiceName: %s  Addr: %s  ID: %s", s.conf.Register.Name, s.conf.GRPC.Addr(), s.sid)
	return nil
}

func setupLogger(c *pkgconfig.AppConfig) *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "relation",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder, // 全路径编码器
	}

	// 设置日志级别
	atom := zap.NewAtomicLevelAt(zapcore.Level(c.Log.V))
	config := zap.Config{
		Level:            atom,                                                  // 日志级别
		Development:      true,                                                  // 开发模式，堆栈跟踪
		Encoding:         c.Log.Format,                                          // 输出格式 console 或 json
		EncoderConfig:    encoderConfig,                                         // 编码器配置
		InitialFields:    map[string]interface{}{"serviceName": "relation_bff"}, // 初始化字段，如：添加一个服务器名称
		OutputPaths:      []string{"stdout"},                                    // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		ErrorOutputPaths: []string{"stderr"},
	}
	// 构建日志
	var err error
	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("logger 初始化失败: %v", err))
	}
	logger.Info("logger 初始化成功")
	return logger
}

func setRabbitMQProvider(c *pkgconfig.AppConfig) *msg_queue.RabbitMQ {
	rmq, err := msg_queue.NewRabbitMQ(fmt.Sprintf("amqp://%s:%s@%s", c.MessageQueue.Username, c.MessageQueue.Password, c.MessageQueue.Addr()))
	if err != nil {
		panic(err)
	}
	return rmq
}
