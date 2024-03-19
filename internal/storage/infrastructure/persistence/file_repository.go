package persistence

import (
	"github.com/cossim/coss-server/internal/storage/domain/entity"
	"gorm.io/gorm"
)

type FileRepo struct {
	db *gorm.DB
}

func NewFileRepo(db *gorm.DB) *FileRepo {
	return &FileRepo{db: db}
}

func (f *FileRepo) Create(file *entity.File) error {
	return f.db.Create(file).Error
}

func (f *FileRepo) Update(file *entity.File) error {
	return f.db.Model(&entity.File{}).Where("id = ?", file.ID).Updates(file).Error
}

func (f *FileRepo) Delete(fileID string) error {
	return f.db.Where("id = ?", fileID).Delete(&entity.File{}).Error
}

func (f *FileRepo) GetByID(fileID string) (*entity.File, error) {
	var file entity.File
	if err := f.db.Where("id = ?", fileID).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}
