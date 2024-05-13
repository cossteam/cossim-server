package persistence

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/msg/domain/entity"
	"github.com/cossim/coss-server/internal/msg/domain/repository"
	"github.com/cossim/coss-server/internal/msg/infra/persistence/converter"
	"github.com/cossim/coss-server/internal/msg/infra/persistence/po"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils/time"
	"gorm.io/gorm"
)

var _ repository.UserMessageRepository = &UserMsgRepo{}

type UserMsgRepo struct {
	db *gorm.DB
}

func NewUserMsgRepo(db *gorm.DB) *UserMsgRepo {
	return &UserMsgRepo{db: db}
}

func (g *UserMsgRepo) Find(ctx context.Context, query *entity.UserMsgQuery) (*entity.UserMsgQueryResult, error) {
	var messages []*entity.UserMessage
	result := &entity.UserMsgQueryResult{}

	db := g.db.Model(&entity.UserMessage{})

	// 根据查询条件过滤消息 ID
	if len(query.MsgIds) > 0 {
		db = db.Where("id IN (?)", query.MsgIds)
	}

	// 根据对话 ID 过滤消息
	if len(query.DialogIds) > 0 {
		db = db.Where("dialog_id IN (?)", query.DialogIds)
	}

	at := time.Now()
	if query.EndAt <= 0 {
		query.EndAt = at
	}

	// 根据时间范围过滤消息
	if query.EndAt > 0 {
		if query.EndAt > at {
			return nil, code.MyCustomErrorCode.CustomMessage("endAt must be less than or equal to the current time")
		}
		db = db.Where("created_at BETWEEN ? AND ?", query.StartAt, query.EndAt)
	}

	if query.Content != "" {
		db = db.Where("content LIKE ?", "%"+query.Content+"%")
	}

	if query.SendID != "" {
		db = db.Where("send_id = ?", query.SendID)
	}

	if entity.IsValidMessageType(query.MsgType) {
		db = db.Where("msg_type = ?", query.MsgType)
	}

	// 查询总消息数
	var totalCount int64
	if err := db.Count(&totalCount).Error; err != nil {
		return nil, err
	}

	result.TotalCount = totalCount

	// 查询剩余消息数
	if query.PageSize > 0 && totalCount > query.PageNum*query.PageSize {
		result.Remaining = totalCount - query.PageNum*query.PageSize
	}

	// 查询总页码
	if query.PageSize > 0 {
		result.TotalPages = (totalCount + query.PageSize - 1) / query.PageSize
	}

	// 查询当前页码
	result.CurrentPage = query.PageNum

	// 分页
	if query.PageNum > 0 {
		offset := (query.PageNum - 1) * query.PageSize
		db = db.Offset(int(offset))
	}

	if query.PageSize > 0 {
		db = db.Limit(int(query.PageSize))
	}

	if query.Sort == "" {
		query.Sort = "desc"
		db = db.Order(fmt.Sprintf("created_at %s", query.Sort))
	}

	// 执行查询
	err := db.Find(&messages).Error
	if err != nil {
		return nil, err
	}

	result.Messages = messages
	result.ReturnedCount = int64(len(result.Messages))

	return result, nil
}

func (g *UserMsgRepo) ReadAllUserMsg(uid string, dialogId uint) error {
	return g.db.Model(&po.UserMessage{}).
		Where("receive_id = ? AND dialog_id = ? AND deleted_at = 0", uid, dialogId).
		Updates(map[string]interface{}{
			"is_read": entity.IsRead,
			"read_at": time.Now(),
		}).Error
}

func (g *UserMsgRepo) InsertUserMessage(message *entity.UserMessage) (*entity.UserMessage, error) {
	um := converter.UserMessageEntityToPO(message)

	if err := g.db.Create(um).Error; err != nil {
		return nil, err
	}
	entityUser := converter.UserMessagePOToEntity(um)
	return entityUser, nil
}

func (g *UserMsgRepo) GetUserMsgList(dialogId uint, sendId string, content string, msgType entity.UserMessageType, pageNumber, pageSize int) ([]*entity.UserMessage, int32, int32) {
	var results []*po.UserMessage

	query := g.db.Model(&po.UserMessage{})
	query = query.Where("dialog_id = ? ", dialogId).Order("created_at DESC")

	var count int64
	err := query.Count(&count).Error
	if err != nil {
		return nil, 0, 0
	}
	if content != "" {
		query = query.Where("content LIKE ?", "%"+content+"%")
	}
	if sendId != "" {
		query = query.Where("send_id = ?", sendId)
	}
	if entity.IsValidMessageType(msgType) {
		query = query.Where("msg_type = ?", msgType)
	}

	offset := (pageNumber - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize).Find(&results)

	resp := converter.UserMessagePOToEntityList(results)

	return resp, int32(count), int32(pageNumber)
}

func (g *UserMsgRepo) GetLastMsgsForUserWithFriends(userID string, friendIDs []string) ([]*entity.UserMessage, error) {
	var userMessages []*po.UserMessage

	result := g.db.
		Where("(send_id = ? AND receive_id IN (?)) OR (send_id IN (?) AND receive_id = ?)",
			userID, friendIDs, friendIDs, userID).
		Group("receive_id").
		Order("created_at DESC").
		Find(&userMessages)

	if result.Error != nil {
		return nil, result.Error
	}
	resp := converter.UserMessagePOToEntityList(userMessages)

	return resp, nil
}

func (g *UserMsgRepo) UpdateUserMessage(msg *entity.UserMessage) error {
	en := converter.UserMessageEntityToPO(msg)
	err := g.db.Model(&en).Updates(en).Error
	return err
}

func (g *UserMsgRepo) LogicalDeleteUserMessage(msgId uint) error {
	return g.db.Model(&po.UserMessage{}).Where("id = ?", msgId).Update("deleted_at", time.Now()).Error
}

func (g *UserMsgRepo) PhysicalDeleteUserMessage(msgId uint) error {
	return g.db.Delete(&po.UserMessage{}, msgId).Error
}

func (g *UserMsgRepo) GetUserMsgByID(msgId uint) (*entity.UserMessage, error) {
	msg := &po.UserMessage{}

	if err := g.db.Model(&po.UserMessage{}).Where("id = ? AND deleted_at = 0", msgId).First(msg).Error; err != nil {
		return nil, err
	}
	resp := converter.UserMessagePOToEntity(msg)
	return resp, nil
}

func (g *UserMsgRepo) UpdateUserMsgColumn(msgId uint, column string, value interface{}) error {
	return g.db.Model(&po.UserMessage{}).Where("id = ?", msgId).Update(column, value).Error
}

func (g *UserMsgRepo) GetUserMsgLabelByDialogId(dialogId uint) ([]*entity.UserMessage, error) {
	var userMessages []*po.UserMessage
	if err := g.db.Model(&po.UserMessage{}).Where("dialog_id = ? AND is_label = ? AND deleted_at = 0", dialogId, entity.IsLabel).Find(&userMessages).Error; err != nil {
		return nil, err
	}
	resp := converter.UserMessagePOToEntityList(userMessages)
	return resp, nil
}

func (g *UserMsgRepo) SetUserMsgsReadStatus(msgIds []uint, dialogId uint) error {
	return g.db.Model(&po.UserMessage{}).
		Where("id IN (?) AND dialog_id = ? AND deleted_at = 0", msgIds, dialogId).
		Updates(map[string]interface{}{
			"is_read": entity.IsRead,
			"read_at": time.Now(),
		}).Error
}

func (g *UserMsgRepo) SetUserMsgReadStatus(msgId uint, isRead entity.ReadType) error {
	dd := g.db.Model(&po.UserMessage{}).Where("id = ? AND deleted_at = 0", msgId)
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

func (g *UserMsgRepo) GetUnreadUserMsgs(uid string, dialogId uint) ([]*entity.UserMessage, error) {
	var userMessages []*po.UserMessage
	if err := g.db.Model(&po.UserMessage{}).Where("dialog_id = ? AND is_read = ? AND receive_id = ? AND deleted_at = 0",
		dialogId,
		entity.NotRead,
		uid).
		Where("type NOT IN (?)", []entity.UserMessageType{entity.MessageTypeLabel, entity.MessageTypeDelete}).
		Find(&userMessages).Error; err != nil {
		return nil, err
	}
	resp := converter.UserMessagePOToEntityList(userMessages)
	return resp, nil
}

func (g *UserMsgRepo) GetBatchUserMsgsBurnAfterReadingMessages(msgIds []uint, dialogID uint) ([]*entity.UserMessage, error) {
	var userMessages []*po.UserMessage
	err := g.db.Model(&po.UserMessage{}).
		Where("dialog_id = ? AND id IN (?) AND is_burn_after_reading = ?", dialogID, msgIds, true).
		Find(&userMessages).Error
	if err != nil {
		return nil, err
	}
	resp := converter.UserMessagePOToEntityList(userMessages)
	return resp, nil
}

func (g *UserMsgRepo) PhysicalDeleteUserMessages(msgIds []uint) error {
	return g.db.Delete(&po.UserMessage{}, msgIds).Error
}

func (g *UserMsgRepo) LogicalDeleteUserMessages(msgIds []uint) error {
	return g.db.Model(&po.UserMessage{}).
		Where("id IN (?)", msgIds).
		Update("deleted_at", time.Now()).Error
}

func (g *UserMsgRepo) GetUserMsgIdAfterMsgList(dialogId uint, msgIds uint) ([]*entity.UserMessage, int64, error) {
	var userMessages []*po.UserMessage
	var total int64
	query := g.db.Model(&po.UserMessage{}).
		Where("dialog_id = ? AND id > ? AND deleted_at = 0", dialogId, msgIds).
		Order("id desc")
	err := g.db.Model(&po.UserMessage{}).
		Where("dialog_id = ?  AND deleted_at = 0", dialogId).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = query.Find(&userMessages).Error
	if err != nil {
		return nil, 0, err
	}
	resp := converter.UserMessagePOToEntityList(userMessages)
	return resp, total, err
}

func (g *UserMsgRepo) GetUserMsgByIDs(msgIds []uint) ([]*entity.UserMessage, error) {
	var userMessages []*po.UserMessage
	err := g.db.Model(&po.UserMessage{}).
		Where("id IN (?) AND deleted_at = 0", msgIds).
		Find(&userMessages).Error
	if err != nil {
		return nil, err
	}
	resp := converter.UserMessagePOToEntityList(userMessages)
	return resp, err
}

func (g *UserMsgRepo) UpdateUserMsgColumnByDialogId(dialogId uint, column string, value interface{}) error {
	return g.db.Model(&po.UserMessage{}).Where("dialog_id = ?", dialogId).Update(column, value).Error
}

func (g *UserMsgRepo) InsertUserMessages(message []*entity.UserMessage) error {
	msg := converter.UserMessageEntityToPOList(message)
	return g.db.Create(&msg).Error
}

func (g *UserMsgRepo) DeleteUserMessagesByDialogID(dialogId uint) error {
	return g.db.Model(&po.UserMessage{}).Where("dialog_id = ?", dialogId).Update("deleted_at", time.Now()).Error
}

func (g *UserMsgRepo) PhysicalDeleteUserMessagesByDialogID(dialogId uint) error {
	return g.db.Where("dialog_id = ?", dialogId).Delete(&po.UserMessage{}).Error
}

func (g *UserMsgRepo) GetUserDialogLastMsgs(dialogId uint, pageNumber, pageSize int) ([]*entity.UserMessage, int64, error) {
	var userMessages []*po.UserMessage
	var total int64
	query := g.db.Model(&po.UserMessage{}).
		Where("dialog_id = ? AND is_burn_after_reading = ? AND deleted_at = 0", dialogId, false).
		Order("id DESC")

	err := g.db.Model(&po.UserMessage{}).
		Where("dialog_id = ?  AND deleted_at = 0", dialogId).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = query.Limit(pageSize).
		Offset(pageSize * (pageNumber - 1)).
		Find(&userMessages).Error
	if err != nil {
		return nil, 0, err
	}
	resp := converter.UserMessagePOToEntityList(userMessages)
	return resp, total, err
}

func (g *UserMsgRepo) GetLastUserMsgsByDialogIDs(dialogIds []uint) ([]*entity.UserMessage, error) {
	userMessages := make([]*po.UserMessage, 0)

	for _, dialogId := range dialogIds {
		var lastMsg po.UserMessage
		g.db.Where("dialog_id = ? AND deleted_at = 0", dialogId).Order("created_at DESC").First(&lastMsg)
		if lastMsg.ID != 0 {
			userMessages = append(userMessages, &lastMsg)
		}
	}
	resp := converter.UserMessagePOToEntityList(userMessages)
	return resp, nil
}

func (g *UserMsgRepo) GetUserMsgIdBeforeMsgList(dialogId uint, msgId uint, pageSize int) ([]*entity.UserMessage, int32, error) {
	var userMessages []*po.UserMessage
	var total int64
	err := g.db.Model(&po.UserMessage{}).Where("dialog_id = ? AND deleted_at = 0", dialogId).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	err = g.db.Model(&po.UserMessage{}).
		Where("dialog_id = ? AND id < ? AND deleted_at = 0", dialogId, msgId).Order("id DESC").
		Limit(pageSize).
		Find(&userMessages).Error
	if err != nil {
		return nil, 0, err
	}

	resp := converter.UserMessagePOToEntityList(userMessages)
	return resp, int32(total), err
}
