package repository

import (
	"context"
	"github.com/cossim/coss-server/internal/msg/domain/entity"
)

type UserMessageRepository interface {
	UserMessageRepositoryBase
	UserMessageRepositoryQuery
}

type UserMessageRepositoryBase interface {
	InsertUserMessage(message *entity.UserMessage) (*entity.UserMessage, error)
	InsertUserMessages(message []*entity.UserMessage) error
	UpdateUserMessage(msg *entity.UserMessage) error
	UpdateUserMsgColumn(msgId uint, column string, value interface{}) error
	SetUserMsgsReadStatus(msgIds []uint, dialogId uint) error
	SetUserMsgReadStatus(msgId uint, readType entity.ReadType) error
	UpdateUserMsgColumnByDialogId(dialogId uint, column string, value interface{}) error
	ReadAllUserMsg(uid string, dialogId uint) error
	LogicalDeleteUserMessages(msgIds []uint) error
	PhysicalDeleteUserMessages(msgIds []uint) error
	PhysicalDeleteUserMessagesByDialogID(dialogId uint) error
	PhysicalDeleteUserMessage(msgId uint) error
	LogicalDeleteUserMessage(msgId uint) error
	DeleteUserMessagesByDialogID(dialogId uint) error
}

type UserMessageRepositoryQuery interface {
	GetLastMsgsForUserWithFriends(userID string, friendIDs []string) ([]*entity.UserMessage, error)
	GetUserMsgList(dialogId uint, sendId string, content string, msgType entity.UserMessageType, pageNumber, pageSize int) ([]*entity.UserMessage, int32, int32)
	GetUserMsgByID(msgId uint) (*entity.UserMessage, error)
	GetUserMsgByIDs(msgIds []uint) ([]*entity.UserMessage, error)
	GetUserMsgLabelByDialogId(dialogId uint) ([]*entity.UserMessage, error)
	GetUnreadUserMsgs(uid string, dialogId uint) ([]*entity.UserMessage, error)
	GetBatchUserMsgsBurnAfterReadingMessages(msgIds []uint, dialogId uint) ([]*entity.UserMessage, error)
	GetUserMsgIdAfterMsgList(dialogId uint, msgIds uint) ([]*entity.UserMessage, int64, error)
	GetUserMsgIdBeforeMsgList(dialogId uint, msgId uint, pageSize int) ([]*entity.UserMessage, int32, error)
	GetUserDialogLastMsgs(dialogId uint, pageNumber, pageSize int) ([]*entity.UserMessage, int64, error)
	GetLastUserMsgsByDialogIDs(dialogIds []uint) ([]*entity.UserMessage, error)
	Find(ctx context.Context, query *entity.UserMsgQuery) (*entity.UserMsgQueryResult, error)
}
