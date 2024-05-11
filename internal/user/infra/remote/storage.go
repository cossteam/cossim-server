package remote

import (
	"bytes"
	"context"
	storagev1 "github.com/cossim/coss-server/internal/storage/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/storage"
	myminio "github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/cossim/coss-server/pkg/utils/avatar"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/o1egl/govatar"
	"image/png"
)

type StorageService interface {
	GenerateAvatar(ctx context.Context) (string, error)
	//UploadAvatar(ctx context.Context, path string, reader bytes.Reader, avatarType string) error
}

var _ StorageService = &storageService{}

type storageService struct {
	client storage.StorageProvider
}

func NewStorageService(client storage.StorageProvider) StorageService {
	return &storageService{client: client}
}

func (s *storageService) UploadAvatar(ctx context.Context, path string, reader bytes.Reader, avatarType string) error {
	bucket, err := myminio.GetBucketName(int(storagev1.FileType_Other))
	if err != nil {
		return err
	}

	fileID := uuid.New().String()
	key := myminio.GenKey(bucket, fileID+".jpeg")
	err = s.client.UploadAvatar(ctx, key, &reader, reader.Size(), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *storageService) GenerateAvatar(ctx context.Context) (string, error) {
	img, err := govatar.Generate(avatar.RandomGender())
	if err != nil {
		return "", err
	}

	// 将图像编码为PNG格式
	var buf bytes.Buffer
	err = png.Encode(&buf, img)
	if err != nil {
		return "", err
	}

	bucket, err := myminio.GetBucketName(int(storagev1.FileType_Other))
	if err != nil {
		return "", err
	}

	reader := bytes.NewReader(buf.Bytes())
	fileID := uuid.New().String()
	path := myminio.GenKey(bucket, fileID+".jpeg")
	if err = s.client.UploadAvatar(ctx, path, reader, reader.Size(), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	}); err != nil {
		return "", err
	}

	return path, nil
}
