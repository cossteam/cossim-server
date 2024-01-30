package service

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/pkg/code"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/cossim/coss-server/service/storage/api/v1"
	"github.com/cossim/coss-server/service/storage/domain/entity"
	"github.com/cossim/coss-server/service/storage/domain/repository"
	"github.com/rs/xid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"strconv"
)

var (
	cfg    *pkgconfig.AppConfig
	logger *zap.Logger
)

func NewService(fr repository.FileRepository, c *pkgconfig.AppConfig) *Service {
	cfg = c
	setupLogger()

	return &Service{
		fr:  fr,
		ac:  c,
		sid: xid.New().String(),
	}
}

type Service struct {
	fr repository.FileRepository
	v1.UnimplementedStorageServiceServer

	ac        *pkgconfig.AppConfig
	discovery discovery.Discovery
	sid       string
}

func (s *Service) Start(discover bool) {
	if !discover {
		return
	}
	d, err := discovery.NewConsulRegistry(s.ac.Register.Addr())
	if err != nil {
		panic(err)
	}
	s.discovery = d
	if err = s.discovery.Register(s.ac.Register.Name, s.ac.GRPC.Addr(), s.sid); err != nil {
		panic(err)
	}
	log.Printf("Service registration successful ServiceName: %s  Addr: %s  ID: %s", s.ac.Register.Name, s.ac.GRPC.Addr(), s.sid)
}

func (s *Service) Stop(discover bool) error {
	if !discover {
		return nil
	}
	if err := s.discovery.Cancel(s.sid); err != nil {
		log.Printf("Failed to cancel service registration: %v", err)
		return err
	}
	log.Printf("Service registration canceled ServiceName: %s  Addr: %s  ID: %s", s.ac.Register.Name, s.ac.GRPC.Addr(), s.sid)
	return nil
}

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
		logger.Error("创建文件记录失败", zap.Error(err))
		return resp, status.Error(codes.Code(code.StorageErrCreateFileRecordFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s *Service) GetFileInfo(ctx context.Context, request *v1.GetFileInfoRequest) (*v1.GetFileInfoResponse, error) {
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

func (s *Service) Delete(ctx context.Context, request *v1.DeleteRequest) (*v1.DeleteResponse, error) {
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

func (s *Service) Restart(discover bool) {
	s.Stop(discover)
	s.Start(discover)
}

func setupLogger() {
	logger = plog.NewDevLogger("storage_svc")
}
