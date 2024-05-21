package converter

import (
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/infra/persistence/po"
)

func GroupRelationEntityToPo(e *entity.GroupRelation) *po.GroupRelation {
	m := &po.GroupRelation{}
	m.ID = e.ID
	m.GroupID = e.GroupID
	m.Identity = uint8(e.Identity)
	m.EntryMethod = uint8(e.EntryMethod)
	m.JoinedAt = e.JoinedAt
	m.MuteEndTime = e.MuteEndTime
	m.UserID = e.UserID
	m.Inviter = e.Inviter
	m.Remark = e.Remark
	m.Label = e.Label
	m.SilentNotification = e.SilentNotification
	m.PrivacyMode = e.PrivacyMode
	return nil
}

func GroupRelationPoToEntity(po *po.GroupRelation) *entity.GroupRelation {
	return &entity.GroupRelation{
		ID:                 po.ID,
		CreatedAt:          po.CreatedAt,
		GroupID:            po.GroupID,
		Identity:           entity.GroupIdentity(po.Identity),
		EntryMethod:        entity.EntryMethod(po.EntryMethod),
		JoinedAt:           po.JoinedAt,
		MuteEndTime:        po.MuteEndTime,
		UserID:             po.UserID,
		Inviter:            po.Inviter,
		Remark:             po.Remark,
		Label:              po.Label,
		SilentNotification: po.SilentNotification,
		PrivacyMode:        po.PrivacyMode,
	}
}
