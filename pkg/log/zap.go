package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(serviceName, format string, level int8) *zap.Logger {
	return setLogger(serviceName, format, zapcore.Level(level))
}

func NewDefaultLogger(serviceName string) *zap.Logger {
	return setLogger(serviceName, "console", zapcore.InfoLevel)
}

func NewDevLogger(serviceName string) *zap.Logger {
	return setLogger(serviceName, "console", zapcore.DebugLevel)
}

func setLogger(serviceName, encoding string, level zapcore.Level) *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "log",
		CallerKey:      "caller",
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.FullCallerEncoder,
	}

	// 设置日志级别
	atom := zap.NewAtomicLevelAt(level)
	config := zap.Config{
		Level:         atom,                                               // 日志级别
		Development:   true,                                               // 开发模式，堆栈跟踪
		Encoding:      encoding,                                           // 输出格式 console 或 json
		EncoderConfig: encoderConfig,                                      // 编码器配置
		InitialFields: map[string]interface{}{"serviceName": serviceName}, // 初始化字段，如：添加一个服务器名称
		OutputPaths:   []string{"stdout"},                                 // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
	}

	// 构建日志
	logger, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize log: %v", err))
	}

	// 添加全局字段
	logger = logger.WithOptions()
	return logger
}
