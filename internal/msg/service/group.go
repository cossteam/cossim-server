package service

import (
	"context"
	"encoding/json"
	"fmt"
	groupApi "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	"github.com/cossim/coss-server/internal/msg/api/http/model"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	pkgtime "github.com/cossim/coss-server/pkg/utils/time"
	pushv1 "github.com/cossim/hipush/api/grpc/v1"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/workflow"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lithammer/shortuuid/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"reflect"
	"sync"
)

// 推送群聊消息
func (s *Service) sendWsGroupMsg(ctx context.Context, uIds []string, driverId string, msg *model.WsGroupMsg) {
	//发送群聊消息
	for _, uid := range uIds {
		m := constants.WsMsg{Uid: uid, DriverId: driverId, Event: constants.SendGroupMessageEvent, SendAt: pkgtime.Now(), Data: msg}
		//查询是否静默通知
		groupRelation, err := s.relationGroupClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
			GroupId: uint32(msg.GroupId),
			UserId:  uid,
		})
		if err != nil {
			s.logger.Error("获取群聊关系失败", zap.Error(err))
			continue
		}

		//判断是否静默通知
		if groupRelation.IsSilent == relationgrpcv1.GroupSilentNotificationType_GroupSilent {
			m.Event = constants.SendSilentGroupMessageEvent
		}

		var is bool
		r, err := s.userLoginClient.GetUserLoginByDriverIdAndUserId(ctx, &usergrpcv1.DriverIdAndUserId{
			DriverId: driverId,
			UserId:   uid,
		})
		if err != nil {
			s.logger.Error("获取用户登录信息失败", zap.Error(err))
		}

		//在线则推送ws
		if _, ok := pool[uid]; ok {
			for _, c := range pool[uid] {
				for _, c2 := range c {
					m.Rid = c2.Rid
					js, _ := json.Marshal(m)
					message, err := Enc.GetSecretMessage(string(js), uid)
					if err != nil {
						s.logger.Error("加密失败", zap.Error(err))
						return
					}
					err = c2.Conn.WriteMessage(websocket.TextMessage, []byte(message))
					if err != nil {
						s.logger.Error("发布ws消息失败", zap.Error(err))
						continue
					}
					appid, ok := s.ac.Push.PlatformAppID[r.Platform]
					if !ok {
						s.logger.Error("platform appid not found", zap.String("platform", r.Platform))
						continue
					}
					if constants.DriverType(r.Platform) != constants.MobileClient {
						s.logger.Info("platform not mobile", zap.String("platform", r.Platform))
						continue
					}
					if is {
						if _, err := s.pushClient.Push(ctx, &pushv1.PushRequest{
							Platform:    r.Platform,
							Tokens:      []string{r.DriverToken},
							Title:       msg.SenderInfo.Name,
							Message:     message,
							AppID:       appid,
							Development: true,
						}); err != nil {
							s.logger.Error("push err", zap.Error(err), zap.String("title", msg.SenderInfo.Name), zap.String("message", message), zap.String("platform", r.Platform), zap.String("driverToken", r.DriverToken), zap.String("appid", appid))
						}
					}
				}
			}
			continue
		}
		//否则推送到消息队列
		//是否传输加密

		if Enc.IsEnable() {

			marshal, err := json.Marshal(m)
			if err != nil {
				s.logger.Error("json解析失败", zap.Error(err))
				return
			}
			message, err := Enc.GetSecretMessage(string(marshal), uid)
			if err != nil {
				return
			}
			err = rabbitMQClient.PublishEncryptedMessage(uid, message)
			if err != nil {
				s.logger.Error("发布消息失败", zap.Error(err))
				return
			}
			return
		}
		err = rabbitMQClient.PublishMessage(uid, m)
		if err != nil {
			s.logger.Error("发布消息失败", zap.Error(err))
			return
		}
	}
}

func (s *Service) SendGroupMsg(ctx context.Context, userID string, driverId string, req *model.SendGroupMsgRequest) (interface{}, error) {

	if !model.IsAllowedConversationType(req.IsBurnAfterReadingType) {
		return nil, code.MsgErrInsertUserMessageFailed
	}

	groupRelation, err := s.relationGroupClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
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

	dialogs, err := s.relationDialogClient.GetDialogByIds(ctx, &relationgrpcv1.GetDialogByIdsRequest{
		DialogIds: []uint32{req.DialogId},
	})
	if err != nil {
		s.logger.Error("获取会话失败", zap.Error(err))
		return nil, err
	}
	if len(dialogs.Dialogs) == 0 {
		return nil, code.DialogErrGetDialogUserByDialogIDAndUserIDFailed
	}

	_, err = s.relationDialogClient.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		DialogId: req.DialogId,
		UserId:   userID,
	})
	if err != nil {
		s.logger.Error("获取用户会话失败", zap.Error(err))
		return nil, code.DialogErrGetDialogUserByDialogIDAndUserIDFailed
	}

	//查询群聊所有用户id
	uids, err := s.relationGroupClient.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{
		GroupId: req.GroupId,
	})

	var message *msggrpcv1.SendGroupMsgResponse

	workflow.InitGrpc(s.dtmGrpcServer, s.dialogGrpcServer, grpc.NewServer())
	gid := shortuuid.New()
	wfName := "send_group_msg_workflow_" + gid
	if err := workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {

		_, err := s.relationDialogClient.BatchCloseOrOpenDialog(ctx, &relationgrpcv1.BatchCloseOrOpenDialogRequest{
			DialogId: req.DialogId,
			Action:   relationgrpcv1.CloseOrOpenDialogType_OPEN,
			UserIds:  uids.UserIds,
		})
		if err != nil {
			return err
		}
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := s.relationDialogClient.BatchCloseOrOpenDialog(ctx, &relationgrpcv1.BatchCloseOrOpenDialogRequest{
				DialogId: req.DialogId,
				Action:   relationgrpcv1.CloseOrOpenDialogType_CLOSE,
				UserIds:  uids.UserIds,
			})
			return err
		})

		message, err = s.msgClient.SendGroupMessage(ctx, &msggrpcv1.SendGroupMsgRequest{
			DialogId:               req.DialogId,
			GroupId:                req.GroupId,
			UserId:                 userID,
			Content:                req.Content,
			Type:                   uint32(req.Type),
			ReplayId:               req.ReplayId,
			AtUsers:                req.AtUsers,
			AtAllUser:              msggrpcv1.AtAllUserType(req.AtAllUser),
			IsBurnAfterReadingType: msggrpcv1.BurnAfterReadingType(req.IsBurnAfterReadingType),
		})
		// 发送成功进行消息推送
		if err != nil {
			s.logger.Error("发送消息失败", zap.Error(err))
			return err
		}
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
		_, err = s.msgClient.SetGroupMessageRead(context.Background(), &msggrpcv1.SetGroupMessagesReadRequest{
			List: list,
		})
		if err != nil {
			return err
		}
		return err
	}); err != nil {
		return "", err
	}
	if err := workflow.Execute(wfName, gid, nil); err != nil {
		return "", err
	}

	//查询发送者信息
	info, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}
	s.sendWsGroupMsg(ctx, uids.UserIds, driverId, &model.WsGroupMsg{
		MsgId:              message.MsgId,
		GroupId:            int64(req.GroupId),
		SenderId:           userID,
		Content:            req.Content,
		MsgType:            uint(req.Type),
		ReplayId:           uint(req.ReplayId),
		SendAt:             pkgtime.Now(),
		DialogId:           req.DialogId,
		AtUsers:            req.AtUsers,
		AtAllUser:          req.AtAllUser,
		IsBurnAfterReading: req.IsBurnAfterReadingType,
		SenderInfo: model.SenderInfo{
			Avatar: info.Avatar,
			Name:   info.NickName,
			UserId: userID,
		},
	})

	if s.cache {
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.updateCacheGroupDialog(req.DialogId, uids.UserIds)
			if err != nil {
				s.logger.Error("更新缓存群聊会话失败", zap.Error(err))
				return
			}
		}()
		wg.Wait()
	}

	return message.MsgId, nil
}

func (s *Service) EditGroupMsg(ctx context.Context, userID string, driverId string, msgID uint32, content string) (interface{}, error) {
	//获取消息
	msginfo, err := s.msgClient.GetGroupMessageById(ctx, &msggrpcv1.GetGroupMsgByIDRequest{
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
	userIds, err := s.relationDialogClient.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
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
	_, err = s.msgClient.EditGroupMessage(ctx, &msggrpcv1.EditGroupMsgRequest{
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
	s.SendMsgToUsers(userIds.UserIds, driverId, constants.EditMsgEvent, msginfo, true)

	if s.cache {
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			err := s.updateCacheGroupDialog(msginfo.DialogId, userIds.UserIds)
			if err != nil {
				s.logger.Error("更新缓存群聊会话失败", zap.Error(err))
				return
			}
		}()
		wg.Wait()
	}

	return msgID, nil
}

func (s *Service) RecallGroupMsg(ctx context.Context, userID string, driverId string, msgID uint32) (interface{}, error) {
	//获取消息
	msginfo, err := s.msgClient.GetGroupMessageById(ctx, &msggrpcv1.GetGroupMsgByIDRequest{
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
	userIds, err := s.relationDialogClient.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
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
	// 调用相应的 gRPC 客户端方法来撤回群消息
	msg, err := s.msgClient.DeleteGroupMessage(ctx, &msggrpcv1.DeleteGroupMsgRequest{
		MsgId: msgID,
	})
	if err != nil {
		s.logger.Error("撤回群消息失败", zap.Error(err))
		return nil, err
	}

	//s.SendMsgToUsers(userIds.UserIds, driverId, constants.RecallMsgEvent, msginfo, true)
	//
	//if s.cache {
	//	wg := sync.WaitGroup{}
	//	wg.Add(1)
	//	go func() {
	//		defer wg.Done()
	//		err := s.updateCacheGroupDialog(msginfo.DialogId, userIds.UserIds)
	//		if err != nil {
	//			s.logger.Error("更新缓存群聊会话失败", zap.Error(err))
	//			return
	//		}
	//	}()
	//	wg.Wait()
	//}

	msg2 := &model.SendGroupMsgRequest{
		DialogId: msginfo.DialogId,
		GroupId:  msginfo.GroupId,
		Content:  msginfo.Content,
		ReplayId: msginfo.Id,
		Type:     model.MessageTypeDelete,
	}
	_, err = s.SendGroupMsg(ctx, userID, driverId, msg2)
	if err != nil {
		return nil, err
	}

	return msg.Id, nil
}

func (s *Service) LabelGroupMessage(ctx context.Context, userID string, driverId string, msgID uint32, label model.LabelMsgType) (interface{}, error) {
	// 获取群聊消息
	msginfo, err := s.msgClient.GetGroupMessageById(ctx, &msggrpcv1.GetGroupMsgByIDRequest{
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
	userIds, err := s.relationDialogClient.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
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
	_, err = s.msgClient.SetGroupMsgLabel(ctx, &msggrpcv1.SetGroupMsgLabelRequest{
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
		ReplayId: msginfo.Id,
		Type:     model.MessageTypeLabel,
	}

	_, err = s.SendGroupMsg(ctx, userID, driverId, msg2)
	if err != nil {
		return nil, err
	}
	//s.SendMsgToUsers(userIds.UserIds, driverId, constants.LabelMsgEvent, msginfo, true)
	return nil, nil
}

func (s *Service) GetGroupLabelMsgList(ctx context.Context, userID string, dialogId uint32) (interface{}, error) {
	_, err := s.relationDialogClient.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		UserId:   userID,
		DialogId: dialogId,
	})
	if err != nil {
		s.logger.Error("获取对话用户失败", zap.Error(err))
		return nil, err
	}

	msgs, err := s.msgClient.GetGroupMsgLabelByDialogId(ctx, &msggrpcv1.GetGroupMsgLabelByDialogIdRequest{
		DialogId: dialogId,
	})
	if err != nil {
		s.logger.Error("获取群聊消息标注失败", zap.Error(err))
		return nil, err
	}

	return msgs, nil
}

func (s *Service) GetGroupMessageList(c *gin.Context, id string, request *model.GroupMsgListRequest) (interface{}, error) {
	_, err := s.groupClient.GetGroupInfoByGid(c, &groupApi.GetGroupInfoRequest{
		Gid: request.GroupId,
	})
	if err != nil {
		s.logger.Error("获取群聊信息失败", zap.Error(err))
		return nil, err
	}

	_, err = s.relationGroupClient.GetGroupRelation(c, &relationgrpcv1.GetGroupRelationRequest{
		GroupId: request.GroupId,
		UserId:  id,
	})
	if err != nil {
		s.logger.Error("获取群聊关系失败", zap.Error(err))
		return nil, err
	}

	msg, err := s.msgClient.GetGroupMessageList(c, &msggrpcv1.GetGroupMsgListRequest{
		GroupId:  int32(request.GroupId),
		UserId:   request.UserId,
		Content:  request.Content,
		Type:     request.Type,
		PageNum:  int32(request.PageNum),
		PageSize: int32(request.PageSize),
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
		msgRead, err := s.msgClient.GetGroupMessageReadByMsgIdAndUserId(c, &msggrpcv1.GetGroupMessageReadByMsgIdAndUserIdRequest{
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
		info, err := s.userClient.UserInfo(c, &usergrpcv1.UserInfoRequest{
			UserId: v.UserId,
		})
		if err != nil {
			return nil, err
		}

		sendRelation, err := s.relationGroupClient.GetGroupRelation(c, &relationgrpcv1.GetGroupRelationRequest{
			GroupId: request.GroupId,
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
			CreatedAt:              v.CreatedAt,
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

func (s *Service) SetGroupMessagesRead(c context.Context, id string, driverId string, request *model.GroupMessageReadRequest) (interface{}, error) {
	_, err := s.groupClient.GetGroupInfoByGid(c, &groupApi.GetGroupInfoRequest{
		Gid: request.GroupId,
	})
	if err != nil {
		s.logger.Error("获取群聊信息失败", zap.Error(err))
		return nil, err
	}

	_, err = s.relationGroupClient.GetGroupRelation(c, &relationgrpcv1.GetGroupRelationRequest{
		GroupId: request.GroupId,
		UserId:  id,
	})
	if err != nil {
		s.logger.Error("获取群聊关系失败", zap.Error(err))
		return nil, err
	}

	_, err = s.relationDialogClient.GetDialogUserByDialogIDAndUserID(c, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		UserId:   id,
		DialogId: request.DialogId,
	})
	if err != nil {
		return nil, err
	}

	msgList := make([]*msggrpcv1.SetGroupMessageReadRequest, 0)
	for _, v := range request.MsgIds {
		userId, _ := s.msgClient.GetGroupMessageReadByMsgIdAndUserId(c, &msggrpcv1.GetGroupMessageReadByMsgIdAndUserIdRequest{
			MsgId:  v,
			UserId: id,
		})
		if userId != nil {
			continue
		}
		msgList = append(msgList, &msggrpcv1.SetGroupMessageReadRequest{
			MsgId:    v,
			GroupId:  request.GroupId,
			DialogId: request.DialogId,
			UserId:   id,
			ReadAt:   pkgtime.Now(),
		})
	}
	if len(msgList) == 0 {
		return nil, nil
	}

	_, err = s.msgClient.SetGroupMessageRead(c, &msggrpcv1.SetGroupMessagesReadRequest{
		List: msgList,
	})
	if err != nil {
		return nil, err
	}

	msgs, err := s.msgClient.GetGroupMessagesByIds(c, &msggrpcv1.GetGroupMessagesByIdsRequest{
		MsgIds:  request.MsgIds,
		GroupId: request.GroupId,
	})
	if err != nil {
		return nil, err
	}

	//给消息发送者推送谁读了消息
	for _, message := range msgs.GroupMessages {
		if message.UserId != id {
			s.SendMsg(message.UserId, driverId, constants.GroupMsgReadEvent, map[string]interface{}{"msg_id": message.Id, "read_user_id": id}, false)
		}
	}

	if s.cache {
		userMsgs, err := s.msgClient.GetGroupUnreadMessages(c, &msggrpcv1.GetGroupUnreadMessagesRequest{
			UserId:   id,
			DialogId: request.DialogId,
		})
		if err != nil {
			return nil, err
		}

		err = s.updateCacheDialogFieldValue(fmt.Sprintf("dialog:%s", id), request.DialogId, "DialogUnreadCount", len(userMsgs.GroupMessages))
		if err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func (s *Service) GetGroupMessageReadersResponse(c context.Context, userId string, msgId uint32, dialogId uint32, groupId uint32) (interface{}, error) {
	_, err := s.groupClient.GetGroupInfoByGid(c, &groupApi.GetGroupInfoRequest{
		Gid: groupId,
	})
	if err != nil {
		s.logger.Error("获取群聊信息失败", zap.Error(err))
		return nil, err
	}

	_, err = s.relationGroupClient.GetGroupRelation(c, &relationgrpcv1.GetGroupRelationRequest{
		GroupId: groupId,
		UserId:  userId,
	})
	if err != nil {
		s.logger.Error("获取群聊关系失败", zap.Error(err))
		return nil, err
	}

	_, err = s.relationDialogClient.GetDialogUserByDialogIDAndUserID(c, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		UserId:   userId,
		DialogId: dialogId,
	})
	if err != nil {
		return nil, err
	}

	us, err := s.msgClient.GetGroupMessageReaders(c, &msggrpcv1.GetGroupMessageReadersRequest{
		MsgId:    msgId,
		GroupId:  groupId,
		DialogId: dialogId,
	})
	if err != nil {
		return nil, err
	}

	info, err := s.userClient.GetBatchUserInfo(c, &usergrpcv1.GetBatchUserInfoRequest{
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

func (s *Service) updateCacheGroupDialog(dialogId uint32, userIds []string) error {
	//获取最后一条消息，更新缓存
	lastMsg, err := s.msgClient.GetLastMsgsByDialogIds(context.Background(), &msggrpcv1.GetLastMsgsByDialogIdsRequest{
		DialogIds: []uint32{dialogId},
	})
	if err != nil {
		return err
	}

	if len(lastMsg.LastMsgs) == 0 {
		return nil
	}
	lm := lastMsg.LastMsgs[0]

	//查询发送者信息
	info, err := s.userClient.UserInfo(context.Background(), &usergrpcv1.UserInfoRequest{
		UserId: lm.SenderId,
	})
	if err != nil {
		return err
	}

	//获取对话信息
	dialogInfo, err := s.relationDialogClient.GetDialogById(context.Background(), &relationgrpcv1.GetDialogByIdRequest{
		DialogId: dialogId,
	})
	if err != nil {
		return err
	}

	if dialogInfo.Type != uint32(relationgrpcv1.DialogType_GROUP_DIALOG) {
		return nil
	}

	ginfo, err := s.groupClient.GetGroupInfoByGid(context.Background(), &groupApi.GetGroupInfoRequest{
		Gid: dialogInfo.GroupId,
	})
	if err != nil {
		return err
	}

	for _, userId := range userIds {
		dialogUser, err := s.relationDialogClient.GetDialogUserByDialogIDAndUserID(context.Background(), &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
			DialogId: dialogId,
			UserId:   userId,
		})
		if err != nil {
			return err
		}

		msgs, err := s.msgClient.GetGroupUnreadMessages(context.Background(), &msggrpcv1.GetGroupUnreadMessagesRequest{
			UserId:   userId,
			DialogId: dialogId,
		})
		if err != nil {
			return err
		}

		dialogName := ginfo.Name

		re := model.UserDialogListResponse{
			DialogId:          dialogId,
			GroupId:           dialogInfo.GroupId,
			DialogType:        model.ConversationType(dialogInfo.Type),
			DialogName:        dialogName,
			DialogAvatar:      ginfo.Avatar,
			DialogUnreadCount: len(msgs.GroupMessages),
			LastMessage: model.Message{
				MsgType:  uint(lm.Type),
				Content:  lm.Content,
				SenderId: lm.SenderId,
				SendTime: lm.CreatedAt,
				MsgId:    uint64(lm.Id),
				SenderInfo: model.SenderInfo{
					UserId: info.UserId,
					Name:   info.NickName,
					Avatar: info.Avatar,
				},
				IsBurnAfterReading: model.BurnAfterReadingType(lm.IsBurnAfterReadingType),
				IsLabel:            model.LabelMsgType(lm.IsLabel),
				ReplayId:           lm.ReplyId,
				AtUsers:            lm.AtUsers,
				AtAllUser:          model.AtAllUserType(lm.AtAllUser),
			},
			DialogCreateAt: dialogInfo.CreateAt,
			TopAt:          int64(dialogUser.TopAt),
		}
		err = s.updateRedisUserDialogList(userId, re)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *Service) updateCacheDialogFieldValue(key string, dialogId uint32, field string, value interface{}) error {
	exists, err := s.redisClient.ExistsKey(key)
	if err != nil {
		return err
	}

	if !exists {
		return nil
	}

	length, err := s.redisClient.GetListLength(key)
	if err != nil {
		return err
	}

	// 每次处理的元素数量
	batchSize := 10

	for start := 0; start < int(length); start += batchSize {
		// 获取当前批次的元素
		values, err := s.redisClient.GetList(key, int64(start), int64(start+batchSize-1))
		if err != nil {
			return err
		}

		if len(values) > 0 {
			for i, v := range values {
				var dialog model.UserDialogListResponse
				err := json.Unmarshal([]byte(v), &dialog)
				if err != nil {
					fmt.Println("Error decoding cached data:", err)
					continue
				}
				if dialog.DialogId == dialogId {
					// 获取结构体的反射值
					valueOfDialog := reflect.ValueOf(&dialog).Elem()

					structField := valueOfDialog.FieldByName(field)

					// 获取字段的反射值
					if structField.IsValid() {
						// 检查字段类型并设置对应类型的值
						switch structField.Kind() {
						case reflect.String:
							structField.SetString(value.(string))
						case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
							// 调整类型断言，确保匹配实际类型
							structField.SetInt(int64(value.(int)))
						// 可以根据需要添加其他类型的处理
						default:
							fmt.Printf("Unsupported field type: %s\n", structField.Kind())
							return nil
						}
					} else {
						fmt.Printf("Field %s not found in UserDialogListResponse\n", field)
						return nil
					}

					// Marshal updated struct and update the list element
					marshal, err := json.Marshal(&dialog)
					if err != nil {
						return err
					}

					err = s.redisClient.UpdateListElement(key, int64(start+i), string(marshal))
					if err != nil {
						return err
					}
					break
				}
			}
		}
	}

	return nil
}
