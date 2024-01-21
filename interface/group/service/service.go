package service

import (
	"fmt"
	"github.com/cossim/coss-server/pkg/config"
	groupgrpcv1 "github.com/cossim/coss-server/service/group/api/v1"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	usergrpcv1 "github.com/cossim/coss-server/service/user/api/v1"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Service struct
type Service struct {
	groupClient          groupgrpcv1.GroupServiceClient
	relationDialogClient relationgrpcv1.DialogServiceClient
	relationGroupClient  relationgrpcv1.GroupRelationServiceClient
	relationUserClient   relationgrpcv1.UserRelationServiceClient
	userClient           usergrpcv1.UserServiceClient

	logger *zap.Logger

	dtmGrpcServer      string
	relationGrpcServer string
	dialogGrpcServer   string
	groupGrpcServer    string
}

func New(c *config.AppConfig) (s *Service) {
	logger := setupLogger(c)

	relationGroupClient := setupRelationGroupGRPCClient(c.Discovers["relation"].Addr)
	relationDialogClient := setupRelationDialogGRPCClient(c.Discovers["relation"].Addr)
	relationUserClient := setupRelationUserGRPCClient(c.Discovers["relation"].Addr)
	groupClient := setupGroupGRPCClient(c.Discovers["group"].Addr)
	userClient := setupUserGRPCClient(c.Discovers["user"].Addr)

	return &Service{
		relationGroupClient:  relationGroupClient,
		relationDialogClient: relationDialogClient,
		relationUserClient:   relationUserClient,
		groupClient:          groupClient,
		userClient:           userClient,

		logger: logger,

		dtmGrpcServer:      c.Dtm.Addr,
		relationGrpcServer: c.Discovers["relation"].Addr,
		dialogGrpcServer:   c.Discovers["relation"].Addr,
		groupGrpcServer:    c.Discovers["group"].Addr,
	}
}

func setupLogger(c *config.AppConfig) *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "group",
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
		Level:            atom,                                                   // 日志级别
		Development:      true,                                                   // 开发模式，堆栈跟踪
		Encoding:         c.Log.Format,                                           // 输出格式 console 或 json
		EncoderConfig:    encoderConfig,                                          // 编码器配置
		InitialFields:    map[string]interface{}{"serviceName": "group_bff_svc"}, // 初始化字段，如：添加一个服务器名称
		OutputPaths:      []string{"stdout"},                                     // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
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

func setupRelationGroupGRPCClient(addr string) relationgrpcv1.GroupRelationServiceClient {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return relationgrpcv1.NewGroupRelationServiceClient(conn)
}

func setupGroupGRPCClient(addr string) groupgrpcv1.GroupServiceClient {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return groupgrpcv1.NewGroupServiceClient(conn)
}

func setupRelationUserGRPCClient(addr string) relationgrpcv1.UserRelationServiceClient {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return relationgrpcv1.NewUserRelationServiceClient(conn)
}

func setupUserGRPCClient(addr string) usergrpcv1.UserServiceClient {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return usergrpcv1.NewUserServiceClient(conn)
}

func setupRelationDialogGRPCClient(addr string) relationgrpcv1.DialogServiceClient {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	return relationgrpcv1.NewDialogServiceClient(conn)
}
