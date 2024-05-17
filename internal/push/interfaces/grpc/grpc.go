package grpc

import (
	"context"
	api "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	"github.com/cossim/coss-server/internal/push/service"
	"github.com/cossim/coss-server/pkg/cache"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/version"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
)

type Handler struct {
	ac          *pkgconfig.AppConfig
	sid         string
	logger      *zap.Logger
	PushService *service.Service
	cache       cache.PushCache
	api.UnimplementedPushServiceServer
}

func (h *Handler) Init(cfg *pkgconfig.AppConfig) error {
	h.ac = cfg
	h.sid = xid.New().String()
	h.logger = plog.NewDefaultLogger("push_service", int8(cfg.Log.Level))
	cache2, err := cache.NewPushCacheRedis(cfg.Redis.Addr(), cfg.Redis.Password, 0)
	if err != nil {
		return err
	}
	h.cache = cache2
	//h.pushService = pushService
	//h.PushService.Init(cfg)
	return nil
}

func (h *Handler) Name() string {
	//TODO implement me
	return "push_service"
}

func (h *Handler) Version() string {
	//TODO implement me
	return version.FullVersion()
}

func (h *Handler) Register(s *grpc.Server) {
	api.RegisterPushServiceServer(s, h)
}

func (h *Handler) RegisterHealth(s *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(s, health.NewServer())

}

func (h *Handler) Stop(ctx context.Context) error {
	return h.cache.DeleteAllCache(ctx)
}

func (h *Handler) DiscoverServices(services map[string]*grpc.ClientConn) error {
	for k, v := range services {
		if err := h.PushService.HandlerGrpcClient(k, v); err != nil {
			h.logger.Error("handler grpc client error", zap.String("name", k), zap.String("addr", v.Target()), zap.Error(err))
		}
	}
	return nil
}
