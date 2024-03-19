package http

import (
	"context"
	"github.com/cossim/coss-server/pkg/code"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/http/middleware"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/server"
	"github.com/cossim/coss-server/pkg/version"
	"github.com/gin-gonic/gin"
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/pretty66/websocketproxy"
	"google.golang.org/grpc"
	"io"
	"net/http"
)

var (
	_ server.HTTPService = &Handler{}

	userServiceURL      = new(string)
	relationServiceURL  = new(string)
	messageServiceURL   = new(string)
	messageWsServiceURL = new(string)
	groupServiceURL     = new(string)
	storageServiceURL   = new(string)
	liveUserServiceURL  = new(string)
	adminServiceURL     = new(string)
)

type Handler struct {
	logger logr.Logger
	cfg    *pkgconfig.AppConfig
}

func (h *Handler) Init(cfg *pkgconfig.AppConfig) error {
	logger := zapr.NewLogger(plog.NewDefaultLogger("gateway", int8(cfg.Log.Level)))
	h.logger = logger
	h.cfg = cfg
	return nil
}

func (h *Handler) Name() string {
	return "gateway"
}

func (h *Handler) Version() string {
	return version.FullVersion()
}

func (h *Handler) RegisterRoute(r gin.IRouter) {
	r.Use(middleware.CORSMiddleware(), middleware.RecoveryMiddleware(), middleware.GRPCErrorMiddleware(plog.NewLogger(h.cfg.Log.Format, int8(h.cfg.Log.Level), true)))
	gateway := r.Group("/api/v1")
	{
		gateway.Any("/user/*path", h.proxyToService(userServiceURL))
		gateway.Any("/relation/*path", h.proxyToService(relationServiceURL))
		gateway.Any("/msg/*path", h.proxyToService(messageServiceURL))
		gateway.Any("/group/*path", h.proxyToService(groupServiceURL))
		gateway.Any("/storage/*path", h.proxyToService(storageServiceURL))
		gateway.Any("/live/*path", h.proxyToService(liveUserServiceURL))
		gateway.Any("/admin/*path", h.proxyToService(adminServiceURL))
	}
}

func (h *Handler) Health(r gin.IRouter) string {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) Stop(ctx context.Context) error {
	return nil
}

func (h *Handler) DiscoverServices(services map[string]*grpc.ClientConn) error {
	for name, conn := range services {
		h.handlerGrpcClient(name, conn)
	}
	return nil
}

func (h *Handler) handlerGrpcClient(serviceName string, conn *grpc.ClientConn) {
	addr := conn.Target()
	switch serviceName {
	case "user_bff":
		*userServiceURL = "http://" + addr
		h.logger.Info("gRPC client for user service initialized", "service", "user", "addr", addr)
	case "relation_bff":
		*relationServiceURL = "http://" + addr
		h.logger.Info("gRPC client for relation service initialized", "service", "relation", "addr", addr)
	case "group_bff":
		*groupServiceURL = "http://" + addr
		h.logger.Info("gRPC client for group service initialized", "service", "group", "addr", addr)
	case "msg_bff":
		*messageServiceURL = "http://" + addr
		*messageWsServiceURL = "ws://" + addr + "/api/v1/msg/ws"
		h.logger.Info("gRPC client for group service initialized", "service", "msg", "addr", addr)
	case "storage_bff":
		*storageServiceURL = "http://" + addr
		h.logger.Info("gRPC client for group service initialized", "service", "storage", "addr", addr)
	case "live_bff":
		*liveUserServiceURL = "http://" + addr
		h.logger.Info("gRPC client for group service initialized", "service", "live", "addr", addr)
	case "admin_bff":
		*adminServiceURL = "http://" + addr
		h.logger.Info("gRPC client for group service initialized", "service", "admin", "addr", addr)
	}
}

func (h *Handler) proxyToService(targetURL *string) gin.HandlerFunc {
	return func(c *gin.Context) {
		h.logger.Info("Received request", "RequestHeader", c.Request.Header, "RequestURL", c.Request.URL.String())
		if c.Request.Header.Get("Upgrade") == "websocket" {
			wp, err := websocketproxy.NewProxy(*messageWsServiceURL, func(r *http.Request) error {
				// 握手时设置cookie, 权限验证
				r.Header.Set("Cookie", "----")
				// 伪装来源
				r.Header.Set("Origin", *messageServiceURL)
				return nil
			})
			if err != nil {
				h.logger.Error(err, "Failed to create websocket proxy")
				c.JSON(http.StatusInternalServerError, gin.H{
					"code": http.StatusInternalServerError,
					"msg":  err.Error(),
					"data": nil,
				})
				return
			}
			wp.Proxy(c.Writer, c.Request)
			return
		}
		// 创建一个代理请求
		proxyReq, err := http.NewRequest(c.Request.Method, *targetURL+c.Request.URL.Path, c.Request.Body)
		if err != nil {
			h.logger.Error(err, "Failed to create proxy request")
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  code.InternalServerError.Message(),
				"data": nil,
			})
			return
		}

		// 添加查询字符串到代理请求的 URL 中
		proxyReq.URL.RawQuery = c.Request.URL.RawQuery

		// 复制请求头信息
		proxyReq.Header = make(http.Header)
		for h, val := range c.Request.Header {
			proxyReq.Header[h] = val
		}
		// 发送代理请求
		client := &http.Client{}
		resp, err := client.Do(proxyReq)
		if err != nil {
			h.logger.Error(err, "Failed to fetch response from service")
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  code.InternalServerError.Message(),
				"data": nil,
			})
			return
		}
		defer resp.Body.Close()

		h.logger.Info("Received response from service", "ResponseHeaders", resp.Header, "TargetURL", *targetURL)

		// 将 BFF 服务的响应返回给客户端
		c.Status(resp.StatusCode)
		for h, val := range resp.Header {
			c.Header(h, val[0])
		}
		c.Writer.WriteHeader(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)
	}
}
