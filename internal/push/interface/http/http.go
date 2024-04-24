package http

import (
	"context"
	"github.com/cossim/coss-server/internal/push/service"
	"github.com/cossim/coss-server/internal/user/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/version"
	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
)

type Handler struct {
	logger       *zap.Logger
	enc          encryption.Encryptor
	PushService  *service.Service
	userCache    cache.UserCache
	socketServer *socketio.Server
}

//var (
//	upgrader = websocket.Upgrader{
//		CheckOrigin: func(r *http.Request) bool {
//			return true
//		},
//	}
//)

func (h *Handler) Init(cfg *pkgconfig.AppConfig) error {
	h.logger = plog.NewDefaultLogger("push_bff", int8(cfg.Log.Level))
	if cfg.Encryption.Enable {
		return h.enc.ReadKeyPair()
	}
	userCache, err := cache.NewUserCacheRedis(cfg.Redis.Addr(), cfg.Redis.Password, 0)
	if err != nil {
		return err
	}
	h.userCache = userCache
	h.socketServer = h.PushService.SocketServer
	h.setupSocketEvent()
	h.enc = encryption.NewEncryptor([]byte(cfg.Encryption.Passphrase), cfg.Encryption.Name, cfg.Encryption.Email, cfg.Encryption.RsaBits, cfg.Encryption.Enable)
	return nil
}

func (h *Handler) Name() string {
	return "push_bff"
}

func (h *Handler) Version() string {
	return version.FullVersion()
}

func (h *Handler) RegisterRoute(r gin.IRouter) {
	u := r.Group("/api/v1/push")
	u.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(h.logger), middleware.EncryptionMiddleware(h.enc), middleware.RecoveryMiddleware())
	u.Use(middleware.AuthMiddleware(h.userCache))
	u.GET("/ws/*any", gin.WrapH(h.socketServer))
	u.POST("/ws/*any", gin.WrapH(h.socketServer))

}

func (h *Handler) Health(r gin.IRouter) string {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) Stop(ctx context.Context) error {
	err := h.socketServer.Close()
	if err != nil {
		return err
	}
	return nil
}

func (h *Handler) DiscoverServices(services map[string]*grpc.ClientConn) error {
	for k, v := range services {
		if err := h.PushService.HandlerGrpcClient(k, v); err != nil {
			h.logger.Error("handler grpc client error", zap.String("name", k), zap.String("addr", v.Target()), zap.Error(err))
		}
	}
	return nil
}

func (h *Handler) setupSocketEvent() {
	if h.socketServer == nil {
		panic("socketio server is nil")
	}
	h.socketServer.OnConnect("/", h.ws)

	h.socketServer.OnEvent("/", "reply", func(s socketio.Conn, msg string) {
		s.Emit("reply", "服务端触发客户端事件： "+msg)
		//h.socketServer.BroadcastToRoom()
	})

	h.socketServer.OnEvent("/", "bye", func(s socketio.Conn) string {
		last := s.Context().(string)
		s.Emit("bye", last)
		s.Close()
		return last
	})

	h.socketServer.OnError("/", func(s socketio.Conn, e error) {
		//h.disconnect(s, "")
		h.logger.Error("socketio error", zap.Error(e))
	})

	h.socketServer.OnDisconnect("/", h.disconnect)

	go func() {
		if err := h.socketServer.Serve(); err != nil {
			log.Fatalf("socketio listen error: %s\n", err)
		}
	}()
}
