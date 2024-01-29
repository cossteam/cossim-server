package http

import (
	"encoding/json"
	"fmt"
	"github.com/cossim/coss-server/interface/msg/service"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

var (
	redisClient *redis.Client

	cfg    *pkgconfig.AppConfig
	logger *zap.Logger
	enc    encryption.Encryptor
	svc    *service.Service
)

func Init(c *pkgconfig.AppConfig, srv *service.Service) {
	cfg = c
	svc = srv

	setupLogger()
	setupEncryption()
	service.Enc = enc
	setupRedis()
	setupGin()
}

func setupEncryption() {
	enc = encryption.NewEncryptor([]byte(cfg.Encryption.Passphrase), cfg.Encryption.Name, cfg.Encryption.Email, cfg.Encryption.RsaBits, cfg.Encryption.Enable)

	err := enc.ReadKeyPair()
	if err != nil {
		logger.Fatal("Failed to ", zap.Error(err))
		return
	}

	readString, err := encryption.GenerateRandomKey(32)
	if err != nil {
		logger.Fatal("Failed to ", zap.Error(err))
	}
	resp, err := enc.SecretMessage("{\n    \"content\": \"enim nostrud\",\n    \"receiver_id\": \"e3798b56-68f7-45f0-911f-147b0418f387\",\n    \"type\": 1,\n    \"dialog_id\":82\n}", enc.GetPublicKey(), []byte(readString))
	if err != nil {
		logger.Fatal("Failed to ", zap.Error(err))
	}
	j, err := json.Marshal(resp)
	if err != nil {
		logger.Fatal("Failed to ", zap.Error(err))
	}
	//保存成文件
	cacheDir := ".cache"
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		err := os.Mkdir(cacheDir, 0755) // 创建文件夹并设置权限
		if err != nil {
			logger.Fatal("Failed to ", zap.Error(err))
		}
	}
	// 保存私钥到文件
	privateKeyFile, err := os.Create(cacheDir + "/data.json")
	if err != nil {
		logger.Fatal("Failed to ", zap.Error(err))
	}

	_, err = privateKeyFile.WriteString(string(j))
	if err != nil {
		privateKeyFile.Close()
		logger.Fatal("Failed to ", zap.Error(err))
	}
	privateKeyFile.Close()
	fmt.Println("加密后消息：", string(j))
}

func setupRedis() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr(),
		Password: cfg.Redis.Password, // no password set
		DB:       0,                  // use default DB
		//Protocol: cfg,
	})
	redisClient = rdb
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
	engine.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(logger), middleware.EncryptionMiddleware(enc), middleware.RecoveryMiddleware())

	// 设置路由
	route(engine)

	// 启动 Gin 服务器
	go func() {
		if err := engine.Run(cfg.HTTP.Addr()); err != nil {
			logger.Fatal("Failed to start Gin server", zap.Error(err))
		}
	}()
}

// @title Swagger Example API
func route(engine *gin.Engine) {
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	u := engine.Group("/api/v1/msg")
	//u.Use(middleware.AuthMiddleware())
	u.GET("/ws", middleware.AuthMiddleware(redisClient), ws)
	u.POST("/send/user", middleware.AuthMiddleware(redisClient), sendUserMsg)
	u.POST("/send/group", middleware.AuthMiddleware(redisClient), sendGroupMsg)
	u.GET("/list/user", middleware.AuthMiddleware(redisClient), getUserMsgList)
	u.GET("/list/group", middleware.AuthMiddleware(redisClient), getGroupMsgList)
	u.GET("/dialog/list", middleware.AuthMiddleware(redisClient), getUserDialogList)
	u.POST("/recall/group", middleware.AuthMiddleware(redisClient), recallGroupMsg)
	u.POST("/recall/user", middleware.AuthMiddleware(redisClient), recallUserMsg)
	u.POST("/edit/group", middleware.AuthMiddleware(redisClient), editGroupMsg)
	u.POST("/edit/user", middleware.AuthMiddleware(redisClient), editUserMsg)
	u.POST("/read/user", middleware.AuthMiddleware(redisClient), readUserMsgs)

	//群聊标注消息
	u.POST("/label/group", middleware.AuthMiddleware(redisClient), labelGroupMessage)
	u.GET("/label/group", middleware.AuthMiddleware(redisClient), getGroupLabelMsgList)
	//私聊标注消息
	u.POST("/label/user", middleware.AuthMiddleware(redisClient), labelUserMessage)
	u.GET("/label/user", middleware.AuthMiddleware(redisClient), getUserLabelMsgList)
	u.POST("/after/get", middleware.AuthMiddleware(redisClient), getDialogAfterMsg)

	u.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("msg")))
}
