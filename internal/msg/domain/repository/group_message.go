package repository

import (
	"github.com/cossim/coss-server/internal/msg/api/grpc/dataTransformers"
	"github.com/cossim/coss-server/internal/msg/domain/entity"
)

type GroupMessageRepository interface {
	GroupMsgReadRepositoryBase
	GroupMsgReadRepositoryQuery
}

type GroupMsgReadRepositoryBase interface {
	InsertGroupMessage(*entity.GroupMessage) (*entity.GroupMessage, error)
	UpdateGroupMsgColumnByDialogId(uint, string, interface{}) error
	UpdateGroupMsgColumn(msgId uint, column string, value interface{}) error
	UpdateGroupMessage(*entity.GroupMessage) error
	PhysicalDeleteGroupMessage(uint) error
	LogicalDeleteGroupMessage(uint) error
	DeleteGroupMessagesByDialogID(uint) error
	PhysicalDeleteGroupMessagesByDialogID(uint) error
	PhysicalDeleteGroupMessages([]uint) error
	LogicalDeleteGroupMessages([]uint) error
}

type GroupMsgReadRepositoryQuery interface {
	GetLastGroupMsgsByDialogIDs([]uint) ([]*entity.GroupMessage, error)
	GetGroupMsgByID(uint) (*entity.GroupMessage, error)
	GetGroupMsgsByIDs([]uint) ([]*entity.GroupMessage, error)
	GetLastMsgsForGroupsWithIDs([]uint) ([]*entity.GroupMessage, error)
	GetGroupMsgList(response dataTransformers.GroupMsgList) (*dataTransformers.GroupMsgListResponse, error)
	GetGroupMsgLabelByDialogId(dialogId uint) ([]*entity.GroupMessage, error)
	GetGroupMsgIdAfterMsgList(dialogId uint, msgIds uint) ([]*entity.GroupMessage, int64, error)
	GetGroupMsgIdBeforeMsgList(dialogId uint, msgId uint, pageSize int) ([]*entity.GroupMessage, int32, error)
	GetGroupMsgIdsByDialogID(dialogId uint) ([]uint, error)
	GetGroupUnreadMsgList(dialogId uint, msgIds []uint) ([]*entity.GroupMessage, error)
	GetGroupDialogLastMsgs(dialogId uint, pageNumber, pageSize int) ([]*entity.GroupMessage, int64, error)
}
