package http

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/interface/gateway/config"
	"github.com/cossim/coss-server/pkg/code"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/cossim/coss-server/pkg/http/middleware"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/pretty66/websocketproxy"
	"go.uber.org/zap"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	logger = plog.NewDevLogger("gateway")
	server *http.Server
	engine *gin.Engine
	dis    discovery.Discovery

	userServiceURL      = new(string)
	relationServiceURL  = new(string)
	messageServiceURL   = new(string)
	messageWsServiceURL = new(string)
	groupServiceURL     = new(string)
	storageServiceURL   = new(string)
	liveUserServiceURL  = new(string)
)

func Start(discover bool) {
	engine = gin.New()
	server = &http.Server{
		Addr:    config.Conf.HTTP.Addr(),
		Handler: engine,
	}

	if discover {
		setDiscovery()
	}
	setupURLs(discover)
	setupGin()

	go func() {
		logger.Info("Gin server is running on port", zap.String("addr", config.Conf.HTTP.Addr()))
		if err := server.ListenAndServe(); err != nil {
			logger.Info("Failed to start Gin server", zap.Error(err))
			return
		}
	}()
}

func Restart(discover bool) error {
	Start(discover)
	return nil
}

func Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}
}

func setDiscovery() {
	d, err := discovery.NewConsulRegistry(config.Conf.Register.Addr())
	if err != nil {
		panic(err)
	}
	dis = d
}

func setupURLs(d bool) {
	if d {
		go discover()
	} else {
		direct()
	}
	//userServiceURL = "http://" + cfg.Discovers["user"].Addr()
	//relationServiceURL = "http://" + cfg.Discovers["relation"].Addr()
	//messageServiceURL = "http://" + cfg.Discovers["msg"].Addr()
	//messageWsServiceURL = "ws://" + cfg.Discovers["msg"].Addr() + "/api/v1/msg/ws"
	//groupServiceURL = "http://" + cfg.Discovers["group"].Addr()
	//storageServiceURL = "http://" + cfg.Discovers["storage"].Addr()
}

func discover() {
	var wg sync.WaitGroup
	type serviceInfo struct {
		ServiceName string
		Addr        string
	}
	ch := make(chan serviceInfo)

	for serviceName, c := range config.Conf.Discovers {
		if c.Direct {
			continue
		}
		wg.Add(1)
		go func(serviceName string, c pkgconfig.ServiceConfig) {
			defer wg.Done()
			for {
				addr, err := dis.Discover(c.Name)
				if err != nil {
					log.Printf("Service discovery failed ServiceName: %s %v\n", c.Name, err)
					time.Sleep(15 * time.Second)
					continue
				}
				log.Printf("Service discovery successful ServiceName: %s  Addr: %s\n", c.Name, addr)

				ch <- serviceInfo{ServiceName: serviceName, Addr: addr}
				break
			}
		}(serviceName, c)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for info := range ch {
		if err := handlerGrpcClient(info.ServiceName, info.Addr); err != nil {
			log.Printf("Failed to initialize gRPC client for service: %s, Error: %v\n", info.ServiceName, err)
		}
	}
}

func direct() {
	for serviceName, _ := range config.Conf.Discovers {
		if err := handlerGrpcClient(serviceName, config.Conf.Discovers[serviceName].Addr()); err != nil {
			panic(err)
		}
	}
}

func handlerGrpcClient(serviceName string, addr string) error {
	switch serviceName {
	case "user":
		*userServiceURL = "http://" + addr
		logger.Info("gRPC client for user service initialized", zap.String("service", "user"), zap.String("addr", addr))
	case "relation":
		*relationServiceURL = "http://" + addr
		logger.Info("gRPC client for relation service initialized", zap.String("service", "relation"), zap.String("addr", addr))
	case "group":
		*groupServiceURL = "http://" + addr
		logger.Info("gRPC client for group service initialized", zap.String("service", "group"), zap.String("addr", addr))
	case "msg":
		*messageServiceURL = "http://" + addr
		*messageWsServiceURL = "ws://" + addr + "/api/v1/msg/ws"
		logger.Info("gRPC client for group service initialized", zap.String("service", "msg"), zap.String("addr", addr))
	case "storage":
		*storageServiceURL = "http://" + addr
		logger.Info("gRPC client for group service initialized", zap.String("service", "storage"), zap.String("addr", addr))
	case "live":
		*liveUserServiceURL = "http://" + addr
		logger.Info("gRPC client for group service initialized", zap.String("service", "live"), zap.String("addr", addr))
	}

	return nil
}

func setupGin() {
	gin.SetMode(gin.ReleaseMode)
	// 添加一些中间件或其他配置
	engine.Use(middleware.CORSMiddleware(), middleware.RecoveryMiddleware(), middleware.GRPCErrorMiddleware(logger))
	// 设置路由
	route(engine)
}

func route(engine *gin.Engine) {
	gateway := engine.Group("/api/v1")
	{
		gateway.Any("/user/*path", proxyToService(userServiceURL))
		gateway.Any("/relation/*path", proxyToService(relationServiceURL))
		gateway.Any("/msg/*path", proxyToService(messageServiceURL))
		gateway.Any("/group/*path", proxyToService(groupServiceURL))
		gateway.Any("/storage/*path", proxyToService(storageServiceURL))
		gateway.Any("/live/*path", proxyToService(liveUserServiceURL))
	}
}

func proxyToService(targetURL *string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("Received request", zap.Any("RequestHeaders", c.Request.Header), zap.String("RequestURL", c.Request.URL.String()))
		if c.Request.Header.Get("Upgrade") == "websocket" {
			wp, err := websocketproxy.NewProxy(*messageWsServiceURL, func(r *http.Request) error {
				// 握手时设置cookie, 权限验证
				r.Header.Set("Cookie", "----")
				// 伪装来源
				r.Header.Set("Origin", *messageServiceURL)
				return nil
			})
			if err != nil {
				fmt.Println("websocketproxy err ", err)
			}
			wp.Proxy(c.Writer, c.Request)
			return
		}
		// 创建一个代理请求
		proxyReq, err := http.NewRequest(c.Request.Method, *targetURL+c.Request.URL.Path, c.Request.Body)
		if err != nil {
			logger.Error("Failed to create proxy request", zap.Error(err))
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
			logger.Error("Failed to fetch response from service", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  code.InternalServerError.Message(),
				"data": nil,
			})
			return
		}
		defer resp.Body.Close()

		logger.Info("Received response from service", zap.Any("ResponseHeaders", resp.Header), zap.String("TargetURL", *targetURL))

		// 将 BFF 服务的响应返回给客户端
		c.Status(resp.StatusCode)
		for h, val := range resp.Header {
			c.Header(h, val[0])
		}
		c.Writer.WriteHeader(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)
	}
}
