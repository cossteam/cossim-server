package service

import (
	"fmt"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/cossim/coss-server/pkg/msg_queue"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	user "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/redis/go-redis/v9"
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
	conf   *pkgconfig.AppConfig
	logger *zap.Logger

	discovery   discovery.Discovery
	userClient  user.UserServiceClient
	relClient   relationgrpcv1.UserRelationServiceClient
	redisClient *redis.Client
	//rabbitMQClient *msg_queue.RabbitMQ

	sid             string
	tokenExpiration time.Duration
}

func New(c *pkgconfig.AppConfig) (s *Service) {
	s = &Service{
		conf:            c,
		sid:             xid.New().String(),
		tokenExpiration: 60 * 60 * 24 * 3 * time.Second,
		//rabbitMQClient:  setRabbitMQProvider(c),
	}

	s.setupLogger()
	s.setupRedis()
	return s
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
	switch serviceName {
	case "user":
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		s.userClient = user.NewUserServiceClient(conn)
		s.logger.Info("gRPC client for user service initialized", zap.String("addr", conn.Target()))
	case "relation":
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}
		s.relClient = relationgrpcv1.NewUserRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("addr", conn.Target()))
	}

	return nil
}

func (s *Service) setupLogger() {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "user",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder, // 全路径编码器
	}

	// 设置日志级别
	atom := zap.NewAtomicLevelAt(zapcore.Level(s.conf.Log.V))
	config := zap.Config{
		Level:            atom,                                              // 日志级别
		Development:      true,                                              // 开发模式，堆栈跟踪
		Encoding:         s.conf.Log.Format,                                 // 输出格式 console 或 json
		EncoderConfig:    encoderConfig,                                     // 编码器配置
		InitialFields:    map[string]interface{}{"serviceName": "user-bff"}, // 初始化字段，如：添加一个服务器名称
		OutputPaths:      []string{"stdout"},                                // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		ErrorOutputPaths: []string{"stderr"},
	}
	// 构建日志
	var err error
	s.logger, err = config.Build()
	if err != nil {
		panic(fmt.Sprintf("log 初始化失败: %v", err))
	}
	s.logger.Info("log 初始化成功")
}

func (s *Service) setupRedis() {
	s.redisClient = redis.NewClient(&redis.Options{
		Addr:     s.conf.Redis.Addr(),
		Password: s.conf.Redis.Password, // no password set
		DB:       0,                     // use default DB
		//Protocol: cfg,
	})
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

func (s *Service) Ping() {
}

func setRabbitMQProvider(c *pkgconfig.AppConfig) *msg_queue.RabbitMQ {
	rmq, err := msg_queue.NewRabbitMQ(fmt.Sprintf("amqp://%s:%s@%s", c.MessageQueue.Username, c.MessageQueue.Password, c.MessageQueue.Addr()))
	if err != nil {
		panic(err)
	}
	return rmq
}
