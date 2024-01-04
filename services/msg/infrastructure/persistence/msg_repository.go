package persistence

import (
	"github.com/cossim/coss-server/services/msg/domain/entity"
	"gorm.io/gorm"
)

type MsgRepo struct {
	db *gorm.DB
}

func NewMsgRepo(db *gorm.DB) *MsgRepo {
	return &MsgRepo{db: db}
}

func (g *MsgRepo) InsertUserMessage(senderId string, receiverId string, msg string, msgType entity.UserMessageType, replyId uint) (*entity.UserMessage, error) {
	content := &entity.UserMessage{
		SendID:    senderId,
		ReceiveID: receiverId,
		Content:   msg,
		Type:      msgType,
		ReplyId:   replyId,
	}
	if err := g.db.Save(content).Error; err != nil {
		return nil, err
	}
	return content, nil
}

func (g *MsgRepo) InsertGroupMessage(uid string, groupId uint, msg string, msgType entity.UserMessageType, replyId uint) (*entity.GroupMessage, error) {
	content := &entity.GroupMessage{
		UID:     uid,
		GroupID: groupId,
		Content: msg,
		Type:    msgType,
		ReplyId: replyId,
	}
	if err := g.db.Save(content).Error; err != nil {
		return nil, err
	}
	return content, nil
}
