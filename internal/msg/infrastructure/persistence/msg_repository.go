package persistence

import (
	"fmt"
	"github.com/cossim/coss-server/internal/msg/api/grpc/dataTransformers"
	"github.com/cossim/coss-server/internal/msg/domain/entity"
	"github.com/cossim/coss-server/internal/msg/domain/repository"
	"github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

var _ repository.MsgRepository = &MsgRepo{}

type MsgRepo struct {
	db *gorm.DB
}

func (g *MsgRepo) GetGroupUnreadMsgList(dialogId uint32, msgIds []uint32) ([]*entity.GroupMessage, error) {
	var msgList []*entity.GroupMessage
	err := g.db.Model(&entity.GroupMessage{}).
		Where("id NOT IN (?) AND dialog_id = ? AND deleted_at = 0", msgIds, dialogId).
		Find(&msgList).Error
	if err != nil {
		return nil, err
	}
	return msgList, nil
}

func (g *MsgRepo) ReadAllUserMsg(uid string, dialogId uint32) error {
	return g.db.Model(&entity.UserMessage{}).
		Where("receive_id = ? AND dialog_id = ? AND deleted_at = 0", uid, dialogId).
		Updates(map[string]interface{}{
			"is_read": entity.IsRead,
			"read_at": time.Now(),
		}).Error
}

func NewMsgRepo(db *gorm.DB) *MsgRepo {
	return &MsgRepo{db: db}
}

func (g *MsgRepo) InsertUserMessage(senderId string, receiverId string, msg string, msgType entity.UserMessageType, replyId uint, dialogId uint, isBurnAfterReading entity.BurnAfterReadingType) (*entity.UserMessage, error) {
	content := &entity.UserMessage{
		SendID:             senderId,
		ReceiveID:          receiverId,
		Content:            msg,
		Type:               msgType,
		ReplyId:            replyId,
		DialogId:           dialogId,
		IsBurnAfterReading: isBurnAfterReading,
	}
	if err := g.db.Save(content).Error; err != nil {
		return nil, err
	}
	return content, nil
}

func (g *MsgRepo) InsertGroupMessage(uid string, groupId uint, msg string, msgType entity.UserMessageType, replyId uint, dialogId uint, isBurnAfterReading entity.BurnAfterReadingType, atUsers []string, atAllUser entity.AtAllUserType) (*entity.GroupMessage, error) {
	content := &entity.GroupMessage{
		UserID:             uid,
		GroupID:            groupId,
		Content:            msg,
		Type:               msgType,
		ReplyId:            replyId,
		DialogId:           dialogId,
		IsBurnAfterReading: isBurnAfterReading,
		AtUsers:            atUsers,
		AtAllUser:          atAllUser,
	}
	if err := g.db.Save(content).Error; err != nil {
		return nil, err
	}
	return content, nil
}

func (g *MsgRepo) GetUserMsgList(userId, friendId string, content string, msgType entity.UserMessageType, pageNumber, pageSize int) ([]entity.UserMessage, int32, int32) {
	var results []entity.UserMessage

	query := g.db.Model(&entity.UserMessage{})
	query = query.Where("(send_id = ? AND receive_id = ?) OR (send_id = ? AND receive_id = ?)", userId, friendId, friendId, userId).Order("created_at DESC")

	if content != "" {
		query = query.Where("content LIKE ?", "%"+content+"%")
	}

	if entity.IsValidMessageType(msgType) {
		query = query.Where("msg_type = ?", msgType)
	}

	offset := (pageNumber - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize).Find(&results)
	var count int64
	// 注意：这里的 query 是一个新的查询，需要重新构建条件
	countQuery := g.db.Model(&entity.UserMessage{})
	countQuery = countQuery.Where("(send_id = ? AND receive_id = ?) OR (send_id = ? AND receive_id = ?)", userId, friendId, friendId, userId)
	if content != "" {
		countQuery = countQuery.Where("content LIKE ?", "%"+content+"%")
	}
	if entity.IsValidMessageType(msgType) {
		countQuery = countQuery.Where("msg_type = ?", msgType)
	}
	countQuery.Count(&count)

	return results, int32(count), int32(pageNumber)

}

func (g *MsgRepo) GetLastMsgsForUserWithFriends(userID string, friendIDs []string) ([]*entity.UserMessage, error) {
	var userMessages []*entity.UserMessage

	result := g.db.
		Where("(send_id = ? AND receive_id IN (?)) OR (send_id IN (?) AND receive_id = ?)",
			userID, friendIDs, friendIDs, userID).
		Group("receive_id").
		Order("created_at DESC").
		Find(&userMessages)

	if result.Error != nil {
		return nil, result.Error
	}

	return userMessages, nil
}

func (g *MsgRepo) GetLastMsgsForGroupsWithIDs(groupIDs []uint) ([]*entity.GroupMessage, error) {
	var groupMessages []*entity.GroupMessage

	result := g.db.
		Where("group_id IN (?)", groupIDs).
		Group("group_id").
		Order("created_at DESC").
		Find(&groupMessages)

	if result.Error != nil {
		return nil, result.Error
	}

	return groupMessages, nil
}

func (g *MsgRepo) UpdateUserMessage(msg entity.UserMessage) error {
	err := g.db.Model(&msg).Updates(msg).Error
	return err
}

func (g *MsgRepo) UpdateGroupMessage(msg entity.GroupMessage) error {
	err := g.db.Model(&msg).Updates(msg).Error
	return err
}

func (g *MsgRepo) LogicalDeleteUserMessage(msgId uint32) error {
	err := g.db.Model(&entity.UserMessage{}).Where("id = ?", msgId).Update("deleted_at", time.Now()).Error
	return err
}

func (g *MsgRepo) LogicalDeleteGroupMessage(msgId uint32) error {
	err := g.db.Model(&entity.GroupMessage{}).Where("id = ?", msgId).Update("deleted_at", time.Now()).Error
	return err
}

func (g *MsgRepo) PhysicalDeleteUserMessage(msgId uint32) error {
	err := g.db.Delete(&entity.UserMessage{}, msgId).Error
	return err
}

func (g *MsgRepo) PhysicalDeleteGroupMessage(msgId uint32) error {
	err := g.db.Delete(&entity.GroupMessage{}, msgId).Error
	return err
}

func (g *MsgRepo) GetUserMsgByID(msgId uint32) (*entity.UserMessage, error) {
	msg := &entity.UserMessage{}
	if err := g.db.Model(&entity.UserMessage{}).Where("id = ? AND deleted_at = 0", msgId).First(msg).Error; err != nil {
		return nil, err
	}
	return msg, nil
}

func (g *MsgRepo) GetGroupMsgByID(msgId uint32) (*entity.GroupMessage, error) {
	msg := &entity.GroupMessage{}
	if err := g.db.Model(&entity.GroupMessage{}).Where("id = ? AND deleted_at = 0", msgId).First(msg).Error; err != nil {
		return nil, err
	}
	return msg, nil
}

func (g *MsgRepo) UpdateUserMsgColumn(msgId uint32, column string, value interface{}) error {
	return g.db.Model(&entity.UserMessage{}).Where("id = ?", msgId).Update(column, value).Error
}

func (g *MsgRepo) UpdateGroupMsgColumn(msgId uint32, column string, value interface{}) error {
	return g.db.Model(&entity.GroupMessage{}).Where("id = ?", msgId).Update(column, value).Error
}

func (g *MsgRepo) GetUserMsgLabelByDialogId(dialogId uint32) ([]*entity.UserMessage, error) {
	var userMessages []*entity.UserMessage
	if err := g.db.Model(&entity.UserMessage{}).Where("dialog_id = ? AND is_label = ? AND deleted_at = 0", dialogId, entity.IsLabel).Find(&userMessages).Error; err != nil {
		return nil, err
	}
	return userMessages, nil
}

func (g *MsgRepo) GetGroupMsgLabelByDialogId(dialogId uint32) ([]*entity.GroupMessage, error) {
	var groupMessages []*entity.GroupMessage
	if err := g.db.Model(&entity.GroupMessage{}).Where("dialog_id = ? AND is_label = ? AND deleted_at = 0", dialogId, entity.IsLabel).Find(&groupMessages).Error; err != nil {
		return nil, err
	}
	return groupMessages, nil
}

func (g *MsgRepo) SetUserMsgsReadStatus(msgIds []uint32, dialogId uint32) error {
	return g.db.Model(&entity.UserMessage{}).
		Where("id IN (?) AND dialog_id = ? AND deleted_at = 0", msgIds, dialogId).
		Updates(map[string]interface{}{
			"is_read": entity.IsRead,
			"read_at": time.Now(),
		}).Error
}

func (g *MsgRepo) SetUserMsgReadStatus(msgId uint32, isRead entity.ReadType) error {
	dd := g.db.Model(&entity.UserMessage{}).Where("id = ? AND deleted_at = 0", msgId)
	if isRead == entity.IsRead {
		return dd.Updates(map[string]interface{}{
			"is_read": entity.IsRead,
			"read_at": time.Now(),
		}).Error
	} else {
		return dd.Updates(map[string]interface{}{
			"is_read": entity.IsRead,
			"read_at": 0,
		}).Error
	}
}

func (g *MsgRepo) GetUnreadUserMsgs(uid string, dialogId uint32) ([]*entity.UserMessage, error) {
	var userMessages []*entity.UserMessage
	if err := g.db.Model(&entity.UserMessage{}).Where("dialog_id = ? AND is_read = ? AND receive_id = ? AND deleted_at = 0",
		dialogId,
		entity.NotRead,
		uid).
		Where("type NOT IN (?)", []entity.UserMessageType{entity.MessageTypeLabel, entity.MessageTypeDelete}).
		Find(&userMessages).Error; err != nil {
		return nil, err
	}
	return userMessages, nil
}

func (g *MsgRepo) GetBatchUserMsgsBurnAfterReadingMessages(msgIds []uint32, dialogID uint32) ([]*entity.UserMessage, error) {
	var userMessages []*entity.UserMessage
	err := g.db.Model(&entity.UserMessage{}).
		Where("dialog_id = ? AND id IN (?) AND is_burn_after_reading = ?", dialogID, msgIds, entity.IsBurnAfterReading).
		Find(&userMessages).Error
	if err != nil {
		return nil, err
	}
	return userMessages, nil
}

func (g *MsgRepo) PhysicalDeleteUserMessages(msgIds []uint32) error {
	return g.db.Delete(&entity.UserMessage{}, msgIds).Error
}

func (g *MsgRepo) PhysicalDeleteGroupMessages(msgIds []uint32) error {
	return g.db.Delete(&entity.GroupMessage{}, msgIds).Error
}

func (g *MsgRepo) LogicalDeleteUserMessages(msgIds []uint32) error {
	return g.db.Model(&entity.UserMessage{}).
		Where("id IN (?)", msgIds).
		Update("deleted_at", time.Now()).Error
}

func (g *MsgRepo) LogicalDeleteGroupMessages(msgIds []uint32) error {
	return g.db.Model(&entity.GroupMessage{}).
		Where("id IN (?)", msgIds).
		Update("deleted_at", time.Now()).Error
}

func (g *MsgRepo) GetUserMsgIdAfterMsgList(dialogId uint32, msgIds uint32) ([]*entity.UserMessage, error) {
	var userMessages []*entity.UserMessage
	err := g.db.Model(&entity.UserMessage{}).
		Where("dialog_id = ? AND id > ? AND deleted_at = 0", dialogId, msgIds).
		Order("id ASC").
		Find(&userMessages).Error
	return userMessages, err
}

func (g *MsgRepo) GetGroupMsgIdAfterMsgList(dialogId uint32, msgIds uint32) ([]*entity.GroupMessage, error) {
	var groupMessages []*entity.GroupMessage
	err := g.db.Model(&entity.GroupMessage{}).
		Where("dialog_id = ? AND id > ? AND deleted_at = 0", dialogId, msgIds).
		Order("id ASC").
		Find(&groupMessages).Error
	return groupMessages, err
}

func (g *MsgRepo) GetGroupMsgList(list dataTransformers.GroupMsgList) (*dataTransformers.GroupMsgListResponse, error) {
	response := &dataTransformers.GroupMsgListResponse{}

	query := g.db.Model(&entity.GroupMessage{}).
		Where("group_id = ? AND deleted_at = 0", list.GroupID)

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

	var groupMessages []entity.GroupMessage
	err = query.Order("id DESC").
		Limit(list.PageSize).
		Offset(list.PageSize * (list.PageNumber - 1)).
		Find(&groupMessages).
		Error
	if err != nil {
		return response, err
	}

	response.GroupMessages = groupMessages
	response.Total = int32(total)
	response.CurrentPage = int32(list.PageNumber)

	return response, nil
}

func (g *MsgRepo) GetGroupMsgsByIDs(msgIds []uint32) ([]*entity.GroupMessage, error) {
	var groupMessages []*entity.GroupMessage
	err := g.db.Model(&entity.GroupMessage{}).
		Where("id IN (?) AND deleted_at = 0", msgIds).
		Find(&groupMessages).Error

	return groupMessages, err
}

func (g *MsgRepo) GetGroupMsgIdsByDialogID(dialogID uint32) ([]uint32, error) {
	var msgIds []uint32
	err := g.db.Model(&entity.GroupMessage{}).
		Where("dialog_id = ? AND deleted_at = 0", dialogID).
		Select("id").
		Where("type NOT IN (?)", []entity.UserMessageType{entity.MessageTypeLabel, entity.MessageTypeDelete}).
		Find(&msgIds).Error
	return msgIds, err
}

func (g *MsgRepo) GetUserMsgByIDs(msgIds []uint32) ([]*entity.UserMessage, error) {
	var userMessages []*entity.UserMessage
	err := g.db.Model(&entity.UserMessage{}).
		Where("id IN (?) AND deleted_at = 0", msgIds).
		Find(&userMessages).Error
	return userMessages, err
}

func (g *MsgRepo) InsertUserMessages(message []entity.UserMessage) error {
	return g.db.Create(&message).Error
}

func (g *MsgRepo) DeleteUserMessagesByDialogID(dialogId uint32) error {
	return g.db.Model(&entity.UserMessage{}).Where("dialog_id = ?", dialogId).Update("deleted_at", time.Now()).Error
}

func (g *MsgRepo) DeleteGroupMessagesByDialogID(dialogId uint32) error {
	return g.db.Model(&entity.GroupMessage{}).Where("dialog_id = ?", dialogId).Update("deleted_at", time.Now()).Error
}

func (g *MsgRepo) UpdateUserMsgColumnByDialogId(dialogId uint32, column string, value interface{}) error {
	return g.db.Model(&entity.UserMessage{}).Where("dialog_id = ?", dialogId).Update(column, value).Error
}

func (g *MsgRepo) UpdateGroupMsgColumnByDialogId(dialogId uint32, column string, value interface{}) error {
	return g.db.Model(&entity.GroupMessage{}).Where("dialog_id = ?", dialogId).Update(column, value).Error
}

func (g *MsgRepo) PhysicalDeleteUserMessagesByDialogID(dialogId uint32) error {
	return g.db.Where("dialog_id = ?", dialogId).Delete(&entity.UserMessage{}).Error

}

func (g *MsgRepo) PhysicalDeleteGroupMessagesByDialogID(dialogId uint32) error {
	return g.db.Where("dialog_id = ?", dialogId).Delete(&entity.GroupMessage{}).Error
}

func (g *MsgRepo) GetUserDialogLastMsgs(dialogId uint32, pageNumber, pageSize int) ([]entity.UserMessage, error) {
	var userMessages []entity.UserMessage
	fmt.Println("entity.NotBurnAfterReading => ", entity.NotBurnAfterReading)
	err := g.db.Model(&entity.UserMessage{}).
		Where("dialog_id = ? AND is_burn_after_reading = ? AND deleted_at = 0", dialogId, entity.NotBurnAfterReading).
		Order("updated_at DESC").
		Limit(pageSize).
		Offset(pageSize * (pageNumber - 1)).
		Find(&userMessages).Error
	if err != nil {
		return nil, err
	}
	return userMessages, err
}

func (g *MsgRepo) GetGroupDialogLastMsgs(dialogId uint32, pageNumber, pageSize int) ([]entity.GroupMessage, error) {
	var groupMessages []entity.GroupMessage
	err := g.db.Model(&entity.GroupMessage{}).
		Where("dialog_id = ? AND deleted_at = 0", dialogId).
		Order("updated_at DESC").
		Limit(pageSize).
		Offset(pageSize * (pageNumber - 1)).
		Find(&groupMessages).Error
	if err != nil {
		return nil, err
	}
	return groupMessages, err
}

func (g *MsgRepo) GetLastUserMsgsByDialogIDs(dialogIds []uint) ([]*entity.UserMessage, error) {
	userMessages := make([]*entity.UserMessage, 0)

	for _, dialogId := range dialogIds {
		var lastMsg entity.UserMessage
		g.db.Where("dialog_id = ? AND deleted_at = 0", dialogId).Order("created_at DESC").First(&lastMsg)
		if lastMsg.ID != 0 {
			userMessages = append(userMessages, &lastMsg)
		}
	}
	return userMessages, nil
}

func (g *MsgRepo) GetLastGroupMsgsByDialogIDs(dialogIds []uint) ([]*entity.GroupMessage, error) {
	groupMessages := make([]*entity.GroupMessage, 0)
	for _, dialogId := range dialogIds {
		var lastMsg entity.GroupMessage
		g.db.Where("dialog_id = ?  AND deleted_at = 0", dialogId).Order("created_at DESC").First(&lastMsg)
		if lastMsg.ID != 0 {
			groupMessages = append(groupMessages, &lastMsg)
		}
	}
	return groupMessages, nil
}
