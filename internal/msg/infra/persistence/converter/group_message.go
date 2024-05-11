package converter

import (
	"github.com/cossim/coss-server/internal/msg/domain/entity"
	"github.com/cossim/coss-server/internal/msg/infra/persistence/po"
)

func GroupMessageEntityToPO(gm *entity.GroupMessage) *po.GroupMessage {
	return &po.GroupMessage{
		DialogId:           gm.DialogID,
		GroupID:            gm.GroupID,
		Type:               uint(gm.Type),
		ReplyId:            gm.ReplyId,
		ReadCount:          gm.ReadCount,
		UserID:             gm.UserID,
		Content:            gm.Content,
		IsLabel:            gm.IsLabel,
		ReplyEmoji:         gm.ReplyEmoji,
		AtAllUser:          uint(gm.AtAllUser),
		AtUsers:            gm.AtUsers,
		IsBurnAfterReading: uint(gm.IsBurnAfterReading),
		BaseModel: po.BaseModel{
			ID:        gm.ID,
			CreatedAt: gm.CreatedAt,
			UpdatedAt: gm.UpdatedAt,
			DeletedAt: gm.DeletedAt,
		},
	}
}

func GroupMessagePOToEntity(model *po.GroupMessage) *entity.GroupMessage {
	return &entity.GroupMessage{
		DialogID:           model.DialogId,
		GroupID:            model.GroupID,
		Type:               entity.UserMessageType(model.Type),
		ReplyId:            model.ReplyId,
		ReadCount:          model.ReadCount,
		UserID:             model.UserID,
		Content:            model.Content,
		IsLabel:            model.IsLabel,
		ReplyEmoji:         model.ReplyEmoji,
		AtAllUser:          entity.AtAllUserType(model.AtAllUser),
		AtUsers:            model.AtUsers,
		IsBurnAfterReading: entity.BurnAfterReadingType(model.IsBurnAfterReading),
		BaseModel: entity.BaseModel{
			ID:        model.ID,
			CreatedAt: model.CreatedAt,
			UpdatedAt: model.UpdatedAt,
			DeletedAt: model.DeletedAt,
		},
	}
}

func GroupMessagePOToEntityList(models []*po.GroupMessage) []*entity.GroupMessage {
	var list []*entity.GroupMessage
	for _, model := range models {
		list = append(list, GroupMessagePOToEntity(model))
	}
	return list
}

func GroupMessageEntityToPOList(models []*entity.GroupMessage) []*po.GroupMessage {
	var list []*po.GroupMessage
	for _, model := range models {
		list = append(list, GroupMessageEntityToPO(model))
	}
	return list
}
