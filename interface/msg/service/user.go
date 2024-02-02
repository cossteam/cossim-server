package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cossim/coss-server/interface/msg/api/model"
	"github.com/cossim/coss-server/interface/msg/config"
	"github.com/cossim/coss-server/pkg/cache"
	"github.com/cossim/coss-server/pkg/code"
	pkgtime "github.com/cossim/coss-server/pkg/utils/time"
	groupApi "github.com/cossim/coss-server/service/group/api/v1"
	msggrpcv1 "github.com/cossim/coss-server/service/msg/api/v1"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	usergrpcv1 "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"log"
	"sort"
)

func (s *Service) SendUserMsg(ctx context.Context, userID string, req *model.SendUserMsgRequest) (interface{}, error) {

	if !model.IsAllowedConversationType(req.IsBurnAfterReadingType) {
		return nil, code.MsgErrInsertUserMessageFailed
	}
	userRelationStatus1, err := s.relationUserClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
		UserId:   userID,
		FriendId: req.ReceiverId,
	})
	if err != nil {
		s.logger.Error("获取用户关系失败", zap.Error(err))
		return nil, err
	}

	if userRelationStatus1.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
		return nil, code.RelationUserErrFriendRelationNotFound
	}

	userRelationStatus2, err := s.relationUserClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
		UserId:   req.ReceiverId,
		FriendId: userID,
	})
	if err != nil {
		s.logger.Error("获取用户关系失败", zap.Error(err))
		return nil, err
	}

	if userRelationStatus2.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
		return nil, code.RelationUserErrFriendRelationNotFound
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
		return nil, err
	}

	message, err := s.msgClient.SendUserMessage(ctx, &msggrpcv1.SendUserMsgRequest{
		DialogId:               req.DialogId,
		SenderId:               userID,
		ReceiverId:             req.ReceiverId,
		Content:                req.Content,
		Type:                   int32(req.Type),
		ReplayId:               uint64(req.ReplayId),
		IsBurnAfterReadingType: msggrpcv1.BurnAfterReadingType(req.IsBurnAfterReadingType),
	})
	if err != nil {
		s.logger.Error("发送消息失败", zap.Error(err))
		return nil, code.MsgErrInsertUserMessageFailed
	}

	s.sendWsUserMsg(userID, req.ReceiverId, req.Content, req.Type, req.ReplayId, req.DialogId, userRelationStatus2.IsSilent, message.MsgId, req.IsBurnAfterReadingType)

	return message, nil
}

// 推送私聊消息
func (s *Service) sendWsUserMsg(senderId, receiverId string, msg string, msgType uint, replayId uint, dialogId uint32, silent relationgrpcv1.UserSilentNotificationType, msgid uint32, isBnr model.BurnAfterReadingType) {
	m := config.WsMsg{Uid: receiverId, Event: config.SendUserMessageEvent, SendAt: pkgtime.Now(),
		Data: &model.WsUserMsg{
			SenderId:           senderId,
			Content:            msg,
			MsgType:            msgType,
			ReplayId:           replayId,
			MsgId:              msgid,
			SendAt:             pkgtime.Now(),
			DialogId:           dialogId,
			IsBurnAfterReading: isBnr,
		},
	}

	//userRelation, err := s.relationClient.GetUserRelation(context.Background(), &relationgrpcv1.GetUserRelationRequest{UserId: receiverId, FriendId: senderId})
	//if err != nil {
	//	log.Error("获取用户关系失败", zap.Error(err))
	//	return
	//}
	//
	//if userRelation.Status != relationgrpcv1.RelationStatus_RELATION_STATUS_ADDED {
	//	log.Error("不是好友")
	//	return
	//}

	//是否静默通知
	if silent == relationgrpcv1.UserSilentNotificationType_UserSilent {
		m.Event = config.SendSilentUserMessageEvent
	}

	//遍历该用户所有客户端
	if _, ok := pool[receiverId]; ok {
		if len(pool[receiverId]) > 0 {
			for _, c := range pool[receiverId] {
				for _, c2 := range c {
					m.Rid = c2.Rid
					js, _ := json.Marshal(m)
					message, err := Enc.GetSecretMessage(string(js), receiverId)
					if err != nil {
						return
					}
					err = c2.Conn.WriteMessage(websocket.TextMessage, []byte(message))
					if err != nil {
						s.logger.Error("send msggrpcv1 err", zap.Error(err))
						return
					}
				}
			}
			return
		}
	}
	if Enc.IsEnable() {
		marshal, err := json.Marshal(m)
		if err != nil {
			s.logger.Error("json解析失败", zap.Error(err))
			return
		}
		message, err := Enc.GetSecretMessage(string(marshal), receiverId)
		if err != nil {
			return
		}
		err = rabbitMQClient.PublishEncryptedMessage(receiverId, message)
		if err != nil {
			s.logger.Error("发布消息失败", zap.Error(err))
			return
		}
		return
	}
	err := rabbitMQClient.PublishMessage(receiverId, m)
	if err != nil {
		s.logger.Error("发布消息失败", zap.Error(err))
		return
	}
}

func (s *Service) GetUserMessageList(ctx context.Context, userID string, req *model.MsgListRequest) (interface{}, error) {
	msg, err := s.msgClient.GetUserMessageList(ctx, &msggrpcv1.GetUserMsgListRequest{
		UserId:   userID,     //当前用户
		FriendId: req.UserId, //好友id
		Content:  req.Content,
		Type:     req.Type,
		PageNum:  int32(req.PageNum),
		PageSize: int32(req.PageSize),
	})
	if err != nil {
		s.logger.Error("获取用户消息列表失败", zap.Error(err))
		return nil, err
	}

	return msg, nil
}

func (s *Service) GetUserDialogList(ctx context.Context, userID string) (interface{}, error) {
	//获取对话id
	ids, err := s.relationDialogClient.GetUserDialogList(ctx, &relationgrpcv1.GetUserDialogListRequest{
		UserId: userID,
	})
	if err != nil {
		s.logger.Error("获取用户会话id", zap.Error(err))
		return nil, err
	}

	//获取对话信息
	infos, err := s.relationDialogClient.GetDialogByIds(ctx, &relationgrpcv1.GetDialogByIdsRequest{
		DialogIds: ids.DialogIds,
	})
	if err != nil {
		s.logger.Error("获取用户会话信息", zap.Error(err))
		return nil, err
	}

	//获取最后一条消息
	dialogIds, err := s.msgClient.GetLastMsgsByDialogIds(ctx, &msggrpcv1.GetLastMsgsByDialogIdsRequest{
		DialogIds: ids.DialogIds,
	})
	if err != nil {
		s.logger.Error("获取消息失败", zap.Error(err))
		return nil, err
	}

	//封装响应数据
	var responseList = make([]model.UserDialogListResponse, 0)
	for _, v := range infos.Dialogs {
		var re model.UserDialogListResponse
		du, err := s.relationDialogClient.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
			DialogId: v.Id,
			UserId:   userID,
		})
		if err != nil {
			s.logger.Error("获取对话用户信息失败", zap.Error(err))
			return nil, err
		}
		re.TopAt = int64(du.TopAt)
		//用户
		if v.Type == 0 {
			users, _ := s.relationDialogClient.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
				DialogId: v.Id,
			})
			if len(users.UserIds) == 0 {
				continue
			}
			for _, id := range users.UserIds {
				if id == userID {
					continue
				}
				info, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
					UserId: id,
				})
				if err != nil {
					fmt.Println(err)
					continue
				}
				re.DialogId = v.Id
				re.DialogAvatar = info.Avatar
				re.DialogName = info.NickName
				re.DialogType = 0
				re.DialogUnreadCount = 10
				re.UserId = info.UserId
				re.DialogCreateAt = v.CreateAt
				break
			}

		} else if v.Type == 1 {
			//群聊
			info, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{
				Gid: v.GroupId,
			})
			if err != nil {
				fmt.Println(err)
				continue
			}
			re.DialogAvatar = info.Avatar
			re.DialogName = info.Name
			re.DialogType = 1
			re.DialogUnreadCount = 10
			//re.UserId = v.OwnerId
			re.GroupId = info.Id
			re.DialogId = v.Id
			re.DialogCreateAt = v.CreateAt
		}
		// 匹配最后一条消息
		for _, msg := range dialogIds.LastMsgs {
			if msg.DialogId == v.Id {
				re.LastMessage = model.Message{
					MsgId:    uint64(msg.Id),
					Content:  msg.Content,
					SenderId: msg.SenderId,
					SendTime: msg.CreatedAt,
					MsgType:  uint(msg.Type),
				}
				break
			}
		}

		responseList = append(responseList, re)
	}
	//根据发送时间排序
	sort.Slice(responseList, func(i, j int) bool {
		return responseList[i].LastMessage.SendTime > responseList[j].LastMessage.SendTime
	})

	return responseList, nil
}

func (s *Service) RecallUserMsg(ctx context.Context, userID string, msgID uint32) (interface{}, error) {
	//获取消息
	msginfo, err := s.msgClient.GetUserMessageById(ctx, &msggrpcv1.GetUserMsgByIDRequest{
		MsgId: msgID,
	})
	if err != nil {
		s.logger.Error("获取消息失败", zap.Error(err))
		return nil, err
	}

	if msginfo.SenderId != userID {
		return nil, code.Unauthorized
	}
	//判断发送时间是否超过两分钟
	if pkgtime.Now()-msginfo.CreatedAt > 120 {
		return nil, code.MsgErrTimeoutExceededCannotRevoke
	}

	// 调用相应的 gRPC 客户端方法来撤回用户消息
	msg, err := s.msgClient.DeleteUserMessage(ctx, &msggrpcv1.DeleteUserMsgRequest{
		MsgId: msgID,
	})
	if err != nil {
		s.logger.Error("撤回消息失败", zap.Error(err))
		return nil, err
	}

	return msg.Id, nil
}

func (s *Service) EditUserMsg(c *gin.Context, userID string, msgID uint32, content string) (interface{}, error) {
	//获取消息
	msginfo, err := s.msgClient.GetUserMessageById(context.Background(), &msggrpcv1.GetUserMsgByIDRequest{
		MsgId: msgID,
	})
	if err != nil {
		s.logger.Error("获取消息失败", zap.Error(err))
		return nil, err
	}

	if msginfo.SenderId != userID {
		return nil, code.Unauthorized
	}

	// 调用相应的 gRPC 客户端方法来编辑用户消息
	_, err = s.msgClient.EditUserMessage(context.Background(), &msggrpcv1.EditUserMsgRequest{
		UserMessage: &msggrpcv1.UserMessage{
			Id:      msgID,
			Content: content,
		},
	})
	if err != nil {
		s.logger.Error("编辑用户消息失败", zap.Error(err))
		return nil, err
	}

	return msgID, nil
}

func (s *Service) ReadUserMsgs(ctx context.Context, userid string, dialogId uint32, msgids []uint32) (interface{}, error) {
	ids, err := s.relationDialogClient.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
		DialogId: dialogId,
	})
	if err != nil {
		s.logger.Error("批量设置私聊消息状态为已读", zap.Error(err))
		return nil, err
	}

	found := false
	for _, v := range ids.UserIds {
		if v == userid {
			found = true
			break
		}
	}
	if !found {
		return nil, code.NotFound
	}

	_, err = s.msgClient.SetUserMsgsReadStatus(ctx, &msggrpcv1.SetUserMsgsReadStatusRequest{
		MsgIds:   msgids,
		DialogId: dialogId,
	})
	if err != nil {
		s.logger.Error("批量设置私聊消息状态为已读", zap.Error(err))
		return nil, err
	}

	return nil, nil
}

func (s *Service) LabelUserMessage(ctx context.Context, userID string, msgID uint32, label model.LabelMsgType) (interface{}, error) {
	// 获取用户消息
	msginfo, err := s.msgClient.GetUserMessageById(ctx, &msggrpcv1.GetUserMsgByIDRequest{
		MsgId: msgID,
	})
	if err != nil {
		s.logger.Error("获取用户消息失败", zap.Error(err))
		return nil, err
	}
	//判断是否在对话内
	userIds, err := s.relationDialogClient.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
		DialogId: msginfo.DialogId,
	})
	if err != nil {
		s.logger.Error("获取用户消息失败", zap.Error(err))
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
		return nil, code.RelationUserErrFriendRelationNotFound
	}

	// 调用 gRPC 客户端方法将用户消息设置为标注状态
	_, err = s.msgClient.SetUserMsgLabel(context.Background(), &msggrpcv1.SetUserMsgLabelRequest{
		MsgId:   msgID,
		IsLabel: msggrpcv1.MsgLabel(label),
	})
	if err != nil {
		s.logger.Error("设置用户消息标注失败", zap.Error(err))
		return nil, err
	}

	return nil, nil
}

func (s *Service) GetUserLabelMsgList(ctx context.Context, userID string, dialogID uint32) (interface{}, error) {
	_, err := s.relationDialogClient.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		UserId:   userID,
		DialogId: dialogID,
	})
	if err != nil {
		s.logger.Error("获取用户对话失败", zap.Error(err))
		return nil, err
	}

	msgs, err := s.msgClient.GetUserMsgLabelByDialogId(ctx, &msggrpcv1.GetUserMsgLabelByDialogIdRequest{
		DialogId: dialogID,
	})
	if err != nil {
		s.logger.Error("获取用户标注消息失败", zap.Error(err))
		return nil, err

	}

	return msgs, nil
}

func (s *Service) Ws(conn *websocket.Conn, uid string, deviceType, token string) {
	defer conn.Close()
	//用户上线
	wsRid++
	messages := rabbitMQClient.GetChannel()
	if messages.IsClosed() {
		log.Fatal("Channel is Closed")
	}
	cli := &client{
		ClientType: deviceType,
		Conn:       conn,
		Uid:        uid,
		Rid:        wsRid,
		queue:      messages,
	}
	if _, ok := pool[uid]; !ok {
		pool[uid] = make(map[string][]*client)
	}

	if s.conf.MultipleDeviceLimit.Enable {
		if _, ok := pool[uid][deviceType]; ok {
			if len(pool[uid][deviceType]) == s.conf.MultipleDeviceLimit.Max {
				return
			}
		}
	}
	//保存到线程池
	cli.wsOnlineClients()

	//todo 加锁
	//更新登录信息
	values, err := cache.GetAllListValues(s.redisClient, cli.Uid)
	if err != nil {
		s.logger.Error("获取用户信息失败1", zap.Error(err))
		return
	}
	list, err := cache.GetUserInfoList(values)

	if err != nil {
		s.logger.Error("获取用户信息失败2", zap.Error(err))
		return
	}
	for _, info := range list {
		if info.Token == token {
			info.Rid = cli.Rid
			nlist := cache.GetUserInfoListToInterfaces(list)
			err := cache.DeleteList(s.redisClient, cli.Uid)
			if err != nil {
				s.logger.Error("获取用户信息失败", zap.Error(err))
				return
			}
			err = cache.AddToList(s.redisClient, cli.Uid, nlist)
			if err != nil {
				s.logger.Error("保存用户信息失败", zap.Error(err))
				return
			}
			break
		}
	}
	//读取客户端消息
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			s.logger.Error("读取消息失败", zap.Error(err))
			//删除redis里的rid
			values, err := cache.GetAllListValues(s.redisClient, cli.Uid)
			if err != nil {
				s.logger.Error("获取用户信息失败1", zap.Error(err))
				return
			}
			list, err := cache.GetUserInfoList(values)

			if err != nil {
				s.logger.Error("获取用户信息失败2", zap.Error(err))
				return
			}
			for _, info := range list {
				if info.Token == token {
					info.Rid = 0
					nlist := cache.GetUserInfoListToInterfaces(list)
					err := cache.DeleteList(s.redisClient, cli.Uid)
					if err != nil {
						s.logger.Error("获取用户信息失败", zap.Error(err))
						return
					}
					err = cache.AddToList(s.redisClient, cli.Uid, nlist)
					if err != nil {
						s.logger.Error("保存用户信息失败", zap.Error(err))
						return
					}
					break
				}
			}
			//用户下线
			cli.wsOfflineClients()
			return
		}
	}
}

// SendMsg 推送消息
func (s *Service) SendMsg(uid string, event config.WSEventType, data interface{}, pushOffline bool) {
	m := config.WsMsg{Uid: uid, Event: event, Rid: 0, Data: data, SendAt: pkgtime.Now()}
	_, ok := pool[uid]
	if !pushOffline && !ok {
		return
	}
	if pushOffline && !ok {
		//不在线则推送到消息队列
		err := rabbitMQClient.PublishMessage(uid, m)
		if err != nil {
			fmt.Println("发布消息失败：", err)
			return
		}
		return
	}
	for _, v := range pool[uid] {
		for _, c := range v {
			m.Rid = c.Rid
			js, _ := json.Marshal(m)
			err := c.Conn.WriteMessage(websocket.TextMessage, js)
			if err != nil {
				s.logger.Error("send msg err", zap.Error(err))
				return
			}
		}
	}
}

// SendMsgToUsers 推送多个用户消息
func (s *Service) SendMsgToUsers(uids []string, event config.WSEventType, data interface{}) {
	for _, uid := range uids {
		s.SendMsg(uid, event, data, true)
	}
}

// 获取对话落后信息
func (s *Service) GetDialogAfterMsg(ctx context.Context, request []model.AfterMsg, userID string) ([]*model.GetDialogAfterMsgResponse, error) {
	var responses = make([]*model.GetDialogAfterMsgResponse, 0)
	dialogIds := make([]uint32, 0)
	for _, v := range request {
		dialogIds = append(dialogIds, v.DialogId)
	}

	//TODO 验证是否在对话内

	infos, err := s.relationDialogClient.GetDialogByIds(ctx, &relationgrpcv1.GetDialogByIdsRequest{
		DialogIds: dialogIds,
	})
	if err != nil {
		s.logger.Error("获取用户会话信息", zap.Error(err))
		return nil, err
	}
	var infos2 = make([]*msggrpcv1.GetGroupMsgIdAfterMsgRequest, 0)
	var infos3 = make([]*msggrpcv1.GetUserMsgIdAfterMsgRequest, 0)
	for i, i2 := range infos.Dialogs {
		if i2.Type == uint32(model.GroupConversation) {
			infos2 = append(infos2, &msggrpcv1.GetGroupMsgIdAfterMsgRequest{
				DialogId: i2.Id,
				MsgId:    request[i].MsgId,
			})
		} else {
			infos3 = append(infos3, &msggrpcv1.GetUserMsgIdAfterMsgRequest{
				DialogId: i2.Id,
				MsgId:    request[i].MsgId,
			})
		}
	}
	//获取群聊消息
	grouplist, err := s.msgClient.GetGroupMsgIdAfterMsgList(ctx, &msggrpcv1.GetGroupMsgIdAfterMsgListRequest{
		List: infos2,
	})
	if err != nil {
		return nil, err
	}
	for _, i2 := range grouplist.Messages {
		msgs := make([]*model.Message, 0)
		for _, i3 := range i2.GroupMessages {
			msg := model.Message{}
			msg.MsgId = uint64(i3.Id)
			msg.MsgType = uint(i3.Type)
			msg.Content = i3.Content
			msg.SenderId = i3.UserId
			msg.SendTime = i3.CreatedAt
			msgs = append(msgs, &msg)
		}
		responses = append(responses, &model.GetDialogAfterMsgResponse{
			DialogId: i2.DialogId,
			Messages: msgs,
		})
	}
	//获取私聊消息
	userlist, err := s.msgClient.GetUserMsgIdAfterMsgList(ctx, &msggrpcv1.GetUserMsgIdAfterMsgListRequest{
		List: infos3,
	})
	if err != nil {
		return nil, err
	}
	for _, i2 := range userlist.Messages {
		msgs := make([]*model.Message, 0)
		for _, i3 := range i2.UserMessages {
			msg := model.Message{}
			msg.MsgId = uint64(i3.Id)
			msg.MsgType = uint(i3.Type)
			msg.Content = i3.Content
			msg.SenderId = i3.SenderId
			msg.SendTime = i3.CreatedAt
			msgs = append(msgs, &msg)
		}
		responses = append(responses, &model.GetDialogAfterMsgResponse{
			DialogId: i2.DialogId,
			Messages: msgs,
		})
	}
	return responses, nil
}
