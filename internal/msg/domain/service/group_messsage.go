package service

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/msg/api/grpc/dataTransformers"
	"github.com/cossim/coss-server/internal/msg/domain/entity"
	"github.com/cossim/coss-server/internal/msg/infra/persistence"
	"github.com/cossim/coss-server/pkg/code"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/utils"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type GroupMsgDomain interface {
	// 发送群消息
	SendGroupMessage(ctx context.Context, msg *entity.GroupMessage) (*entity.GroupMessage, error)
	// 发送群消息回滚
	SendGroupMessageRevert(ctx context.Context, id uint) error
	// 获取群聊对话最新消息
	GetGroupLastMessageList(ctx context.Context, dialogID uint, pageSize, pageNum int) ([]*entity.GroupMessage, int64, error)
	// 根据群组id获取最后一条消息
	GetLastMsgsForGroupsWithIDs(ctx context.Context, ids []uint) ([]*entity.GroupMessage, error)
	// 获取群消息列表
	GetGroupMessageList(ctx context.Context, msg *entity.GroupMessage, pageSize, pageNum int) ([]*entity.GroupMessage, int64, error)
	// 编辑群消息
	EditGroupMessage(ctx context.Context, msg *entity.GroupMessage) error
	// 撤回群消息
	DeleteGroupMessage(ctx context.Context, id uint, isPhysical bool) error
	// 根据消息id获取群消息
	GetGroupMessageById(ctx context.Context, id uint) (*entity.GroupMessage, error)
	// 根据多个消息id获取群消息
	GetGroupMessagesByIds(ctx context.Context, ids []uint) ([]*entity.GroupMessage, error)
	// 设置群消息标注状态
	SetGroupMsgLabel(ctx context.Context, id uint, isLabel bool) error
	// 根据对话id获取群消息标注信息
	GetGroupMsgLabelByDialogId(ctx context.Context, dialogID uint) ([]*entity.GroupMessage, error)
	// 根据对话id与msgId查询msgId之后的群消息
	GetGroupMsgIdAfterMsgList(ctx context.Context, msgID, dialogID uint) ([]*entity.GroupMessage, int64, error)
	// 根据对话id删除群聊消息
	DeleteGroupMessageByDialogId(ctx context.Context, dialogID uint, isPhysical bool) error
	// 根据对话id删除群聊消息回滚
	DeleteGroupMessageByDialogIdRollback(ctx context.Context, dialogID uint) error
	// 获取群聊未读消息
	GetGroupUnreadMessages(ctx context.Context, dialogID uint, userID string) ([]*entity.GroupMessage, error)
	// 根据对话id获取最后一条消息
	GetLastMsgsByDialogIds(ctx context.Context, dialogIDs []uint) ([]*entity.GroupMessage, error)
}

type GroupMsgDomainImpl struct {
	db   *gorm.DB
	ac   *pkgconfig.AppConfig
	repo *persistence.Repositories
}

func NewGroupMsgDomain(db *gorm.DB, ac *pkgconfig.AppConfig) GroupMsgDomain {
	return &GroupMsgDomainImpl{
		db:   db,
		ac:   ac,
		repo: persistence.NewRepositories(db),
	}
}

func (g GroupMsgDomainImpl) SendGroupMessage(ctx context.Context, msg *entity.GroupMessage) (*entity.GroupMessage, error) {
	mg, err := g.repo.Gmr.InsertGroupMessage(msg)
	if err != nil {
		return nil, status.Error(codes.Code(code.MsgErrInsertGroupMessageFailed.Code()), err.Error())
	}
	return mg, nil
}

func (g GroupMsgDomainImpl) SendGroupMessageRevert(ctx context.Context, id uint) error {
	if err := g.repo.Gmr.PhysicalDeleteGroupMessage(id); err != nil {
		return status.Error(codes.Code(code.MsgErrDeleteGroupMessageFailed.Code()), err.Error())
	}
	return nil
}

func (g GroupMsgDomainImpl) GetGroupLastMessageList(ctx context.Context, dialogID uint, pageSize, pageNum int) ([]*entity.GroupMessage, int64, error) {
	msgs, total, err := g.repo.Gmr.GetGroupDialogLastMsgs(dialogID, pageNum, pageSize)
	if err != nil {
		return nil, 0, status.Error(codes.Code(code.GetMsgErrGetUserMsgByIDFailed.Code()), err.Error())
	}
	return msgs, total, nil
}

func (g GroupMsgDomainImpl) GetLastMsgsForGroupsWithIDs(ctx context.Context, ids []uint) ([]*entity.GroupMessage, error) {
	msgs, err := g.repo.Gmr.GetLastMsgsForGroupsWithIDs(ids)
	if err != nil {
		return nil, status.Error(codes.Code(code.MsgErrGetLastMsgsForGroupsWithIDs.Code()), err.Error())
	}
	return msgs, nil
}

func (g GroupMsgDomainImpl) GetGroupMessageList(ctx context.Context, msg *entity.GroupMessage, pageSize, pageNum int) ([]*entity.GroupMessage, int64, error) {
	if msg.ID != 0 {
		list, total, err := g.repo.Gmr.GetGroupMsgIdBeforeMsgList(msg.DialogID, msg.ID, pageSize)
		if err != nil {
			return nil, 0, err
		}

		return list, int64(total), nil
	}
	list, err := g.repo.Gmr.GetGroupMsgList(dataTransformers.GroupMsgList{
		DialogId:   msg.DialogID,
		Content:    msg.Content,
		UserID:     msg.UserID,
		MsgType:    msg.Type,
		PageNumber: pageNum,
		PageSize:   pageSize,
	})
	if err != nil {
		return nil, 0, status.Error(codes.Code(code.MsgErrGetGroupMsgListFailed.Code()), err.Error())
	}
	return list.GroupMessages, int64(list.Total), nil
}

func (g GroupMsgDomainImpl) EditGroupMessage(ctx context.Context, msg *entity.GroupMessage) error {
	if err := g.repo.Gmr.UpdateGroupMessage(msg); err != nil {
		return status.Error(codes.Code(code.MsgErrEditGroupMessageFailed.Code()), err.Error())
	}
	return nil
}

func (g GroupMsgDomainImpl) DeleteGroupMessage(ctx context.Context, id uint, isPhysical bool) error {
	if isPhysical {
		if err := g.repo.Gmr.PhysicalDeleteGroupMessage(id); err != nil {
			return status.Error(codes.Code(code.MsgErrDeleteGroupMessageFailed.Code()), err.Error())
		}
		return nil
	}
	if err := g.repo.Gmr.LogicalDeleteGroupMessage(id); err != nil {
		return status.Error(codes.Code(code.MsgErrDeleteGroupMessageFailed.Code()), err.Error())
	}
	return nil
}

func (g GroupMsgDomainImpl) GetGroupMessageById(ctx context.Context, id uint) (*entity.GroupMessage, error) {
	msg, err := g.repo.Gmr.GetGroupMsgByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.Code(code.GetMsgErrGetGroupMsgByIDFailed.Code()), err.Error())
		}
		return nil, err
	}
	return msg, nil
}

func (g GroupMsgDomainImpl) GetGroupMessagesByIds(ctx context.Context, ids []uint) ([]*entity.GroupMessage, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	msgs, err := g.repo.Gmr.GetGroupMsgsByIDs(ids)
	if err != nil {
		return nil, status.Error(codes.Code(code.GetMsgErrGetGroupMsgByIDFailed.Code()), err.Error())
	}
	return msgs, nil
}

func (g GroupMsgDomainImpl) SetGroupMsgLabel(ctx context.Context, id uint, isLabel bool) error {
	if err := g.repo.Gmr.UpdateGroupMsgColumn(id, "is_label", isLabel); err != nil {
		return status.Error(codes.Code(code.SetMsgErrSetGroupMsgLabelFailed.Code()), err.Error())
	}
	return nil
}

func (g GroupMsgDomainImpl) GetGroupMsgLabelByDialogId(ctx context.Context, dialogID uint) ([]*entity.GroupMessage, error) {
	msgs, err := g.repo.Gmr.GetGroupMsgLabelByDialogId(dialogID)
	if err != nil {
		return nil, status.Error(codes.Code(code.GetMsgErrGetGroupMsgLabelByDialogIdFailed.Code()), err.Error())
	}
	return msgs, nil
}

func (g GroupMsgDomainImpl) GetGroupMsgIdAfterMsgList(ctx context.Context, msgID, dialogID uint) ([]*entity.GroupMessage, int64, error) {
	list, total, err := g.repo.Gmr.GetGroupMsgIdAfterMsgList(dialogID, msgID)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (g GroupMsgDomainImpl) DeleteGroupMessageByDialogId(ctx context.Context, dialogID uint, isPhysical bool) error {
	if isPhysical {
		err := g.repo.Gmr.PhysicalDeleteGroupMessagesByDialogID(dialogID)
		if err != nil {
			return status.Error(codes.Aborted, fmt.Sprintf("failed to delete group msg: %v", err))
		}
		return nil
	}
	err := g.repo.Gmr.DeleteGroupMessagesByDialogID(dialogID)
	if err != nil {
		return status.Error(codes.Aborted, fmt.Sprintf("failed to delete group msg: %v", err))
	}
	return nil
}

func (g GroupMsgDomainImpl) DeleteGroupMessageByDialogIdRollback(ctx context.Context, dialogID uint) error {
	err := g.repo.Gmr.UpdateGroupMsgColumnByDialogId(dialogID, "deleted_at", 0)
	if err != nil {
		return status.Error(codes.Code(code.MsgErrDeleteGroupMessageFailed.Code()), err.Error())
	}
	return nil
}

func (g GroupMsgDomainImpl) GetGroupUnreadMessages(ctx context.Context, dialogID uint, userID string) ([]*entity.GroupMessage, error) {
	ids, err := g.repo.Gmr.GetGroupMsgIdsByDialogID(dialogID)
	if err != nil {
		return nil, err
	}

	//获取已读消息id
	rids, err := g.repo.Gmrr.GetGroupMsgUserReadIdsByDialogID(dialogID, userID)
	if err != nil {
		return nil, err
	}

	//求差集
	msgIds := utils.SliceDifference(ids, rids)

	msgs, err := g.repo.Gmr.GetGroupMsgsByIDs(msgIds)
	if err != nil {
		return nil, err
	}

	return msgs, nil
}

func (g GroupMsgDomainImpl) GetLastMsgsByDialogIds(ctx context.Context, dialogIDs []uint) ([]*entity.GroupMessage, error) {
	ds, err := g.repo.Gmr.GetLastGroupMsgsByDialogIDs(dialogIDs)
	if err != nil {
		return nil, err
	}
	return ds, nil
}
