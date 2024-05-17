package repository

import (
	"context"
	"github.com/cossim/coss-server/internal/admin/domain/entity"
)

type AdminRepository interface {
	InsertAdmin(ctx context.Context, admin *entity.Admin) error
	InsertAndUpdateAdmin(ctx context.Context, admin *entity.Admin) error
	GetAdminByID(ctx context.Context, id uint) (*entity.Admin, error)
	Find(ctx context.Context, query *entity.Query) ([]*entity.Admin, error)
}
