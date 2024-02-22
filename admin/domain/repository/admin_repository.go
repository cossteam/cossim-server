package repository

import "github.com/cossim/coss-server/admin/domain/entity"

type AdminRepository interface {
	InsertAdmin(admin *entity.Admin) error
	InsertAndUpdateAdmin(admin *entity.Admin) error
	GetAdminByID(id uint) (*entity.Admin, error)
}
