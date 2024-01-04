package repository

import "github.com/cossim/coss-server/services/msg/domain/entity"

type MsgRepository interface {
	InsertUserMessage(senderId string, receiverId string, msg string, msgType entity.UserMessageType, replyId uint) (*entity.UserMessage, error)
	InsertGroupMessage(uid string, groupId uint, msg string, msgType entity.UserMessageType, replyId uint) (*entity.GroupMessage, error)
}
