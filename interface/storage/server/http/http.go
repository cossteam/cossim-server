package http

import (
	"encoding/json"
	"fmt"
	conf "github.com/cossim/coss-server/interface/storage/config"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	"github.com/cossim/coss-server/pkg/storage"
	"github.com/cossim/coss-server/pkg/storage/minio"
	storagev1 "github.com/cossim/coss-server/service/storage/api/v1"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"os"
)

var (
	sp            storage.StorageProvider
	storageClient storagev1.StorageServiceClient
	redisClient   *redis.Client
	cfg           *config.AppConfig
	logger        *zap.Logger
	downloadURL   = "/api/v1/storage/files/download"
	enc           encryption.Encryptor
)

func Init(c *config.AppConfig) {
	cfg = c
	setupRedis()
	setupLogger()
	setupEncryption()
	setupStorageClient()
	setMinIOProvider()
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
	resp, err := enc.SecretMessage("{\n    \"user_id\": \"e3798b56-68f7-45f0-911f-147b0418f387\",\n    \"action\": 0,\n    \"e2e_public_key\": \"ex Ut ad incididunt occaecat\"\n}", enc.GetPublicKey(), []byte(readString))
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
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password, // no password set
		DB:       0,                  // use default DB
		//Protocol: cfg,
	})
	redisClient = rdb
}

func setupStorageClient() {
	var err error
	storageConn, err := grpc.Dial(cfg.Discovers["storage"].Addr, grpc.WithInsecure())
	if err != nil {
		logger.Fatal("Failed to connect to gRPC server", zap.Error(err))
	}

	storageClient = storagev1.NewStorageServiceClient(storageConn)
}

func setMinIOProvider() {
	var err error
	sp, err = minio.NewMinIOStorage(conf.MinioConf.Endpoint, conf.MinioConf.AccessKey, conf.MinioConf.SecretKey, conf.MinioConf.SSL)
	if err != nil {
		panic(err)
	}
}

func setupLogger() {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "storage",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder, // 全路径编码器
	}

	// 设置日志级别
	atom := zap.NewAtomicLevelAt(zapcore.Level(cfg.Log.V))
	c := zap.Config{
		Level:            atom,                                                 // 日志级别
		Development:      true,                                                 // 开发模式，堆栈跟踪
		Encoding:         "console",                                            // 输出格式 console 或 json
		EncoderConfig:    encoderConfig,                                        // 编码器配置
		InitialFields:    map[string]interface{}{"serviceName": "storage-bff"}, // 初始化字段，如：添加一个服务器名称
		OutputPaths:      []string{"stdout"},                                   // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		ErrorOutputPaths: []string{"stderr"},
	}

	var err error
	logger, err = c.Build()
	if err != nil {
		panic(fmt.Sprintf("logger初始化失败: %v", err))
	}
	logger.Info("logger初始化成功")
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
		if err := engine.Run(cfg.HTTP.Addr); err != nil {
			logger.Fatal("Failed to start Gin server", zap.Error(err))
		}
	}()
}

// @title coss-storage-bff服务
func route(engine *gin.Engine) {
	// 添加不同的中间件给不同的路由组
	// 比如除了swagger路径外其他的路径添加了身份验证中间件
	api := engine.Group("/api/v1/storage")
	//api.Use(middleware.AuthMiddleware())

	api.GET("/files/download/:type/:id", download)
	api.GET("/files/:id", getFileInfo)
	api.POST("/files", upload)
	api.DELETE("/files/:id", deleteFile)

	// 为Swagger路径添加不需要身份验证的中间件
	swagger := engine.Group("/api/v1/storage/swagger")
	swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("storage")))
}
