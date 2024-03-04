package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"time"
)

func GinLogger(lg *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		//query := c.Request.URL.RawQuery
		c.Next()
		// 获取当前日志级别
		logLevel := lg.Level()

		// 根据zap日志级别设置gin的日志级别
		switch logLevel {
		case zap.DebugLevel:
			gin.SetMode(gin.DebugMode)
		case zap.InfoLevel:
			gin.SetMode(gin.ReleaseMode)
		case zap.WarnLevel, zap.ErrorLevel, zap.DPanicLevel, zap.PanicLevel, zap.FatalLevel:
			gin.SetMode(gin.ReleaseMode)
		default:
			gin.SetMode(gin.ReleaseMode)
		}

		cost := time.Since(start)
		lg.Info(fmt.Sprintf("%s | %d | %.6fms | %s | %s %s",
			start.Format("2006/01/02 - 15:04:05"),
			c.Writer.Status(),
			float64(cost.Nanoseconds())/float64(time.Millisecond),
			c.ClientIP(),
			c.Request.Method,
			path,
		))
	}
}
