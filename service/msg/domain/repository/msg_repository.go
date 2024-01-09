package repository

import (
	"github.com/cossim/coss-server/service/msg/domain/entity"
)

type MsgRepository interface {
	InsertUserMessage(senderId string, receiverId string, msg string, msgType entity.UserMessageType, replyId uint) (*entity.UserMessage, error)
	InsertGroupMessage(uid string, groupId uint, msg string, msgType entity.UserMessageType, replyId uint) (*entity.GroupMessage, error)
	GetUserMsgList(uid, friendId string, content string, msgType entity.UserMessageType, pageNumber, pageSize int) ([]entity.UserMessage, int32, int32)
	GetLastMsgsForUserWithFriends(userID string, friendIDs []string) ([]*entity.UserMessage, error)
	GetLastMsgsForGroupsWithIDs(groupIDs []uint) ([]*entity.GroupMessage, error)
}
