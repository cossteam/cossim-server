package http

import (
	"context"
	"github.com/cossim/coss-server/internal/admin/service"
	authv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/manager/server"
	"github.com/cossim/coss-server/pkg/version"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
	"strconv"
)

var (
	_ server.HTTPService = &Handler{}
)

type Handler struct {
	svc         *service.Service
	redisClient *cache.RedisClient
	logger      *zap.Logger
	enc         encryption.Encryptor
	db          *gorm.DB
	jwtKey      string
	authService authv1.UserAuthServiceClient
}

func (h *Handler) Init(cfg *pkgconfig.AppConfig) error {
	h.setupRedisClient(cfg)
	h.logger = plog.NewDefaultLogger("admin_bff", int8(cfg.Log.Level))
	if cfg.Encryption.Enable {
		if err := h.enc.ReadKeyPair(); err != nil {
			return err
		}
	}
	mysql, err := db.NewMySQL(cfg.MySQL.Address, strconv.Itoa(cfg.MySQL.Port), cfg.MySQL.Username, cfg.MySQL.Password, cfg.MySQL.Database, int64(cfg.Log.Level), cfg.MySQL.Opts)
	if err != nil {
		return err
	}

	h.db, err = mysql.GetConnection()
	if err != nil {
		return err
	}

	h.enc = encryption.NewEncryptor([]byte(cfg.Encryption.Passphrase), cfg.Encryption.Name, cfg.Encryption.Email, cfg.Encryption.RsaBits, cfg.Encryption.Enable)
	h.svc = service.New(cfg)
	return nil
}

func (h *Handler) Name() string {
	return "admin_bff"
}

func (h *Handler) Version() string {
	return version.FullVersion()
}

// @title CossApi
func (h *Handler) RegisterRoute(r gin.IRouter) {
	u := r.Group("/api/v1/admin")
	u.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(h.logger), middleware.RecoveryMiddleware())
	//u.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("admin")))
	u.Use(middleware.AdminAuthMiddleware(h.redisClient, h.db, h.jwtKey))

	u.Use(middleware.EncryptionMiddleware(h.enc))
	u.POST("/notification/send_all", h.sendAllNotification)
}

func (h *Handler) Health(r gin.IRouter) string {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) Stop(ctx context.Context) error {
	return h.svc.Stop(ctx)
}

func (h *Handler) DiscoverServices(services map[string]*grpc.ClientConn) error {
	for k, v := range services {
		if err := h.svc.HandlerGrpcClient(k, v); err != nil {
			h.logger.Error("handler grpc client error", zap.String("name", k), zap.String("addr", v.Target()), zap.Error(err))
		}
	}
	return nil
}

func (h *Handler) setupRedisClient(cfg *pkgconfig.AppConfig) {
	h.redisClient = cache.NewRedisClient(cfg.Redis.Addr(), cfg.Redis.Password)
}
