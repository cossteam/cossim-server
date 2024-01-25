package service

import (
	"fmt"
	config2 "github.com/cossim/coss-server/interface/msg/config"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/msg_queue"
	groupgrpcv1 "github.com/cossim/coss-server/service/group/api/v1"
	msggrpcv1 "github.com/cossim/coss-server/service/msg/api/v1"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	usergrpcv1 "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/goccy/go-json"
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

var (
	rabbitMQClient *msg_queue.RabbitMQ

	wsRid   int64 = 0 //全局客户端id
	wsMutex       = sync.Mutex{}
	Enc     encryption.Encryptor
	pool    = make(map[string]map[string][]*client)
)

// Service struct
type Service struct {
	relationUserClient   relationgrpcv1.UserRelationServiceClient
	relationGroupClient  relationgrpcv1.GroupRelationServiceClient
	relationDialogClient relationgrpcv1.DialogServiceClient
	userClient           usergrpcv1.UserServiceClient
	groupClient          groupgrpcv1.GroupServiceClient
	msgClient            msggrpcv1.MsgServiceClient
	redisClient          *redis.Client

	//mqClient *msg_queue.RabbitMQ

	logger    *zap.Logger
	sid       string
	discovery discovery.Discovery
	conf      *pkgconfig.AppConfig
	//Enc       encryption.Encryptor

	//pool  map[string]map[string][]*client
}

func New(c *pkgconfig.AppConfig) (s *Service) {
	mqClient, err := msg_queue.NewRabbitMQ(fmt.Sprintf("amqp://%s:%s@%s", c.MessageQueue.Username, c.MessageQueue.Password, c.MessageQueue.Addr()))
	if err != nil {
		panic(err)
	}
	rabbitMQClient = mqClient
	return &Service{
		conf:        c,
		logger:      setupLogger(c),
		sid:         xid.New().String(),
		redisClient: setupRedis(c),
		//Enc:       setupEncryption(),
		//rabbitMQClient: mqClient,
		//pool:     make(map[string]map[string][]*client),
		//mqClient: mqClient,
		//pool:     make(map[string]map[string][]*client),
	}
}

func setupLogger(c *pkgconfig.AppConfig) *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
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
		Level:            atom,                                             // 日志级别
		Development:      true,                                             // 开发模式，堆栈跟踪
		Encoding:         c.Log.Format,                                     // 输出格式 console 或 json
		EncoderConfig:    encoderConfig,                                    // 编码器配置
		InitialFields:    map[string]interface{}{"serviceName": "msg-bff"}, // 初始化字段，如：添加一个服务器名称
		OutputPaths:      []string{"stdout"},                               // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
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

func (s *Service) Start(discover bool) {
	// 监听服务消息队列
	go s.ListenQueue()

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

func (s *Service) ListenQueue() {
	if rabbitMQClient.GetConnection().IsClosed() {
		panic("mqClient Connection is closed")
	}
	msgs, err := rabbitMQClient.ConsumeServiceMessages(msg_queue.MsgService, msg_queue.Service_Exchange)
	if err != nil {
		panic(err)
	}
	go func() {
		//监听队列
		for msg := range msgs {
			var msg_query msg_queue.ServiceQueueMsg
			err := json.Unmarshal(msg.Body, &msg_query)
			if err != nil {
				fmt.Println("解析消息失败：", err)
				return
			}
			switch msg_query.Action {
			case msg_queue.SendMessage:
				mmap, ok := msg_query.Data.(map[string]interface{})
				if !ok {
					fmt.Println("解析消息失败：")
					return
				}
				//map解析成结构体
				jsonData, err := json.Marshal(mmap)
				if err != nil {
					fmt.Println("Failed to marshal map to JSON:", err)
					return
				}
				var wsm config2.WsMsg
				err = json.Unmarshal(jsonData, &wsm)
				if err != nil {
					fmt.Println("解析消息失败：", err)
					return
				}
				s.SendMsg(wsm.Uid, wsm.Event, wsm.Data)
			}
		}
	}()
}

func (s *Service) Stop(discover bool) {
	//关闭队列
	rabbitMQClient.Close()
	if discover {
		if err := s.discovery.Cancel(s.sid); err != nil {
			log.Printf("Failed to cancel service registration: %v", err)
			return
		}
		log.Printf("Service registration canceled ServiceName: %s  Addr: %s  ID: %s", s.conf.Register.Name, s.conf.HTTP.Addr(), s.sid)
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
		s.userClient = usergrpcv1.NewUserServiceClient(conn)
		s.logger.Info("gRPC client for user service initialized", zap.String("service", "user"), zap.String("addr", conn.Target()))
	case "relation":
		s.relationUserClient = relationgrpcv1.NewUserRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "userRelation"), zap.String("addr", conn.Target()))

		s.relationGroupClient = relationgrpcv1.NewGroupRelationServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "groupRelation"), zap.String("addr", conn.Target()))

		s.relationDialogClient = relationgrpcv1.NewDialogServiceClient(conn)
		s.logger.Info("gRPC client for relation service initialized", zap.String("service", "dialogRelation"), zap.String("addr", conn.Target()))
	case "group":
		s.groupClient = groupgrpcv1.NewGroupServiceClient(conn)
		s.logger.Info("gRPC client for group service initialized", zap.String("service", "group"), zap.String("addr", conn.Target()))
	case "msg":
		s.msgClient = msggrpcv1.NewMsgServiceClient(conn)
		s.logger.Info("gRPC client for group service initialized", zap.String("service", "msg"), zap.String("addr", conn.Target()))
	}

	return nil
}

func setupRedis(cfg *pkgconfig.AppConfig) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password, // no password set
		DB:       0,                  // use default DB
		//Protocol: cfg,
	})
}
