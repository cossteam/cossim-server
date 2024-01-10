package repository

import (
	"context"
)

type ProviderRepository interface {
	Upload(context context.Context, content string) error
	Download(context context.Context, fileID string) (string, error)
	Delete(context context.Context, fileID string) error
}
