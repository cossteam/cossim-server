package persistence

import (
	"context"
	"github.com/cossim/coss-server/internal/admin/domain/entity"
	"github.com/cossim/coss-server/internal/admin/infra/persistence/converter"
	"github.com/cossim/coss-server/internal/admin/infra/persistence/po"
	"gorm.io/gorm"
)

type AdminRepository struct {
	db *gorm.DB
}

func NewAdminRepo(db *gorm.DB) *AdminRepository {
	return &AdminRepository{db: db}
}

func (a AdminRepository) InsertAdmin(ctx context.Context, admin *entity.Admin) error {
	//TODO implement me
	panic("implement me")
}

func (a AdminRepository) InsertAndUpdateAdmin(ctx context.Context, admin *entity.Admin) error {
	ad := converter.AdminEntityToPO(admin)

	return a.db.Where(po.Admin{UserId: ad.UserId}).Assign(ad).FirstOrCreate(ad).Error
}

func (a AdminRepository) GetAdminByID(ctx context.Context, id uint) (*entity.Admin, error) {
	var ad *po.Admin
	if err := a.db.Where("id = ?", id).First(ad).Error; err != nil {
		return nil, err
	}
	admin := converter.AdminPOToEntity(ad)
	return admin, nil
}

func (a AdminRepository) Find(ctx context.Context, query *entity.Query) ([]*entity.Admin, error) {
	var admins []po.Admin

	c := a.db.Model(&po.Admin{})
	if query.UserId != nil {
		c = c.Where("user_id = ?", *query.UserId)
	}
	if query.Role != nil {
		c = c.Where("role = ?", *query.Role)
	}
	if query.Status != nil {
		c = c.Where("status = ?", *query.Status)
	}
	if query.ID != nil {
		c = c.Where("id = ?", *query.ID)
	}
	if err := c.Find(&admins).Error; err != nil {
		return nil, err
	}

	var adminsEntity []*entity.Admin
	for _, admin := range admins {
		adminsEntity = append(adminsEntity, converter.AdminPOToEntity(&admin))
	}
	return adminsEntity, nil

}
