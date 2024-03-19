package persistence

import (
	"github.com/cossim/coss-server/internal/msg/domain/entity"
	"gorm.io/gorm"
)

type GroupMsgReadRepo struct {
	db *gorm.DB
}

func NewGroupMsgReadRepo(db *gorm.DB) *GroupMsgReadRepo {
	return &GroupMsgReadRepo{
		db: db,
	}
}

func (g GroupMsgReadRepo) GetGroupMsgReadByMsgID(msgId uint32) ([]*entity.GroupMessageRead, error) {
	var groupMsgReads []*entity.GroupMessageRead
	err := g.db.Table("group_message_read").Where("msg_id = ?", msgId).Find(&groupMsgReads).Error
	return groupMsgReads, err
}

func (g GroupMsgReadRepo) SetGroupMsgReadByMsgs(read []*entity.GroupMessageRead) error {
	for _, r := range read {
		err := g.db.
			Where(entity.GroupMessageRead{MsgId: r.MsgId, DialogId: r.DialogId, GroupID: r.GroupID, UserId: r.UserId}).
			Assign(entity.GroupMessageRead{ReadAt: r.ReadAt}).
			FirstOrCreate(r).Error
		if err != nil {
			return err
		}
	}
	return nil
}

func (g GroupMsgReadRepo) GetGroupMsgReadUserIdsByMsgID(msgId uint32) ([]string, error) {
	var userIds []string
	err := g.db.Model(entity.GroupMessageRead{}).Where("msg_id = ?", msgId).Pluck("distinct(user_id)", &userIds).Error
	return userIds, err
}

func (g GroupMsgReadRepo) GetGroupMsgReadByMsgIDAndUserID(msgId uint32, userId string) (*entity.GroupMessageRead, error) {
	var groupMsgRead entity.GroupMessageRead
	err := g.db.
		Where(entity.GroupMessageRead{MsgId: uint(msgId), UserId: userId}).
		First(&groupMsgRead).Error
	return &groupMsgRead, err
}

func (g GroupMsgReadRepo) GetGroupMsgUserReadIdsByDialogID(dialogID uint32, userId string) ([]uint32, error) {
	var msgIds []uint32
	err := g.db.Model(entity.GroupMessageRead{}).Where("dialog_id = ? and user_id = ?", dialogID, userId).Pluck("distinct(msg_id)", &msgIds).Error
	return msgIds, err
}
