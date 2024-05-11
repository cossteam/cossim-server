package service

import (
	"context"
	groupApi "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	"github.com/cossim/coss-server/internal/msg/api/http/model"
	pushv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils"
	pkgtime "github.com/cossim/coss-server/pkg/utils/time"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/workflow"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/lithammer/shortuuid/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// 推送群聊消息
func (s *Service) sendWsGroupMsg(ctx context.Context, uIds []string, driverId string, msg *pushv1.SendWsGroupMsg) {
	bytes, err := utils.StructToBytes(msg)
	if err != nil {
		return
	}

	//发送群聊消息
	for _, uid := range uIds {
		m := &pushv1.WsMsg{Uid: uid, DriverId: driverId, Event: pushv1.WSEventType_SendGroupMessageEvent, PushOffline: true, SendAt: pkgtime.Now(), Data: &any.Any{Value: bytes}}

		//查询是否静默通知
		groupRelation, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
			GroupId: uint32(msg.GroupId),
			UserId:  uid,
		})
		if err != nil {
			s.logger.Error("获取群聊关系失败", zap.Error(err))
			continue
		}

		//判断是否静默通知
		if groupRelation.IsSilent == relationgrpcv1.GroupSilentNotificationType_GroupSilent {
			m.Event = pushv1.WSEventType_SendSilentGroupMessageEvent
		}

		bytes2, err := utils.StructToBytes(m)
		if err != nil {
			return
		}

		_, err = s.pushService.Push(ctx, &pushv1.PushRequest{
			Type: pushv1.Type_Ws,
			Data: bytes2,
		})
		if err != nil {
			s.logger.Error("发送消息失败", zap.Error(err))
		}

	}
}

func (s *Service) SendGroupMsg(ctx context.Context, userID string, driverId string, req *model.SendGroupMsgRequest) (interface{}, error) {
	group, err := s.groupService.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{
		Gid: req.GroupId,
	})
	if err != nil {
		s.logger.Error("获取群聊信息失败", zap.Error(err))
		return nil, err
	}

	if group.Status != groupgrpcv1.GroupStatus_GROUP_STATUS_NORMAL {
		return nil, code.GroupErrGroupStatusNotAvailable
	}

	if group.SilenceTime != 0 {
		return nil, code.GroupErrGroupIsSilence
	}

	if !model.IsAllowedConversationType(req.IsBurnAfterReadingType) {
		return nil, code.MsgErrInsertUserMessageFailed
	}

	groupRelation, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		GroupId: req.GroupId,
		UserId:  userID,
	})
	if err != nil {
		s.logger.Error("获取群聊关系失败", zap.Error(err))
		return nil, err
	}

	if groupRelation.MuteEndTime > pkgtime.Now() && groupRelation.MuteEndTime != 0 {
		return nil, code.GroupErrUserIsMuted
	}

	dialogs, err := s.relationDialogService.GetDialogByIds(ctx, &relationgrpcv1.GetDialogByIdsRequest{
		DialogIds: []uint32{req.DialogId},
	})
	if err != nil {
		s.logger.Error("获取会话失败", zap.Error(err))
		return nil, err
	}
	if len(dialogs.Dialogs) == 0 {
		return nil, code.DialogErrGetDialogUserByDialogIDAndUserIDFailed
	}

	_, err = s.relationDialogService.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		DialogId: req.DialogId,
		UserId:   userID,
	})
	if err != nil {
		s.logger.Error("获取用户会话失败", zap.Error(err))
		return nil, code.DialogErrGetDialogUserByDialogIDAndUserIDFailed
	}

	//查询群聊所有用户id
	uids, err := s.relationGroupService.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{
		GroupId: req.GroupId,
	})

	var message *msggrpcv1.SendGroupMsgResponse

	workflow.InitGrpc(s.dtmGrpcServer, s.relationServiceAddr, grpc.NewServer())
	gid := shortuuid.New()
	wfName := "send_group_msg_workflow_" + gid
	if err := workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {

		_, err := s.relationDialogService.BatchCloseOrOpenDialog(ctx, &relationgrpcv1.BatchCloseOrOpenDialogRequest{
			DialogId: req.DialogId,
			Action:   relationgrpcv1.CloseOrOpenDialogType_OPEN,
			UserIds:  uids.UserIds,
		})
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
		}

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := s.relationDialogService.BatchCloseOrOpenDialog(ctx, &relationgrpcv1.BatchCloseOrOpenDialogRequest{
				DialogId: req.DialogId,
				Action:   relationgrpcv1.CloseOrOpenDialogType_CLOSE,
				UserIds:  uids.UserIds,
			})
			return err
		})

		message, err = s.msgService.SendGroupMessage(ctx, &msggrpcv1.SendGroupMsgRequest{
			DialogId:               req.DialogId,
			GroupId:                req.GroupId,
			UserId:                 userID,
			Content:                req.Content,
			Type:                   uint32(req.Type),
			ReplyId:                req.ReplyId,
			AtUsers:                req.AtUsers,
			AtAllUser:              msggrpcv1.AtAllUserType(req.AtAllUser),
			IsBurnAfterReadingType: msggrpcv1.BurnAfterReadingType(req.IsBurnAfterReadingType),
		})
		// 发送成功进行消息推送
		if err != nil {
			s.logger.Error("发送消息失败", zap.Error(err))
			return status.Error(codes.Aborted, err.Error())
		}

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := s.msgService.SendGroupMessageRevert(wf.Context, &msggrpcv1.MsgIdRequest{MsgId: message.MsgId})

			return err
		})
		// 发送成功后添加自己的已读记录
		data2 := &msggrpcv1.SetGroupMessageReadRequest{
			MsgId:    message.MsgId,
			GroupId:  message.GroupId,
			DialogId: req.DialogId,
			UserId:   userID,
			ReadAt:   pkgtime.Now(),
		}

		var list []*msggrpcv1.SetGroupMessageReadRequest
		list = append(list, data2)
		_, err = s.msgGroupService.SetGroupMessageRead(context.Background(), &msggrpcv1.SetGroupMessagesReadRequest{
			List: list,
		})
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
		}
		return err
	}); err != nil {
		return "", err
	}

	if err := workflow.Execute(wfName, gid, nil); err != nil {
		return "", code.MsgErrInsertGroupMessageFailed
	}

	//查询发送者信息
	info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}

	resp := &model.SendGroupMsgResponse{
		MsgId: message.MsgId,
	}

	if req.ReplyId != 0 {
		msg, err := s.msgService.GetGroupMessageById(ctx, &msggrpcv1.GetGroupMsgByIDRequest{
			MsgId: req.ReplyId,
		})
		if err != nil {
			return nil, err
		}

		userInfo, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
			UserId: msg.UserId,
		})
		if err != nil {
			return nil, err
		}

		resp.ReplyMsg = &model.Message{
			MsgType:  uint(msg.Type),
			Content:  msg.Content,
			SenderId: msg.UserId,
			SendAt:   msg.GetCreatedAt(),
			MsgId:    uint64(msg.Id),
			SenderInfo: model.SenderInfo{
				UserId: userInfo.UserId,
				Name:   userInfo.NickName,
				Avatar: userInfo.Avatar,
			},
			IsBurnAfterReading: model.BurnAfterReadingType(msg.IsBurnAfterReadingType),
			IsLabel:            model.LabelMsgType(msg.IsLabel),
			ReplyId:            msg.ReplyId,
		}
	}

	rmsg := &pushv1.MessageInfo{}
	if resp.ReplyMsg != nil {
		rmsg = &pushv1.MessageInfo{
			GroupId:  resp.ReplyMsg.GroupId,
			MsgType:  uint32(resp.ReplyMsg.MsgType),
			Content:  resp.ReplyMsg.Content,
			SenderId: resp.ReplyMsg.SenderId,
			SendAt:   resp.ReplyMsg.SendAt,
			MsgId:    resp.ReplyMsg.MsgId,
			SenderInfo: &pushv1.SenderInfo{
				UserId: resp.ReplyMsg.SenderInfo.UserId,
				Avatar: resp.ReplyMsg.SenderInfo.Avatar,
				Name:   resp.ReplyMsg.SenderInfo.Name,
			},
			ReceiverInfo: &pushv1.SenderInfo{
				UserId: resp.ReplyMsg.ReceiverInfo.UserId,
				Avatar: resp.ReplyMsg.ReceiverInfo.Avatar,
				Name:   resp.ReplyMsg.ReceiverInfo.Name,
			},
			AtAllUser:          uint64(resp.ReplyMsg.AtAllUser),
			AtUsers:            resp.ReplyMsg.AtUsers,
			IsBurnAfterReading: uint64(resp.ReplyMsg.IsBurnAfterReading),
			IsLabel:            int32(resp.ReplyMsg.IsLabel),
			ReplyId:            resp.ReplyMsg.ReplyId,
			IsRead:             resp.ReplyMsg.IsRead,
			ReadAt:             resp.ReplyMsg.ReadAt,
		}
	}
	s.sendWsGroupMsg(ctx, uids.UserIds, driverId, &pushv1.SendWsGroupMsg{
		MsgId:              message.MsgId,
		GroupId:            int64(req.GroupId),
		SenderId:           userID,
		Content:            req.Content,
		MsgType:            uint32(req.Type),
		ReplyId:            req.ReplyId,
		SendAt:             pkgtime.Now(),
		DialogId:           req.DialogId,
		AtUsers:            req.AtUsers,
		AtAllUsers:         uint32(req.AtAllUser),
		IsBurnAfterReading: uint32(req.IsBurnAfterReadingType),
		SenderInfo: &pushv1.SenderInfo{
			Avatar: info.Avatar,
			Name:   info.NickName,
			UserId: userID,
		},
		ReplyMsg: rmsg,
	})

	return resp, nil
}

func (s *Service) EditGroupMsg(ctx context.Context, userID string, driverId string, msgID uint32, content string) (interface{}, error) {
	//获取消息
	msginfo, err := s.msgService.GetGroupMessageById(ctx, &msggrpcv1.GetGroupMsgByIDRequest{
		MsgId: msgID,
	})
	if err != nil {
		s.logger.Error("获取消息失败", zap.Error(err))
		return nil, err
	}
	if msginfo.UserId != userID {
		return nil, code.Unauthorized
	}

	//判断是否在对话内
	userIds, err := s.relationDialogService.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
		DialogId: msginfo.DialogId,
	})
	if err != nil {
		s.logger.Error("获取用户对话信息失败", zap.Error(err))
		return nil, err
	}
	var found bool
	for _, user := range userIds.UserIds {
		if user == userID {
			found = true
			break
		}
	}
	if !found {
		return nil, code.DialogErrGetDialogByIdFailed
	}

	// 调用相应的 gRPC 客户端方法来编辑群消息
	_, err = s.msgService.EditGroupMessage(ctx, &msggrpcv1.EditGroupMsgRequest{
		GroupMessage: &msggrpcv1.GroupMessage{
			Id:      msgID,
			Content: content,
		},
	})
	if err != nil {
		s.logger.Error("编辑群消息失败", zap.Error(err))
		return nil, err
	}

	msginfo.Content = content
	s.SendMsgToUsers(userIds.UserIds, driverId, pushv1.WSEventType_EditMsgEvent, msginfo, true)

	return msgID, nil
}

func (s *Service) RecallGroupMsg(ctx context.Context, userID string, driverId string, msgID uint32) (interface{}, error) {
	//获取消息
	msginfo, err := s.msgService.GetGroupMessageById(ctx, &msggrpcv1.GetGroupMsgByIDRequest{
		MsgId: msgID,
	})
	if err != nil {
		s.logger.Error("获取消息失败", zap.Error(err))
		return nil, err
	}

	if model.IsPromptMessageType(model.UserMessageType(msginfo.Type)) {
		return nil, code.MsgErrDeleteGroupMessageFailed
	}

	if msginfo.UserId != userID {
		return nil, code.Unauthorized
	}

	//判断发送时间是否超过两分钟
	if pkgtime.IsTimeDifferenceGreaterThanTwoMinutes(pkgtime.Now(), msginfo.CreatedAt) {
		return nil, code.MsgErrTimeoutExceededCannotRevoke
	}

	//判断是否在对话内
	userIds, err := s.relationDialogService.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
		DialogId: msginfo.DialogId,
	})
	if err != nil {
		s.logger.Error("获取用户对话信息失败", zap.Error(err))
		return nil, err
	}
	var found bool
	for _, user := range userIds.UserIds {
		if user == userID {
			found = true
			break
		}
	}
	if !found {
		return nil, code.DialogErrGetDialogByIdFailed
	}

	msg2 := &model.SendGroupMsgRequest{
		DialogId: msginfo.DialogId,
		GroupId:  msginfo.GroupId,
		Content:  msginfo.Content,
		ReplyId:  msginfo.Id,
		Type:     model.MessageTypeDelete,
	}
	_, err = s.SendGroupMsg(ctx, userID, driverId, msg2)
	if err != nil {
		return nil, err
	}

	// 调用相应的 gRPC 客户端方法来撤回群消息
	msg, err := s.msgService.DeleteGroupMessage(ctx, &msggrpcv1.DeleteGroupMsgRequest{
		MsgId: msgID,
	})
	if err != nil {
		s.logger.Error("撤回群消息失败", zap.Error(err))
		return nil, err
	}

	return msg.Id, nil
}

func (s *Service) LabelGroupMessage(ctx context.Context, userID string, driverId string, msgID uint32, label model.LabelMsgType) (interface{}, error) {
	// 获取群聊消息
	msginfo, err := s.msgService.GetGroupMessageById(ctx, &msggrpcv1.GetGroupMsgByIDRequest{
		MsgId: msgID,
	})
	if err != nil {
		s.logger.Error("获取群聊消息失败", zap.Error(err))
		return nil, err
	}

	if model.IsPromptMessageType(model.UserMessageType(msginfo.Type)) {
		return nil, code.SetMsgErrSetGroupMsgLabelFailed
	}

	//判断是否在对话内
	userIds, err := s.relationDialogService.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
		DialogId: msginfo.DialogId,
	})
	if err != nil {
		s.logger.Error("获取对话用户失败", zap.Error(err))
		return nil, err
	}

	found := false
	for _, v := range userIds.UserIds {
		if v == userID {
			found = true
			break
		}
	}
	if !found {
		return nil, code.RelationGroupErrNotInGroup
	}

	// 调用 gRPC 客户端方法将群聊消息设置为标注状态
	_, err = s.msgService.SetGroupMsgLabel(ctx, &msggrpcv1.SetGroupMsgLabelRequest{
		MsgId:   msgID,
		IsLabel: msggrpcv1.MsgLabel(label),
	})
	if err != nil {
		s.logger.Error("设置群聊消息标注失败", zap.Error(err))
		return nil, err
	}

	msginfo.IsLabel = msggrpcv1.MsgLabel(label)
	msg2 := &model.SendGroupMsgRequest{
		DialogId: msginfo.DialogId,
		GroupId:  msginfo.GroupId,
		Content:  msginfo.Content,
		ReplyId:  msginfo.Id,
		Type:     model.MessageTypeLabel,
	}

	if label == model.NotLabel {
		msg2.Type = model.MessageTypeCancelLabel
	}

	_, err = s.SendGroupMsg(ctx, userID, driverId, msg2)
	if err != nil {
		return nil, err
	}
	//s.SendMsgToUsers(userIds.UserIds, driverId, constants.LabelMsgEvent, msginfo, true)
	return nil, nil
}

func (s *Service) GetGroupLabelMsgList(ctx context.Context, userID string, dialogId uint32) (interface{}, error) {
	_, err := s.relationDialogService.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		UserId:   userID,
		DialogId: dialogId,
	})
	if err != nil {
		s.logger.Error("获取对话用户失败", zap.Error(err))
		return nil, err
	}

	msgs, err := s.msgService.GetGroupMsgLabelByDialogId(ctx, &msggrpcv1.GetGroupMsgLabelByDialogIdRequest{
		DialogId: dialogId,
	})
	if err != nil {
		s.logger.Error("获取群聊消息标注失败", zap.Error(err))
		return nil, err
	}

	return msgs, nil
}

func (s *Service) GetGroupMessageList(c *gin.Context, id string, request *model.MsgListRequest) (interface{}, error) {
	//查询对话信息
	byId, err := s.relationDialogService.GetDialogById(c, &relationgrpcv1.GetDialogByIdRequest{
		DialogId: request.DialogId,
	})
	if err != nil {
		return nil, err
	}

	if byId.GroupId == 0 {
		return nil, code.MsgErrGetGroupMsgListFailed
	}

	_, err = s.groupService.GetGroupInfoByGid(c, &groupApi.GetGroupInfoRequest{
		Gid: byId.GroupId,
	})
	if err != nil {
		s.logger.Error("获取群聊信息失败", zap.Error(err))
		return nil, err
	}

	_, err = s.relationGroupService.GetGroupRelation(c, &relationgrpcv1.GetGroupRelationRequest{
		GroupId: byId.GroupId,
		UserId:  id,
	})
	if err != nil {
		s.logger.Error("获取群聊关系失败", zap.Error(err))
		return nil, err
	}

	msg, err := s.msgService.GetGroupMessageList(c, &msggrpcv1.GetGroupMsgListRequest{
		DialogId: request.DialogId,
		UserId:   request.UserId,
		Content:  request.Content,
		Type:     request.Type,
		PageNum:  int32(request.PageNum),
		PageSize: int32(request.PageSize),
		MsgId:    uint64(request.MsgId),
	})
	if err != nil {
		s.logger.Error("获取群聊消息列表失败", zap.Error(err))
		return nil, err
	}

	resp := &model.GetGroupMsgListResponse{}
	resp.CurrentPage = msg.CurrentPage
	resp.Total = msg.Total

	msgList := make([]*model.GroupMessage, 0)
	for _, v := range msg.GroupMessages {
		ReadAt := 0
		isRead := 0
		//查询是否已读
		msgRead, err := s.msgGroupService.GetGroupMessageReadByMsgIdAndUserId(c, &msggrpcv1.GetGroupMessageReadByMsgIdAndUserIdRequest{
			MsgId:  v.Id,
			UserId: request.UserId,
		})
		if err != nil {
			s.logger.Error("获取群聊消息是否已读失败", zap.Error(err))
		}
		if msgRead != nil {
			ReadAt = int(msgRead.ReadAt)
			isRead = 1
		}

		//查询信息
		info, err := s.userService.UserInfo(c, &usergrpcv1.UserInfoRequest{
			UserId: v.UserId,
		})
		if err != nil {
			return nil, err
		}

		sendRelation, err := s.relationGroupService.GetGroupRelation(c, &relationgrpcv1.GetGroupRelationRequest{
			GroupId: byId.GroupId,
			UserId:  v.UserId,
		})
		if err != nil {
			s.logger.Error("获取群聊关系失败", zap.Error(err))
			return nil, err
		}

		name := info.NickName
		if sendRelation != nil && sendRelation.Remark != "" {
			name = sendRelation.Remark
		}

		msgList = append(msgList, &model.GroupMessage{
			MsgId:                  v.Id,
			Content:                v.Content,
			GroupId:                v.GroupId,
			Type:                   v.Type,
			SendAt:                 v.CreatedAt,
			DialogId:               v.DialogId,
			IsLabel:                model.LabelMsgType(v.IsLabel),
			ReadCount:              v.ReadCount,
			ReplyId:                v.ReplyId,
			UserId:                 v.UserId,
			AtUsers:                v.AtUsers,
			ReadAt:                 int64(ReadAt),
			IsRead:                 int32(isRead),
			AtAllUser:              model.AtAllUserType(v.AtAllUser),
			IsBurnAfterReadingType: model.BurnAfterReadingType(v.IsBurnAfterReadingType),
			SenderInfo: model.SenderInfo{
				Name:   name,
				UserId: info.UserId,
				Avatar: info.Avatar,
			},
		})
	}
	resp.GroupMessages = msgList

	return resp, nil
}

func (s *Service) SetGroupMessagesRead(c context.Context, uid string, driverId string, request *model.GroupMessageReadRequest) (interface{}, error) {
	dialog, err := s.relationDialogService.GetDialogById(c, &relationgrpcv1.GetDialogByIdRequest{
		DialogId: request.DialogId,
	})
	if err != nil {
		s.logger.Error("获取对话失败", zap.Error(err))
		return nil, err
	}

	info, err := s.userService.UserInfo(c, &usergrpcv1.UserInfoRequest{
		UserId: uid,
	})
	if err != nil {
		return nil, err
	}

	if dialog.Type != uint32(model.GroupConversation) && dialog.GroupId == 0 {
		return nil, code.DialogErrTypeNotSupport
	}

	_, err = s.groupService.GetGroupInfoByGid(c, &groupApi.GetGroupInfoRequest{
		Gid: dialog.GroupId,
	})
	if err != nil {
		s.logger.Error("获取群聊信息失败", zap.Error(err))
		return nil, err
	}

	_, err = s.relationGroupService.GetGroupRelation(c, &relationgrpcv1.GetGroupRelationRequest{
		GroupId: dialog.GroupId,
		UserId:  uid,
	})
	if err != nil {
		s.logger.Error("获取群聊关系失败", zap.Error(err))
		return nil, err
	}

	_, err = s.relationDialogService.GetDialogUserByDialogIDAndUserID(c, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		UserId:   uid,
		DialogId: request.DialogId,
	})
	if err != nil {
		return nil, err
	}

	if request.ReadAll {
		resp1, err := s.msgService.ReadAllGroupMsg(c, &msggrpcv1.ReadAllGroupMsgRequest{
			DialogId: request.DialogId,
			UserId:   uid,
		})
		if err != nil {
			s.logger.Error("设置群聊消息已读失败", zap.Error(err))
			return nil, err
		}

		//给消息发送者推送谁读了消息
		for _, v := range resp1.UnreadGroupMessage {
			if v.UserId != uid {
				data := map[string]interface{}{"msg_id": v.MsgId, "operator_info": &model.SenderInfo{
					Name:   info.NickName,
					UserId: info.UserId,
					Avatar: info.Avatar,
				}}
				bytes, err := utils.StructToBytes(data)
				if err != nil {
					s.logger.Error("push err:", zap.Error(err))
					continue
				}
				msg := &pushv1.WsMsg{Uid: v.UserId, Event: pushv1.WSEventType_GroupMsgReadEvent, Data: &any.Any{Value: bytes}}
				bytes2, err := utils.StructToBytes(msg)
				if err != nil {
					s.logger.Error("push err:", zap.Error(err))
					continue
				}
				_, err = s.pushService.Push(c, &pushv1.PushRequest{
					Type: pushv1.Type_Ws,
					Data: bytes2,
				})
				if err != nil {
					s.logger.Error("push err:", zap.Error(err))
					continue
				}
			}
		}

		return nil, nil
	}

	msgList := make([]*msggrpcv1.SetGroupMessageReadRequest, 0)
	for _, v := range request.MsgIds {
		userId, _ := s.msgGroupService.GetGroupMessageReadByMsgIdAndUserId(c, &msggrpcv1.GetGroupMessageReadByMsgIdAndUserIdRequest{
			MsgId:  v,
			UserId: uid,
		})
		if userId != nil {
			continue
		}
		msgList = append(msgList, &msggrpcv1.SetGroupMessageReadRequest{
			MsgId:    v,
			GroupId:  dialog.GroupId,
			DialogId: request.DialogId,
			UserId:   uid,
			ReadAt:   pkgtime.Now(),
		})
	}
	if len(msgList) == 0 {
		return nil, nil
	}

	_, err = s.msgGroupService.SetGroupMessageRead(c, &msggrpcv1.SetGroupMessagesReadRequest{
		List: msgList,
	})
	if err != nil {
		return nil, err
	}

	msgs, err := s.msgService.GetGroupMessagesByIds(c, &msggrpcv1.GetGroupMessagesByIdsRequest{
		MsgIds:  request.MsgIds,
		GroupId: dialog.GroupId,
	})
	if err != nil {
		return nil, err
	}

	//给消息发送者推送谁读了消息
	for _, message := range msgs.GroupMessages {
		if message.UserId != uid {
			s.SendMsg(message.UserId, driverId, pushv1.WSEventType_GroupMsgReadEvent, map[string]interface{}{"msg_id": message.Id, "operator_info": &model.SenderInfo{
				Name:   info.NickName,
				UserId: info.UserId,
				Avatar: info.Avatar,
			}}, false)
		}
	}

	return nil, nil
}

func (s *Service) GetGroupMessageReadersResponse(c context.Context, userId string, msgId uint32, dialogId uint32, groupId uint32) (interface{}, error) {
	_, err := s.groupService.GetGroupInfoByGid(c, &groupApi.GetGroupInfoRequest{
		Gid: groupId,
	})
	if err != nil {
		s.logger.Error("获取群聊信息失败", zap.Error(err))
		return nil, err
	}

	_, err = s.relationGroupService.GetGroupRelation(c, &relationgrpcv1.GetGroupRelationRequest{
		GroupId: groupId,
		UserId:  userId,
	})
	if err != nil {
		s.logger.Error("获取群聊关系失败", zap.Error(err))
		return nil, err
	}

	_, err = s.relationDialogService.GetDialogUserByDialogIDAndUserID(c, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		UserId:   userId,
		DialogId: dialogId,
	})
	if err != nil {
		return nil, err
	}

	us, err := s.msgGroupService.GetGroupMessageReaders(c, &msggrpcv1.GetGroupMessageReadersRequest{
		MsgId:    msgId,
		GroupId:  groupId,
		DialogId: dialogId,
	})
	if err != nil {
		return nil, err
	}

	info, err := s.userService.GetBatchUserInfo(c, &usergrpcv1.GetBatchUserInfoRequest{
		UserIds: us.UserIds,
	})
	if err != nil {
		return nil, err
	}
	response := make([]model.GetGroupMessageReadersResponse, 0)

	if len(us.UserIds) == len(info.Users) {
		for _, v := range us.UserIds {
			for _, v1 := range info.Users {
				if v == v1.UserId {
					response = append(response, model.GetGroupMessageReadersResponse{
						UserId: v1.UserId,
						Avatar: v1.Avatar,
						Name:   v1.NickName,
					})
				}
			}
		}
	}

	return response, nil
}

//
//func (s *Service) updateCacheGroupDialog(dialogId uint32, userIds []string) error {
//	//获取最后一条消息，更新缓存
//	lastMsg, err := s.msgService.GetLastMsgsByDialogIds(context.Background(), &msggrpcv1.GetLastMsgsByDialogIdsRequest{
//		DialogIds: []uint32{dialogId},
//	})
//	if err != nil {
//		return err
//	}
//
//	if len(lastMsg.LastMsgs) == 0 {
//		return nil
//	}
//	lm := lastMsg.LastMsgs[0]
//
//	//查询发送者信息
//	info, err := s.userService.UserInfo(context.Background(), &usergrpcv1.UserInfoRequest{
//		UserID: lm.SenderId,
//	})
//	if err != nil {
//		return err
//	}
//
//	//获取对话信息
//	dialogInfo, err := s.relationDialogService.GetDialogById(context.Background(), &relationgrpcv1.GetDialogByIdRequest{
//		DialogId: dialogId,
//	})
//	if err != nil {
//		return err
//	}
//
//	if dialogInfo.Type != uint32(relationgrpcv1.DialogType_GROUP_DIALOG) {
//		return nil
//	}
//
//	ginfo, err := s.groupService.GetGroupInfoByGid(context.Background(), &groupApi.GetGroupInfoRequest{
//		Gid: dialogInfo.GroupId,
//	})
//	if err != nil {
//		return err
//	}
//
//	for _, userId := range userIds {
//		dialogUser, err := s.relationDialogService.GetDialogUserByDialogIDAndUserID(context.Background(), &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
//			DialogId: dialogId,
//			UserID:   userId,
//		})
//		if err != nil {
//			return err
//		}
//
//		msgs, err := s.msgService.GetGroupUnreadMessages(context.Background(), &msggrpcv1.GetGroupUnreadMessagesRequest{
//			UserID:   userId,
//			DialogId: dialogId,
//		})
//		if err != nil {
//			return err
//		}
//
//		dialogName := ginfo.Name
//
//		re := model.UserDialogListResponse{
//			DialogId:          dialogId,
//			GroupId:           dialogInfo.GroupId,
//			DialogType:        model.ConversationType(dialogInfo.Type),
//			DialogName:        dialogName,
//			DialogAvatar:      ginfo.Avatar,
//			DialogUnreadCount: len(msgs.GroupMessages),
//			LastMessage: model.Message{
//				MsgType:  uint(lm.Type),
//				Content:  lm.Content,
//				SenderId: lm.SenderId,
//				SendAt:   lm.CreatedAt,
//				MsgId:    uint64(lm.Id),
//				SenderInfo: model.SenderInfo{
//					UserID: info.UserID,
//					Name:   info.NickName,
//					Avatar: info.Avatar,
//				},
//				IsBurnAfterReading: model.BurnAfterReadingType(lm.IsBurnAfterReadingType),
//				IsLabel:            model.LabelMsgType(lm.IsLabel),
//				ReplyId:            lm.ReplyId,
//				AtUsers:            lm.AtUsers,
//				AtAllUser:          model.AtAllUserType(lm.AtAllUser),
//			},
//			DialogCreateAt: dialogInfo.CreateAt,
//			TopAt:          int64(dialogUser.TopAt),
//		}
//		err = s.updateRedisUserDialogList(userId, re)
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func (s *Service) updateCacheDialogFieldValue(key string, dialogId uint32, field string, value interfaces{}) error {
//	exists, err := s.redisClient.ExistsKey(key)
//	if err != nil {
//		return err
//	}
//
//	if !exists {
//		return nil
//	}
//
//	length, err := s.redisClient.GetListLength(key)
//	if err != nil {
//		return err
//	}
//
//	// 每次处理的元素数量
//	batchSize := 10
//
//	for start := 0; start < int(length); start += batchSize {
//		// 获取当前批次的元素
//		values, err := s.redisClient.GetList(key, int64(start), int64(start+batchSize-1))
//		if err != nil {
//			return err
//		}
//
//		if len(values) > 0 {
//			for i, v := range values {
//				var dialog model.UserDialogListResponse
//				err := json.Unmarshal([]byte(v), &dialog)
//				if err != nil {
//					fmt.Println("Error decoding cached data:", err)
//					continue
//				}
//				if dialog.DialogId == dialogId {
//					// 获取结构体的反射值
//					valueOfDialog := reflect.ValueOf(&dialog).Elem()
//
//					structField := valueOfDialog.FieldByName(field)
//
//					// 获取字段的反射值
//					if structField.IsValid() {
//						// 检查字段类型并设置对应类型的值
//						switch structField.Kind() {
//						case reflect.String:
//							structField.SetString(value.(string))
//						case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
//							// 调整类型断言，确保匹配实际类型
//							structField.SetInt(int64(value.(int)))
//						// 可以根据需要添加其他类型的处理
//						default:
//							fmt.Printf("Unsupported field type: %s\n", structField.Kind())
//							return nil
//						}
//					} else {
//						fmt.Printf("Field %s not found in UserDialogListResponse\n", field)
//						return nil
//					}
//
//					// Marshal updated struct and update the list element
//					marshal, err := json.Marshal(&dialog)
//					if err != nil {
//						return err
//					}
//
//					err = s.redisClient.UpdateListElement(key, int64(start+i), string(marshal))
//					if err != nil {
//						return err
//					}
//					break
//				}
//			}
//		}
//	}
//
//	return nil
//}
