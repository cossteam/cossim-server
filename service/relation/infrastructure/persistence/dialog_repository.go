package persistence

import (
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
	//for _, dialogUser := range dialogUsers {
	//	fmt.Println(dialogUser.GroupId)
	//}
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

func (g *DialogRepo) DeleteDialogByIds(dialogIDs []uint) error {
	return g.db.Where("id IN (?)", dialogIDs).Delete(&entity.Dialog{}).Error
}

func (g *DialogRepo) DeleteDialogByDialogID(dialogID uint) error {
	return g.db.Where("id = ?", dialogID).Delete(&entity.Dialog{}).Error
}

func (g *DialogRepo) DeleteDialogUserByDialogID(dialogID uint) error {
	return g.db.Where("dialog_id = ?", dialogID).Unscoped().Delete(&entity.DialogUser{}).Error
}
