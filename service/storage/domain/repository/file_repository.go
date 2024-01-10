package repository

import (
	"github.com/cossim/coss-server/service/storage/domain/entity"
)

type FileRepository interface {
	Create(file *entity.File) error
	Update(file *entity.File) error
	Delete(fileID string) error
	GetByID(fileID string) (*entity.File, error)
}
