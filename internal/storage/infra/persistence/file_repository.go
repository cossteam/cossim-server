package persistence

import (
	"github.com/cossim/coss-server/internal/storage/domain/entity"
	"github.com/cossim/coss-server/internal/storage/infra/persistence/converter"
	"github.com/cossim/coss-server/internal/storage/infra/persistence/po"
	"gorm.io/gorm"
)

type FileRepo struct {
	db *gorm.DB
}

func NewFileRepo(db *gorm.DB) *FileRepo {
	return &FileRepo{db: db}
}

func (f *FileRepo) Create(file *entity.File) error {
	model := converter.FileEntityToPO(file)
	return f.db.Create(model).Error
}

func (f *FileRepo) Update(file *entity.File) error {
	model := converter.FileEntityToPO(file)
	return f.db.Where("id = ?", model.ID).Updates(model).Error
}

func (f *FileRepo) Delete(fileID string) error {
	return f.db.Where("id = ?", fileID).Delete(&po.File{}).Error
}

func (f *FileRepo) GetByID(fileID string) (*entity.File, error) {
	model := &po.File{}
	if err := f.db.Where("id = ?", fileID).First(model).Error; err != nil {
		return nil, err
	}

	file := converter.FilePOToEntity(model)

	return file, nil
}
