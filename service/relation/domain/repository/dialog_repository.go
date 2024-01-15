package repository

import (
	"github.com/cossim/coss-server/service/relation/domain/entity"
)

type DialogRepository interface {
	CreateDialog(ownerID string, dialogType entity.DialogType, groupID uint) (*entity.Dialog, error)
	JoinDialog(dialogID uint, userID string) (*entity.DialogUser, error)
	GetUserDialogs(userID string) ([]uint, error)
	GetDialogsByIDs(dialogIDs []uint) ([]*entity.Dialog, error)
	GetDialogUsersByDialogID(dialogID uint) ([]*entity.DialogUser, error)
	GetDialogUserByDialogIDAndUserID(dialogID uint, userID string) (*entity.DialogUser, error)
	GetDialogByGroupId(groupId uint) (*entity.Dialog, error)

	DeleteDialogByIds(dialogIDs []uint) error
	DeleteDialogByDialogID(dialogID uint) error
	DeleteDialogUserByDialogID(dialogID uint) error
	DeleteDialogUserByDialogIDAndUserID(dialogID uint, userID string) error
}
