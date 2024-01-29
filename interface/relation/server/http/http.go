package http

import (
	"fmt"
	"github.com/cossim/coss-server/interface/relation/service"
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

	cfg    *pkgconfig.AppConfig
	logger *zap.Logger
	svc    *service.Service
	enc    encryption.Encryptor
)

func Init(c *pkgconfig.AppConfig, service *service.Service) {
	cfg = c
	svc = service

	setupLogger()
	setupRedis()
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

	//readString, err := encryption.GenerateRandomKey(32)
	//if err != nil {
	//	logger.Fatal("Failed to ", zap.Error(err))
	//}
	//resp, err := enc.SecretMessage("{\n    \"user_id\": \"e3798b56-68f7-45f0-911f-147b0418f387\",\n    \"action\": 0,\n    \"e2e_public_key\": \"ex Ut ad incididunt occaecat\"\n}", enc.GetPublicKey(), []byte(readString))
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
		Level:            atom,                                                  // 日志级别
		Development:      true,                                                  // 开发模式，堆栈跟踪
		Encoding:         cfg.Log.Format,                                        // 输出格式 console 或 json
		EncoderConfig:    encoderConfig,                                         // 编码器配置
		InitialFields:    map[string]interface{}{"serviceName": "relation_bff"}, // 初始化字段，如：添加一个服务器名称
		OutputPaths:      []string{"stdout"},                                    // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
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

// @title coss-relation-bff服务

func route(engine *gin.Engine) {
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	// 添加不同的中间件给不同的路由组
	// 比如除了swagger路径外其他的路径添加了身份验证中间件
	api := engine.Group("/api/v1/relation")
	api.Use(middleware.AuthMiddleware(redisClient))

	u := api.Group("/user")
	u.GET("/friend_list", friendList)
	u.GET("/blacklist", blackList)
	u.GET("/request_list", userRequestList)
	u.POST("/add_friend", addFriend)
	u.POST("/manage_friend", manageFriend)
	u.POST("/delete_friend", deleteFriend)
	u.POST("/add_blacklist", addBlacklist)
	u.POST("/delete_blacklist", deleteBlacklist)
	u.POST("/switch/e2e/key", switchUserE2EPublicKey)
	//设置用户静默通知
	u.POST("/silent", setUserSilentNotification)

	g := api.Group("/group")
	g.GET("/member", getGroupMember)
	g.GET("/request_list", groupRequestList)
	// 邀请好友加入群聊
	g.POST("/invite", inviteGroup)
	// 申请加入群聊
	g.POST("/join", joinGroup)
	// 用户加入或拒绝群聊
	g.POST("/manage_join", manageJoinGroup)
	//获取用户群聊列表
	g.GET("/list", getUserGroupList)
	// 退出群聊
	g.POST("quit", quitGroup)
	//群聊设置消息静默
	g.POST("/silent", setGroupSilentNotification)

	gg := api.Group("/group/admin")
	// 管理员管理群聊申请
	gg.POST("/manage/join", adminManageJoinGroup)
	// 管理员移除群聊成员
	gg.POST("/manage/remove", removeUserFromGroup)

	d := api.Group("/dialog")
	d.POST("/top", topOrCancelTopDialog)
	d.POST("/show", closeOrOpenDialog)

	// 为Swagger路径添加不需要身份验证的中间件
	swagger := engine.Group("/api/v1/relation/swagger")
	swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("relation")))
}
