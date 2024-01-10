package service

import (
	"context"
	v1 "github.com/cossim/coss-server/service/msg/api/v1"
	"github.com/cossim/coss-server/service/msg/domain/entity"
)

func (s *Service) SendUserMessage(ctx context.Context, request *v1.SendUserMsgRequest) (*v1.SendUserMsgResponse, error) {
	resp := &v1.SendUserMsgResponse{}

	_, err := s.mr.InsertUserMessage(request.GetSenderId(), request.GetReceiverId(), request.GetContent(), entity.UserMessageType(request.GetType()), uint(request.GetReplayId()), uint(request.GetDialogId()))
	if err != nil {
		return resp, err
	}
	return resp, err
}

func (s *Service) SendGroupMessage(ctx context.Context, request *v1.SendGroupMsgRequest) (*v1.SendGroupMsgResponse, error) {
	resp := &v1.SendGroupMsgResponse{}

	ums, err := s.mr.InsertGroupMessage(request.GetUserId(), uint(request.GetGroupId()), request.GetContent(), entity.UserMessageType(request.GetType()), uint(request.GetReplayId()), uint(request.GetDialogId()))
	if err != nil {
		return resp, err
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
			CreatedAt:  m.CreatedAt.Unix(),
		})
	}
	resp.Total = total
	resp.CurrentPage = current

	return resp, nil
}

func (s *Service) GetLastMsgsForUserWithFriends(ctx context.Context, in *v1.UserMsgsRequest) (*v1.UserMessages, error) {
	resp := &v1.UserMessages{}
	msgs, err := s.mr.GetLastMsgsForUserWithFriends(in.UserId, in.FriendId)
	if err != nil {
		return resp, err
	}
	nmsgs := make([]*v1.UserMessage, 0)
	for _, m := range msgs {
		nmsgs = append(nmsgs, &v1.UserMessage{
			Id:        uint32(m.ID),
			Content:   m.Content,
			Type:      uint32(m.Type),
			ReplayId:  uint64(m.ReplyId),
			ReadAt:    m.ReadAt,
			CreatedAt: m.CreatedAt.Unix(),
		})
	}
	resp.UserMessages = nmsgs
	return resp, nil
}

func (s *Service) GetLastMsgsForGroupsWithIDs(ctx context.Context, in *v1.GroupMsgsRequest) (*v1.GroupMessages, error) {
	resp := &v1.GroupMessages{}
	ids := make([]uint, 0)
	if len(in.GroupId) > 0 {
		for _, id := range in.GroupId {
			ids = append(ids, uint(id))
		}
	}
	msgs, err := s.mr.GetLastMsgsForGroupsWithIDs(ids)
	if err != nil {
		return resp, err
	}
	nmsgs := make([]*v1.GroupMessage, 0)
	for _, m := range msgs {
		nmsgs = append(nmsgs, &v1.GroupMessage{
			Id:        uint32(m.ID),
			Uid:       m.UID,
			Content:   m.Content,
			Type:      uint32(m.Type),
			ReplyId:   uint32(m.ReplyId),
			ReadCount: int32(m.ReadCount),
			CreatedAt: m.CreatedAt.Unix(),
		})
	}
	resp.GroupMessages = nmsgs
	return resp, nil
}

func (s *Service) GetLastMsgsByDialogIds(ctx context.Context, in *v1.GetLastMsgsByDialogIdsRequest) (*v1.GetLastMsgsResponse, error) {
	resp := &v1.GetLastMsgsResponse{}

	ids := make([]uint, 0)
	if len(in.DialogIds) > 0 {
		for _, id := range in.DialogIds {
			ids = append(ids, uint(id))
		}
	}
	result, err := s.mr.GetLastMsgsByDialogIDs(ids)
	if err != nil {
		return resp, err
	}

	if len(result) > 0 {
		for _, m := range result {
			resp.LastMsgs = append(resp.LastMsgs, &v1.LastMsg{
				Id:        uint32(m.ID),
				Type:      uint32(m.Type),
				CreatedAt: m.CreateAt,
				Content:   m.Content,
				DialogId:  uint32(m.DialogId),
			})
		}
	}
	return resp, nil
}
