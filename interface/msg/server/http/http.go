package http

import (
	"fmt"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/http/middleware"
	"github.com/cossim/coss-server/pkg/msg_queue"
	groupApi "github.com/cossim/coss-server/service/group/api/v1"
	msgApi "github.com/cossim/coss-server/service/msg/api/v1"
	relationApi "github.com/cossim/coss-server/service/relation/api/v1"
	userApi "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"time"
)

var (
	msgClient       msgApi.MsgServiceClient
	relationClient  relationApi.UserRelationServiceClient
	redisClient     *redis.Client
	userGroupClient relationApi.GroupRelationServiceClient
	groupClient     groupApi.GroupServiceClient
	rabbitMQClient  *msg_queue.RabbitMQ
	userClient      userApi.UserServiceClient
	dialogClient    relationApi.DialogServiceClient
	cfg             *config.AppConfig
	logger          *zap.Logger
)

func Init(c *config.AppConfig) {
	cfg = c
	setupLogger()
	setupDialogGRPCClient()
	setupMsgGRPCClient()
	setupUserGRPCClient()
	setupRedis()
	setupGroupGRPCClient()
	setRabbitMQProvider()
	setupRelationGRPCClient()
	setupGin()
}
func setupRedis() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password, // no password set
		DB:       0,                  // use default DB
		//Protocol: cfg,
	})
	redisClient = rdb
}
func setRabbitMQProvider() {
	rmq, err := msg_queue.NewRabbitMQ(fmt.Sprintf("amqp://%s:%s@%s", cfg.MessageQueue.Username, cfg.MessageQueue.Password, cfg.MessageQueue.Addr))
	fmt.Println("cfg.MessageQueue => ", cfg.MessageQueue)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	rabbitMQClient = rmq
}
func setupGroupGRPCClient() {
	var err error
	groupConn, err := grpc.Dial(cfg.Discovers["group"].Addr, grpc.WithInsecure())
	if err != nil {
		logger.Fatal("Failed to connect to gRPC server", zap.Error(err))
	}

	groupClient = groupApi.NewGroupServiceClient(groupConn)
}

func setupDialogGRPCClient() {
	var err error
	msgConn, err := grpc.Dial(cfg.Discovers["relation"].Addr, grpc.WithInsecure())
	if err != nil {
		logger.Fatal("Failed to connect to gRPC server", zap.Error(err))
	}

	dialogClient = relationApi.NewDialogServiceClient(msgConn)
}
func setupRelationGRPCClient() {
	var err error
	relationConn, err := grpc.Dial(cfg.Discovers["relation"].Addr, grpc.WithInsecure())
	if err != nil {
		logger.Fatal("Failed to connect to gRPC server", zap.Error(err))
	}

	relationClient = relationApi.NewUserRelationServiceClient(relationConn)
	userGroupClient = relationApi.NewGroupRelationServiceClient(relationConn)
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
	u.POST("/send/user", middleware.AuthMiddleware(redisClient), sendUserMsg)
	u.POST("/send/group", middleware.AuthMiddleware(redisClient), sendGroupMsg)
	u.GET("/list/user", middleware.AuthMiddleware(redisClient), getUserMsgList)
	u.GET("/dialog/list", middleware.AuthMiddleware(redisClient), getUserDialogList)
	u.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("msg")))
}
