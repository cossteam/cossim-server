package grpc

//
//import (
//	"context"
//	"fmt"
//	v1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
//	"github.com/cossim/coss-server/pkg/code"
//	"google.golang.org/grpc/codes"
//	"google.golang.org/grpc/status"
//)
//
//func (s *Handler) SendGroupMessage(ctx context.Context, request *v1.SendGroupMsgRequest) (*v1.SendGroupMsgResponse, error) {
//	resp := &v1.SendGroupMsgResponse{}
//
//	//ums, err := s.gmr.InsertGroupMessage(&entity.GroupMessage{
//	//	DialogId:  uint(request.DialogId),
//	//	GroupID:   uint(request.GroupId),
//	//	Type:      entity.UserMessageType(request.Type),
//	//	ReplyId:   uint(request.ReplyId),
//	//	UserID:    request.UserId,
//	//	Content:   request.Content,
//	//	AtAllUser: entity.AtAllUserType(request.AtAllUser),
//	//	AtUsers:   request.AtUsers,
//	//})
//	//if err != nil {
//	//	return resp, status.Error(codes.Code(code.MsgErrInsertGroupMessageFailed.Code()), err.Error())
//	//}
//
//	return &v1.SendGroupMsgResponse{
//		MsgId:   uint32(ums.ID),
//		GroupId: uint32(ums.GroupID),
//	}, nil
//}
//
//func (s *Handler) SendGroupMessageRevert(ctx context.Context, request *v1.MsgIdRequest) (*v1.SendGroupMsgRevertResponse, error) {
//	resp := &v1.SendGroupMsgRevertResponse{}
//	//if err := s.gmr.PhysicalDeleteGroupMessage(request.MsgId); err != nil {
//	//	return resp, status.Error(codes.Code(code.MsgErrDeleteGroupMessageFailed.Code()), err.Error())
//	//}
//	return resp, nil
//}
//
//func (s *Handler) GetLastMsgsForGroupsWithIDs(ctx context.Context, request *v1.GroupMsgsRequest) (*v1.GroupMessages, error) {
//	resp := &v1.GroupMessages{}
//	//ids := make([]uint, 0)
//	//if len(request.GroupId) > 0 {
//	//	for _, id := range request.GroupId {
//	//		ids = append(ids, uint(id))
//	//	}
//	//}
//	//msgs, err := s.gmr.GetLastMsgsForGroupsWithIDs(ids)
//	//if err != nil {
//	//	return resp, status.Error(codes.Code(code.MsgErrGetLastMsgsForGroupsWithIDs.Code()), err.Error())
//	//}
//	nmsgs := make([]*v1.GroupMessage, 0)
//	for _, m := range msgs {
//		nmsgs = append(nmsgs, &v1.GroupMessage{
//			Id:        uint32(m.ID),
//			UserId:    m.UserID,
//			Content:   m.Content,
//			Type:      uint32(m.Type),
//			ReplyId:   uint32(m.ReplyId),
//			ReadCount: int32(m.ReadCount),
//			CreatedAt: m.CreatedAt,
//		})
//	}
//	resp.GroupMessages = nmsgs
//	return resp, nil
//}
//
//func (s *Handler) EditGroupMessage(ctx context.Context, request *v1.EditGroupMsgRequest) (*v1.GroupMessage, error) {
//	var resp = &v1.GroupMessage{}
//	//msg := entity.GroupMessage{
//	//	BaseModel: entity.BaseModel{
//	//		ID: uint(request.GroupMessage.Id),
//	//	},
//	//	Content: request.GroupMessage.Content,
//	//	ReplyId: uint(request.GroupMessage.ReplyId),
//	//	UserID:  request.GroupMessage.UserId,
//	//	GroupID: uint(request.GroupMessage.GroupId),
//	//	Type:    entity.UserMessageType(request.GroupMessage.Type),
//	//	IsLabel: uint(request.GroupMessage.IsLabel),
//	//}
//	//if err := s.gmr.UpdateGroupMessage(msg); err != nil {
//	//	return resp, status.Error(codes.Code(code.MsgErrEditGroupMessageFailed.Code()), err.Error())
//	//}
//	resp = &v1.GroupMessage{
//		Id:        uint32(msg.ID),
//		UserId:    msg.UserID,
//		Content:   msg.Content,
//		Type:      uint32(int32(msg.Type)),
//		ReplyId:   uint32(msg.ReplyId),
//		GroupId:   uint32(msg.GroupID),
//		ReadCount: int32(msg.ReadCount),
//	}
//	//
//	//if s.cacheEnable {
//	//	if err := s.cache.DeleteLastMessage(ctx, request.GroupMessage.DialogId); err != nil {
//	//		log.Printf("delete last message failed: %v", err)
//	//	}
//	//}
//
//	return resp, nil
//}
//
//func (s *Handler) DeleteGroupMessage(ctx context.Context, request *v1.DeleteGroupMsgRequest) (*v1.GroupMessage, error) {
//	var resp = &v1.GroupMessage{}
//	if err := s.gmr.LogicalDeleteGroupMessage(request.MsgId); err != nil {
//		return resp, status.Error(codes.Code(code.MsgErrDeleteGroupMessageFailed.Code()), err.Error())
//	}
//
//	//if s.cacheEnable {
//	//	if err := s.cache.DeleteLastMessage(ctx, request.DialogId); err != nil {
//	//		log.Printf("delete last message failed: %v", err)
//	//	}
//	//}
//
//	return &v1.GroupMessage{
//		Id: request.MsgId,
//	}, nil
//}
//
//func (s *Handler) GetGroupMessageById(ctx context.Context, in *v1.GetGroupMsgByIDRequest) (*v1.GroupMessage, error) {
//	var resp = &v1.GroupMessage{}
//	//msg, err := s.gmr.GetGroupMsgByID(in.MsgId)
//	//if err != nil {
//	//	if err == gorm.ErrRecordNotFound {
//	//		return resp, status.Error(codes.Code(code.GetMsgErrGetGroupMsgByIDFailed.Code()), err.Error())
//	//	}
//	//	return resp, err
//	//}
//	return &v1.GroupMessage{
//		Id:        uint32(msg.ID),
//		UserId:    msg.UserID,
//		DialogId:  uint32(msg.DialogId),
//		Content:   msg.Content,
//		Type:      uint32(int32(msg.Type)),
//		ReplyId:   uint32(msg.ReplyId),
//		GroupId:   uint32(msg.GroupID),
//		ReadCount: int32(msg.ReadCount),
//		CreatedAt: msg.CreatedAt,
//	}, nil
//}
//
//func (s *Handler) SetGroupMsgLabel(ctx context.Context, in *v1.SetGroupMsgLabelRequest) (*v1.SetGroupMsgLabelResponse, error) {
//	var resp = &v1.SetGroupMsgLabelResponse{}
//	if err := s.gmr.UpdateGroupMsgColumn(in.MsgId, "is_label", in.IsLabel); err != nil {
//		return resp, status.Error(codes.Code(code.SetMsgErrSetGroupMsgLabelFailed.Code()), err.Error())
//	}
//	resp.MsgId = in.MsgId
//	return resp, nil
//}
//
//func (s *Handler) GetGroupMsgLabelByDialogId(ctx context.Context, in *v1.GetGroupMsgLabelByDialogIdRequest) (*v1.GetGroupMsgLabelByDialogIdResponse, error) {
//	var resp = &v1.GetGroupMsgLabelByDialogIdResponse{}
//	//msgs, err := s.gmr.GetGroupMsgLabelByDialogId(in.DialogId)
//	//if err != nil {
//	//	return resp, status.Error(codes.Code(code.GetMsgErrGetGroupMsgLabelByDialogIdFailed.Code()), err.Error())
//	//}
//
//	var msglist = make([]*v1.GroupMessage, 0)
//	for _, msg := range msgs {
//		msglist = append(msglist, &v1.GroupMessage{
//			Id:        uint32(msg.ID),
//			UserId:    msg.UserID,
//			Content:   msg.Content,
//			Type:      uint32(int32(msg.Type)),
//			ReplyId:   uint32(msg.ReplyId),
//			GroupId:   uint32(msg.GroupID),
//			ReadCount: int32(msg.ReadCount),
//		})
//	}
//	resp.DialogId = in.DialogId
//	resp.MsgList = msglist
//	return resp, nil
//}
//
//func (s *Handler) GetGroupMsgIdAfterMsgList(ctx context.Context, in *v1.GetGroupMsgIdAfterMsgListRequest) (*v1.GetGroupMsgIdAfterMsgListResponse, error) {
//	var resp = &v1.GetGroupMsgIdAfterMsgListResponse{}
//	//if len(in.List) > 0 {
//	//	for _, v := range in.List {
//	//		list, total, err := s.gmr.GetGroupMsgIdAfterMsgList(v.DialogId, v.MsgId)
//	//		if err != nil {
//	//			return nil, err
//	//		}
//	//		if len(list) > 0 {
//	//			mlist := &v1.GetGroupMsgIdAfterMsgResponse{}
//	//			for _, msg := range list {
//	//				mlist.GroupMessages = append(mlist.GroupMessages, &v1.GroupMessage{
//	//					Id:        uint32(msg.ID),
//	//					UserId:    msg.UserID,
//	//					Content:   msg.Content,
//	//					Type:      uint32(int32(msg.Type)),
//	//					ReplyId:   uint32(msg.ReplyId),
//	//					GroupId:   uint32(msg.GroupID),
//	//					ReadCount: int32(msg.ReadCount),
//	//				})
//	//			}
//	//			mlist.Total = uint64(total)
//	//			mlist.DialogId = v.DialogId
//	//			resp.Messages = append(resp.Messages, mlist)
//	//		}
//	//	}
//	//}
//	return resp, nil
//}
//
//func (s *Handler) GetGroupMessageList(ctx context.Context, in *v1.GetGroupMsgListRequest) (*v1.GetGroupMsgListResponse, error) {
//	var resp = &v1.GetGroupMsgListResponse{}
//	if in.MsgId != 0 {
//		//list, total, err := s.gmr.GetGroupMsgIdBeforeMsgList(in.DialogId, uint32(in.MsgId), int(in.PageSize))
//		//if err != nil {
//		//	return nil, err
//		//}
//
//		if len(list) > 0 {
//			for _, msg := range list {
//				resp.GroupMessages = append(resp.GroupMessages, &v1.GroupMessage{
//					Id:        uint32(msg.ID),
//					UserId:    msg.UserID,
//					Content:   msg.Content,
//					Type:      uint32(int32(msg.Type)),
//					ReplyId:   uint32(msg.ReplyId),
//					GroupId:   uint32(msg.GroupID),
//					ReadCount: int32(msg.ReadCount),
//					CreatedAt: msg.CreatedAt,
//					DialogId:  uint32(msg.DialogId),
//					IsLabel:   v1.MsgLabel(msg.IsLabel),
//					AtUsers:   msg.AtUsers,
//					AtAllUser: v1.AtAllUserType(msg.AtAllUser),
//				})
//			}
//		}
//		resp.Total = total
//		return resp, nil
//	}
//	//list, err := s.gmr.GetGroupMsgList(dataTransformers.GroupMsgList{
//	//	DialogId:   in.DialogId,
//	//	Content:    in.Content,
//	//	UserID:     in.UserId,
//	//	MsgType:    entity.UserMessageType(in.Type),
//	//	PageNumber: int(in.PageNum),
//	//	PageSize:   int(in.PageSize),
//	//})
//	//if err != nil {
//	//	return nil, status.Error(codes.Code(code.MsgErrGetGroupMsgListFailed.Code()), err.Error())
//	//}
//
//	resp.Total = list.Total
//	resp.CurrentPage = list.CurrentPage
//
//	if len(list.GroupMessages) > 0 {
//		for _, msg := range list.GroupMessages {
//			resp.GroupMessages = append(resp.GroupMessages, &v1.GroupMessage{
//				Id:        uint32(msg.ID),
//				UserId:    msg.UserID,
//				Content:   msg.Content,
//				Type:      uint32(int32(msg.Type)),
//				ReplyId:   uint32(msg.ReplyId),
//				GroupId:   uint32(msg.GroupID),
//				ReadCount: int32(msg.ReadCount),
//				CreatedAt: msg.CreatedAt,
//				DialogId:  uint32(msg.DialogId),
//				IsLabel:   v1.MsgLabel(msg.IsLabel),
//				AtUsers:   msg.AtUsers,
//				AtAllUser: v1.AtAllUserType(msg.AtAllUser),
//			})
//		}
//	}
//
//	return resp, nil
//}
//
//func (s *Handler) GetGroupMessagesByIds(ctx context.Context, in *v1.GetGroupMessagesByIdsRequest) (*v1.GetGroupMessagesByIdsResponse, error) {
//	resp := &v1.GetGroupMessagesByIdsResponse{}
//	//msgs, err := s.gmr.GetGroupMsgsByIDs(in.MsgIds)
//	//if err != nil {
//	//	return nil, status.Error(codes.Code(code.GetMsgErrGetGroupMsgByIDFailed.Code()), err.Error())
//	//}
//	if len(msgs) > 0 {
//		for _, msg := range msgs {
//			resp.GroupMessages = append(resp.GroupMessages, &v1.GroupMessage{
//				Id:                     uint32(msg.ID),
//				UserId:                 msg.UserID,
//				Content:                msg.Content,
//				Type:                   uint32(int32(msg.Type)),
//				ReplyId:                uint32(msg.ReplyId),
//				GroupId:                uint32(msg.GroupID),
//				ReadCount:              int32(msg.ReadCount),
//				DialogId:               uint32(msg.DialogId),
//				IsLabel:                v1.MsgLabel(msg.IsLabel),
//				IsBurnAfterReadingType: v1.BurnAfterReadingType(msg.IsBurnAfterReading),
//				AtUsers:                msg.AtUsers,
//				AtAllUser:              v1.AtAllUserType(msg.AtAllUser),
//				CreatedAt:              msg.CreatedAt,
//			})
//		}
//	}
//	return resp, nil
//}
//
//func (s *Handler) GetGroupUnreadMessages(ctx context.Context, in *v1.GetGroupUnreadMessagesRequest) (*v1.GetGroupUnreadMessagesResponse, error) {
//	resp := &v1.GetGroupUnreadMessagesResponse{}
//	//获取群聊消息id
//	//去除不需要的消息类型
//	//ids, err := s.gmr.GetGroupMsgIdsByDialogID(in.DialogId)
//	//if err != nil {
//	//	return nil, err
//	//}
//	//
//	////获取已读消息id
//	//rids, err := s.gmrr.GetGroupMsgUserReadIdsByDialogID(in.DialogId, in.UserId)
//	//if err != nil {
//	//	return nil, err
//	//}
//	//
//	////求差集
//	//msgIds := utils.SliceDifference(ids, rids)
//	//
//	//msgs, err := s.gmr.GetGroupMsgsByIDs(msgIds)
//	//if err != nil {
//	//	return nil, err
//	//}
//	//if len(msgs) > 0 {
//	//	for _, msg := range msgs {
//	//		resp.GroupMessages = append(resp.GroupMessages, &v1.GroupMessage{
//	//			Id:                     uint32(msg.ID),
//	//			UserId:                 msg.UserID,
//	//			Content:                msg.Content,
//	//			Type:                   uint32(msg.Type),
//	//			ReplyId:                uint32(msg.ReplyId),
//	//			GroupId:                uint32(msg.GroupID),
//	//			ReadCount:              int32(msg.ReadCount),
//	//			DialogId:               uint32(msg.DialogId),
//	//			IsLabel:                v1.MsgLabel(msg.IsLabel),
//	//			IsBurnAfterReadingType: v1.BurnAfterReadingType(msg.IsBurnAfterReading),
//	//			AtUsers:                msg.AtUsers,
//	//			AtAllUser:              v1.AtAllUserType(msg.AtAllUser),
//	//			CreatedAt:              msg.CreatedAt,
//	//		})
//	//	}
//	//}
//	//查询指定消息
//	return resp, nil
//}
//
//func (s *Handler) DeleteGroupMessageByDialogId(ctx context.Context, in *v1.DeleteGroupMsgByDialogIdRequest) (*v1.DeleteGroupMsgByDialogIdResponse, error) {
//	resp := &v1.DeleteGroupMsgByDialogIdResponse{}
//	//err := s.gmr.DeleteGroupMessagesByDialogID(in.DialogId)
//	//if err != nil {
//	//	return resp, status.Error(codes.Aborted, fmt.Sprintf("failed to delete group msg: %v", err))
//	//}
//	return resp, err
//}
//
//func (s *Handler) ConfirmDeleteGroupMessageByDialogId(ctx context.Context, in *v1.DeleteGroupMsgByDialogIdRequest) (*v1.DeleteGroupMsgByDialogIdResponse, error) {
//	resp := &v1.DeleteGroupMsgByDialogIdResponse{}
//	err := s.gmr.PhysicalDeleteGroupMessagesByDialogID(in.DialogId)
//	if err != nil {
//		return resp, status.Error(codes.Aborted, fmt.Sprintf("failed to delete group msg: %v", err))
//	}
//	return resp, err
//}
//
//func (s *Handler) DeleteGroupMessageByDialogIdRollback(ctx context.Context, in *v1.DeleteUserMsgByDialogIdRequest) (*v1.DeleteGroupMsgByDialogIdResponse, error) {
//	resp := &v1.DeleteGroupMsgByDialogIdResponse{}
//	//err := s.gmr.UpdateGroupMsgColumnByDialogId(in.DialogId, "deleted_at", 0)
//	//if err != nil {
//	//	return resp, status.Error(codes.Code(code.MsgErrDeleteGroupMessageFailed.Code()), err.Error())
//	//}
//	return resp, err
//}
//
//func (s *Handler) GetGroupLastMessageList(ctx context.Context, request *v1.GetLastMsgListRequest) (*v1.GroupMessages, error) {
//	resp := &v1.GroupMessages{}
//	//msgs, total, err := s.gmr.GetGroupDialogLastMsgs(request.DialogId, int(request.PageNum), int(request.PageSize))
//	//if err != nil {
//	//	return nil, status.Error(codes.Code(code.GetMsgErrGetUserMsgByIDFailed.Code()), err.Error())
//	//}
//	if len(msgs) > 0 {
//		for _, msg := range msgs {
//			resp.GroupMessages = append(resp.GroupMessages, &v1.GroupMessage{
//				Id:                     uint32(msg.ID),
//				UserId:                 msg.UserID,
//				Content:                msg.Content,
//				Type:                   uint32(msg.Type),
//				ReplyId:                uint32(msg.ReplyId),
//				GroupId:                uint32(msg.GroupID),
//				ReadCount:              int32(msg.ReadCount),
//				DialogId:               uint32(msg.DialogId),
//				IsLabel:                v1.MsgLabel(msg.IsLabel),
//				IsBurnAfterReadingType: v1.BurnAfterReadingType(msg.IsBurnAfterReading),
//				AtUsers:                msg.AtUsers,
//				AtAllUser:              v1.AtAllUserType(msg.AtAllUser),
//				CreatedAt:              msg.CreatedAt,
//			})
//		}
//	}
//	resp.Total = uint64(total)
//	return resp, nil
//}
