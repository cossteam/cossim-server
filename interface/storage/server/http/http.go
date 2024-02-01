package http

import (
	"context"
	"github.com/cossim/coss-server/interface/storage/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/storage"
	"github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/cossim/coss-server/pkg/utils/os"
	storagev1 "github.com/cossim/coss-server/service/storage/api/v1"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/xid"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"net/http"
	"time"
)

var (
	sp            storage.StorageProvider
	storageClient storagev1.StorageServiceClient
	redisClient   *redis.Client

	logger      *zap.Logger
	downloadURL = "/api/v1/storage/files/download"
	enc         encryption.Encryptor

	discover discovery.Discovery
	sid      string

	server         *http.Server
	engine         *gin.Engine
	gatewayAddress string
	gatewayPort    string
	appPath        string
)

func Start(dis bool) {
	engine = gin.New()
	server = &http.Server{
		Addr:    config.Conf.HTTP.Addr(),
		Handler: engine,
	}

	setupRedis()
	setupLogger()
	setupEncryption()
	setupStorageClient()
	setLoadSystem()
	setMinIOProvider()
	setupGin()

	if enc == nil {
		logger.Fatal("Failed to setup encryption")
		return
	}
	if redisClient == nil {
		logger.Fatal("Failed to setup redis")
		return
	}

	if dis {
		setupDiscovery()
	}

	go func() {
		logger.Info("Gin server is running on port", zap.String("addr", config.Conf.HTTP.Addr()))
		if err := server.ListenAndServe(); err != nil {
			logger.Info("Failed to start Gin server", zap.Error(err))
			return
		}
	}()
}

func Stop(dis bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	if dis {
		if err := discover.Cancel(sid); err != nil {
			log.Printf("Failed to cancel service registration: %v", err)
			return
		}
		log.Printf("Service registration canceled ServiceName: %s  Addr: %s  ID: %s", config.Conf.Register.Name, config.Conf.GRPC.Addr(), sid)
	}
}

func Restart(dis bool) error {
	Start(dis)
	return nil
}

func setupDiscovery() {
	d, err := discovery.NewConsulRegistry(config.Conf.Register.Addr())
	if err != nil {
		panic(err)
	}
	discover = d
	sid = xid.New().String()
	if err = d.RegisterHTTP(config.Conf.Register.Name, config.Conf.HTTP.Addr(), sid); err != nil {
		panic(err)
	}
	logger.Info("Service register success", zap.String("name", config.Conf.Register.Name), zap.String("addr", config.Conf.HTTP.Addr()), zap.String("id", sid))
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
		Addr:     config.Conf.Redis.Addr(),
		Password: config.Conf.Redis.Password, // no password set
		DB:       0,                          // use default DB
		//Protocol: cfg,
	})
	redisClient = rdb
}

func setupStorageClient() {
	conn, err := grpc.Dial(config.Conf.Discovers["storage"].Addr(), grpc.WithInsecure())
	if err != nil {
		logger.Fatal("Failed to connect to gRPC server", zap.Error(err))
	}

	storageClient = storagev1.NewStorageServiceClient(conn)
}

func setMinIOProvider() {
	var err error
	sp, err = minio.NewMinIOStorage(config.Conf.OSS["minio"].Addr(), config.Conf.OSS["minio"].AccessKey, config.Conf.OSS["minio"].SecretKey, config.Conf.OSS["minio"].SSL)
	if err != nil {
		panic(err)
	}
}

func setupLogger() {
	logger = plog.NewDevLogger("storage_bff")
}

func setupGin() {
	// 添加一些中间件或其他配置
	engine.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(logger), middleware.EncryptionMiddleware(enc), middleware.RecoveryMiddleware())
	// 设置路由
	route(engine)
}

// @title coss-storage-bff服务
func route(engine *gin.Engine) {
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	// 添加不同的中间件给不同的路由组
	// 比如除了swagger路径外其他的路径添加了身份验证中间件
	api := engine.Group("/api/v1/storage")
	//api.Use(middleware.AuthMiddleware())

	api.GET("/files/download/:type/:id", download)
	api.GET("/files/:id", getFileInfo)
	api.POST("/files", upload)
	api.DELETE("/files/:id", deleteFile)
	api.GET("/files/multipart/key", getMultipartKey)
	api.POST("/files/multipart/upload", uploadMultipart)
	api.POST("/files/multipart/complete", completeUploadMultipart)

	// 为Swagger路径添加不需要身份验证的中间件
	swagger := engine.Group("/api/v1/storage/swagger")
	swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("storage")))
}

func setLoadSystem() {

	env := config.Conf.SystemConfig.Environment
	if env == "" {
		env = "dev"
	}

	switch env {
	case "prod":
		path := config.Conf.SystemConfig.AvatarFilePath
		if path == "" {
			path = "/.catch/"
		}
		appPath = path

		gatewayAdd := config.Conf.SystemConfig.GatewayAddress
		if gatewayAdd == "" {
			gatewayAdd = "43.229.28.107"
		}

		gatewayAddress = gatewayAdd

		gatewayPo := config.Conf.SystemConfig.GatewayPort
		if gatewayPo == "" {
			gatewayPo = "8080"
		}
		gatewayPort = gatewayPo
	default:
		path := config.Conf.SystemConfig.AvatarFilePathDev
		if path == "" {
			npath, err := os.GetPackagePath()
			if err != nil {
				panic(err)
			}
			path = npath + "deploy/docker/config/common/"
		}
		appPath = path

		gatewayAdd := config.Conf.SystemConfig.GatewayAddressDev
		if gatewayAdd == "" {
			gatewayAdd = "127.0.0.1"
		}

		gatewayAddress = gatewayAdd

		gatewayPo := config.Conf.SystemConfig.GatewayPortDev
		if gatewayPo == "" {
			gatewayPo = "8080"
		}
		gatewayPort = gatewayPo
	}

}
