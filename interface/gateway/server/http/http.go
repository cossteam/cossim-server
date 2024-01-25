package http

import (
	"fmt"
	"github.com/cossim/coss-server/pkg/code"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/cossim/coss-server/pkg/http/middleware"
	"github.com/gin-gonic/gin"
	"github.com/pretty66/websocketproxy"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

var (
	cfg    *pkgconfig.AppConfig
	logger *zap.Logger

	dis discovery.Discovery

	userServiceURL      string
	relationServiceURL  string
	messageServiceURL   string
	messageWsServiceURL string
	groupServiceURL     string
	storageServiceURL   string
)

func Init(c *pkgconfig.AppConfig, discover bool) {
	cfg = c

	setupLogger()
	setDiscovery(discover)
	setupURLs(discover)
	setupGin()
}

func setDiscovery(f bool) {
	if !f {
		return
	}
	d, err := discovery.NewConsulRegistry(cfg.Register.Addr())
	if err != nil {
		panic(err)
	}
	dis = d
}

func setupLogger() {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
		EncodeTime:     zapcore.ISO8601TimeEncoder,    // ISO8601 UTC 时间格式
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder, // 全路径编码器
	}

	// 设置日志级别
	atom := zap.NewAtomicLevelAt(zapcore.Level(cfg.Log.V))
	config := zap.Config{
		Level:            atom,                                              // 日志级别
		Development:      true,                                              // 开发模式，堆栈跟踪
		Encoding:         cfg.Log.Format,                                    // 输出格式 console 或 json
		EncoderConfig:    encoderConfig,                                     // 编码器配置
		InitialFields:    map[string]interface{}{"serviceName": "user-bff"}, // 初始化字段，如：添加一个服务器名称
		OutputPaths:      []string{"stdout"},                                // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		ErrorOutputPaths: []string{"stderr"},
	}
	// 构建日志
	var err error
	logger, err = config.Build()
	if err != nil {
		panic(fmt.Sprintf("log 初始化失败: %v", err))
	}
	logger.Info("log 初始化成功")
	logger.Info("无法获取网址",
		zap.String("url", "http://www.baidu.com"),
		zap.Int("attempt", 3),
		zap.Duration("backoff", time.Second),
	)
}

func setupURLs(d bool) {
	if cfg == nil {
		panic("Config not initialized")
	}
	if d {
		discover()
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

	for serviceName, c := range cfg.Discovers {
		wg.Add(1)
		go func(serviceName string, c pkgconfig.ServiceConfig) {
			defer wg.Done()
			for {
				addr, err := dis.Discover(c.Name)
				if err != nil {
					log.Printf("Service discovery failed ServiceName: %s %v\n", c.Name, err)
					time.Sleep(5 * time.Second)
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
	for serviceName, _ := range cfg.Discovers {
		if err := handlerGrpcClient(serviceName, cfg.Discovers[serviceName].Addr()); err != nil {
			panic(err)
		}
	}
}

func handlerGrpcClient(serviceName string, addr string) error {
	switch serviceName {
	case "user":
		userServiceURL = "http://" + addr
		logger.Info("gRPC client for user service initialized", zap.String("service", "user"), zap.String("addr", addr))
	case "relation":
		relationServiceURL = "http://" + addr
		logger.Info("gRPC client for relation service initialized", zap.String("service", "relation"), zap.String("addr", addr))
	case "group":
		groupServiceURL = "http://" + addr
		logger.Info("gRPC client for group service initialized", zap.String("service", "group"), zap.String("addr", addr))
	case "msg":
		messageServiceURL = "http://" + addr
		messageWsServiceURL = "ws://" + addr + "/api/v1/msg/ws"
		logger.Info("gRPC client for group service initialized", zap.String("service", "msg"), zap.String("addr", addr))
	case "storage":
		storageServiceURL = "http://" + addr
		logger.Info("gRPC client for group service initialized", zap.String("service", "storage"), zap.String("addr", addr))
	}

	return nil
}

func setupGin() {
	if cfg == nil {
		panic("Config not initialized")
	}

	// 初始化 gin engine
	engine := gin.New()

	// 添加一些中间件或其他配置
	engine.Use(middleware.CORSMiddleware(), middleware.RecoveryMiddleware(), middleware.GRPCErrorMiddleware(logger))

	// 设置路由
	route(engine)

	// 启动 Gin 服务器
	go func() {
		if err := engine.Run(cfg.HTTP.Addr()); err != nil {
			logger.Fatal("Failed to start Gin server", zap.Error(err))
		}
	}()
}

func route(engine *gin.Engine) {
	gateway := engine.Group("/api/v1")
	{
		gateway.Any("/user/*path", proxyToService(userServiceURL))
		gateway.Any("/relation/*path", proxyToService(relationServiceURL))
		gateway.Any("/msg/*path", proxyToService(messageServiceURL))
		gateway.Any("/group/*path", proxyToService(groupServiceURL))
		gateway.Any("/storage/*path", proxyToService(storageServiceURL))
	}
}

func proxyToService(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info("Received request", zap.Any("RequestHeaders", c.Request.Header), zap.String("RequestURL", c.Request.URL.String()))
		if c.Request.Header.Get("Upgrade") == "websocket" {
			wp, err := websocketproxy.NewProxy(messageWsServiceURL, func(r *http.Request) error {
				// 握手时设置cookie, 权限验证
				r.Header.Set("Cookie", "----")
				// 伪装来源
				r.Header.Set("Origin", messageServiceURL)
				return nil
			})
			if err != nil {
				fmt.Println("websocketproxy err ", err)
			}
			wp.Proxy(c.Writer, c.Request)
			return
		}
		// 创建一个代理请求
		proxyReq, err := http.NewRequest(c.Request.Method, targetURL+c.Request.URL.Path, c.Request.Body)
		if err != nil {
			logger.Error("Failed to create proxy request", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  code.InternalServerError,
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch response from service"})
			return
		}
		defer resp.Body.Close()

		logger.Info("Received response from service", zap.Any("ResponseHeaders", resp.Header), zap.String("TargetURL", targetURL))

		// 将 BFF 服务的响应返回给客户端
		c.Status(resp.StatusCode)
		for h, val := range resp.Header {
			c.Header(h, val[0])
		}
		c.Writer.WriteHeader(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)
	}
}
