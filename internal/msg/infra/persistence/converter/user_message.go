package converter

import (
	"github.com/cossim/coss-server/internal/msg/domain/entity"
	"github.com/cossim/coss-server/internal/msg/infra/persistence/po"
)

func UserMessagePOToEntity(um *po.UserMessage) *entity.UserMessage {
	return &entity.UserMessage{
		Type:               entity.UserMessageType(um.Type),
		DialogId:           um.DialogId,
		IsRead:             entity.ReadType(um.IsRead),
		ReplyId:            um.ReplyId,
		ReadAt:             um.ReadAt,
		ReceiveID:          um.ReceiveID,
		SendID:             um.SendID,
		Content:            um.Content,
		IsLabel:            um.IsLabel,
		IsBurnAfterReading: entity.BurnAfterReadingType(um.IsBurnAfterReading),
		ReplyEmoji:         um.ReplyEmoji,
		BaseModel: entity.BaseModel{
			ID:        um.ID,
			CreatedAt: um.CreatedAt,
			UpdatedAt: um.UpdatedAt,
			DeletedAt: um.DeletedAt,
		},
	}
}

func UserMessageEntityToPO(um *entity.UserMessage) *po.UserMessage {
	return &po.UserMessage{
		Type:               uint(um.Type),
		DialogId:           um.DialogId,
		IsRead:             uint(um.IsRead),
		ReplyId:            um.ReplyId,
		ReadAt:             um.ReadAt,
		ReceiveID:          um.ReceiveID,
		SendID:             um.SendID,
		Content:            um.Content,
		IsLabel:            um.IsLabel,
		IsBurnAfterReading: uint(um.IsBurnAfterReading),
		ReplyEmoji:         um.ReplyEmoji,
		BaseModel: po.BaseModel{
			ID:        um.ID,
			CreatedAt: um.CreatedAt,
			UpdatedAt: um.UpdatedAt,
			DeletedAt: um.DeletedAt,
		},
	}
}

func UserMessagePOToEntityList(models []*po.UserMessage) []*entity.UserMessage {
	var list []*entity.UserMessage
	for _, model := range models {
		list = append(list, UserMessagePOToEntity(model))
	}
	return list
}

func UserMessageEntityToPOList(models []*entity.UserMessage) []*po.UserMessage {
	var list []*po.UserMessage
	for _, model := range models {
		list = append(list, UserMessageEntityToPO(model))
	}
	return list
}
