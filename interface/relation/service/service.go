package service

import (
	"fmt"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/msg_queue"
	group "github.com/cossim/coss-server/service/group/api/v1"
	relation "github.com/cossim/coss-server/service/relation/api/v1"
	user "github.com/cossim/coss-server/service/user/api/v1"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Service struct
type Service struct {
	dialogClient        relation.DialogServiceClient
	groupRelationClient relation.GroupRelationServiceClient
	userRelationClient  relation.UserRelationServiceClient
	userClient          user.UserServiceClient
	groupClient         group.GroupServiceClient
	rabbitMQClient      *msg_queue.RabbitMQ
	logger              *zap.Logger

	dtmGrpcServer      string
	relationGrpcServer string
	dialogGrpcServer   string
}

func New(c *config.AppConfig) *Service {
	logger := setupLogger(c)

	dialogClient := setupDialogGRPCClient(c.Discovers["relation"].Addr)
	groupRelationClient := setupGROUPRelationGRPCClient(c.Discovers["relation"].Addr)
	userRelationClient := setupUserRelationGRPCClient(c.Discovers["relation"].Addr)
	userClient := setupUserGRPCClient(c.Discovers["user"].Addr)
	groupClient := setupGroupGRPCClient(c.Discovers["group"].Addr)
	rabbitMQClient := setRabbitMQProvider(c)

	return &Service{
		dialogClient:        dialogClient,
		groupRelationClient: groupRelationClient,
		userRelationClient:  userRelationClient,
		userClient:          userClient,
		groupClient:         groupClient,
		rabbitMQClient:      rabbitMQClient,
		logger:              logger,

		dtmGrpcServer:      c.Dtm.Addr,
		relationGrpcServer: c.Discovers["relation"].Addr,
		dialogGrpcServer:   c.Discovers["relation"].Addr,
	}
}

func setupLogger(c *config.AppConfig) *zap.Logger {
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

func setupGroupGRPCClient(addr string) group.GroupServiceClient {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return group.NewGroupServiceClient(conn)
}

func setupUserRelationGRPCClient(addr string) relation.UserRelationServiceClient {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return relation.NewUserRelationServiceClient(conn)
}

func setupGROUPRelationGRPCClient(addr string) relation.GroupRelationServiceClient {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return relation.NewGroupRelationServiceClient(conn)
}

func setupDialogGRPCClient(addr string) relation.DialogServiceClient {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return relation.NewDialogServiceClient(conn)
}

func setRabbitMQProvider(c *config.AppConfig) *msg_queue.RabbitMQ {
	rmq, err := msg_queue.NewRabbitMQ(fmt.Sprintf("amqp://%s:%s@%s", c.MessageQueue.Username, c.MessageQueue.Password, c.MessageQueue.Addr))
	if err != nil {
		panic(err)
	}
	return rmq
}

func setupUserGRPCClient(addr string) user.UserServiceClient {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return user.NewUserServiceClient(conn)
}
