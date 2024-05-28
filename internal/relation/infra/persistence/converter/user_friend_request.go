package converter

import (
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/infra/persistence/po"
)

func UserFriendRequestEntityToPo(e *entity.UserFriendRequest) *po.UserFriendRequest {
	m := &po.UserFriendRequest{}
	m.ID = e.ID
	m.SenderID = e.SenderID
	m.ReceiverID = e.RecipientID
	m.Remark = e.Remark
	m.OwnerID = e.OwnerID
	m.Status = uint8(e.Status)
	m.ExpiredAt = e.ExpiredAt
	return m
}

func UserFriendRequestPoToEntity(po *po.UserFriendRequest) *entity.UserFriendRequest {
	return &entity.UserFriendRequest{
		ID:          po.ID,
		CreatedAt:   po.CreatedAt,
		SenderID:    po.SenderID,
		RecipientID: po.ReceiverID,
		Remark:      po.Remark,
		OwnerID:     po.OwnerID,
		Status:      entity.RequestStatus(po.Status),
		ExpiredAt:   po.ExpiredAt,
	}
}
