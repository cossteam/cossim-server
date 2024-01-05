package http

import (
	"fmt"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/http/middleware"
	userApi "github.com/cossim/coss-server/services/user/api/v1"

	msgApi "github.com/cossim/coss-server/services/msg/api/v1"
	relationApi "github.com/cossim/coss-server/services/relation/api/v1"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"time"
)

var (
	msgClient      msgApi.MsgServiceClient
	relationClient relationApi.RelationServiceClient
	userClient     userApi.UserServiceClient
	cfg            *config.AppConfig
	logger         *zap.Logger
)

func Init(c *config.AppConfig) {
	cfg = c
	setupLogger()
	setupMsgGRPCClient()
	setupUserGRPCClient()
	setupRelationGRPCClient()
	setupGin()
}

func setupRelationGRPCClient() {
	var err error
	relationConn, err := grpc.Dial(cfg.Discovers["relation"].Addr, grpc.WithInsecure())
	if err != nil {
		logger.Fatal("Failed to connect to gRPC server", zap.Error(err))
	}

	relationClient = relationApi.NewRelationServiceClient(relationConn)
}

func setupMsgGRPCClient() {
	var err error
	msgConn, err := grpc.Dial(cfg.Discovers["msg"].Addr, grpc.WithInsecure())
	if err != nil {
		logger.Fatal("Failed to connect to gRPC server", zap.Error(err))
	}

	msgClient = msgApi.NewMsgServiceClient(msgConn)
}
func setupUserGRPCClient() {
	var err error
	userConn, err := grpc.Dial(cfg.Discovers["user"].Addr, grpc.WithInsecure())
	if err != nil {
		logger.Fatal("Failed to connect to gRPC server", zap.Error(err))
	}

	userClient = userApi.NewUserServiceClient(userConn)
}

func setupLogger() {
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
	atom := zap.NewAtomicLevelAt(zapcore.Level(cfg.Log.V))
	config := zap.Config{
		Level:            atom,                                             // 日志级别
		Development:      true,                                             // 开发模式，堆栈跟踪
		Encoding:         cfg.Log.Format,                                   // 输出格式 console 或 json
		EncoderConfig:    encoderConfig,                                    // 编码器配置
		InitialFields:    map[string]interface{}{"serviceName": "msg-bff"}, // 初始化字段，如：添加一个服务器名称
		OutputPaths:      []string{"stdout"},                               // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		ErrorOutputPaths: []string{"stderr"},
	}
	// 构建日志
	var err error
	logger, err = config.Build()
	if err != nil {
		panic(fmt.Sprintf("log 初始化失败: %v", err))
	}
	logger.Info("log 初始化成功")
	logger.Info("无法获取网址",
		zap.String("url", "http://www.baidu.com"),
		zap.Int("attempt", 3),
		zap.Duration("backoff", time.Second),
	)
}

func setupGin() {
	if cfg == nil {
		panic("Config not initialized")
	}

	// 初始化 gin engine
	engine := gin.New()

	// 添加一些中间件或其他配置
	engine.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(logger), middleware.RecoveryMiddleware())

	// 设置路由
	route(engine)

	// 启动 Gin 服务器
	go func() {
		if err := engine.Run(cfg.HTTP.Addr); err != nil {
			logger.Fatal("Failed to start Gin server", zap.Error(err))
		}
	}()
}

// @title Swagger Example API
func route(engine *gin.Engine) {
	u := engine.Group("/api/v1/msg")
	//u.Use(middleware.AuthMiddleware())
	u.GET("/ws", ws)
	u.POST("/send/user", middleware.AuthMiddleware(), sendUserMsg)
	u.POST("/send/group", middleware.AuthMiddleware(), sendGroupMsg)
	u.GET("/list/user", middleware.AuthMiddleware(), getUserMsgList)
	u.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("msg")))
}
