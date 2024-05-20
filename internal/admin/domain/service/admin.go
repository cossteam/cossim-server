package service

import (
	"context"
	"github.com/cossim/coss-server/internal/admin/domain/entity"
	"github.com/cossim/coss-server/internal/admin/infra/persistence"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"gorm.io/gorm"
)

type AdminDomain interface {
	GetAdminByID(ctx context.Context, id uint) (*entity.Admin, error)
	InsertAdmin(ctx context.Context, admin *entity.Admin) error
	InsertAndUpdateAdmin(ctx context.Context, admin *entity.Admin) error
	Find(ctx context.Context, query *entity.Query) ([]*entity.Admin, error)
}

type AdminDomainImpl struct {
	db *gorm.DB
	ar *persistence.Repositories
}

func NewAdminDomain(db *gorm.DB, cfg *pkgconfig.AppConfig) AdminDomain {
	resp := &AdminDomainImpl{
		db: db,
		ar: persistence.NewRepositories(db),
	}
	resp.ar.Automigrate()
	return resp
}

func (a *AdminDomainImpl) GetAdminByID(ctx context.Context, id uint) (*entity.Admin, error) {
	byID, err := a.ar.Ar.GetAdminByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return byID, nil
}

func (a *AdminDomainImpl) InsertAdmin(ctx context.Context, admin *entity.Admin) error {
	err := a.ar.Ar.InsertAdmin(ctx, admin)
	if err != nil {
		return err
	}
	return nil
}

func (a *AdminDomainImpl) InsertAndUpdateAdmin(ctx context.Context, admin *entity.Admin) error {
	err := a.ar.Ar.InsertAndUpdateAdmin(ctx, admin)
	if err != nil {
		return err
	}
	return nil
}

func (a *AdminDomainImpl) Find(ctx context.Context, query *entity.Query) ([]*entity.Admin, error) {
	find, err := a.ar.Ar.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	return find, nil
}
