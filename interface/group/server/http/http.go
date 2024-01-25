package http

import (
	"fmt"
	"github.com/cossim/coss-server/interface/group/service"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	redisClient *redis.Client
	cfg         *pkgconfig.AppConfig
	logger      *zap.Logger
	enc         encryption.Encryptor
	svc         *service.Service
)

func Init(c *pkgconfig.AppConfig, service *service.Service) {
	cfg = c
	svc = service

	setupLogger()
	setupEncryption()
	setupGin()
}

func setupEncryption() {
	enc = encryption.NewEncryptor([]byte(cfg.Encryption.Passphrase), cfg.Encryption.Name, cfg.Encryption.Email, cfg.Encryption.RsaBits, cfg.Encryption.Enable)

	err := enc.ReadKeyPair()
	if err != nil {
		logger.Fatal("Failed to ", zap.Error(err))
		return
	}
	//
	//readString, err := encryption.GenerateRandomKey(32)
	//if err != nil {
	//	logger.Fatal("Failed to ", zap.Error(err))
	//}
	//resp, err := enc.SecretMessage("{\n    \"name\": \"置手提织则表\",\n    \"type\": 1,\n    \"max_members_limit\": 75,\n    \"avatar\": \"http://dummyimage.com/100x100\"\n}", enc.GetPublicKey(), []byte(readString))
	//if err != nil {
	//	logger.Fatal("Failed to ", zap.Error(err))
	//}
	//j, err := json.Marshal(resp)
	//if err != nil {
	//	logger.Fatal("Failed to ", zap.Error(err))
	//}
	////保存成文件
	//cacheDir := ".cache"
	//if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
	//	err := os.Mkdir(cacheDir, 0755) // 创建文件夹并设置权限
	//	if err != nil {
	//		logger.Fatal("Failed to ", zap.Error(err))
	//	}
	//}
	//// 保存私钥到文件
	//privateKeyFile, err := os.Create(cacheDir + "/data.json")
	//if err != nil {
	//	logger.Fatal("Failed to ", zap.Error(err))
	//}
	//
	//_, err = privateKeyFile.WriteString(string(j))
	//if err != nil {
	//	privateKeyFile.Close()
	//	logger.Fatal("Failed to ", zap.Error(err))
	//}
	//privateKeyFile.Close()
	//fmt.Println("加密后消息：", string(j))
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
		InitialFields:    map[string]interface{}{"serviceName": "msg_bff"}, // 初始化字段，如：添加一个服务器名称
		OutputPaths:      []string{"stdout"},                               // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		ErrorOutputPaths: []string{"stderr"},
	}
	// 构建日志
	var err error
	logger, err = config.Build()
	if err != nil {
		panic(fmt.Sprintf("logger 初始化失败: %v", err))
	}
	logger.Info("logger 初始化成功")
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

// @title coss-user服务

func route(engine *gin.Engine) {
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	g := engine.Group("/api/v1/group")
	// 获取群聊信息
	g.GET("/info", middleware.AuthMiddleware(redisClient), getGroupInfoByGid)
	// 创建群聊
	g.POST("/create", middleware.AuthMiddleware(redisClient), createGroup)
	// 删除群聊
	g.POST("/delete", middleware.AuthMiddleware(redisClient), deleteGroup)
	// 更新群聊信息
	g.POST("/update", middleware.AuthMiddleware(redisClient), updateGroup)
	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("group")))
	//u.POST("/logout", handleLogout)
}
