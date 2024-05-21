package converter

import (
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/infra/persistence/po"
)

func GroupJoinRequestEntityToPo(e *entity.GroupJoinRequest) *po.GroupJoinRequest {
	m := &po.GroupJoinRequest{}
	m.ID = e.ID
	m.GroupID = e.GroupID
	m.InviterAt = e.InviterAt
	m.ExpiredAt = e.ExpiredAt
	m.UserID = e.UserID
	m.Inviter = e.Inviter
	m.Remark = e.Remark
	m.OwnerID = e.OwnerID
	m.Status = uint8(e.Status)
	return m
}

func GroupJoinRequestPoToEntity(po *po.GroupJoinRequest) *entity.GroupJoinRequest {
	e := &entity.GroupJoinRequest{}
	e.ID = po.ID
	e.CreatedAt = po.CreatedAt
	e.InviterAt = po.InviterAt
	e.ExpiredAt = po.ExpiredAt
	e.GroupID = po.GroupID
	e.UserID = po.UserID
	e.Inviter = po.Inviter
	e.Remark = po.Remark
	e.OwnerID = po.OwnerID
	e.Status = entity.RequestStatus(po.Status)
	return e
}
