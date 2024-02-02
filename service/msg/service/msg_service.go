package service

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/service/msg/api/dataTransformers"
	v1 "github.com/cossim/coss-server/service/msg/api/v1"
	"github.com/cossim/coss-server/service/msg/domain/entity"
	"github.com/cossim/coss-server/service/msg/infrastructure/persistence"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

func (s *Service) SendUserMessage(ctx context.Context, request *v1.SendUserMsgRequest) (*v1.SendUserMsgResponse, error) {
	resp := &v1.SendUserMsgResponse{}

	msg, err := s.mr.InsertUserMessage(request.GetSenderId(), request.GetReceiverId(), request.GetContent(), entity.UserMessageType(request.GetType()), uint(request.GetReplayId()), uint(request.GetDialogId()))
	if err != nil {
		return resp, status.Error(codes.Code(code.MsgErrInsertUserMessageFailed.Code()), err.Error())
	}
	resp.MsgId = uint32(msg.ID)
	return resp, err
}

func (s *Service) SendGroupMessage(ctx context.Context, request *v1.SendGroupMsgRequest) (*v1.SendGroupMsgResponse, error) {
	resp := &v1.SendGroupMsgResponse{}

	ums, err := s.mr.InsertGroupMessage(request.GetUserId(), uint(request.GetGroupId()), request.GetContent(), entity.UserMessageType(request.GetType()), uint(request.GetReplayId()), uint(request.GetDialogId()))
	if err != nil {
		return resp, status.Error(codes.Code(code.MsgErrInsertGroupMessageFailed.Code()), err.Error())
	}
	return &v1.SendGroupMsgResponse{
		MsgId:   uint32(ums.ID),
		GroupId: uint32(ums.GroupID),
	}, nil
}

func (s *Service) GetUserMessageList(ctx context.Context, request *v1.GetUserMsgListRequest) (*v1.GetUserMsgListResponse, error) {
	resp := &v1.GetUserMsgListResponse{}

	ums, total, current := s.mr.GetUserMsgList(request.GetUserId(), request.GetFriendId(), request.GetContent(), entity.UserMessageType(request.GetType()), int(request.GetPageNum()), int(request.GetPageSize()))

	for _, m := range ums {
		resp.UserMessages = append(resp.UserMessages, &v1.UserMessage{
			Id:         uint32(m.ID),
			SenderId:   m.SendID,
			ReceiverId: m.ReceiveID,
			Content:    m.Content,
			Type:       uint32(int32(m.Type)),
			ReplayId:   uint64(m.ReplyId),
			IsRead:     int32(m.IsRead),
			ReadAt:     m.ReadAt,
			CreatedAt:  m.CreatedAt,
		})
	}
	resp.Total = total
	resp.CurrentPage = current

	return resp, nil
}

func (s *Service) GetLastMsgsForUserWithFriends(ctx context.Context, request *v1.UserMsgsRequest) (*v1.UserMessages, error) {
	resp := &v1.UserMessages{}
	msgs, err := s.mr.GetLastMsgsForUserWithFriends(request.UserId, request.FriendId)
	if err != nil {
		return resp, status.Error(codes.Code(code.MsgErrGetLastMsgsForUserWithFriends.Code()), err.Error())
	}
	nmsgs := make([]*v1.UserMessage, 0)
	for _, m := range msgs {
		nmsgs = append(nmsgs, &v1.UserMessage{
			Id:        uint32(m.ID),
			Content:   m.Content,
			Type:      uint32(m.Type),
			ReplayId:  uint64(m.ReplyId),
			ReadAt:    m.ReadAt,
			CreatedAt: m.CreatedAt,
		})
	}
	resp.UserMessages = nmsgs
	return resp, nil
}

func (s *Service) GetLastMsgsForGroupsWithIDs(ctx context.Context, request *v1.GroupMsgsRequest) (*v1.GroupMessages, error) {
	resp := &v1.GroupMessages{}
	ids := make([]uint, 0)
	if len(request.GroupId) > 0 {
		for _, id := range request.GroupId {
			ids = append(ids, uint(id))
		}
	}
	msgs, err := s.mr.GetLastMsgsForGroupsWithIDs(ids)
	if err != nil {
		return resp, status.Error(codes.Code(code.MsgErrGetLastMsgsForGroupsWithIDs.Code()), err.Error())
	}
	nmsgs := make([]*v1.GroupMessage, 0)
	for _, m := range msgs {
		nmsgs = append(nmsgs, &v1.GroupMessage{
			Id:        uint32(m.ID),
			UserId:    m.UserID,
			Content:   m.Content,
			Type:      uint32(m.Type),
			ReplyId:   uint32(m.ReplyId),
			ReadCount: int32(m.ReadCount),
			CreatedAt: m.CreatedAt,
		})
	}
	resp.GroupMessages = nmsgs
	return resp, nil
}

func (s *Service) GetLastMsgsByDialogIds(ctx context.Context, request *v1.GetLastMsgsByDialogIdsRequest) (*v1.GetLastMsgsResponse, error) {
	resp := &v1.GetLastMsgsResponse{}

	ids := make([]uint, 0)
	if len(request.DialogIds) > 0 {
		for _, id := range request.DialogIds {
			ids = append(ids, uint(id))
		}
	}
	result, err := s.mr.GetLastMsgsByDialogIDs(ids)
	fmt.Println("result", result)
	if err != nil {
		return resp, status.Error(codes.Code(code.MsgErrGetLastMsgsByDialogIds.Code()), err.Error())
	}

	if len(result) > 0 {
		for _, m := range result {
			resp.LastMsgs = append(resp.LastMsgs, &v1.LastMsg{
				Id:        uint32(m.ID),
				Type:      uint32(m.Type),
				CreatedAt: m.CreateAt,
				Content:   m.Content,
				SenderId:  m.SenderId,
				DialogId:  uint32(m.DialogId),
			})
		}
	}
	return resp, nil
}

func (s *Service) EditUserMessage(ctx context.Context, request *v1.EditUserMsgRequest) (*v1.UserMessage, error) {
	var resp = &v1.UserMessage{}
	msg := entity.UserMessage{
		BaseModel: entity.BaseModel{
			ID: uint(request.UserMessage.Id),
		},
		Content:   request.UserMessage.Content,
		ReplyId:   uint(request.UserMessage.ReplayId),
		SendID:    request.UserMessage.SenderId,
		ReceiveID: request.UserMessage.ReceiverId,
		Type:      entity.UserMessageType(request.UserMessage.Type),
		IsLabel:   uint(request.UserMessage.IsLabel),
	}
	if err := s.mr.UpdateUserMessage(msg); err != nil {
		return resp, status.Error(codes.Code(code.MsgErrEditUserMessageFailed.Code()), err.Error())
	}

	return resp, nil
}

func (s *Service) DeleteUserMessage(ctx context.Context, request *v1.DeleteUserMsgRequest) (*v1.UserMessage, error) {
	var resp = &v1.UserMessage{}
	err := s.mr.LogicalDeleteUserMessage(request.MsgId)
	if err != nil {
		return resp, status.Error(codes.Code(code.MsgErrDeleteUserMessageFailed.Code()), err.Error())
	}
	return &v1.UserMessage{
		Id: request.MsgId,
	}, nil
}

func (s *Service) EditGroupMessage(ctx context.Context, request *v1.EditGroupMsgRequest) (*v1.GroupMessage, error) {
	var resp = &v1.GroupMessage{}
	msg := entity.GroupMessage{
		BaseModel: entity.BaseModel{
			ID: uint(request.GroupMessage.Id),
		},
		Content: request.GroupMessage.Content,
		ReplyId: uint(request.GroupMessage.ReplyId),
		UserID:  request.GroupMessage.UserId,
		GroupID: uint(request.GroupMessage.GroupId),
		Type:    entity.UserMessageType(request.GroupMessage.Type),
		IsLabel: uint(request.GroupMessage.IsLabel),
	}
	if err := s.mr.UpdateGroupMessage(msg); err != nil {
		return resp, status.Error(codes.Code(code.MsgErrEditGroupMessageFailed.Code()), err.Error())
	}
	resp = &v1.GroupMessage{
		Id:        uint32(msg.ID),
		UserId:    msg.UserID,
		Content:   msg.Content,
		Type:      uint32(int32(msg.Type)),
		ReplyId:   uint32(msg.ReplyId),
		GroupId:   uint32(msg.GroupID),
		ReadCount: int32(msg.ReadCount),
	}
	return resp, nil
}

func (s *Service) DeleteGroupMessage(ctx context.Context, request *v1.DeleteGroupMsgRequest) (*v1.GroupMessage, error) {
	var resp = &v1.GroupMessage{}
	if err := s.mr.LogicalDeleteGroupMessage(request.MsgId); err != nil {
		return resp, status.Error(codes.Code(code.MsgErrDeleteGroupMessageFailed.Code()), err.Error())
	}
	return &v1.GroupMessage{
		Id: request.MsgId,
	}, nil
}

func (s *Service) GetUserMessageById(ctx context.Context, in *v1.GetUserMsgByIDRequest) (*v1.UserMessage, error) {
	var resp = &v1.UserMessage{}
	msg, err := s.mr.GetUserMsgByID(in.MsgId)
	if err != nil {
		return resp, status.Error(codes.Code(code.GetMsgErrGetUserMsgByIDFailed.Code()), err.Error())
	}

	resp = &v1.UserMessage{
		DialogId:   uint32(msg.DialogId),
		Id:         uint32(msg.ID),
		Content:    msg.Content,
		Type:       uint32(int32(msg.Type)),
		ReplayId:   uint64(msg.ReplyId),
		SenderId:   msg.SendID,
		ReceiverId: msg.ReceiveID,
		CreatedAt:  msg.CreatedAt,
	}
	return resp, nil
}

func (s *Service) GetGroupMessageById(ctx context.Context, in *v1.GetGroupMsgByIDRequest) (*v1.GroupMessage, error) {
	var resp = &v1.GroupMessage{}
	msg, err := s.mr.GetGroupMsgByID(in.MsgId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return resp, status.Error(codes.Code(code.GetMsgErrGetGroupMsgByIDFailed.Code()), err.Error())
		}
		return resp, err
	}
	return &v1.GroupMessage{
		Id:        uint32(msg.ID),
		UserId:    msg.UserID,
		Content:   msg.Content,
		Type:      uint32(int32(msg.Type)),
		ReplyId:   uint32(msg.ReplyId),
		GroupId:   uint32(msg.GroupID),
		ReadCount: int32(msg.ReadCount),
		CreatedAt: msg.CreatedAt,
	}, nil
}

func (s *Service) SetUserMsgLabel(ctx context.Context, in *v1.SetUserMsgLabelRequest) (*v1.SetUserMsgLabelResponse, error) {
	var resp = &v1.SetUserMsgLabelResponse{}
	if err := s.mr.UpdateUserMsgColumn(in.MsgId, "is_label", in.IsLabel); err != nil {
		return resp, status.Error(codes.Code(code.SetMsgErrSetUserMsgLabelFailed.Code()), err.Error())
	}
	resp.MsgId = in.MsgId
	return resp, nil
}

func (s *Service) SetGroupMsgLabel(ctx context.Context, in *v1.SetGroupMsgLabelRequest) (*v1.SetGroupMsgLabelResponse, error) {
	var resp = &v1.SetGroupMsgLabelResponse{}
	if err := s.mr.UpdateGroupMsgColumn(in.MsgId, "is_label", in.IsLabel); err != nil {
		return resp, status.Error(codes.Code(code.SetMsgErrSetGroupMsgLabelFailed.Code()), err.Error())
	}
	resp.MsgId = in.MsgId
	return resp, nil
}

func (s *Service) GetUserMsgLabelByDialogId(ctx context.Context, in *v1.GetUserMsgLabelByDialogIdRequest) (*v1.GetUserMsgLabelByDialogIdResponse, error) {
	var resp = &v1.GetUserMsgLabelByDialogIdResponse{}

	msgs, err := s.mr.GetUserMsgLabelByDialogId(in.DialogId)
	if err != nil {
		return resp, status.Error(codes.Code(code.GetMsgErrGetUserMsgLabelByDialogIdFailed.Code()), err.Error())
	}

	var msglist = make([]*v1.UserMessage, 0)
	for _, msg := range msgs {
		msglist = append(msglist, &v1.UserMessage{
			Id:         uint32(msg.ID),
			Content:    msg.Content,
			Type:       uint32(msg.Type),
			ReplayId:   uint64(msg.ReplyId),
			SenderId:   msg.SendID,
			ReceiverId: msg.ReceiveID,
			CreatedAt:  msg.CreatedAt,
		})
	}
	resp.MsgList = msglist
	resp.DialogId = in.DialogId
	return resp, nil
}

func (s *Service) GetGroupMsgLabelByDialogId(ctx context.Context, in *v1.GetGroupMsgLabelByDialogIdRequest) (*v1.GetGroupMsgLabelByDialogIdResponse, error) {
	var resp = &v1.GetGroupMsgLabelByDialogIdResponse{}
	msgs, err := s.mr.GetGroupMsgLabelByDialogId(in.DialogId)
	if err != nil {
		return resp, status.Error(codes.Code(code.GetMsgErrGetGroupMsgLabelByDialogIdFailed.Code()), err.Error())
	}

	var msglist = make([]*v1.GroupMessage, 0)
	for _, msg := range msgs {
		msglist = append(msglist, &v1.GroupMessage{
			Id:        uint32(msg.ID),
			UserId:    msg.UserID,
			Content:   msg.Content,
			Type:      uint32(int32(msg.Type)),
			ReplyId:   uint32(msg.ReplyId),
			GroupId:   uint32(msg.GroupID),
			ReadCount: int32(msg.ReadCount),
		})
	}
	resp.DialogId = in.DialogId
	resp.MsgList = msglist
	return resp, nil
}

func (s *Service) SetUserMsgsReadStatus(ctx context.Context, in *v1.SetUserMsgsReadStatusRequest) (*v1.SetUserMsgsReadStatusResponse, error) {
	var resp = &v1.SetUserMsgsReadStatusResponse{}
	//获取阅后即焚消息id
	messages, err := s.mr.GetBatchUserMsgsBurnAfterReadingMessages(in.MsgIds, in.DialogId)
	if err != nil {
		return nil, err
	}
	rids := make([]uint32, 0)
	if len(messages) > 0 {
		for _, msg := range messages {
			rids = append(rids, uint32(msg.ID))
		}
	}
	err = s.db.Transaction(func(tx *gorm.DB) error {
		npo := persistence.NewRepositories(tx)
		if err := npo.Mr.SetUserMsgsReadStatus(in.MsgIds, in.DialogId); err != nil {
			return err
		}
		if len(rids) > 0 {
			err := npo.Mr.LogicalDeleteUserMessages(rids)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return resp, status.Error(codes.Code(code.SetMsgErrSetUserMsgsReadStatusFailed.Code()), err.Error())
	}

	//删除消息
	return resp, nil
}

// 修改用户消息的已读状态
func (s *Service) SetUserMsgReadStatus(ctx context.Context, in *v1.SetUserMsgReadStatusRequest) (*v1.SetUserMsgReadStatusResponse, error) {
	var resp = &v1.SetUserMsgReadStatusResponse{}
	if err := s.mr.SetUserMsgReadStatus(in.MsgId, entity.ReadType(in.IsRead)); err != nil {
		return resp, status.Error(codes.Code(code.SetMsgErrSetUserMsgReadStatusFailed.Code()), err.Error())
	}
	return resp, nil
}

func (s *Service) GetUnreadUserMsgs(ctx context.Context, in *v1.GetUnreadUserMsgsRequest) (*v1.GetUnreadUserMsgsResponse, error) {
	var resp = &v1.GetUnreadUserMsgsResponse{}
	msgs, err := s.mr.GetUnreadUserMsgs(in.UserId, in.DialogId)
	if err != nil {
		return resp, status.Error(codes.Code(code.GetMsgErrGetUnreadUserMsgsFailed.Code()), err.Error())
	}
	var list []*v1.UserMessage
	for _, msg := range msgs {
		list = append(list, &v1.UserMessage{
			Id:         uint32(msg.ID),
			Content:    msg.Content,
			Type:       uint32(msg.Type),
			ReplayId:   uint64(msg.ReplyId),
			SenderId:   msg.SendID,
			ReceiverId: msg.ReceiveID,
			CreatedAt:  msg.CreatedAt,
		})
	}
	resp.UserMessages = list
	return resp, nil
}

func (s *Service) GetUserMsgIdAfterMsgList(ctx context.Context, in *v1.GetUserMsgIdAfterMsgListRequest) (*v1.GetUserMsgIdAfterMsgListResponse, error) {
	var resp = &v1.GetUserMsgIdAfterMsgListResponse{}
	if len(in.List) > 0 {
		for _, v := range in.List {
			list, err := s.mr.GetUserMsgIdAfterMsgList(v.DialogId, v.MsgId)
			if err != nil {
				return nil, err
			}
			if len(list) > 0 {
				mlist := &v1.GetUserMsgIdAfterMsgResponse{}
				for _, msg := range list {
					mlist.UserMessages = append(mlist.UserMessages, &v1.UserMessage{
						Id:         uint32(msg.ID),
						Content:    msg.Content,
						Type:       uint32(msg.Type),
						ReplayId:   uint64(msg.ReplyId),
						SenderId:   msg.SendID,
						ReceiverId: msg.ReceiveID,
						CreatedAt:  msg.CreatedAt,
					})
				}
				mlist.DialogId = v.DialogId
				resp.Messages = append(resp.Messages, mlist)
			}
		}
	}
	return resp, nil
}

func (s *Service) GetGroupMsgIdAfterMsgList(ctx context.Context, in *v1.GetGroupMsgIdAfterMsgListRequest) (*v1.GetGroupMsgIdAfterMsgListResponse, error) {
	var resp = &v1.GetGroupMsgIdAfterMsgListResponse{}
	if len(in.List) > 0 {
		for _, v := range in.List {
			list, err := s.mr.GetGroupMsgIdAfterMsgList(v.DialogId, v.MsgId)
			if err != nil {
				return nil, err
			}
			if len(list) > 0 {
				mlist := &v1.GetGroupMsgIdAfterMsgResponse{}
				for _, msg := range list {
					mlist.GroupMessages = append(mlist.GroupMessages, &v1.GroupMessage{
						Id:        uint32(msg.ID),
						UserId:    msg.UserID,
						Content:   msg.Content,
						Type:      uint32(int32(msg.Type)),
						ReplyId:   uint32(msg.ReplyId),
						GroupId:   uint32(msg.GroupID),
						ReadCount: int32(msg.ReadCount),
					})
				}
				mlist.DialogId = v.DialogId
				resp.Messages = append(resp.Messages, mlist)
			}
		}
	}
	return resp, nil
}

func (s *Service) GetGroupMessageList(ctx context.Context, in *v1.GetGroupMsgListRequest) (*v1.GetGroupMsgListResponse, error) {
	var resp = &v1.GetGroupMsgListResponse{}
	list, err := s.mr.GetGroupMsgList(dataTransformers.GroupMsgList{
		GroupID:    uint32(in.GroupId),
		Content:    in.Content,
		UserID:     in.UserId,
		MsgType:    entity.UserMessageType(in.Type),
		PageNumber: int(in.PageNum),
		PageSize:   int(in.PageSize),
	})
	if err != nil {
		return nil, status.Error(codes.Code(code.MsgErrGetGroupMsgListFailed.Code()), err.Error())
	}

	resp.Total = list.Total
	resp.CurrentPage = list.CurrentPage

	if len(list.GroupMessages) > 0 {
		for _, msg := range list.GroupMessages {
			resp.GroupMessages = append(resp.GroupMessages, &v1.GroupMessage{
				Id:        uint32(msg.ID),
				UserId:    msg.UserID,
				Content:   msg.Content,
				Type:      uint32(int32(msg.Type)),
				ReplyId:   uint32(msg.ReplyId),
				GroupId:   uint32(msg.GroupID),
				ReadCount: int32(msg.ReadCount),
				CreatedAt: msg.CreatedAt,
				DialogId:  uint32(msg.DialogId),
				IsLabel:   v1.MsgLabel(msg.IsLabel),
			})
		}
	}

	return resp, nil
}
