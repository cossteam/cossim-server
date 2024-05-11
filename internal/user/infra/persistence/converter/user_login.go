package converter

import (
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/infra/persistence/po"
)

func UserLoginEntityToPO(e *entity.UserLogin) *po.UserLogin {
	return &po.UserLogin{
		BaseModel: po.BaseModel{
			ID:        e.ID,
			CreatedAt: e.CreatedAt,
			//UpdatedAt: e.UpdatedAt,
			//DeletedAt: e.DeletedAt,
		},
		UserId:      e.UserID,
		LoginCount:  e.LoginCount,
		LastAt:      e.LastAt,
		Token:       e.Token,
		DriverId:    e.DriverID,
		DriverToken: e.DriverToken,
		Platform:    e.Platform,
	}
}

func UserLoginPOToEntity(po *po.UserLogin) *entity.UserLogin {
	return &entity.UserLogin{
		ID:          po.ID,
		CreatedAt:   po.CreatedAt,
		UserID:      po.UserId,
		LoginCount:  po.LoginCount,
		LastAt:      po.LastAt,
		Token:       po.Token,
		DriverID:    po.DriverId,
		DriverToken: po.DriverToken,
		Platform:    po.Platform,
		DriverType:  "",
		ClientIP:    "",
		Rid:         "",
	}
}
