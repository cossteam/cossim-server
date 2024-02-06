package service

import (
	"context"
	"github.com/cossim/coss-server/pkg/code"
	v1 "github.com/cossim/coss-server/service/msg/api/v1"
	"github.com/cossim/coss-server/service/msg/domain/entity"
	"github.com/cossim/coss-server/service/msg/infrastructure/persistence"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func (s *Service) SetGroupMessageRead(ctx context.Context, in *v1.SetGroupMessagesReadRequest) (*v1.SetGroupMessageReadResponse, error) {
	resp := &v1.SetGroupMessageReadResponse{}
	var reads []*entity.GroupMessageRead
	msgids := make([]uint32, 0)

	if len(in.List) > 0 {
		reads = make([]*entity.GroupMessageRead, len(in.List))
		for i, _ := range in.List {
			reads[i] = &entity.GroupMessageRead{
				DialogId: uint(in.List[i].DialogId),
				MsgId:    uint(in.List[i].MsgId),
				UserId:   in.List[i].UserId,
				GroupID:  uint(in.List[i].GroupId),
				ReadAt:   in.List[i].ReadAt,
			}
			msgids = append(msgids, in.List[i].MsgId)
		}
	}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		npo := persistence.NewRepositories(tx)
		msgs, err := npo.Mr.GetGroupMsgsByIDs(msgids)

		//修改已读数量
		for _, v := range msgs {
			v.ReadCount++
			err := npo.Mr.UpdateGroupMsgColumn(uint32(v.ID), "read_count", v.ReadCount)
			if err != nil {
				return err
			}
		}

		err = npo.Gmrr.SetGroupMsgReadByMsgs(reads)
		if err != nil {
			return status.Error(codes.Code(code.GroupErrSetGroupMsgReadFailed.Code()), err.Error())
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *Service) GetGroupMessageReaders(ctx context.Context, in *v1.GetGroupMessageReadersRequest) (*v1.GetGroupMessageReadersResponse, error) {
	resp := &v1.GetGroupMessageReadersResponse{}
	msgs, err := s.gmrr.GetGroupMsgReadUserIdsByMsgID(in.MsgId)
	if err != nil {
		return nil, status.Error(codes.Code(code.GroupErrGetGroupMsgReadersFailed.Code()), err.Error())
	}
	resp.UserIds = msgs

	return resp, err
}

func (s *Service) GetGroupMessageReadByMsgIdAndUserId(ctx context.Context, in *v1.GetGroupMessageReadByMsgIdAndUserIdRequest) (*v1.GetGroupMessageReadByMsgIdAndUserIdResponse, error) {
	resp := &v1.GetGroupMessageReadByMsgIdAndUserIdResponse{}
	msg, err := s.gmrr.GetGroupMsgReadByMsgIDAndUserID(in.MsgId, in.UserId)
	if err != nil {
		return nil, status.Error(codes.Code(code.GroupErrGetGroupMsgReadByMsgIdAndUserIdFailed.Code()), err.Error())
	}
	resp.ReadAt = msg.ReadAt
	return resp, err
}
