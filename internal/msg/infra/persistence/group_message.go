package persistence

import (
	"github.com/cossim/coss-server/internal/msg/api/grpc/dataTransformers"
	"github.com/cossim/coss-server/internal/msg/domain/entity"
	"github.com/cossim/coss-server/internal/msg/domain/repository"
	"github.com/cossim/coss-server/internal/msg/infra/persistence/converter"
	"github.com/cossim/coss-server/internal/msg/infra/persistence/po"
	"github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

var _ repository.GroupMessageRepository = &GroupMsgRepo{}

type GroupMsgRepo struct {
	db *gorm.DB
}

func NewGroupMsgRepo(db *gorm.DB) *GroupMsgRepo {
	return &GroupMsgRepo{db: db}
}

func (g *GroupMsgRepo) GetGroupUnreadMsgList(dialogId uint, msgIds []uint) ([]*entity.GroupMessage, error) {
	var msgList []*po.GroupMessage
	err := g.db.Model(&po.GroupMessage{}).
		Where("id NOT IN (?) AND dialog_id = ? AND deleted_at = 0", msgIds, dialogId).
		Find(&msgList).Error
	if err != nil {
		return nil, err
	}
	resp := converter.GroupMessagePOToEntityList(msgList)
	return resp, nil
}

func (g *GroupMsgRepo) InsertGroupMessage(msg *entity.GroupMessage) (*entity.GroupMessage, error) {
	gm := converter.GroupMessageEntityToPO(msg)
	if err := g.db.Create(gm).Error; err != nil {
		return nil, err
	}

	newMsg := converter.GroupMessagePOToEntity(gm)
	return newMsg, nil
}

func (g *GroupMsgRepo) GetLastMsgsForGroupsWithIDs(groupIDs []uint) ([]*entity.GroupMessage, error) {
	var groupMessages []*po.GroupMessage

	result := g.db.
		Where("group_id IN (?)", groupIDs).
		Group("group_id").
		Order("created_at DESC").
		Find(&groupMessages)

	if result.Error != nil {
		return nil, result.Error
	}

	resp := converter.GroupMessagePOToEntityList(groupMessages)
	return resp, nil
}

func (g *GroupMsgRepo) UpdateGroupMessage(msg *entity.GroupMessage) error {
	po := converter.GroupMessageEntityToPO(msg)
	err := g.db.Model(&po).Updates(po).Error
	return err
}

func (g *GroupMsgRepo) LogicalDeleteGroupMessage(msgId uint) error {
	err := g.db.Model(&po.GroupMessage{}).Where("id = ?", msgId).Update("deleted_at", time.Now()).Error
	return err
}

func (g *GroupMsgRepo) PhysicalDeleteGroupMessage(msgId uint) error {
	err := g.db.Delete(&po.GroupMessage{}, msgId).Error
	return err
}

func (g *GroupMsgRepo) GetGroupMsgByID(msgId uint) (*entity.GroupMessage, error) {
	msg := &po.GroupMessage{}
	if err := g.db.Model(&po.GroupMessage{}).Where("id = ? AND deleted_at = 0", msgId).First(msg).Error; err != nil {
		return nil, err
	}
	resp := converter.GroupMessagePOToEntity(msg)
	return resp, nil
}

func (g *GroupMsgRepo) UpdateGroupMsgColumn(msgId uint, column string, value interface{}) error {
	return g.db.Model(&po.GroupMessage{}).Where("id = ?", msgId).Update(column, value).Error
}

func (g *GroupMsgRepo) GetGroupMsgLabelByDialogId(dialogId uint) ([]*entity.GroupMessage, error) {
	var groupMessages []*po.GroupMessage
	if err := g.db.Model(&po.GroupMessage{}).Where("dialog_id = ? AND is_label = ? AND deleted_at = 0", dialogId, entity.IsLabel).Find(&groupMessages).Error; err != nil {
		return nil, err
	}
	resp := converter.GroupMessagePOToEntityList(groupMessages)
	return resp, nil
}

func (g *GroupMsgRepo) PhysicalDeleteGroupMessages(msgIds []uint) error {
	return g.db.Delete(&po.GroupMessage{}, msgIds).Error
}

func (g *GroupMsgRepo) LogicalDeleteGroupMessages(msgIds []uint) error {
	return g.db.Model(&po.GroupMessage{}).
		Where("id IN (?)", msgIds).
		Update("deleted_at", time.Now()).Error
}

func (g *GroupMsgRepo) GetGroupMsgIdAfterMsgList(dialogId uint, msgIds uint) ([]*entity.GroupMessage, int64, error) {
	var groupMessages []*po.GroupMessage
	var total int64

	query := g.db.Model(&po.GroupMessage{}).
		Where("dialog_id = ? AND id > ? AND deleted_at = 0", dialogId, msgIds).
		Order("id desc")
	err := g.db.Model(&po.GroupMessage{}).
		Where("dialog_id = ? AND deleted_at = 0", dialogId, msgIds).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Find(&groupMessages).Error
	if err != nil {
		return nil, 0, err
	}

	resp := converter.GroupMessagePOToEntityList(groupMessages)
	return resp, total, err
}

func (g *GroupMsgRepo) GetGroupMsgList(list dataTransformers.GroupMsgList) (*dataTransformers.GroupMsgListResponse, error) {
	response := &dataTransformers.GroupMsgListResponse{}

	query := g.db.Model(&po.GroupMessage{}).
		Where("dialog_id = ? AND deleted_at = 0", list.DialogId)

	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return response, err
	}
	if list.UserID != "" {
		query = query.Where("user_id = ?", list.UserID)
	}

	if list.Content != "" {
		query = query.Where("content LIKE ?", "%"+list.Content+"%")
	}

	if list.MsgType != 0 {
		query = query.Where("msg_type = ?", list.MsgType)
	}

	var groupMessages []*po.GroupMessage
	err = query.Order("id DESC").
		Limit(list.PageSize).
		Offset(list.PageSize * (list.PageNumber - 1)).
		Find(&groupMessages).
		Error
	if err != nil {
		return response, err
	}

	resp := converter.GroupMessagePOToEntityList(groupMessages)

	response.GroupMessages = resp
	response.Total = int32(total)
	response.CurrentPage = int32(list.PageNumber)

	return response, nil
}

func (g *GroupMsgRepo) GetGroupMsgsByIDs(msgIds []uint) ([]*entity.GroupMessage, error) {
	if len(msgIds) == 0 {
		return nil, nil
	}
	var groupMessages []*po.GroupMessage
	err := g.db.Model(&po.GroupMessage{}).
		Where("id IN (?) AND deleted_at = 0", msgIds).
		Find(&groupMessages).Error
	if err != nil {
		return nil, err
	}

	resp := converter.GroupMessagePOToEntityList(groupMessages)
	return resp, err
}

func (g *GroupMsgRepo) GetGroupMsgIdsByDialogID(dialogID uint) ([]uint, error) {
	var msgIds []uint
	err := g.db.Model(&po.GroupMessage{}).
		Where("dialog_id = ? AND deleted_at = 0", dialogID).
		Select("id").
		Where("type NOT IN (?)", []entity.UserMessageType{entity.MessageTypeLabel, entity.MessageTypeDelete}).
		Find(&msgIds).Error
	if err != nil {
		return nil, err
	}

	return msgIds, err
}

func (g *GroupMsgRepo) DeleteGroupMessagesByDialogID(dialogId uint) error {
	return g.db.Model(&po.GroupMessage{}).Where("dialog_id = ?", dialogId).Update("deleted_at", time.Now()).Error
}

func (g *GroupMsgRepo) UpdateGroupMsgColumnByDialogId(dialogId uint, column string, value interface{}) error {
	return g.db.Model(&po.GroupMessage{}).Where("dialog_id = ?", dialogId).Update(column, value).Error
}

func (g *GroupMsgRepo) PhysicalDeleteGroupMessagesByDialogID(dialogId uint) error {
	return g.db.Where("dialog_id = ?", dialogId).Delete(&po.GroupMessage{}).Error
}

func (g *GroupMsgRepo) GetGroupDialogLastMsgs(dialogId uint, pageNumber, pageSize int) ([]*entity.GroupMessage, int64, error) {
	var groupMessages []*po.GroupMessage
	var total int64
	query := g.db.Model(&po.GroupMessage{}).
		Where("dialog_id = ? AND deleted_at = 0", dialogId).
		Order("id DESC")

	err := g.db.Model(&po.GroupMessage{}).
		Where("dialog_id = ? AND deleted_at = 0", dialogId).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Limit(pageSize).
		Offset(pageSize * (pageNumber - 1)).
		Find(&groupMessages).Error
	if err != nil {
		return nil, 0, err
	}

	resp := converter.GroupMessagePOToEntityList(groupMessages)
	return resp, total, err
}

func (g *GroupMsgRepo) GetLastGroupMsgsByDialogIDs(dialogIds []uint) ([]*entity.GroupMessage, error) {
	groupMessages := make([]*po.GroupMessage, 0)
	for _, dialogId := range dialogIds {
		var lastMsg po.GroupMessage
		g.db.Where("dialog_id = ?  AND deleted_at = 0", dialogId).Order("created_at DESC").First(&lastMsg)
		if lastMsg.ID != 0 {
			groupMessages = append(groupMessages, &lastMsg)
		}
	}

	resp := converter.GroupMessagePOToEntityList(groupMessages)
	return resp, nil
}

func (g *GroupMsgRepo) GetGroupMsgIdBeforeMsgList(dialogId uint, msgId uint, pageSize int) ([]*entity.GroupMessage, int32, error) {
	var groupMessages []*po.GroupMessage
	var total int64
	err := g.db.Model(&po.GroupMessage{}).Where("dialog_id = ? AND deleted_at = 0", dialogId).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = g.db.Model(&po.GroupMessage{}).
		Where("dialog_id = ? AND id < ? AND deleted_at = 0", dialogId, msgId).Order("id DESC").
		Limit(pageSize).
		Find(&groupMessages).Error
	if err != nil {
		return nil, 0, err
	}

	resp := converter.GroupMessagePOToEntityList(groupMessages)
	return resp, int32(total), err
}
