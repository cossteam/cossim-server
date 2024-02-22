package persistence

import (
	"github.com/cossim/coss-server/admin/domain/entity"
	"gorm.io/gorm"
)

type AdminRepository struct {
	db *gorm.DB
}

func NewAdminRepo(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (a AdminRepository) InsertAdmin(admin *entity.Admin) error {
	//TODO implement me
	panic("implement me")
}

func (a AdminRepository) InsertAndUpdateAdmin(admin *entity.Admin) error {
	return a.db.Where(entity.Admin{UserId: admin.UserId}).Assign(admin).FirstOrCreate(admin).Error
}

func (a AdminRepository) GetAdminByID(id uint) (*entity.Admin, error) {
	var admin *entity.Admin
	if err := a.db.Where("id = ?", id).First(&admin).Error; err != nil {
		return nil, err
	}
	return admin, nil
}
