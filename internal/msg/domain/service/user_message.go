package service

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/msg/domain/entity"
	"github.com/cossim/coss-server/internal/msg/infra/persistence"
	"github.com/cossim/coss-server/pkg/code"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
	"time"
)

type UserMsgDomain interface {
	// 发送私聊消息
	SendUserMessage(ctx context.Context, message *entity.UserMessage) (*entity.UserMessage, error)
	// 发送私聊消息
	SendUserMessageRevert(ctx context.Context, id uint) error
	// 群发私聊消息
	SendMultiUserMessage(ctx context.Context, list []*entity.UserMessage) error
	// 获取用户消息列表
	GetUserMessageList(ctx context.Context, message *entity.UserMessage, pageSize, pageNum int, startAt, endAt int64) ([]*entity.UserMessage, int64, error)
	// 获取私聊对话最新消息
	GetUserLastMessageList(ctx context.Context, dialogID uint, pageSize, pageNum int) ([]*entity.UserMessage, int64, error)
	//根据好友id获取最后一条消息
	GetLastMsgsForUserWithFriends(ctx context.Context, userID string, friendIDs []string) ([]*entity.UserMessage, error)
	// 根据对话id获取最后一条消息
	GetLastMsgsByDialogIds(ctx context.Context, dialogIDs []uint) ([]*entity.UserMessage, error)
	// 编辑私聊消息
	EditUserMessage(ctx context.Context, message *entity.UserMessage) error
	// 根据对话id与msgId查询msgId之后的私聊消息
	GetUserMsgIdAfterMsgList(ctx context.Context, msgID uint, dialogID uint) ([]*entity.UserMessage, int64, error)
	// 根据消息id获取私聊消息
	GetUserMessageById(ctx context.Context, id uint) (*entity.UserMessage, error)
	// 根据多个消息id获取私聊消息
	GetUserMessagesByIds(ctx context.Context, ids []uint) ([]*entity.UserMessage, error)
	// 根据对话id获取私聊标注信息
	GetUserMsgLabelByDialogId(ctx context.Context, dialogID uint) ([]*entity.UserMessage, error)
	// 一键已读用户消息
	ReadAllUserMsg(ctx context.Context, dialogID uint, userID string) error
	// 批量设置私聊消息id为已读
	SetUserMsgsReadStatus(ctx context.Context, ids []uint, dialogID uint, openBurnAfterReadingTimeOut int64) error
	// 修改指定私聊消息的已读状态
	SetUserMsgReadStatus(ctx context.Context, id uint, isRead entity.ReadType) error
	// 获取私聊对话未读消息
	GetUnreadUserMsgs(ctx context.Context, dialogID uint, userID string) ([]*entity.UserMessage, error)
	// 根据对话id删除私聊消息
	DeleteUserMessageByDialogId(ctx context.Context, dialogID uint, isPhysical bool) error
	// 根据对话id删除私聊消息回滚
	DeleteUserMessageByDialogIdRollback(ctx context.Context, dialogID uint) error
	// 根据消息id删除私聊消息
	DeleteUserMessageById(ctx context.Context, id uint, isPhysical bool) error
	//修改消息标注状态
	SetUserMsgLabel(ctx context.Context, id uint, isLabel bool) error
}

type UserMsgDomainImpl struct {
	db   *gorm.DB
	ac   *pkgconfig.AppConfig
	repo *persistence.Repositories
}

func NewUserMsgDomain(db *gorm.DB, ac *pkgconfig.AppConfig) UserMsgDomain {
	return &UserMsgDomainImpl{
		db:   db,
		ac:   ac,
		repo: persistence.NewRepositories(db),
	}
}

func (u *UserMsgDomainImpl) SendUserMessage(ctx context.Context, message *entity.UserMessage) (*entity.UserMessage, error) {
	msg, err := u.repo.Umr.InsertUserMessage(message)
	if err != nil {
		return nil, status.Error(codes.Code(code.MsgErrInsertUserMessageFailed.Code()), err.Error())
	}
	return msg, nil
}

func (u *UserMsgDomainImpl) SendUserMessageRevert(ctx context.Context, id uint) error {
	err := u.repo.Umr.PhysicalDeleteUserMessage(id)
	if err != nil {
		return status.Error(codes.Code(code.MsgErrDeleteUserMessageFailed.Code()), err.Error())
	}
	return nil
}

func (u *UserMsgDomainImpl) SendMultiUserMessage(ctx context.Context, messages []*entity.UserMessage) error {
	err := u.repo.Umr.InsertUserMessages(messages)
	if err != nil {
		return status.Error(codes.Code(code.MsgErrSendMultipleFailed.Code()), err.Error())
	}
	return nil
}

func (u *UserMsgDomainImpl) SetUserMsgLabel(ctx context.Context, id uint, isLabel bool) error {
	err := u.repo.Umr.UpdateUserMsgColumn(id, "is_label", isLabel)
	if err != nil {
		return status.Error(codes.Code(code.SetMsgErrSetUserMsgLabelFailed.Code()), err.Error())
	}
	return nil
}

func (u *UserMsgDomainImpl) GetUserMessageList(ctx context.Context, message *entity.UserMessage, pageSize, pageNum int, startAt, endAt int64) ([]*entity.UserMessage, int64, error) {
	if message.ID != 0 {
		res, err := u.repo.Umr.Find(ctx, &entity.UserMsgQuery{
			DialogIds: []uint32{uint32(message.DialogId)},
			MsgIds:    []uint32{uint32(message.ID)},
			PageSize:  int64(pageSize),
		})
		if err != nil {
			return nil, 0, status.Error(codes.Code(code.MsgErrGetUserMessageListFailed.Code()), err.Error())
		}
		return res.Messages, res.TotalCount, nil
	}
	res, err := u.repo.Umr.Find(ctx, &entity.UserMsgQuery{
		DialogIds: []uint32{uint32(message.DialogId)},
		MsgType:   message.Type,
		PageNum:   int64(pageNum),
		PageSize:  int64(pageSize),
		Content:   message.Content,
		SendID:    message.SendID,
		StartAt:   startAt,
		EndAt:     endAt,
	})
	if err != nil {
		return nil, 0, status.Error(codes.Code(code.MsgErrGetUserMessageListFailed.Code()), err.Error())
	}

	return res.Messages, res.TotalCount, nil
}

func (u *UserMsgDomainImpl) GetUserLastMessageList(ctx context.Context, dialogID uint, pageSize, pageNum int) ([]*entity.UserMessage, int64, error) {
	msgs, total, err := u.repo.Umr.GetUserDialogLastMsgs(dialogID, pageNum, pageSize)
	if err != nil {
		return nil, 0, status.Error(codes.Code(code.GetMsgErrGetUserMsgByIDFailed.Code()), err.Error())
	}
	return msgs, total, nil
}

func (u *UserMsgDomainImpl) GetLastMsgsForUserWithFriends(ctx context.Context, userID string, friendIDs []string) ([]*entity.UserMessage, error) {
	msgs, err := u.repo.Umr.GetLastMsgsForUserWithFriends(userID, friendIDs)
	if err != nil {
		return nil, status.Error(codes.Code(code.MsgErrGetLastMsgsForUserWithFriends.Code()), err.Error())
	}
	return msgs, nil
}

func (u *UserMsgDomainImpl) GetLastMsgsByDialogIds(ctx context.Context, dialogIDs []uint) ([]*entity.UserMessage, error) {
	result2, err := u.repo.Umr.GetLastUserMsgsByDialogIDs(dialogIDs)
	if err != nil {
		return nil, status.Error(codes.Code(code.MsgErrGetLastMsgsByDialogIds.Code()), err.Error())
	}
	return result2, nil
}

func (u *UserMsgDomainImpl) EditUserMessage(ctx context.Context, message *entity.UserMessage) error {
	if err := u.repo.Umr.UpdateUserMessage(message); err != nil {
		return status.Error(codes.Code(code.MsgErrEditUserMessageFailed.Code()), err.Error())
	}
	return nil
}

func (u *UserMsgDomainImpl) GetUserMsgIdAfterMsgList(ctx context.Context, msgID uint, dialogID uint) ([]*entity.UserMessage, int64, error) {
	list, total, err := u.repo.Umr.GetUserMsgIdAfterMsgList(dialogID, msgID)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil

}

func (u *UserMsgDomainImpl) GetUserMessageById(ctx context.Context, id uint) (*entity.UserMessage, error) {
	msg, err := u.repo.Umr.GetUserMsgByID(id)
	if err != nil {
		return nil, status.Error(codes.Code(code.GetMsgErrGetUserMsgByIDFailed.Code()), err.Error())
	}
	return msg, nil
}

func (u *UserMsgDomainImpl) GetUserMessagesByIds(ctx context.Context, ids []uint) ([]*entity.UserMessage, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	msgs, err := u.repo.Umr.GetUserMsgByIDs(ids)
	if err != nil {
		return nil, status.Error(codes.Code(code.GetMsgErrGetUserMsgByIDFailed.Code()), err.Error())
	}
	return msgs, nil
}

func (u *UserMsgDomainImpl) GetUserMsgLabelByDialogId(ctx context.Context, dialogID uint) ([]*entity.UserMessage, error) {
	msgs, err := u.repo.Umr.GetUserMsgLabelByDialogId(dialogID)
	if err != nil {
		return nil, status.Error(codes.Code(code.GetMsgErrGetUserMsgLabelByDialogIdFailed.Code()), err.Error())
	}
	return msgs, nil
}

func (u *UserMsgDomainImpl) ReadAllUserMsg(ctx context.Context, dialogID uint, userID string) error {
	if err := u.repo.Umr.ReadAllUserMsg(userID, dialogID); err != nil {
		return status.Error(codes.Code(code.SetMsgErrSetUserMsgsReadStatusFailed.Code()), err.Error())
	}
	return nil
}

func (u *UserMsgDomainImpl) SetUserMsgsReadStatus(ctx context.Context, ids []uint, dialogID uint, openBurnAfterReadingTimeOut int64) error {
	//获取阅后即焚消息id
	messages, err := u.repo.Umr.GetBatchUserMsgsBurnAfterReadingMessages(ids, dialogID)
	if err != nil {
		return err
	}
	rids := make([]uint, 0)
	if len(messages) > 0 {
		for _, msg := range messages {
			rids = append(rids, uint(msg.ID))
		}
	}
	if openBurnAfterReadingTimeOut == 0 {
		openBurnAfterReadingTimeOut = 10
	}
	err = u.db.Transaction(func(tx *gorm.DB) error {
		npo := persistence.NewRepositories(tx)
		if err := npo.Umr.SetUserMsgsReadStatus(ids, dialogID); err != nil {
			return err
		}
		if len(rids) > 0 {
			//起一个携程，定时器根据超时时间删除
			go func(rpo *persistence.Repositories) {
				time.Sleep(time.Duration(openBurnAfterReadingTimeOut) * time.Second)
				err := rpo.Umr.LogicalDeleteUserMessages(rids)
				if err != nil {
					fmt.Println(err.Error())
					return
				}
			}(npo)
		}

		return nil
	})
	if err != nil {
		return status.Error(codes.Code(code.SetMsgErrSetUserMsgsReadStatusFailed.Code()), err.Error())
	}
	return nil
}

func (u *UserMsgDomainImpl) SetUserMsgReadStatus(ctx context.Context, id uint, isRead entity.ReadType) error {
	if err := u.repo.Umr.SetUserMsgReadStatus(id, isRead); err != nil {
		return status.Error(codes.Code(code.SetMsgErrSetUserMsgReadStatusFailed.Code()), err.Error())
	}
	return nil
}

func (u *UserMsgDomainImpl) GetUnreadUserMsgs(ctx context.Context, dialogID uint, userID string) ([]*entity.UserMessage, error) {
	msgs, err := u.repo.Umr.GetUnreadUserMsgs(userID, dialogID)
	if err != nil {
		return nil, status.Error(codes.Code(code.GetMsgErrGetUnreadUserMsgsFailed.Code()), err.Error())
	}
	return msgs, nil
}

func (u *UserMsgDomainImpl) DeleteUserMessageByDialogId(ctx context.Context, dialogID uint, isPhysical bool) error {
	if isPhysical {
		err := u.repo.Umr.PhysicalDeleteUserMessagesByDialogID(dialogID)
		if err != nil {
			return status.Error(codes.Aborted, fmt.Sprintf("failed to delete user msg: %v", err))
		}
	}
	err := u.repo.Umr.DeleteUserMessagesByDialogID(dialogID)
	if err != nil {
		return status.Error(codes.Aborted, fmt.Sprintf("failed to delete user msg: %v", err))
	}
	return nil
}

func (u *UserMsgDomainImpl) DeleteUserMessageByDialogIdRollback(ctx context.Context, dialogID uint) error {
	err := u.repo.Umr.UpdateUserMsgColumnByDialogId(dialogID, "deleted_at", 0)
	if err != nil {
		return status.Error(codes.Code(code.MsgErrDeleteUserMessageFailed.Code()), err.Error())
	}
	return nil
}

func (u *UserMsgDomainImpl) DeleteUserMessageById(ctx context.Context, id uint, isPhysical bool) error {
	if isPhysical {
		if err := u.repo.Umr.PhysicalDeleteUserMessage(id); err != nil {
			return status.Error(codes.Code(code.MsgErrDeleteUserMessageFailed.Code()), err.Error())
		}
	}
	if err := u.repo.Umr.LogicalDeleteUserMessage(id); err != nil {
		return status.Error(codes.Code(code.MsgErrDeleteUserMessageFailed.Code()), err.Error())
	}
	return nil
}
