package service

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/pkg/code"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/cossim/coss-server/pkg/version"
	"github.com/cossim/coss-server/service/storage/api/v1"
	api "github.com/cossim/coss-server/service/storage/api/v1"
	"github.com/cossim/coss-server/service/storage/domain/entity"
	"github.com/cossim/coss-server/service/storage/domain/repository"
	"github.com/cossim/coss-server/service/storage/infrastructure/persistence"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
	"strconv"
)

type Service struct {
	logger *zap.Logger
	ac     *pkgconfig.AppConfig
	fr     repository.FileRepository
	v1.UnimplementedStorageServiceServer
}

func (s *Service) Init(cfg *pkgconfig.AppConfig) error {
	dbConn, err := db.NewMySQLFromDSN(cfg.MySQL.DSN).GetConnection()
	if err != nil {
		return err
	}

	infra := persistence.NewRepositories(dbConn)
	if err = infra.Automigrate(); err != nil {
		return err
	}
	s.fr = infra.FR
	s.ac = cfg
	s.logger = plog.NewDevLogger("storage_service")
	return nil
}

func (s *Service) Name() string {
	//TODO implement me
	return "storage_service"
}

func (s *Service) Version() string { return version.FullVersion() }

func (s *Service) Register(srv *grpc.Server) {
	api.RegisterStorageServiceServer(srv, s)
}

func (s *Service) RegisterHealth(srv *grpc.Server) {
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())
}

func (s *Service) Stop(ctx context.Context) error { return nil }

func (s *Service) DiscoverServices(services map[string]*grpc.ClientConn) error { return nil }

func (s *Service) Upload(ctx context.Context, request *v1.UploadRequest) (*v1.UploadResponse, error) {
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
		s.logger.Error("创建文件记录失败", zap.Error(err))
		return resp, status.Error(codes.Code(code.StorageErrCreateFileRecordFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s *Service) GetFileInfo(ctx context.Context, request *v1.GetFileInfoRequest) (*v1.GetFileInfoResponse, error) {
	file, err := s.fr.GetByID(request.FileID)
	if err != nil {
		s.logger.Error("查询文件信息失败", zap.Error(err))
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

func (s *Service) Delete(ctx context.Context, request *v1.DeleteRequest) (*v1.DeleteResponse, error) {
	fmt.Println("request.FileID => ", request.FileID)
	// 根据文件 ID 删除文件
	err := s.fr.Delete(request.FileID)
	if err != nil {
		s.logger.Error("删除文件记录失败", zap.Error(err))
		return nil, status.Error(codes.Code(code.StorageErrDeleteFileFailed.Code()), err.Error())
	}

	// 返回删除成功的响应
	return &v1.DeleteResponse{}, nil
}
