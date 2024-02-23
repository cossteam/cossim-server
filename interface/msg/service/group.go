package service

import (
	"context"
	"encoding/json"
	"github.com/cossim/coss-server/interface/msg/api/model"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	pkgtime "github.com/cossim/coss-server/pkg/utils/time"
	groupApi "github.com/cossim/coss-server/service/group/api/v1"
	msggrpcv1 "github.com/cossim/coss-server/service/msg/api/v1"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	usergrpcv1 "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
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

	message, err := s.msgClient.SendGroupMessage(ctx, &msggrpcv1.SendGroupMsgRequest{
		DialogId:               req.DialogId,
		GroupId:                req.GroupId,
		UserId:                 userID,
		Content:                req.Content,
		Type:                   req.Type,
		ReplayId:               req.ReplayId,
		AtUsers:                req.AtUsers,
		AtAllUser:              msggrpcv1.AtAllUserType(req.AtAllUser),
		IsBurnAfterReadingType: msggrpcv1.BurnAfterReadingType(req.IsBurnAfterReadingType),
	})
	// 发送成功进行消息推送
	if err != nil {
		s.logger.Error("发送消息失败", zap.Error(err))
		return nil, err
	}
	//查询群聊所有用户id
	uids, err := s.relationGroupClient.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{
		GroupId: req.GroupId,
	})

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

	//TODO 开携程去推送
	s.SendMsgToUsers(userIds.UserIds, driverId, constants.EditMsgEvent, msginfo, true)

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

	if msginfo.UserId != userID {
		return nil, code.Unauthorized
	}
	//判断发送时间是否超过两分钟
	if pkgtime.Now()-msginfo.CreatedAt > 120 {
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

	s.SendMsgToUsers(userIds.UserIds, driverId, constants.RecallMsgEvent, msginfo, true)

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

	s.SendMsgToUsers(userIds.UserIds, driverId, constants.LabelMsgEvent, msginfo, true)
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

	resp, err := s.msgClient.GetGroupMessageList(c, &msggrpcv1.GetGroupMsgListRequest{
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
		userId, _ := s.groupMsgClient.GetGroupMessageReadByMsgIdAndUserId(c, &msggrpcv1.GetGroupMessageReadByMsgIdAndUserIdRequest{
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

	_, err = s.groupMsgClient.SetGroupMessageRead(c, &msggrpcv1.SetGroupMessagesReadRequest{
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

	us, err := s.groupMsgClient.GetGroupMessageReaders(c, &msggrpcv1.GetGroupMessageReadersRequest{
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
