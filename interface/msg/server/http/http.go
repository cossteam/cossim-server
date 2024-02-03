package http

import (
	"context"
	"github.com/cossim/coss-server/interface/msg/config"
	"github.com/cossim/coss-server/interface/msg/service"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	"github.com/cossim/coss-server/pkg/log"
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

	logger *zap.Logger
	enc    encryption.Encryptor
	svc    *service.Service

	server *http.Server
	engine *gin.Engine
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

	if enc == nil {
		logger.Fatal("Failed to setup encryption")
		return
	}
	if redisClient == nil {
		logger.Fatal("Failed to setup redis")
		return
	}
	setupGin()

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
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}
	redisClient.Close()
}

func setupEncryption() {
	enc = encryption.NewEncryptor([]byte(config.Conf.Encryption.Passphrase), config.Conf.Encryption.Name, config.Conf.Encryption.Email, config.Conf.Encryption.RsaBits, config.Conf.Encryption.Enable)

	err := enc.ReadKeyPair()
	if err != nil {
		logger.Fatal("Failed to ", zap.Error(err))
		return
	}

	//readString, err := encryption.GenerateRandomKey(32)
	//if err != nil {
	//	logger.Fatal("Failed to ", zap.Error(err))
	//}
	//resp, err := enc.SecretMessage("{\n    \"content\": \"enim nostrud\",\n    \"receiver_id\": \"e3798b56-68f7-45f0-911f-147b0418f387\",\n    \"type\": 1,\n    \"dialog_id\":82\n}", enc.GetPublicKey(), []byte(readString))
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
		Addr:     config.Conf.Redis.Addr(),
		Password: config.Conf.Redis.Password, // no password set
		DB:       0,                          // use default DB
		//Protocol: cfg,
	})
	redisClient = rdb
}

func setupLogger() {
	logger = log.NewDevLogger("msg_bff")
}

func setupGin() {
	gin.SetMode(gin.ReleaseMode)
	// 添加一些中间件或其他配置
	engine.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(logger), middleware.EncryptionMiddleware(enc), middleware.RecoveryMiddleware())

	// 设置路由
	route(engine)
}

// @title Swagger Example API
func route(engine *gin.Engine) {
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	u := engine.Group("/api/v1/msg")
	//u.Use(middleware.AuthMiddleware())
	u.GET("/ws", middleware.AuthMiddleware(redisClient), ws)
	u.POST("/send/user_relation", middleware.AuthMiddleware(redisClient), sendUserMsg)
	u.POST("/send/group_relation", middleware.AuthMiddleware(redisClient), sendGroupMsg)
	u.GET("/list/user_relation", middleware.AuthMiddleware(redisClient), getUserMsgList)
	u.GET("/list/group_relation", middleware.AuthMiddleware(redisClient), getGroupMsgList)
	u.GET("/dialog/list", middleware.AuthMiddleware(redisClient), getUserDialogList)
	u.POST("/recall/group_relation", middleware.AuthMiddleware(redisClient), recallGroupMsg)
	u.POST("/recall/user_relation", middleware.AuthMiddleware(redisClient), recallUserMsg)
	u.POST("/edit/group_relation", middleware.AuthMiddleware(redisClient), editGroupMsg)
	u.POST("/edit/user_relation", middleware.AuthMiddleware(redisClient), editUserMsg)
	u.POST("/read/user_relation", middleware.AuthMiddleware(redisClient), readUserMsgs)

	//群聊标注消息
	u.POST("/label/group_relation", middleware.AuthMiddleware(redisClient), labelGroupMessage)
	u.GET("/label/group_relation", middleware.AuthMiddleware(redisClient), getGroupLabelMsgList)
	//私聊标注消息
	u.POST("/label/user_relation", middleware.AuthMiddleware(redisClient), labelUserMessage)
	u.GET("/label/user_relation", middleware.AuthMiddleware(redisClient), getUserLabelMsgList)
	u.POST("/after/get", middleware.AuthMiddleware(redisClient), getDialogAfterMsg)

	u.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("msg")))
}
