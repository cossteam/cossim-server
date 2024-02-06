package http

import (
	"context"
	"encoding/json"
	"github.com/cossim/coss-server/interface/live/api/model"
	"github.com/cossim/coss-server/interface/live/config"
	"github.com/cossim/coss-server/interface/live/service"
	"github.com/cossim/coss-server/pkg/cache"
	"github.com/cossim/coss-server/pkg/code"
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

type Handler struct {
	svc         *service.Service
	redisClient *redis.Client

	logger *zap.Logger
	enc    encryption.Encryptor
	engine *gin.Engine
	server *http.Server
}

func NewHandler(svc *service.Service) *Handler {
	engine := gin.New()
	return &Handler{
		svc:    svc,
		logger: log.NewDevLogger("live"),
		enc:    setupEncryption(),
		engine: engine,
		server: &http.Server{
			Addr:    config.Conf.HTTP.Addr(),
			Handler: engine,
		},
		redisClient: redis.NewClient(&redis.Options{
			Addr:     config.Conf.Redis.Addr(),
			Password: config.Conf.Redis.Password, // no password set
			DB:       0,                          // use default DB
		}),
	}
}

func (h *Handler) Start() {
	gin.SetMode(gin.ReleaseMode)
	// 添加一些中间件或其他配置
	h.engine.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(h.logger), middleware.EncryptionMiddleware(h.enc), middleware.RecoveryMiddleware())
	// 设置路由
	h.route()
	go func() {
		h.logger.Info("Gin server is running on port", zap.String("addr", config.Conf.HTTP.Addr()))
		if err := h.server.ListenAndServe(); err != nil {
			h.logger.Info("Failed to start Gin server", zap.Error(err))
			return
		}
	}()
}

func (h *Handler) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := h.server.Shutdown(ctx); err != nil {
		h.logger.Fatal("Server forced to shutdown", zap.Error(err))
	}
	if h.redisClient != nil {
		h.redisClient.Close()
	}
}

// @title Swagger Example API
func (h *Handler) route() {
	h.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	u := h.engine.Group("/api/v1/live/user")
	u.Use(middleware.AuthMiddleware(h.redisClient))
	u.GET("/show", h.UserShow)
	u.POST("/create", h.UserCreate)
	u.POST("/join", h.UserJoin)
	u.POST("/reject", h.UserReject)
	u.POST("/leave", h.UserLeave)

	g := h.engine.Group("/api/v1/live/group")
	g.Use(middleware.AuthMiddleware(h.redisClient))
	g.GET("/show", h.GroupShow)
	g.POST("/create", h.GroupCreate)
	g.POST("/join", h.GroupJoin)
	g.POST("/reject", h.GroupReject)
	g.POST("/leave", h.GroupLeave)

	// 为Swagger路径添加不需要身份验证的中间件
	swagger := h.engine.Group("/api/v1/live/swagger")
	swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("live")))
}

func (h *Handler) getRoomInfoFromRedis(roomID string) (*model.UserRoomInfo, error) {
	room, err := cache.GetKey(h.redisClient, roomID)
	if err != nil {
		return nil, code.LiveErrCallNotFound.Reason(err)
	}
	data := &model.UserRoomInfo{}
	if err = json.Unmarshal([]byte(room.(string)), &data); err != nil {
		return nil, err
	}
	return data, nil
}

func setupEncryption() encryption.Encryptor {
	enc := encryption.NewEncryptor([]byte(config.Conf.Encryption.Passphrase), config.Conf.Encryption.Name, config.Conf.Encryption.Email, config.Conf.Encryption.RsaBits, config.Conf.Encryption.Enable)
	err := enc.ReadKeyPair()
	if err != nil {
		panic(err)
	}
	return enc
}
