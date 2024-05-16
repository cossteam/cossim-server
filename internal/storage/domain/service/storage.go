package service

import (
	"context"
	"github.com/cossim/coss-server/internal/storage/domain/entity"
	"github.com/cossim/coss-server/internal/storage/infra/persistence"
	"github.com/cossim/coss-server/pkg/code"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/storage/minio"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type StorageDomain interface {
	Upload(context.Context, *entity.File) error
	GetFileInfo(context.Context, string) (*entity.File, error)
	Delete(context.Context, string) error
}

type StorageDomainImpl struct {
	db   *gorm.DB
	ac   *pkgconfig.AppConfig
	repo *persistence.Repositories
}

func NewStorageDomain(db *gorm.DB, ac *pkgconfig.AppConfig) StorageDomain {
	return &StorageDomainImpl{
		db:   db,
		ac:   ac,
		repo: persistence.NewRepositories(db),
	}
}

func (s *StorageDomainImpl) Upload(ctx context.Context, file *entity.File) error {
	_, fileName, err := minio.ParseKey(file.Path)
	if err != nil {
		return status.Error(codes.Code(code.StorageErrParseFilePathFailed.Code()), err.Error())
	}

	newfile := &entity.File{
		ID:      fileName,
		Name:    file.Name,
		Owner:   file.Owner,
		Content: file.Content,
		Path:    file.Path,
		Type:    file.Type,
		//Action:   entity.Pending,
		Provider: file.Provider,
		Size:     file.Size,
	}

	if err = s.repo.FR.Create(newfile); err != nil {
		return status.Error(codes.Code(code.StorageErrCreateFileRecordFailed.Code()), err.Error())
	}

	return nil
}

func (s *StorageDomainImpl) GetFileInfo(ctx context.Context, u string) (*entity.File, error) {
	file, err := s.repo.FR.GetByID(u)
	if err != nil {
		return nil, status.Error(codes.Code(code.StorageErrGetFileInfoFailed.Code()), err.Error())
	}
	return file, nil
}

func (s *StorageDomainImpl) Delete(ctx context.Context, u string) error {
	err := s.repo.FR.Delete(u)
	if err != nil {
		return status.Error(codes.Code(code.StorageErrDeleteFileFailed.Code()), err.Error())
	}
	return nil
}
