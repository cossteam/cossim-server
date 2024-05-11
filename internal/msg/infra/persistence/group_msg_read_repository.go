package persistence

import (
	"github.com/cossim/coss-server/internal/msg/domain/entity"
	"github.com/cossim/coss-server/internal/msg/domain/repository"
	"gorm.io/gorm"
)

var _ repository.GroupMsgReadRepository = &GroupMsgReadRepo{}

type GroupMsgReadRepo struct {
	db *gorm.DB
}

func NewGroupMsgReadRepo(db *gorm.DB) *GroupMsgReadRepo {
	return &GroupMsgReadRepo{
		db: db,
	}
}

func (g *GroupMsgReadRepo) GetGroupMsgReadByMsgID(msgId uint) ([]*entity.GroupMessageRead, error) {
	var groupMsgReads []*entity.GroupMessageRead
	err := g.db.Table("group_message_read").Where("msg_id = ?", msgId).Find(&groupMsgReads).Error
	return groupMsgReads, err
}

func (g *GroupMsgReadRepo) SetGroupMsgReadByMsgs(read []*entity.GroupMessageRead) error {
	for _, r := range read {
		err := g.db.
			Where(entity.GroupMessageRead{MsgID: r.MsgID, DialogID: r.DialogID, GroupID: r.GroupID, UserID: r.UserID}).
			Assign(entity.GroupMessageRead{ReadAt: r.ReadAt}).
			FirstOrCreate(r).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *GroupMsgReadRepo) GetGroupMsgReadUserIdsByMsgID(msgId uint) ([]string, error) {
	var userIds []string
	err := g.db.Model(entity.GroupMessageRead{}).Where("msg_id = ?", msgId).Pluck("distinct(user_id)", &userIds).Error
	return userIds, err
}

func (g *GroupMsgReadRepo) GetGroupMsgReadByMsgIDAndUserID(msgId uint, userId string) (*entity.GroupMessageRead, error) {
	var groupMsgRead entity.GroupMessageRead
	err := g.db.
		Where(entity.GroupMessageRead{MsgID: uint(msgId), UserID: userId}).
		First(&groupMsgRead).Error
	return &groupMsgRead, err
}

func (g *GroupMsgReadRepo) GetGroupMsgUserReadIdsByDialogID(dialogID uint, userId string) ([]uint, error) {
	var msgIds []uint
	err := g.db.Model(entity.GroupMessageRead{}).Where("dialog_id = ? and user_id = ?", dialogID, userId).Pluck("distinct(msg_id)", &msgIds).Error
	return msgIds, err
}
