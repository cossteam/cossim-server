package storage

import (
	"context"
	"github.com/minio/minio-go/v7"
	"io"
	"net/url"
)

// StorageProvider 定义了存储接口
type StorageProvider interface {
	Upload(ctx context.Context, key string, reader io.Reader, size int64, opt minio.PutObjectOptions) (*url.URL, error)
	GetUrl(ctx context.Context, key string) (string, error)
	Delete(ctx context.Context, key string) error
	NewMultipartUpload(ctx context.Context, key string, opt minio.PutObjectOptions) (string, error)
	GenUploadPartSignedUrl(ctx context.Context, key string, uploadId string, partNumber int, partSize int64) (string, error)
	CompleteMultipartUpload(ctx context.Context, key string, uploadId string) (*url.URL, error)
	UploadPart(ctx context.Context, key string, uploadId string, partNumber int, reader io.Reader, size int64, opt minio.PutObjectPartOptions) error
	AbortMultipartUpload(ctx context.Context, key string, uploadId string) error
	GetObjectInfo(ctx context.Context, key string, opt minio.GetObjectOptions) (minio.ObjectInfo, error)
	UploadAvatar(ctx context.Context, key string, reader io.Reader, size int64, opt minio.PutObjectOptions) error
}
