package grpc

//
//import (
//	"context"
//	v1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
//	"github.com/cossim/coss-server/internal/msg/domain/entity"
//)
//
//func (s *Handler) SetGroupMessageRead(ctx context.Context, in *v1.SetGroupMessagesReadRequest) (*v1.SetGroupMessageReadResponse, error) {
//	resp := &v1.SetGroupMessageReadResponse{}
//	var reads []*entity.GroupMessageRead
//	msgids := make([]uint32, 0)
//
//	//if len(in.List) > 0 {
//	//	reads = make([]*entity.GroupMessageRead, len(in.List))
//	//	for i, _ := range in.List {
//	//		reads[i] = &entity.GroupMessageRead{
//	//			DialogId: uint(in.List[i].DialogId),
//	//			MsgId:    uint(in.List[i].MsgId),
//	//			UserID:   in.List[i].UserId,
//	//			GroupID:  uint(in.List[i].GroupId),
//	//			ReadAt:   in.List[i].ReadAt,
//	//		}
//	//		msgids = append(msgids, in.List[i].MsgId)
//	//	}
//	//}
//	//err := s.db.Transaction(func(tx *gorm.DB) error {
//	//	npo := persistence.NewRepositories(tx)
//	//	msgs, err := npo.Gmr.GetGroupMsgsByIDs(msgids)
//	//	//修改已读数量
//	//	for _, v := range msgs {
//	//		v.ReadCount++
//	//		err := npo.Gmr.UpdateGroupMsgColumn(uint32(v.ID), "read_count", v.ReadCount)
//	//		if err != nil {
//	//			return err
//	//		}
//	//	}
//	//	err = npo.Gmrr.SetGroupMsgReadByMsgs(reads)
//	//	if err != nil {
//	//		return status.Error(codes.Code(code.GroupErrSetGroupMsgReadFailed.Code()), err.Error())
//	//	}
//	//	return nil
//	//})
//	//if err != nil {
//	//	return nil, err
//	//}
//
//	return resp, nil
//}
//
//func (s *Handler) GetGroupMessageReaders(ctx context.Context, in *v1.GetGroupMessageReadersRequest) (*v1.GetGroupMessageReadersResponse, error) {
//	resp := &v1.GetGroupMessageReadersResponse{}
//	//msgs, err := s.gmrr.GetGroupMsgReadUserIdsByMsgID(in.MsgId)
//	//if err != nil {
//	//	return nil, status.Error(codes.Code(code.GroupErrGetGroupMsgReadersFailed.Code()), err.Error())
//	//}
//	resp.UserIds = msgs
//
//	return resp, err
//}
//
//func (s *Handler) GetGroupMessageReadByMsgIdAndUserId(ctx context.Context, in *v1.GetGroupMessageReadByMsgIdAndUserIdRequest) (*v1.GetGroupMessageReadByMsgIdAndUserIdResponse, error) {
//	resp := &v1.GetGroupMessageReadByMsgIdAndUserIdResponse{}
//	//msg, err := s.gmrr.GetGroupMsgReadByMsgIDAndUserID(in.MsgId, in.UserId)
//	//if err != nil {
//	//	return nil, status.Error(codes.Code(code.GroupErrGetGroupMsgReadByMsgIdAndUserIdFailed.Code()), err.Error())
//	//}
//	resp.ReadAt = msg.ReadAt
//	return resp, err
//}
//
//func (s *Handler) ReadAllGroupMsg(ctx context.Context, request *v1.ReadAllGroupMsgRequest) (*v1.ReadAllGroupMsgResponse, error) {
//	resp := &v1.ReadAllGroupMsgResponse{}
//	//
//	//msgids, err := s.gmrr.GetGroupMsgUserReadIdsByDialogID(request.DialogId, request.UserId)
//	//if err != nil {
//	//	return resp, status.Error(codes.Code(code.GroupErrGetGroupMsgReadByMsgIdAndUserIdFailed.Code()), err.Error())
//	//}
//	//
//	//list, err := s.gmr.GetGroupUnreadMsgList(request.DialogId, msgids)
//	//if err != nil {
//	//	return resp, err
//	//}
//	//var reads []*entity.GroupMessageRead
//	//
//	//if len(list) > 0 {
//	//	reads = make([]*entity.GroupMessageRead, len(list))
//	//	for k, v := range list {
//	//		reads[k] = &entity.GroupMessageRead{
//	//			DialogId: v.DialogId,
//	//			MsgId:    v.ID,
//	//			UserID:   request.UserId,
//	//			GroupID:  v.GroupID,
//	//			ReadAt:   time.Now(),
//	//		}
//	//		resp.UnreadGroupMessage = append(resp.UnreadGroupMessage, &v1.UnreadGroupMessage{
//	//			MsgId:  uint32(v.ID),
//	//			UserId: v.UserID,
//	//		})
//	//	}
//	//}
//	//
//	//if err := s.gmrr.SetGroupMsgReadByMsgs(reads); err != nil {
//	//	return resp, status.Error(codes.Code(code.GroupErrGetGroupMsgReadByMsgIdAndUserIdFailed.Code()), err.Error())
//	//}
//
//	return resp, nil
//}
