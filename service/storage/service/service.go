package service

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/cossim/coss-server/service/storage/api/v1"
	"github.com/cossim/coss-server/service/storage/domain/entity"
	"github.com/cossim/coss-server/service/storage/domain/repository"
	"github.com/cossim/coss-server/service/storage/infrastructure/persistence"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
)

var (
	cfg    *config.AppConfig
	logger *zap.Logger
)

func NewService(c *config.AppConfig, repo *persistence.Repositories) *Service {
	cfg = c
	setupLogger()

	return &Service{
		fr: repo.FR,
	}
}

type Service struct {
	fr repository.FileRepository
	v1.UnimplementedStorageServiceServer
}

func (s Service) Upload(ctx context.Context, request *v1.UploadRequest) (*v1.UploadResponse, error) {
	resp := &v1.UploadResponse{}

	_, fileName, err := minio.ParseKey(request.Path)
	if err != nil {
		return resp, status.Error(codes.Code(code.StorageErrParseFilePathFailed.Code()), err.Error())
	}
	file := &entity.File{
		ID:      fileName,
		Name:    request.FileName,
		Owner:   request.UserID,
		Content: request.Url,
		Path:    request.Path,
		Type:    entity.FileType(request.Type),
		//Action:   entity.Pending,
		Provider: request.Provider,
		Size:     request.Size,
	}

	if err = s.fr.Create(file); err != nil {
		logger.Error("创建文件记录失败", zap.Error(err))
		return resp, status.Error(codes.Code(code.StorageErrCreateFileRecordFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s Service) GetFileInfo(ctx context.Context, request *v1.GetFileInfoRequest) (*v1.GetFileInfoResponse, error) {
	file, err := s.fr.GetByID(request.FileID)
	if err != nil {
		logger.Error("查询文件信息失败", zap.Error(err))
		return nil, status.Error(codes.Code(code.StorageErrGetFileInfoFailed.Code()), err.Error())
	}

	return &v1.GetFileInfoResponse{
		FileID:    file.ID,
		FileName:  file.Name,
		Size:      file.Size,
		Url:       file.Content,
		Path:      file.Path,
		Type:      v1.FileType(file.Type),
		CreatedAt: strconv.FormatInt(file.CreatedAt, 10),
		UpdatedAt: strconv.FormatInt(file.UpdatedAt, 10),
	}, nil
}

func (s Service) Delete(ctx context.Context, request *v1.DeleteRequest) (*v1.DeleteResponse, error) {
	fmt.Println("request.FileID => ", request.FileID)
	// 根据文件 ID 删除文件
	err := s.fr.Delete(request.FileID)
	if err != nil {
		logger.Error("删除文件记录失败", zap.Error(err))
		return nil, status.Error(codes.Code(code.StorageErrDeleteFileFailed.Code()), err.Error())
	}

	// 返回删除成功的响应
	return &v1.DeleteResponse{}, nil
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
	c := zap.Config{
		Level:            atom,                                                    // 日志级别
		Development:      true,                                                    // 开发模式，堆栈跟踪
		Encoding:         "console",                                               // 输出格式 console 或 json
		EncoderConfig:    encoderConfig,                                           // 编码器配置
		InitialFields:    map[string]interface{}{"serviceName": "upload_service"}, // 初始化字段，如：添加一个服务器名称
		OutputPaths:      []string{"stdout"},                                      // 输出到指定文件 stdout（标准输出，正常颜色） stderr（错误输出，红色）
		ErrorOutputPaths: []string{"stderr"},
	}
	// 构建日志
	var err error
	logger, err = c.Build()
	if err != nil {
		panic(fmt.Sprintf("logger初始化失败: %v", err))
	}
	logger.Info("logger初始化成功")
}
