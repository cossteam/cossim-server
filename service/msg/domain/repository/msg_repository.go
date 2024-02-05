package repository

import (
	"github.com/cossim/coss-server/service/msg/api/dataTransformers"
	"github.com/cossim/coss-server/service/msg/domain/entity"
)

type MsgRepository interface {
	InsertUserMessage(senderId string, receiverId string, msg string, msgType entity.UserMessageType, replyId uint, dialogId uint, isBurnAfterReading entity.BurnAfterReadingType) (*entity.UserMessage, error)
	InsertGroupMessage(uid string, groupId uint, msg string, msgType entity.UserMessageType, replyId uint, dialogId uint, isBurnAfterReading entity.BurnAfterReadingType, atUsers []string, atAllUser entity.AtAllUserType) (*entity.GroupMessage, error)
	GetGroupMsgList(response dataTransformers.GroupMsgList) (*dataTransformers.GroupMsgListResponse, error)
	GetUserMsgList(uid, friendId string, content string, msgType entity.UserMessageType, pageNumber, pageSize int) ([]entity.UserMessage, int32, int32)
	GetLastMsgsForUserWithFriends(userID string, friendIDs []string) ([]*entity.UserMessage, error)
	GetLastMsgsForGroupsWithIDs(groupIDs []uint) ([]*entity.GroupMessage, error)
	GetLastMsgsByDialogIDs(dialogIds []uint) ([]dataTransformers.LastMessage, error)
	UpdateUserMessage(msg entity.UserMessage) error
	UpdateGroupMessage(msg entity.GroupMessage) error
	GetUserMsgByID(msgId uint32) (*entity.UserMessage, error)
	GetGroupMsgByID(msgId uint32) (*entity.GroupMessage, error)
	UpdateUserMsgColumn(msgId uint32, column string, value interface{}) error
	UpdateGroupMsgColumn(msgId uint32, column string, value interface{}) error
	GetUserMsgLabelByDialogId(dialogId uint32) ([]*entity.UserMessage, error)
	GetGroupMsgLabelByDialogId(dialogId uint32) ([]*entity.GroupMessage, error)
	SetUserMsgsReadStatus(msgIds []uint32, dialogId uint32) error
	SetUserMsgReadStatus(msgId uint32, readType entity.ReadType) error
	GetUnreadUserMsgs(uid string, dialogId uint32) ([]*entity.UserMessage, error)

	//批量查询阅后即焚消息id
	GetBatchUserMsgsBurnAfterReadingMessages(msgIds []uint32, dialogId uint32) ([]*entity.UserMessage, error)

	GetUserMsgIdAfterMsgList(dialogId uint32, msgIds uint32) ([]*entity.UserMessage, error)
	GetGroupMsgIdAfterMsgList(dialogId uint32, msgIds uint32) ([]*entity.GroupMessage, error)

	//删除
	PhysicalDeleteGroupMessage(msgId uint32) error
	PhysicalDeleteUserMessage(msgId uint32) error
	LogicalDeleteGroupMessage(msgId uint32) error
	LogicalDeleteUserMessage(msgId uint32) error
	//批量删除私聊消息
	PhysicalDeleteUserMessages(msgIds []uint32) error
	PhysicalDeleteGroupMessages(msgIds []uint32) error
	LogicalDeleteUserMessages(msgIds []uint32) error
	LogicalDeleteGroupMessages(msgIds []uint32) error
}
