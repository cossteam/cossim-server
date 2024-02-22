package repository

import (
	"github.com/cossim/coss-server/service/relation/domain/entity"
)

type DialogRepository interface {
	CreateDialog(ownerID string, dialogType entity.DialogType, groupID uint) (*entity.Dialog, error)
	JoinDialog(dialogID uint, userID string) (*entity.DialogUser, error)
	JoinBatchDialog(dialogID uint, userIDs []string) ([]*entity.DialogUser, error)
	GetUserDialogs(userID string) ([]uint, error)
	GetDialogsByIDs(dialogIDs []uint) ([]*entity.Dialog, error)
	GetDialogById(dialogID uint) (*entity.Dialog, error)

	GetDialogUsersByDialogID(dialogID uint) ([]*entity.DialogUser, error)
	GetDialogUserByDialogIDAndUserID(dialogID uint, userID string) (*entity.DialogUser, error)
	GetDialogByGroupId(groupId uint) (*entity.Dialog, error)
	GetDialogByGroupIds(groupIds []uint) ([]*entity.Dialog, error)

	DeleteDialogByIds(dialogIDs []uint) error
	RealDeleteDialogById(dialogID uint) error
	DeleteDialogByDialogID(dialogID uint) error
	DeleteDialogUserByDialogID(dialogID uint) error
	DeleteDialogUserByDialogIDAndUserID(dialogID uint, userID []string) error

	// UpdateDialogByDialogID 根据会话ID更新会话信息
	UpdateDialogByDialogID(dialogID uint, updateFields map[string]interface{}) error
	// UpdateDialogUserByDialogID 根据会话ID更新会话所有用户信息
	UpdateDialogUserByDialogID(dialogID uint, updateFields map[string]interface{}) error
	// UpdateDialogUserByDialogIDAndUserID 根据会话ID和用户ID更新会话成员信息
	UpdateDialogUserByDialogIDAndUserID(dialogID uint, userID string, updateFields map[string]interface{}) error
	// UpdateDialogColumnByDialogID 根据会话ID更新会话信息
	UpdateDialogUserColumnByDialogIDAndUserId(dialogID uint, userID string, column string, value interface{}) error
}
