package converter

import (
	"github.com/cossim/coss-server/internal/relation/domain/entity"
	"github.com/cossim/coss-server/internal/relation/infra/persistence/po"
)

func UserRelationEntityToPo(u *entity.UserRelation) *po.UserRelation {
	m := &po.UserRelation{}
	m.ID = u.ID
	m.UserID = u.UserID
	m.FriendID = u.FriendID
	m.Status = uint(u.Status)
	m.DialogId = u.DialogId
	m.Remark = u.Remark
	m.Label = u.Label
	m.SilentNotification = u.SilentNotification
	m.OpenBurnAfterReading = u.OpenBurnAfterReading
	m.BurnAfterReadingTimeOut = u.BurnAfterReadingTimeOut
	return nil
}

func UserRelationPoToEntity(po *po.UserRelation) *entity.UserRelation {
	return &entity.UserRelation{
		ID:                      po.ID,
		CreatedAt:               po.CreatedAt,
		UserID:                  po.UserID,
		FriendID:                po.FriendID,
		Status:                  entity.UserRelationStatus(po.Status),
		DialogId:                po.DialogId,
		Remark:                  po.Remark,
		Label:                   po.Label,
		SilentNotification:      po.SilentNotification,
		OpenBurnAfterReading:    po.OpenBurnAfterReading,
		BurnAfterReadingTimeOut: po.BurnAfterReadingTimeOut,
	}
}
