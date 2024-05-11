package service

import (
	"context"
	"github.com/cossim/coss-server/internal/msg/domain/entity"
	"github.com/cossim/coss-server/internal/msg/infra/persistence"
	"github.com/cossim/coss-server/pkg/code"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/utils/time"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type GroupMsgReadDomain interface {
	SetGroupMessageRead(ctx context.Context, msg []*entity.GroupMessageRead) error
	// 获取指定消息已读人员
	GetGroupMessageReaders(ctx context.Context, msgID, dialogID, groupID uint) ([]string, error)
	// 根据消息id与用户id查询是否已经读取过
	GetGroupMessageReadByMsgIdAndUserId(ctx context.Context, msgID uint, userID string) (*entity.GroupMessageRead, error)
	//读取所有群聊消息
	ReadAllGroupMsg(ctx context.Context, dialogID uint, userID string) ([]*entity.GroupMessageRead, error)
}

type GroupMsgReadImpl struct {
	db   *gorm.DB
	ac   *pkgconfig.AppConfig
	repo *persistence.Repositories
}

func NewGroupMsgReadDomain(db *gorm.DB, ac *pkgconfig.AppConfig) GroupMsgReadDomain {
	return &GroupMsgReadImpl{
		db:   db,
		ac:   ac,
		repo: persistence.NewRepositories(db),
	}
}

func (g *GroupMsgReadImpl) SetGroupMessageRead(ctx context.Context, read []*entity.GroupMessageRead) error {

	msgIds := make([]uint, 0)
	for _, v := range read {
		msgIds = append(msgIds, v.ID)
	}

	err := g.db.Transaction(func(tx *gorm.DB) error {
		npo := persistence.NewRepositories(tx)
		msgs, err := npo.Gmr.GetGroupMsgsByIDs(msgIds)
		//修改已读数量
		for _, v := range msgs {
			v.ReadCount++
			err := npo.Gmr.UpdateGroupMsgColumn(v.ID, "read_count", v.ReadCount)
			if err != nil {
				return err
			}
		}
		err = npo.Gmrr.SetGroupMsgReadByMsgs(read)
		if err != nil {
			return status.Error(codes.Code(code.GroupErrSetGroupMsgReadFailed.Code()), err.Error())
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (g *GroupMsgReadImpl) GetGroupMessageReaders(ctx context.Context, msgID, dialogID, groupID uint) ([]string, error) {
	msgs, err := g.repo.Gmrr.GetGroupMsgReadUserIdsByMsgID(msgID)
	if err != nil {
		return nil, status.Error(codes.Code(code.GroupErrGetGroupMsgReadersFailed.Code()), err.Error())
	}
	return msgs, nil
}

func (g *GroupMsgReadImpl) GetGroupMessageReadByMsgIdAndUserId(ctx context.Context, msgID uint, userID string) (*entity.GroupMessageRead, error) {
	msg, err := g.repo.Gmrr.GetGroupMsgReadByMsgIDAndUserID(msgID, userID)
	if err != nil {
		return nil, status.Error(codes.Code(code.GroupErrGetGroupMsgReadByMsgIdAndUserIdFailed.Code()), err.Error())
	}
	return msg, nil
}

func (g *GroupMsgReadImpl) ReadAllGroupMsg(ctx context.Context, dialogID uint, userID string) ([]*entity.GroupMessageRead, error) {

	msgids, err := g.repo.Gmrr.GetGroupMsgUserReadIdsByDialogID(dialogID, userID)
	if err != nil {
		return nil, status.Error(codes.Code(code.GroupErrGetGroupMsgReadByMsgIdAndUserIdFailed.Code()), err.Error())
	}

	list, err := g.repo.Gmr.GetGroupUnreadMsgList(dialogID, msgids)
	if err != nil {
		return nil, status.Error(codes.Code(code.GroupErrGetGroupMsgReadByMsgIdAndUserIdFailed.Code()), err.Error())
	}
	var reads []*entity.GroupMessageRead

	if len(list) > 0 {
		reads = make([]*entity.GroupMessageRead, len(list))
		for k, v := range list {
			reads[k] = &entity.GroupMessageRead{
				DialogID: v.DialogID,
				MsgID:    v.ID,
				UserID:   userID,
				GroupID:  v.GroupID,
				ReadAt:   time.Now(),
			}
		}
	}

	if err := g.repo.Gmrr.SetGroupMsgReadByMsgs(reads); err != nil {
		return nil, status.Error(codes.Code(code.GroupErrGetGroupMsgReadByMsgIdAndUserIdFailed.Code()), err.Error())
	}
	return reads, nil
}
