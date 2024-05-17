package converter

import (
	"github.com/cossim/coss-server/internal/admin/domain/entity"
	"github.com/cossim/coss-server/internal/admin/infra/persistence/po"
)

func AdminEntityToPO(e *entity.Admin) *po.Admin {
	return &po.Admin{
		ID:        e.ID,
		UserId:    e.UserId,
		Status:    uint(e.Status),
		CreatedAt: e.CreatedAt,
		Role:      uint(e.Role),
	}
}

func AdminPOToEntity(po *po.Admin) *entity.Admin {
	return &entity.Admin{
		ID:        po.ID,
		UserId:    po.UserId,
		Role:      entity.Role(po.Role),
		Status:    entity.AdminStatus(po.Status),
		CreatedAt: po.CreatedAt,
	}
}
