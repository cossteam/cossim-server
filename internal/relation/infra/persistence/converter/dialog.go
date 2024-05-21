package converter

import (
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/infra/persistence/po"
)

func DialogEntityToPo(e *entity.Dialog) *po.Dialog {
	m := &po.Dialog{}
	m.ID = e.ID
	m.OwnerId = e.OwnerId
	m.Type = uint8(e.Type)
	m.GroupId = e.GroupId
	return m
}

func DialogEntityPoToEntity(po *po.Dialog) *entity.Dialog {
	return &entity.Dialog{
		ID:        po.ID,
		CreatedAt: po.CreatedAt,
		OwnerId:   po.OwnerId,
		Type:      entity.DialogType(po.Type),
		GroupId:   po.GroupId,
	}
}

func DialogUserEntityToPo(e *entity.DialogUser) *po.DialogUser {
	m := &po.DialogUser{}
	m.ID = e.ID
	m.DialogId = e.DialogId
	m.UserId = e.UserId
	m.IsShow = e.IsShow
	m.TopAt = e.TopAt
	return m
}

func DialogUserPoToEntity(m *po.DialogUser) *entity.DialogUser {
	return &entity.DialogUser{
		ID:        m.ID,
		CreatedAt: m.CreatedAt,
		DialogId:  m.DialogId,
		UserId:    m.UserId,
		IsShow:    m.IsShow,
		TopAt:     m.TopAt,
	}
}
