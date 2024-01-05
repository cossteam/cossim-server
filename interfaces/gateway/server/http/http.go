package http

import (
	"fmt"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/http/middleware"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"net/http"
	"time"
)

var (
	cfg    *config.AppConfig
	logger *zap.Logger
)

func Init(c *config.AppConfig) {
	cfg = c

	setupLogger()
	setupGin()
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
		if err := engine.Run(cfg.HTTP.Addr); err != nil {
			logger.Fatal("Failed to start Gin server", zap.Error(err))
		}
	}()
}

func route(engine *gin.Engine) {
	gateway := engine.Group("/api/v1")
	{
		gateway.Any("/user/*path", proxyToService("http://127.0.0.1:8083"))
		gateway.Any("/relation/*path", proxyToService("http://127.0.0.1:8082"))
		gateway.Any("/msg/*path", proxyToService("http://127.0.0.1:8081"))
	}
}

func proxyToService(targetURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 创建一个代理请求
		proxyReq, err := http.NewRequest(c.Request.Method, targetURL+c.Request.URL.Path, c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create proxy request"})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch response from service"})
			return
		}
		defer resp.Body.Close()

		// 将 BFF 服务的响应返回给客户端
		c.Status(resp.StatusCode)
		for h, val := range resp.Header {
			c.Header(h, val[0])
		}
		c.Writer.WriteHeader(resp.StatusCode)
		io.Copy(c.Writer, resp.Body)
	}
}