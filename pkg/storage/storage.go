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
}
