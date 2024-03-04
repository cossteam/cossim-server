package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewLogger(format string, level int8, stdout bool) *zap.Logger {
	return setLogger(format, zapcore.Level(level), stdout)
}

func NewDefaultLogger(serviceName string, level int8) *zap.Logger {
	logger := setLogger("console", zapcore.Level(level), true)
	logger.With(zap.String("serviceName", serviceName))
	return logger
}

func setLogger(encoding string, level zapcore.Level, stdout bool) *zap.Logger {
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
		Level:         atom,               // 日志级别
		Development:   true,               // 开发模式，堆栈跟踪
		Encoding:      encoding,           // 输出格式 console 或 json
		EncoderConfig: encoderConfig,      // 编码器配置
		OutputPaths:   []string{"stdout"}, // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
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
