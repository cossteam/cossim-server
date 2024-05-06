package http

import (
	"context"
	grpcinter "github.com/cossim/coss-server/internal/storage/interface/grpc"
	"github.com/cossim/coss-server/internal/storage/service"
	"github.com/cossim/coss-server/internal/user/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/manager/server"
	"github.com/cossim/coss-server/pkg/version"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	_ server.HTTPService = &Handler{}
)

type Handler struct {
	StorageClient *grpcinter.Handler
	logger        *zap.Logger
	enc           encryption.Encryptor
	svc           *service.Service
	minioAddr     string
	userCache     cache.UserCache
	jwtSecret     string
}

func (h *Handler) Init(cfg *pkgconfig.AppConfig) error {
	h.logger = plog.NewDefaultLogger("storage_bff", int8(cfg.Log.Level))
	h.minioAddr = cfg.OSS.Addr()

	if cfg.Encryption.Enable {
		return h.enc.ReadKeyPair()
	}
	userCache, err := cache.NewUserCacheRedis(cfg.Redis.Addr(), cfg.Redis.Password, 0)
	if err != nil {
		return err
	}
	h.userCache = userCache
	h.enc = encryption.NewEncryptor([]byte(cfg.Encryption.Passphrase), cfg.Encryption.Name, cfg.Encryption.Email, cfg.Encryption.RsaBits, cfg.Encryption.Enable)
	h.svc = service.New(cfg, h.StorageClient)
	h.jwtSecret = cfg.SystemConfig.JwtSecret
	return nil
}

func (h *Handler) Name() string {
	return "storage_bff"
}

func (h *Handler) Version() string {
	return version.FullVersion()
}

// @title CossApi

func (h *Handler) RegisterRoute(r gin.IRouter) {
	r.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(h.logger), middleware.EncryptionMiddleware(h.enc), middleware.RecoveryMiddleware())
	api := r.Group("/api/v1/storage")
	api.GET("/files/download/:type/:id", h.download)
	api.Use(middleware.AuthMiddleware(h.userCache, h.jwtSecret))

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
	return nil
}

var (
	downloadURL     = "/api/v1/storage/files/download"
	systemEnableSSL bool
	gatewayAddress  string
	gatewayPort     string
)
