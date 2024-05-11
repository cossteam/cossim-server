package converter

import (
	"github.com/cossim/coss-server/internal/group/domain/entity"
	"github.com/cossim/coss-server/internal/group/infra/persistence/po"
)

func GroupEntityToPO(e *entity.Group) *po.Group {
	var m *po.Group
	//if err := e.Validate(); err != nil {
	//	return err
	//}
	m.ID = e.ID
	m.Type = uint(e.Type)
	m.Status = uint(e.Status)
	m.MaxMembersLimit = e.MaxMembersLimit
	m.CreatorID = e.CreatorID
	m.Name = e.Name
	m.Avatar = e.Avatar
	m.SilenceTime = e.SilenceTime
	m.JoinApprove = e.JoinApprove
	m.Encrypt = e.Encrypt
	return m
}

func UserPOToEntity(po *po.Group) *entity.Group {
	return &entity.Group{
		ID:              po.ID,
		CreatedAt:       po.CreatedAt,
		Type:            entity.Type(po.Type),
		Status:          entity.Status(po.Status),
		MaxMembersLimit: po.MaxMembersLimit,
		CreatorID:       po.CreatorID,
		Name:            po.Name,
		Avatar:          po.Avatar,
		SilenceTime:     po.SilenceTime,
		JoinApprove:     po.JoinApprove,
		Encrypt:         po.Encrypt,
	}
}
