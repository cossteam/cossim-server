package http

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/interface/group/config"
	"github.com/cossim/coss-server/interface/group/service"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"net/http"
	"time"
)

var (
	redisClient *redis.Client
	logger      *zap.Logger
	enc         encryption.Encryptor
	svc         *service.Service
	server      *http.Server
	engine      *gin.Engine
)

func Start(service *service.Service) {
	svc = service
	engine = gin.New()
	server = &http.Server{
		Addr:    config.Conf.HTTP.Addr(),
		Handler: engine,
	}

	setupLogger()
	setupEncryption()
	setupRedis()
	setupGin()

	if enc == nil {
		logger.Fatal("Failed to setup encryption")
		return
	}
	if redisClient == nil {
		logger.Fatal("Failed to setup redis")
		return
	}
	go func() {
		logger.Info("Gin server is running on port", zap.String("addr", config.Conf.HTTP.Addr()))
		if err := server.ListenAndServe(); err != nil {
			logger.Info("Failed to start Gin server", zap.Error(err))
			return
		}
	}()
}

func Restart(service *service.Service) error {
	Start(service)
	return nil
}

func Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal(fmt.Sprintf("Server forced to shutdown: %v", err))
	}

	redisClient.Close()
}

func setupRedis() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     config.Conf.Redis.Addr(),
		Password: config.Conf.Redis.Password, // no password set
		DB:       0,                          // use default DB
		//Protocol: cfg,
	})
	redisClient = rdb
}

func setupLogger() {
	logger = plog.NewDevLogger("group_bff")
}

func setupGin() {
	gin.SetMode(gin.ReleaseMode)
	// 添加一些中间件或其他配置
	engine.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(logger), middleware.EncryptionMiddleware(enc), middleware.RecoveryMiddleware())

	// 设置路由
	route(engine)
}

// @title coss-user服务

func route(engine *gin.Engine) {
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	g := engine.Group("/api/v1/group_relation")
	// 获取群聊信息
	g.GET("/info", middleware.AuthMiddleware(redisClient), getGroupInfoByGid)
	// 创建群聊
	g.POST("/create", middleware.AuthMiddleware(redisClient), createGroup)
	// 删除群聊
	g.POST("/delete", middleware.AuthMiddleware(redisClient), deleteGroup)
	// 更新群聊信息
	g.POST("/update", middleware.AuthMiddleware(redisClient), updateGroup)
	g.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("group_relation")))
	//u.POST("/logout", handleLogout)
}

func setupEncryption() {
	enc = encryption.NewEncryptor([]byte(config.Conf.Encryption.Passphrase), config.Conf.Encryption.Name, config.Conf.Encryption.Email, config.Conf.Encryption.RsaBits, config.Conf.Encryption.Enable)

	err := enc.ReadKeyPair()
	if err != nil {
		logger.Fatal("Failed to ", zap.Error(err))
		return
	}
	//
	//readString, err := encryption.GenerateRandomKey(32)
	//if err != nil {
	//	log.Fatal("Failed to ", zap.Error(err))
	//}
	//resp, err := enc.SecretMessage("{\n    \"name\": \"置手提织则表\",\n    \"type\": 1,\n    \"max_members_limit\": 75,\n    \"avatar\": \"http://dummyimage.com/100x100\"\n}", enc.GetPublicKey(), []byte(readString))
	//if err != nil {
	//	log.Fatal("Failed to ", zap.Error(err))
	//}
	//j, err := json.Marshal(resp)
	//if err != nil {
	//	log.Fatal("Failed to ", zap.Error(err))
	//}
	////保存成文件
	//cacheDir := ".cache"
	//if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
	//	err := os.Mkdir(cacheDir, 0755) // 创建文件夹并设置权限
	//	if err != nil {
	//		log.Fatal("Failed to ", zap.Error(err))
	//	}
	//}
	//// 保存私钥到文件
	//privateKeyFile, err := os.Create(cacheDir + "/data.json")
	//if err != nil {
	//	log.Fatal("Failed to ", zap.Error(err))
	//}
	//
	//_, err = privateKeyFile.WriteString(string(j))
	//if err != nil {
	//	privateKeyFile.Stop()
	//	log.Fatal("Failed to ", zap.Error(err))
	//}
	//privateKeyFile.Stop()
	//fmt.Println("加密后消息：", string(j))
}
