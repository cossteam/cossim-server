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
	PhysicalDeleteGroupMessage(msgId uint32) error
	PhysicalDeleteUserMessage(msgId uint32) error
	LogicalDeleteGroupMessage(msgId uint32) error
	LogicalDeleteUserMessage(msgId uint32) error
	GetUserMsgByID(msgId uint32) (*entity.UserMessage, error)
	GetGroupMsgByID(msgId uint32) (*entity.GroupMessage, error)
	UpdateUserMsgColumn(msgId uint32, column string, value interface{}) error
	UpdateGroupMsgColumn(msgId uint32, column string, value interface{}) error
}
