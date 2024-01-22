package repository

import (
	"github.com/cossim/coss-server/service/msg/api/dataTransformers"
	"github.com/cossim/coss-server/service/msg/domain/entity"
)

type MsgRepository interface {
	InsertUserMessage(senderId string, receiverId string, msg string, msgType entity.UserMessageType, replyId uint, dialogId uint) (*entity.UserMessage, error)
	InsertGroupMessage(uid string, groupId uint, msg string, msgType entity.UserMessageType, replyId uint, dialogId uint) (*entity.GroupMessage, error)
	GetUserMsgList(uid, friendId string, content string, msgType entity.UserMessageType, pageNumber, pageSize int) ([]entity.UserMessage, int32, int32)
	GetLastMsgsForUserWithFriends(userID string, friendIDs []string) ([]*entity.UserMessage, error)
	GetLastMsgsForGroupsWithIDs(groupIDs []uint) ([]*entity.GroupMessage, error)
	GetLastMsgsByDialogIDs(dialogIds []uint) ([]dataTransformers.LastMessage, error)
	UpdateUserMessage(msg entity.UserMessage) error
	UpdateGroupMessage(msg entity.GroupMessage) error
	PhysicalDeleteGroupMessage(msgId uint64) error
	PhysicalDeleteUserMessage(msgId uint64) error
	LogicalDeleteGroupMessage(msgId uint64) error
	LogicalDeleteUserMessage(msgId uint64) error
}
