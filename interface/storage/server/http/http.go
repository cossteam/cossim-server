package http

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/pkg/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/server"
	"github.com/cossim/coss-server/pkg/storage"
	"github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/cossim/coss-server/pkg/utils/os"
	"github.com/cossim/coss-server/pkg/version"
	storagev1 "github.com/cossim/coss-server/service/storage/api/v1"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	_ server.HTTPService = &Handler{}
)

type Handler struct {
	sp            storage.StorageProvider
	storageClient storagev1.StorageServiceClient
	redisClient   *cache.RedisClient
	logger        *zap.Logger
	enc           encryption.Encryptor
	discover      discovery.Registry
	sid           string
	minioAddr     string
}

func (h *Handler) Init(cfg *pkgconfig.AppConfig) error {
	h.setupRedisClient(cfg)
	h.logger = plog.NewDefaultLogger("storage_bff", int8(cfg.Log.Level))
	sp, err := minio.NewMinIOStorage(cfg.OSS.Addr(), cfg.OSS.AccessKey, cfg.OSS.SecretKey, cfg.OSS.SSL)
	if err != nil {
		return err
	}
	h.sp = sp
	setLoadSystem(cfg)
	h.minioAddr = cfg.OSS.Addr()
	h.enc = encryption.NewEncryptor([]byte(cfg.Encryption.Passphrase), cfg.Encryption.Name, cfg.Encryption.Email, cfg.Encryption.RsaBits, cfg.Encryption.Enable)
	return h.enc.ReadKeyPair()
}

func (h *Handler) Name() string {
	return "storage_bff"
}

func (h *Handler) Version() string {
	return version.FullVersion()
}

func (h *Handler) setupRedisClient(cfg *pkgconfig.AppConfig) {
	h.redisClient = cache.NewRedisClient(cfg.Redis.Addr(), cfg.Redis.Password)
}

// @title storage服务

func (h *Handler) RegisterRoute(r gin.IRouter) {
	r.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(h.logger), middleware.EncryptionMiddleware(h.enc), middleware.RecoveryMiddleware())
	api := r.Group("/api/v1/storage")
	// 为Swagger路径添加不需要身份验证的中间件
	api.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("storage")))
	api.GET("/files/download/:type/:id", h.download)

	api.Use(middleware.AuthMiddleware(h.redisClient))
	api.GET("/files/:id", h.getFileInfo)
	api.POST("/files", h.upload)
	api.DELETE("/files/:id", h.deleteFile)
	api.GET("/files/multipart/key", h.getMultipartKey)
	api.POST("/files/multipart/upload", h.uploadMultipart)
	api.POST("/files/multipart/complete", h.completeUploadMultipart)
	api.POST("/files/multipart/abort", h.abortUploadMultipart)
}

func (h *Handler) Health(r gin.IRouter) string {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) Stop(ctx context.Context) error {
	return nil
}

func (h *Handler) DiscoverServices(services map[string]*grpc.ClientConn) error {
	if _, ok := services["storage_service"]; !ok {
		return errors.New("storage service not found")
	}
	h.storageClient = storagev1.NewStorageServiceClient(services["storage_service"])
	return nil
}

var (
	downloadURL     = "/api/v1/storage/files/download"
	systemEnableSSL bool
	gatewayAddress  string
	gatewayPort     string
	appPath         string
)

func setLoadSystem(ac *pkgconfig.AppConfig) {
	env := ac.SystemConfig.Environment
	if env == "" {
		env = "dev"
	}

	switch env {
	case "prod":
		path := ac.SystemConfig.AvatarFilePath
		if path == "" {
			path = "/.catch/"
		}
		appPath = path

		gatewayAdd := ac.SystemConfig.GatewayAddress
		if gatewayAdd == "" {
			gatewayAdd = "43.229.28.107"
		}

		gatewayAddress = gatewayAdd

		gatewayPo := ac.SystemConfig.GatewayPort
		if gatewayPo == "" {
			gatewayPo = "8080"
		}
		gatewayPort = gatewayPo
	default:
		path := ac.SystemConfig.AvatarFilePathDev
		if path == "" {
			npath, err := os.GetPackagePath()
			if err != nil {
				panic(err)
			}
			path = npath + "deploy/docker/config/common/"
		}
		appPath = path

		gatewayAdd := ac.SystemConfig.GatewayAddressDev
		if gatewayAdd == "" {
			gatewayAdd = "127.0.0.1"
		}

		gatewayAddress = gatewayAdd

		gatewayPo := ac.SystemConfig.GatewayPortDev
		if gatewayPo == "" {
			gatewayPo = "8080"
		}
		gatewayPort = gatewayPo
	}
	if !ac.SystemConfig.Ssl {
		gatewayAddress = gatewayAddress + ":" + gatewayPort
	}
	systemEnableSSL = ac.SystemConfig.Ssl
}
