package grpc

import (
	"context"
	api "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	"github.com/cossim/coss-server/internal/msg/domain/entity"
)

func (s *Handler) SendUserMessage(ctx context.Context, request *api.SendUserMsgRequest) (*api.SendUserMsgResponse, error) {
	resp := &api.SendUserMsgResponse{}
	msg, err := s.umd.SendUserMessage(ctx, &entity.UserMessage{
		Type:               entity.UserMessageType(request.Type),
		DialogId:           uint(request.DialogId),
		ReplyId:            uint(request.ReplyId),
		ReceiveID:          request.ReceiverId,
		SendID:             request.SenderId,
		Content:            request.Content,
		IsBurnAfterReading: request.IsBurnAfterReadingType,
	})
	if err != nil {
		return resp, err
	}
	resp.MsgId = uint32(msg.ID)
	return resp, nil
}

func (s *Handler) SendMultiUserMessage(ctx context.Context, request *api.SendMultiUserMsgRequest) (*api.SendMultiUserMsgResponse, error) {
	resp := &api.SendMultiUserMsgResponse{}
	messages := make([]*entity.UserMessage, 0)
	for _, msg := range request.MsgList {
		messages = append(messages, &entity.UserMessage{
			Type:               entity.UserMessageType(msg.Type),
			DialogId:           uint(msg.DialogId),
			ReplyId:            uint(msg.ReplyId),
			ReceiveID:          msg.ReceiverId,
			SendID:             msg.SenderId,
			Content:            msg.Content,
			IsBurnAfterReading: msg.IsBurnAfterReadingType,
		})
	}
	err := s.umd.SendMultiUserMessage(ctx, messages)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *Handler) ConfirmDeleteUserMessageByDialogId(ctx context.Context, request *api.DeleteUserMsgByDialogIdRequest) (*api.DeleteUserMsgByDialogIdResponse, error) {
	resp := &api.DeleteUserMsgByDialogIdResponse{}
	err := s.umd.DeleteUserMessageByDialogId(ctx, uint(request.DialogId), true)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *Handler) DeleteUserMessageById(ctx context.Context, request *api.DeleteUserMsgByIDRequest) (*api.DeleteUserMsgByIDResponse, error) {
	resp := &api.DeleteUserMsgByIDResponse{}
	err := s.umd.DeleteUserMessageById(ctx, uint(request.ID), request.IsPhysical)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

func (s *Handler) DeleteUserMessageByIDs(ctx context.Context, request *api.DeleteUserMessageByIdsRequest) (*api.DeleteUserMsgByIDResponse, error) {
	resp := &api.DeleteUserMsgByIDResponse{}
	if len(request.IDs) == 0 {
		return resp, nil
	}
	ids := make([]uint, 0)
	for _, id := range request.IDs {
		ids = append(ids, uint(id))
	}
	err := s.umd.DeleteUserMessageByIds(ctx, ids, request.IsPhysical)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

//
//import (
//	"context"
//	"fmt"
//	v1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
//	"github.com/cossim/coss-server/internal/msg/domain/entity"
//	"github.com/cossim/coss-server/internal/msg/infra/persistence"
//	"github.com/cossim/coss-server/pkg/code"
//	"google.golang.org/grpc/codes"
//	"google.golang.org/grpc/status"
//	"gorm.io/gorm"
//	"time"
//)
//
//func (s *Handler) SendUserMessage(ctx context.Context, request *v1.SendUserMsgRequest) (*v1.SendUserMsgResponse, error) {
//	resp := &v1.SendUserMsgResponse{}
//
//	//msg, err := s.umr.InsertUserMessage(&entity.UserMessage{
//	//	Type:               entity.UserMessageType(request.Type),
//	//	DialogID:           uint(request.DialogID),
//	//	ReplyId:            uint(request.ReplyId),
//	//	ReceiveID:          request.ReceiverId,
//	//	SendID:             request.SenderId,
//	//	Content:            request.Content,
//	//	IsBurnAfterReading: entity.BurnAfterReadingType(request.IsBurnAfterReadingType),
//	//})
//	//if err != nil {
//	//	return resp, status.Error(codes.Code(code.MsgErrInsertUserMessageFailed.Code()), err.Error())
//	//}
//	resp.MsgId = uint32(msg.ID)
//
//	//if s.cacheEnable {
//	//	fmt.Println("DeleteLastMessage => ", request.DialogID)
//	//	if err := s.cache.DeleteLastMessage(ctx, request.DialogID); err != nil {
//	//		log.Printf("delete last message failed: %v", err)
//	//	}
//	//}
//
//	return resp, err
//}
//
//func (s *Handler) SendUserMessageRevert(ctx context.Context, request *v1.MsgIdRequest) (*v1.SendUserMsgRevertResponse, error) {
//	resp := &v1.SendUserMsgRevertResponse{}
//	//if err := s.umr.PhysicalDeleteUserMessage(request.MsgId); err != nil {
//	//	return resp, status.Error(codes.Code(code.MsgErrDeleteUserMessageFailed.Code()), err.Error())
//	//}
//	return resp, nil
//}
//
//func (s *Handler) GetUserMessageList(ctx context.Context, request *v1.GetUserMsgListRequest) (*v1.GetUserMsgListResponse, error) {
//	resp := &v1.GetUserMsgListResponse{}
//	//if request.MsgId != 0 {
//	//	//list, total, err := s.umr.GetUserMsgIdBeforeMsgList(request.DialogID, uint32(request.MsgId), int(request.PageSize))
//	//	//if err != nil {
//	//	//	return nil, err
//	//	//}
//	//
//	//	res, err := s.umr.Find(ctx, &entity.UserMsgQuery{
//	//		DialogIds: []uint32{request.DialogID},
//	//		MsgIds:    []uint32{uint32(request.MsgId)},
//	//		PageSize:  int64(request.PageSize),
//	//	})
//	//	if err != nil {
//	//		return nil, err
//	//	}
//	//
//	//	for _, m := range res.Messages {
//	//		resp.UserMessages = append(resp.UserMessages, &v1.UserMessage{
//	//			ID:         uint32(m.ID),
//	//			SenderId:   m.SendID,
//	//			ReceiverId: m.ReceiveID,
//	//			Content:    m.Content,
//	//			Type:       uint32(int32(m.Type)),
//	//			ReplyId:    uint64(m.ReplyId),
//	//			IsRead:     int32(m.IsRead),
//	//			ReadAt:     m.ReadAt,
//	//			CreatedAt:  m.CreatedAt,
//	//			DialogID:   uint32(m.DialogID),
//	//		})
//	//	}
//
//		resp.Total = int32(res.TotalCount)
//		return resp, nil
//	}
//	//ums, total, current := s.umr.GetUserMsgList(request.DialogID, request.ID, request.GetContent(), entity.UserMessageType(request.GetType()), int(request.GetPageNum()), int(request.GetPageSize()))
//	//res, err := s.umr.Find(ctx, &entity.UserMsgQuery{
//	//	DialogIds: []uint32{request.DialogID},
//	//	MsgType:   entity.UserMessageType(request.Type),
//	//	PageNum:   int64(request.PageNum),
//	//	PageSize:  int64(request.PageSize),
//	//	Content:   request.Content,
//	//	SendID:    request.ID,
//	//	StartAt:   int64(request.StartAt),
//	//	EndAt:     int64(request.EndAt),
//	//})
//	//if err != nil {
//	//	return nil, err
//	//}
//	//
//	//for _, m := range res.Messages {
//	//	resp.UserMessages = append(resp.UserMessages, &v1.UserMessage{
//	//		ID:         uint32(m.ID),
//	//		SenderId:   m.SendID,
//	//		ReceiverId: m.ReceiveID,
//	//		Content:    m.Content,
//	//		Type:       uint32(int32(m.Type)),
//	//		ReplyId:    uint64(m.ReplyId),
//	//		IsRead:     int32(m.IsRead),
//	//		ReadAt:     m.ReadAt,
//	//		CreatedAt:  m.CreatedAt,
//	//		DialogID:   uint32(m.DialogID),
//	//	})
//	//}
//	//resp.Total = int32(res.TotalCount)
//	//resp.CurrentPage = int32(res.CurrentPage)
//
//	return resp, nil
//}
//
//func (s *Handler) EditUserMessage(ctx context.Context, request *v1.EditUserMsgRequest) (*v1.UserMessage, error) {
//	var resp = &v1.UserMessage{}
//	msg := entity.UserMessage{
//		BaseModel: entity.BaseModel{
//			ID: uint(request.UserMessage.ID),
//		},
//		Content:   request.UserMessage.Content,
//		ReplyId:   uint(request.UserMessage.ReplyId),
//		SendID:    request.UserMessage.SenderId,
//		ReceiveID: request.UserMessage.ReceiverId,
//		Type:      entity.UserMessageType(request.UserMessage.Type),
//		IsLabel:   uint(request.UserMessage.IsLabel),
//	}
//	//if err := s.umr.UpdateUserMessage(msg); err != nil {
//	//	return resp, status.Error(codes.Code(code.MsgErrEditUserMessageFailed.Code()), err.Error())
//	//}
//	//
//	//if s.cacheEnable {
//	//	if err := s.cache.DeleteLastMessage(ctx, request.UserMessage.DialogID); err != nil {
//	//		log.Printf("delete last message failed: %v", err)
//	//	}
//	//}
//
//	return resp, nil
//}
//
//func (s *Handler) DeleteUserMessage(ctx context.Context, request *v1.DeleteUserMsgRequest) (*v1.UserMessage, error) {
//	var resp = &v1.UserMessage{}
//	err := s.umr.LogicalDeleteUserMessage(request.MsgId)
//	if err != nil {
//		return resp, status.Error(codes.Code(code.MsgErrDeleteUserMessageFailed.Code()), err.Error())
//	}
//
//	//if s.cacheEnable {
//	//	if err := s.cache.DeleteLastMessage(ctx, request.DialogID); err != nil {
//	//		log.Printf("delete last message failed: %v", err)
//	//	}
//	//}
//
//	return &v1.UserMessage{
//		ID: request.MsgId,
//	}, nil
//}
//
//func (s *Handler) GetUserMessageById(ctx context.Context, in *v1.GetUserMsgByIDRequest) (*v1.UserMessage, error) {
//	var resp = &v1.UserMessage{}
//	//msg, err := s.umr.GetUserMsgByID(in.MsgId)
//	//if err != nil {
//	//	return resp, status.Error(codes.Code(code.GetMsgErrGetUserMsgByIDFailed.Code()), err.Error())
//	//}
//
//	resp = &v1.UserMessage{
//		DialogID:   uint32(msg.DialogID),
//		ID:         uint32(msg.ID),
//		Content:    msg.Content,
//		Type:       uint32(int32(msg.Type)),
//		ReplyId:    uint64(msg.ReplyId),
//		SenderId:   msg.SendID,
//		ReceiverId: msg.ReceiveID,
//		CreatedAt:  msg.CreatedAt,
//	}
//	return resp, nil
//}
//
//func (s *Handler) SetUserMsgLabel(ctx context.Context, in *v1.SetUserMsgLabelRequest) (*v1.SetUserMsgLabelResponse, error) {
//	var resp = &v1.SetUserMsgLabelResponse{}
//	//if err := s.umr.UpdateUserMsgColumn(in.MsgId, "is_label", in.IsLabel); err != nil {
//	//	return resp, status.Error(codes.Code(code.SetMsgErrSetUserMsgLabelFailed.Code()), err.Error())
//	//}
//	resp.MsgId = in.MsgId
//	return resp, nil
//}
//
//func (s *Handler) GetUserMsgLabelByDialogId(ctx context.Context, in *v1.GetUserMsgLabelByDialogIdRequest) (*v1.GetUserMsgLabelByDialogIdResponse, error) {
//	var resp = &v1.GetUserMsgLabelByDialogIdResponse{}
//
//	//msgs, err := s.umr.GetUserMsgLabelByDialogId(in.DialogID)
//	//if err != nil {
//	//	return resp, status.Error(codes.Code(code.GetMsgErrGetUserMsgLabelByDialogIdFailed.Code()), err.Error())
//	//}
//
//	var msglist = make([]*v1.UserMessage, 0)
//	for _, msg := range msgs {
//		msglist = append(msglist, &v1.UserMessage{
//			ID:         uint32(msg.ID),
//			Content:    msg.Content,
//			Type:       uint32(msg.Type),
//			ReplyId:    uint64(msg.ReplyId),
//			SenderId:   msg.SendID,
//			ReceiverId: msg.ReceiveID,
//			CreatedAt:  msg.CreatedAt,
//		})
//	}
//	resp.MsgList = msglist
//	resp.DialogID = in.DialogID
//	return resp, nil
//}
//
//func (s *Handler) SetUserMsgsReadStatus(ctx context.Context, in *v1.SetUserMsgsReadStatusRequest) (*v1.SetUserMsgsReadStatusResponse, error) {
//	var resp = &v1.SetUserMsgsReadStatusResponse{}
//	////获取阅后即焚消息id
//	//messages, err := s.umr.GetBatchUserMsgsBurnAfterReadingMessages(in.MsgIds, in.DialogID)
//	//if err != nil {
//	//	return nil, err
//	//}
//	//rids := make([]uint32, 0)
//	//if len(messages) > 0 {
//	//	for _, msg := range messages {
//	//		rids = append(rids, uint32(msg.ID))
//	//	}
//	//}
//	//if in.OpenBurnAfterReadingTimeOut == 0 {
//	//	in.OpenBurnAfterReadingTimeOut = 10
//	//}
//	//err = s.db.Transaction(func(tx *gorm.DB) error {
//	//	npo := persistence.NewRepositories(tx)
//	//	if err := npo.Umr.SetUserMsgsReadStatus(in.MsgIds, in.DialogID); err != nil {
//	//		return err
//	//	}
//	//	if len(rids) > 0 {
//	//		//起一个携程，定时器根据超时时间删除
//	//		go func(rpo *persistence.Repositories) {
//	//			time.Sleep(time.Duration(in.OpenBurnAfterReadingTimeOut) * time.Second)
//	//			err := rpo.Umr.LogicalDeleteUserMessages(rids)
//	//			if err != nil {
//	//				fmt.Println(err.Error())
//	//				return
//	//			}
//	//		}(npo)
//	//	}
//	//
//	//	return nil
//	//})
//	//if err != nil {
//	//	return resp, status.Error(codes.Code(code.SetMsgErrSetUserMsgsReadStatusFailed.Code()), err.Error())
//	//}
//
//	//删除消息
//	return resp, nil
//}
//
//// 修改用户消息的已读状态
//func (s *Handler) SetUserMsgReadStatus(ctx context.Context, in *v1.SetUserMsgReadStatusRequest) (*v1.SetUserMsgReadStatusResponse, error) {
//	var resp = &v1.SetUserMsgReadStatusResponse{}
//	//if err := s.umr.SetUserMsgReadStatus(in.MsgId, entity.ReadType(in.IsRead)); err != nil {
//	//	return resp, status.Error(codes.Code(code.SetMsgErrSetUserMsgReadStatusFailed.Code()), err.Error())
//	//}
//	return resp, nil
//}
//
//func (s *Handler) GetUnreadUserMsgs(ctx context.Context, in *v1.GetUnreadUserMsgsRequest) (*v1.GetUnreadUserMsgsResponse, error) {
//	var resp = &v1.GetUnreadUserMsgsResponse{}
//	//msgs, err := s.umr.GetUnreadUserMsgs(in.ID, in.DialogID)
//	//if err != nil {
//	//	return resp, status.Error(codes.Code(code.GetMsgErrGetUnreadUserMsgsFailed.Code()), err.Error())
//	//}
//	var list []*v1.UserMessage
//	for _, msg := range msgs {
//		list = append(list, &v1.UserMessage{
//			ID:         uint32(msg.ID),
//			Content:    msg.Content,
//			Type:       uint32(msg.Type),
//			ReplyId:    uint64(msg.ReplyId),
//			SenderId:   msg.SendID,
//			ReceiverId: msg.ReceiveID,
//			CreatedAt:  msg.CreatedAt,
//			ReadAt:     msg.ReadAt,
//			IsRead:     int32(msg.IsRead),
//			DialogID:   uint32(msg.DialogID),
//		})
//	}
//	resp.UserMessages = list
//	return resp, nil
//}
//
//func (s *Handler) GetUserMsgIdAfterMsgList(ctx context.Context, in *v1.GetUserMsgIdAfterMsgListRequest) (*v1.GetUserMsgIdAfterMsgListResponse, error) {
//	var resp = &v1.GetUserMsgIdAfterMsgListResponse{}
//	//if len(in.List) > 0 {
//	//	for _, v := range in.List {
//	//		list, total, err := s.umr.GetUserMsgIdAfterMsgList(v.DialogID, v.MsgId)
//	//		if err != nil {
//	//			return nil, err
//	//		}
//	//		if len(list) > 0 {
//	//			mlist := &v1.GetUserMsgIdAfterMsgResponse{}
//	//			for _, msg := range list {
//	//				mlist.UserMessages = append(mlist.UserMessages, &v1.UserMessage{
//	//					ID:         uint32(msg.ID),
//	//					Content:    msg.Content,
//	//					Type:       uint32(msg.Type),
//	//					ReplyId:    uint64(msg.ReplyId),
//	//					SenderId:   msg.SendID,
//	//					ReceiverId: msg.ReceiveID,
//	//					CreatedAt:  msg.CreatedAt,
//	//				})
//	//			}
//	//			mlist.Total = uint64(total)
//	//			mlist.DialogID = v.DialogID
//	//			resp.Messages = append(resp.Messages, mlist)
//	//		}
//	//	}
//	//}
//	return resp, nil
//}
//
//func (s *Handler) GetUserMessagesByIds(ctx context.Context, in *v1.GetUserMessagesByIdsRequest) (*v1.GetUserMessagesByIdsResponse, error) {
//	resp := &v1.GetUserMessagesByIdsResponse{}
//	//msgs, err := s.umr.GetUserMsgByIDs(in.MsgIds)
//	//if err != nil {
//	//	return nil, status.Error(codes.Code(code.GetMsgErrGetUserMsgByIDFailed.Code()), err.Error())
//	//}
//	if len(msgs) > 0 {
//		for _, msg := range msgs {
//			resp.UserMessages = append(resp.UserMessages, &v1.UserMessage{
//				ID:                     uint32(msg.ID),
//				SenderId:               msg.SendID,
//				ReceiverId:             msg.ReceiveID,
//				ReadAt:                 msg.ReadAt,
//				IsRead:                 int32(msg.IsRead),
//				Content:                msg.Content,
//				Type:                   uint32(msg.Type),
//				ReplyId:                uint64(msg.ReplyId),
//				DialogID:               uint32(msg.DialogID),
//				IsLabel:                v1.MsgLabel(msg.IsLabel),
//				IsBurnAfterReadingType: v1.BurnAfterReadingType(msg.IsBurnAfterReading),
//				CreatedAt:              msg.CreatedAt,
//			})
//		}
//	}
//	return resp, nil
//}
//
//func (s *Handler) SendMultiUserMessage(ctx context.Context, in *v1.SendMultiUserMsgRequest) (*v1.SendMultiUserMsgResponse, error) {
//	resp := &v1.SendMultiUserMsgResponse{}
//	if len(in.MsgList) > 0 {
//		list := make([]*entity.UserMessage, 0)
//		for _, msg := range in.MsgList {
//			list = append(list, &entity.UserMessage{
//				SendID:    msg.SenderId,
//				ReceiveID: msg.ReceiverId,
//				Content:   msg.Content,
//				Type:      entity.UserMessageType(msg.Type),
//				//ReplyId:            uint(msg.ReplyId),
//				DialogID: uint(msg.DialogID),
//				//IsBurnAfterReading: entity.BurnAfterReadingType(msg.IsBurnAfterReadingType),
//			})
//		}
//		//err := s.umr.InsertUserMessages(list)
//		//if err != nil {
//		//	return nil, status.Error(codes.Code(code.MsgErrSendMultipleFailed.Code()), err.Error())
//		//}
//	}
//	return resp, nil
//}
//
//func (s *Handler) DeleteUserMessageByDialogId(ctx context.Context, in *v1.DeleteUserMsgByDialogIdRequest) (*v1.DeleteUserMsgByDialogIdResponse, error) {
//	resp := &v1.DeleteUserMsgByDialogIdResponse{}
//	//err := s.umr.DeleteUserMessagesByDialogID(in.DialogID)
//	//if err != nil {
//	//	return resp, status.Error(codes.Aborted, fmt.Sprintf("failed to delete user msg: %v", err))
//	//}
//	return resp, err
//}
//
//func (s *Handler) ConfirmDeleteUserMessageByDialogId(ctx context.Context, in *v1.DeleteUserMsgByDialogIdRequest) (*v1.DeleteUserMsgByDialogIdResponse, error) {
//	resp := &v1.DeleteUserMsgByDialogIdResponse{}
//	err := s.umr.PhysicalDeleteUserMessagesByDialogID(in.DialogID)
//	if err != nil {
//		return resp, status.Error(codes.Aborted, fmt.Sprintf("failed to delete user msg: %v", err))
//	}
//	return resp, err
//}
//
//func (s *Handler) DeleteUserMessageByDialogIdRollback(ctx context.Context, in *v1.DeleteUserMsgByDialogIdRequest) (*v1.DeleteUserMsgByDialogIdResponse, error) {
//	resp := &v1.DeleteUserMsgByDialogIdResponse{}
//	//err := s.umr.UpdateUserMsgColumnByDialogId(in.DialogID, "deleted_at", 0)
//	//if err != nil {
//	//	return resp, status.Error(codes.Code(code.MsgErrDeleteUserMessageFailed.Code()), err.Error())
//	//}
//	return resp, err
//}
//
//func (s *Handler) GetUserLastMessageList(ctx context.Context, request *v1.GetLastMsgListRequest) (*v1.UserMessages, error) {
//	resp := &v1.UserMessages{}
//	//msgs, total, err := s.umr.GetUserDialogLastMsgs(request.DialogID, int(request.PageNum), int(request.PageSize))
//	//if err != nil {
//	//	return nil, status.Error(codes.Code(code.GetMsgErrGetUserMsgByIDFailed.Code()), err.Error())
//	//}
//	if len(msgs) > 0 {
//		for _, msg := range msgs {
//			resp.UserMessages = append(resp.UserMessages, &v1.UserMessage{
//				ID:                     uint32(msg.ID),
//				SenderId:               msg.SendID,
//				ReceiverId:             msg.ReceiveID,
//				ReadAt:                 msg.ReadAt,
//				IsRead:                 int32(msg.IsRead),
//				Content:                msg.Content,
//				Type:                   uint32(msg.Type),
//				ReplyId:                uint64(msg.ReplyId),
//				DialogID:               uint32(msg.DialogID),
//				IsLabel:                v1.MsgLabel(msg.IsLabel),
//				IsBurnAfterReadingType: v1.BurnAfterReadingType(msg.IsBurnAfterReading),
//				CreatedAt:              msg.CreatedAt,
//			})
//		}
//	}
//	resp.Total = uint64(total)
//	return resp, nil
//}
//
//func (s *Handler) ReadAllUserMsg(ctx context.Context, request *v1.ReadAllUserMsgRequest) (*v1.ReadAllUserMsgResponse, error) {
//	resp := &v1.ReadAllUserMsgResponse{}
//	//if err := s.umr.ReadAllUserMsg(request.ID, request.DialogID); err != nil {
//	//	return nil, status.Error(codes.Code(code.SetMsgErrSetUserMsgsReadStatusFailed.Code()), err.Error())
//	//}
//	return resp, nil
//}
//
//func (s *Handler) DeleteUserMessageById(ctx context.Context, request *v1.DeleteUserMsgByIDRequest) (*v1.DeleteUserMsgByIDResponse, error) {
//	resp := &v1.DeleteUserMsgByIDResponse{}
//	//if err := s.umr.PhysicalDeleteUserMessage(request.ID); err != nil {
//	//	return resp, status.Error(codes.Code(code.MsgErrDeleteUserMessageFailed.Code()), err.Error())
//	//}
//	return resp, nil
//}
//
//func (s *Handler) GetLastMsgsForUserWithFriends(ctx context.Context, request *v1.UserMsgsRequest) (*v1.UserMessages, error) {
//	resp := &v1.UserMessages{}
//	//msgs, err := s.umr.GetLastMsgsForUserWithFriends(request.ID, request.FriendId)
//	//if err != nil {
//	//	return resp, status.Error(codes.Code(code.MsgErrGetLastMsgsForUserWithFriends.Code()), err.Error())
//	//}
//	nmsgs := make([]*v1.UserMessage, 0)
//	for _, m := range msgs {
//		nmsgs = append(nmsgs, &v1.UserMessage{
//			ID:        uint32(m.ID),
//			Content:   m.Content,
//			Type:      uint32(m.Type),
//			ReplyId:   uint64(m.ReplyId),
//			ReadAt:    m.ReadAt,
//			CreatedAt: m.CreatedAt,
//		})
//	}
//	resp.UserMessages = nmsgs
//	return resp, nil
//}
//
//func (s *Handler) GetLastUserMsgsByDialogIds(ctx context.Context, request *v1.GetLastMsgsByDialogIdsRequest) (*v1.GetLastMsgsResponse, error) {
//	resp := &v1.GetLastMsgsResponse{}
//
//	//ids := make([]uint, 0)
//	//if len(request.DialogIds) > 0 {
//	//	for _, id := range request.DialogIds {
//	//		ids = append(ids, uint(id))
//	//	}
//	//}
//	//
//	////获取群聊对话最后一条消息
//	//result1, err := s.gmr.GetLastGroupMsgsByDialogIDs(ids)
//	//if err != nil {
//	//	return resp, status.Error(codes.Code(code.MsgErrGetLastMsgsByDialogIds.Code()), err.Error())
//	//}
//	//
//	//result2, err := s.umr.GetLastUserMsgsByDialogIDs(ids)
//	//if err != nil {
//	//	return resp, status.Error(codes.Code(code.MsgErrGetLastMsgsByDialogIds.Code()), err.Error())
//	//}
//	//
//	//for _, m := range result1 {
//	//	//查询是否已读
//	//	read, err := s.gmrr.GetGroupMsgReadByMsgIDAndUserID(uint32(m.ID), request.ID)
//	//	if err != nil && err != gorm.ErrRecordNotFound {
//	//		return resp, status.Error(codes.Code(code.MsgErrGetLastMsgsByDialogIds.Code()), err.Error())
//	//	}
//	//
//	//	isRead := v1.ReadType_NotRead
//	//	if read.ReadAt != 0 {
//	//		isRead = v1.ReadType_IsRead
//	//	}
//	//
//	//	resp.LastMsgs = append(resp.LastMsgs, &v1.LastMsg{
//	//		ID:                     uint32(m.ID),
//	//		Type:                   uint32(m.Type),
//	//		CreatedAt:              m.CreatedAt,
//	//		Content:                m.Content,
//	//		SenderId:               m.ID,
//	//		DialogID:               uint32(m.DialogID),
//	//		IsBurnAfterReadingType: v1.BurnAfterReadingType(m.IsBurnAfterReading),
//	//		AtUsers:                m.AtUsers,
//	//		AtAllUser:              v1.AtAllUserType(m.AtAllUser),
//	//		IsLabel:                v1.MsgLabel(m.IsLabel),
//	//		ReplyId:                uint32(m.ReplyId),
//	//		IsRead:                 int32(isRead),
//	//		GroupID:                uint32(m.GroupID),
//	//		ReadAt:                 read.ReadAt,
//	//	})
//	//}
//	//
//	//for _, m := range result2 {
//	//	resp.LastMsgs = append(resp.LastMsgs, &v1.LastMsg{
//	//		ID:                     uint32(m.ID),
//	//		Type:                   uint32(m.Type),
//	//		CreatedAt:              m.CreatedAt,
//	//		Content:                m.Content,
//	//		SenderId:               m.SendID,
//	//		DialogID:               uint32(m.DialogID),
//	//		ReceiverId:             m.ReceiveID,
//	//		IsBurnAfterReadingType: v1.BurnAfterReadingType(m.IsBurnAfterReading),
//	//		IsLabel:                v1.MsgLabel(m.IsLabel),
//	//		ReplyId:                uint32(m.ReplyId),
//	//		IsRead:                 int32(m.IsRead),
//	//		ReadAt:                 m.ReadAt,
//	//	})
//	//}
//
//	//if s.cacheEnable {
//	//	for _, v := range resp.LastMsgs {
//	//		if err := s.cache.SetLastMessage(ctx, v.DialogID, v, cache.MsgExpireTime); err != nil {
//	//			log.Printf("set last message to cache failed, err: %v", err)
//	//		}
//	//	}
//	//}
//
//	return resp, nil
//}
