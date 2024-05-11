package converter

import (
	"github.com/cossim/coss-server/internal/msg/domain/entity"
	"github.com/cossim/coss-server/internal/msg/infra/persistence/po"
)

func GroupMessageReadPOToEntity(gmr *po.GroupMessageRead) *entity.GroupMessageRead {
	return &entity.GroupMessageRead{
		MsgID:    gmr.MsgId,
		DialogID: gmr.DialogId,
		GroupID:  gmr.GroupID,
		ReadAt:   gmr.ReadAt,
		UserID:   gmr.UserID,
		BaseModel: entity.BaseModel{
			ID:        gmr.ID,
			CreatedAt: gmr.CreatedAt,
			UpdatedAt: gmr.UpdatedAt,
			DeletedAt: gmr.DeletedAt,
		},
	}
}

func GroupMessageReadEntityToPO(gmr *entity.GroupMessageRead) *po.GroupMessageRead {
	return &po.GroupMessageRead{
		MsgId:    gmr.MsgID,
		DialogId: gmr.DialogID,
		GroupID:  gmr.GroupID,
		ReadAt:   gmr.ReadAt,
		UserID:   gmr.UserID,
		BaseModel: po.BaseModel{
			ID:        gmr.ID,
			CreatedAt: gmr.CreatedAt,
			UpdatedAt: gmr.UpdatedAt,
			DeletedAt: gmr.DeletedAt,
		},
	}
}
