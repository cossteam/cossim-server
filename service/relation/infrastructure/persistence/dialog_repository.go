package persistence

import (
	"github.com/cossim/coss-server/pkg/utils/time"
	"github.com/cossim/coss-server/service/relation/domain/entity"
	"gorm.io/gorm"
)

type DialogRepo struct {
	db *gorm.DB
}

func NewDialogRepo(db *gorm.DB) *DialogRepo {
	return &DialogRepo{db: db}
}

// 创建对话
func (g *DialogRepo) CreateDialog(ownerID string, dialogType entity.DialogType, groupID uint) (*entity.Dialog, error) {

	dialog := &entity.Dialog{
		OwnerId: ownerID,
		Type:    dialogType,
		GroupId: groupID,
	}
	if err := g.db.Create(dialog).Error; err != nil {
		return nil, err
	}

	return dialog, nil
}

// 加入对话
func (g *DialogRepo) JoinDialog(dialogID uint, userID string) (*entity.DialogUser, error) {

	dialogUser := &entity.DialogUser{
		DialogId: dialogID,
		UserId:   userID,
		IsShow:   int(entity.IsShow), // 默认已加入
	}

	if err := g.db.Create(dialogUser).Error; err != nil {
		return nil, err
	}

	return dialogUser, nil
}

func (g *DialogRepo) JoinBatchDialog(dialogID uint, userIDs []string) ([]*entity.DialogUser, error) {
	dialogUsers := make([]*entity.DialogUser, len(userIDs))
	for i, userID := range userIDs {
		dialogUsers[i] = &entity.DialogUser{
			DialogId: dialogID,
			UserId:   userID,
			IsShow:   int(entity.IsShow), // 默认已加入
		}
	}

	if err := g.db.CreateInBatches(dialogUsers, len(userIDs)).Error; err != nil {
		return nil, err
	}

	return dialogUsers, nil
}

// 查询用户对话列表
func (g *DialogRepo) GetUserDialogs(userID string) ([]uint, error) {
	var dialogs []uint
	if err := g.db.Model(&entity.DialogUser{}).
		Where("user_id = ? AND is_show = ?", userID, entity.IsShow).
		Pluck("dialog_id", &dialogs).Error; err != nil {
		return nil, err
	}
	return dialogs, nil
}

func (g *DialogRepo) GetDialogsByIDs(dialogIDs []uint) ([]*entity.Dialog, error) {
	var dialogUsers []*entity.Dialog
	if err := g.db.Model(&entity.Dialog{}).Where("id IN (?)", dialogIDs).Find(&dialogUsers).Error; err != nil {
		return nil, err
	}
	return dialogUsers, nil
}

func (g *DialogRepo) GetDialogUsersByDialogID(dialogID uint) ([]*entity.DialogUser, error) {
	var dialogUsers []*entity.DialogUser
	if err := g.db.Model(&entity.DialogUser{}).Where("dialog_id =?", dialogID).Find(&dialogUsers).Error; err != nil {
		return nil, err
	}
	return dialogUsers, nil
}
func (g *DialogRepo) GetDialogUserByDialogIDAndUserID(dialogID uint, userID string) (*entity.DialogUser, error) {
	var DialogUser *entity.DialogUser
	if err := g.db.Model(&entity.DialogUser{}).Where("dialog_id = ? AND user_id = ?", dialogID, userID).First(&DialogUser).Error; err != nil {
		return nil, err
	}
	return DialogUser, nil
}

func (g *DialogRepo) GetDialogByGroupId(groupId uint) (*entity.Dialog, error) {
	var dialog *entity.Dialog
	if err := g.db.Model(&entity.Dialog{}).Where("group_id = ?", groupId).First(&dialog).Error; err != nil {
		return nil, err
	}
	return dialog, nil
}

func (g *DialogRepo) DeleteDialogByIds(dialogIDs []uint) error {
	//return g.db.Model(&entity.Dialog{}).Where("id IN (?)", dialogIDs).Update("deleted_at", time.Unix(time.Now(), 0).Format(time.DateTime)).Error
	return g.db.Model(&entity.Dialog{}).Where("id IN (?)", dialogIDs).Update("deleted_at", time.Now()).Error
}

func (g *DialogRepo) DeleteDialogByDialogID(dialogID uint) error {
	//return g.db.Model(&entity.Dialog{}).Where("id = ?", dialogID).Update("deleted_at", time.Unix(time.Now(), 0).Format(time.DateTime)).Error
	return g.db.Model(&entity.Dialog{}).Where("id = ?", dialogID).Update("deleted_at", time.Now()).Error
}

func (g *DialogRepo) DeleteDialogUserByDialogID(dialogID uint) error {
	//return g.db.Model(&entity.DialogUser{}).Where("dialog_id = ?", dialogID).Unscoped().Update("deleted_at", time.Now()).Error
	return g.db.Model(&entity.DialogUser{}).Where("dialog_id = ?", dialogID).Unscoped().Update("deleted_at", time.Now()).Error
}

func (g *DialogRepo) DeleteDialogUserByDialogIDAndUserID(dialogID uint, userIDs []string) error {
	//return g.db.Model(&entity.DialogUser{}).Where("dialog_id = ? AND user_id = ?", dialogID, userID).Update("deleted_at", time.Now()).Error
	return g.db.Model(&entity.DialogUser{}).Where("dialog_id = ? AND user_id IN (?)", dialogID, userIDs).Update("deleted_at", time.Now()).Error
}

func (g *DialogRepo) GetDialogByGroupIds(groupIds []uint) ([]*entity.Dialog, error) {
	var dialogs []*entity.Dialog
	if err := g.db.Model(&entity.Dialog{}).Where("group_id IN (?)", groupIds).Find(&dialogs).Error; err != nil {
		return nil, err
	}
	return dialogs, nil
}

func (g *DialogRepo) UpdateDialogByDialogID(dialogID uint, updateFields map[string]interface{}) error {
	return g.db.Model(&entity.Dialog{}).Where("id = ?", dialogID).Unscoped().Updates(updateFields).Error
}

func (g *DialogRepo) UpdateDialogUserByDialogID(dialogID uint, updateFields map[string]interface{}) error {
	return g.db.Model(&entity.DialogUser{}).Where("dialog_id = ?", dialogID).Unscoped().Updates(updateFields).Error
}

func (g *DialogRepo) UpdateDialogUserByDialogIDAndUserID(dialogID uint, userID string, updateFields map[string]interface{}) error {
	return g.db.Model(&entity.DialogUser{}).Where("dialog_id = ? AND user_id = ?", dialogID, userID).Updates(updateFields).Error
}
