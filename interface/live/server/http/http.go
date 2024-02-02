package http

import (
	"context"
	"github.com/cossim/coss-server/interface/live/api/dto"
	"github.com/cossim/coss-server/interface/live/api/model"
	"github.com/cossim/coss-server/interface/live/config"
	"github.com/cossim/coss-server/interface/live/service"
	"github.com/cossim/coss-server/pkg/cache"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/encryption"
	pkghttp "github.com/cossim/coss-server/pkg/http"
	"github.com/cossim/coss-server/pkg/http/middleware"
	"github.com/cossim/coss-server/pkg/http/response"
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
	//u.GET("/getToken", h.getJoinToken)
	u.POST("/create", h.Create)
	u.POST("/join", h.Join)
	u.GET("/show", h.Show)
	u.POST("/leave", h.Leave)

	// 为Swagger路径添加不需要身份验证的中间件
	swagger := h.engine.Group("/api/v1/live/swagger")
	swagger.GET("/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler(), ginSwagger.InstanceName("live")))
}

// Create
// @Summary 创建用户通话
// @Description 创建用户通话
// @Tags liveUser
// @Security Bearer
// @Param Authorization header string true "Bearer JWT"
// @param request body dto.UserCallRequest true "request"
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.Response{data=map[string]string} "url=webRtc服务器地址 token=加入通话的token room_name=房间名称 room_id=房间id"
// @Router /live/user/create [post]
func (h *Handler) Create(c *gin.Context) {
	req := new(dto.UserCallRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数验证失败", nil)
		return
	}

	userID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		h.logger.Error("token解析失败", zap.Error(err))
		response.Fail(c, "token解析失败", nil)
		return
	}

	resp, err := h.svc.CreateUserCall(c, userID, req.UserID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "创建通话成功", resp)
}

func (h *Handler) Join(c *gin.Context) {
	req := new(dto.UserJoinRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	//room, err := h.getRoomInfoFromRedis(req.Room)
	//if err != nil {
	//	c.Error(err)
	//	return
	//}
	//
	//if req.UserID != room.RecipientID || req.UserID != room.SenderID {
	//	c.Error(code.Unauthorized.Reason(err))
	//	return
	//}
	//
	//if len(room.Participants)+1 > int(room.MaxParticipants) {
	//	c.Error(code.LiveErrMaxParticipantsExceeded.Reason(err))
	//	return
	//}

	resp, err := h.svc.UserJoinRoom(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "加入通话成功", resp)
}

// Show
// @Summary 获取通话房间信息
// @Description 获取通话房间信息
// @Tags liveUser
// @Security Bearer
// @Param Authorization header string true "Bearer JWT"
// @Param user_id query string true "用户id"
// @Param room query string true "房间名"
// @Produce  json
// @Success		200 {object} dto.Response{data=dto.UserShowResponse} "sid=房间id identity=用户id state=用户状态 (0=加入中 1=已加入 2=已连接 3=断开连接 ) joined_at=加入时间 name=用户名称 is_publisher=是否是创建者"
// @Router /live/user/show [get]
func (h *Handler) Show(c *gin.Context) {
	userID := c.Query("user_id")
	rid := c.Query("room")
	if userID == "" || rid == "" {
		response.SetFail(c, code.InvalidParameter.Message(), nil)
		return
	}
	//
	//room, err := h.getRoomInfoFromRedis(rid)
	//if err != nil {
	//	c.Error(err)
	//	return
	//}
	//
	//if userID != room.RecipientID || userID != room.SenderID {
	//	c.Error(code.Unauthorized.Reason(err))
	//	return
	//}

	resp, err := h.svc.GetUserRoom(c, userID, rid)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "获取通话信息成功", resp)
}

// Leave
// @Summary 结束通话
// @Description 结束通话
// @Tags liveUser
// @Security Bearer
// @Param Authorization header string true "Bearer JWT"
// @param request body dto.UserLeaveRequest true "request"
// @Accept  json
// @Produce  json
// @Success		200 {object} dto.Response{}
// @Router /live/user/leave [post]
func (h *Handler) Leave(c *gin.Context) {
	req := new(dto.UserLeaveRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	//room, err := h.getRoomInfoFromRedis(req.Room)
	//if err != nil {
	//	c.Error(err)
	//	return
	//}
	//
	//if req.UserID != room.RecipientID || req.UserID != room.SenderID {
	//	c.Error(code.Unauthorized.Reason(err))
	//	return
	//}

	resp, err := h.svc.UserLeaveRoom(c, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "结束通话成功", resp)
}

func (h *Handler) getJoinToken(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		response.SetFail(c, code.InvalidParameter.Message(), nil)
		return
	}

	resp, err := h.svc.GetJoinToken(c, userID, "")
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取token成功", resp)
}

func (h *Handler) getRoomInfoFromRedis(roomID string) (*model.RoomInfo, error) {
	room, err := cache.GetKey(h.redisClient, roomID)
	if err != nil {
		return nil, code.NotFound.Reason(err)
	}
	data := &model.RoomInfo{}
	if err = data.FromMap(room); err != nil {
		return nil, code.InvalidParameter.Reason(err)
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
