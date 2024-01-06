package service

import (
	"github.com/cossim/coss-server/service/msg/api/dataTransformers"
	"github.com/cossim/coss-server/service/msg/domain/entity"
	"github.com/cossim/coss-server/service/msg/domain/repository"
)

type MsgService struct {
	ur repository.MsgRepository
}

func NewMsgService(ur repository.MsgRepository) *MsgService {
	return &MsgService{
		ur: ur,
	}
}

func (g MsgService) SendUserMessage(senderId string, receiverId string, msg string, msgType uint, replyId uint) (*entity.UserMessage, error) {
	um, err := g.ur.InsertUserMessage(senderId, receiverId, msg, entity.UserMessageType(msgType), replyId)
	if err != nil {
		return nil, err
	}
	return um, err
}

func (g MsgService) GetUserMessageList(userId string, friendId string, content string, msgType entity.UserMessageType, pageNumber, pageSize int) dataTransformers.UserMsgListResponse {
	ums, total, current := g.ur.GetUserMsgList(userId, friendId, content, msgType, pageNumber, pageSize)

	return dataTransformers.UserMsgListResponse{
		Total:        total,
		CurrentPage:  current,
		UserMessages: ums,
	}
}
