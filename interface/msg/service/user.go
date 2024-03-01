package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cossim/coss-server/interface/msg/api/model"
	"github.com/cossim/coss-server/pkg/cache"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/utils"
	pkgtime "github.com/cossim/coss-server/pkg/utils/time"
	groupApi "github.com/cossim/coss-server/service/group/api/v1"
	msggrpcv1 "github.com/cossim/coss-server/service/msg/api/v1"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	usergrpcv1 "github.com/cossim/coss-server/service/user/api/v1"
	pushv1 "github.com/cossim/hipush/api/grpc/v1"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/workflow"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/lithammer/shortuuid/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"log"
	"sort"
	"sync"
	"time"
)

func (s *Service) SendUserMsg(ctx context.Context, userID string, driverId string, req *model.SendUserMsgRequest) (interface{}, error) {
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

	dialogs, err := s.relationDialogClient.GetDialogById(ctx, &relationgrpcv1.GetDialogByIdRequest{
		DialogId: req.DialogId,
	})
	if err != nil {
		s.logger.Error("获取会话失败", zap.Error(err))
		return nil, err
	}

	if dialogs.Type != uint32(relationgrpcv1.DialogType_USER_DIALOG) {
		return nil, code.DialogErrGetDialogByIdFailed
	}

	_, err = s.relationDialogClient.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		DialogId: req.DialogId,
		UserId:   userID,
	})
	if err != nil {
		s.logger.Error("获取用户会话失败", zap.Error(err))
		return nil, err
	}

	dialogUser2, err := s.relationDialogClient.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		DialogId: req.DialogId,
		UserId:   req.ReceiverId,
	})
	if err != nil {
		s.logger.Error("获取用户会话失败", zap.Error(err))
		return nil, err
	}

	var message *msggrpcv1.SendUserMsgResponse
	workflow.InitGrpc(s.dtmGrpcServer, s.dialogGrpcServer, grpc.NewServer())
	gid := shortuuid.New()
	wfName := "send_user_msg_workflow_" + gid
	if err := workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		_, err := s.relationDialogClient.CloseOrOpenDialog(ctx, &relationgrpcv1.CloseOrOpenDialogRequest{
			DialogId: req.DialogId,
			Action:   relationgrpcv1.CloseOrOpenDialogType_OPEN,
			UserId:   userID,
		})
		if err != nil {
			return err
		}
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := s.relationDialogClient.CloseOrOpenDialog(ctx, &relationgrpcv1.CloseOrOpenDialogRequest{
				DialogId: req.DialogId,
				Action:   relationgrpcv1.CloseOrOpenDialogType_CLOSE,
				UserId:   userID,
			})
			return err
		})

		_, err = s.relationDialogClient.CloseOrOpenDialog(ctx, &relationgrpcv1.CloseOrOpenDialogRequest{
			DialogId: req.DialogId,
			Action:   relationgrpcv1.CloseOrOpenDialogType_OPEN,
			UserId:   dialogUser2.UserId,
		})
		if err != nil {
			return err
		}
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := s.relationDialogClient.CloseOrOpenDialog(ctx, &relationgrpcv1.CloseOrOpenDialogRequest{
				DialogId: req.DialogId,
				Action:   relationgrpcv1.CloseOrOpenDialogType_CLOSE,
				UserId:   dialogUser2.UserId,
			})
			return err
		})

		message, err = s.msgClient.SendUserMessage(ctx, &msggrpcv1.SendUserMsgRequest{
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
			return code.MsgErrInsertUserMessageFailed
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

	s.sendWsUserMsg(userID, req.ReceiverId, driverId, userRelationStatus2.IsSilent, &model.WsUserMsg{
		SenderId:           userID,
		Content:            req.Content,
		MsgType:            req.Type,
		ReplayId:           req.ReplayId,
		MsgId:              message.MsgId,
		ReceiverId:         req.ReceiverId,
		SendAt:             pkgtime.Now(),
		DialogId:           req.DialogId,
		IsBurnAfterReading: req.IsBurnAfterReadingType,
		SenderInfo: model.SenderInfo{
			Avatar: info.Avatar,
			Name:   info.NickName,
			UserId: userID,
		},
	})

	return message, nil
}

// 推送私聊消息
func (s *Service) sendWsUserMsg(senderId, receiverId, driverId string, silent relationgrpcv1.UserSilentNotificationType, msg *model.WsUserMsg) {
	sendFlag := false
	receFlag := false

	m := constants.WsMsg{Uid: receiverId, DriverId: driverId, Event: constants.SendUserMessageEvent, SendAt: pkgtime.Now(),
		Data: msg,
	}

	var is bool
	r, err := s.userLoginClient.GetUserLoginByUserId(context.Background(), &usergrpcv1.GetUserLoginByUserIdRequest{
		UserId: receiverId,
	})
	if err == nil {
		if r.Platform != "" && r.DriverToken != "" && err == nil {
			is = true
		}
	} else {
		s.logger.Error("获取用户登录信息失败", zap.Error(err))
	}

	//是否静默通知
	if silent == relationgrpcv1.UserSilentNotificationType_UserSilent {
		m.Event = constants.SendSilentUserMessageEvent
	}

	//接受者不为系统则推送
	if !constants.IsSystemUser(constants.SystemUser(receiverId)) {
		//遍历该用户所有客户端
		if _, ok := pool[receiverId]; ok {
			if len(pool[receiverId]) > 0 {
				receFlag = true
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
						if is {
							appid, ok := s.ac.Push.PlatformAppID[r.Platform]
							if !ok {
								s.logger.Error("platform appid not found", zap.String("platform", r.Platform))
								continue
							}
							if constants.DriverType(r.ClientType) != constants.MobileClient {
								s.logger.Info("client type not mobile", zap.String("client type", r.ClientType))
								continue
							}
							text, err := utils.ExtractText(msg.Content)
							if err != nil {
								s.logger.Error("extract html text err", zap.Error(err))
								continue
							}
							s.logger.Info("push message", zap.String("title", msg.SenderInfo.Name), zap.String("message", message), zap.String("platform", r.Platform), zap.String("driverToken", r.DriverToken), zap.String("appid", appid))
							if _, err := s.pushClient.Push(context.Background(), &pushv1.PushRequest{
								Platform:    r.Platform,
								Tokens:      []string{r.DriverToken},
								Title:       msg.SenderInfo.Name,
								Message:     text,
								AppID:       appid,
								Topic:       appid,
								Development: true,
							}); err != nil {
								s.logger.Error("push err", zap.Error(err), zap.String("title", msg.SenderInfo.Name), zap.String("message", message), zap.String("platform", r.Platform), zap.String("driverToken", r.DriverToken), zap.String("appid", appid))
							}
						}
					}
				}
			}
		}
	}

	if _, ok := pool[senderId]; ok {
		if len(pool[senderId]) > 0 {
			sendFlag = true
			for _, c := range pool[senderId] {
				for _, c2 := range c {
					m.Rid = c2.Rid
					js, _ := json.Marshal(m)
					message, err := Enc.GetSecretMessage(string(js), senderId)
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
		}
	}
	if Enc.IsEnable() {
		marshal, err := json.Marshal(m)
		if err != nil {
			s.logger.Error("json解析失败", zap.Error(err))
			return
		}
		if !receFlag && !constants.IsSystemUser(constants.SystemUser(receiverId)) {
			message, err := Enc.GetSecretMessage(string(marshal), receiverId)
			if err != nil {
				s.logger.Error("加密消息失败", zap.Error(err))
				return
			}
			err = rabbitMQClient.PublishEncryptedMessage(receiverId, message)
			if err != nil {
				s.logger.Error("发布消息失败", zap.Error(err))
				return
			}
		}

		if !sendFlag {
			message, err := Enc.GetSecretMessage(string(marshal), senderId)
			if err != nil {
				s.logger.Error("加密消息失败", zap.Error(err))
				return
			}
			err = rabbitMQClient.PublishEncryptedMessage(senderId, message)
			if err != nil {
				s.logger.Error("发布消息失败", zap.Error(err))
				return
			}
		}
		return
	}
	if !receFlag && !constants.IsSystemUser(constants.SystemUser(receiverId)) {
		err := rabbitMQClient.PublishMessage(receiverId, m)
		if err != nil {
			s.logger.Error("发布消息失败", zap.Error(err))
			return
		}
	}
	if !sendFlag {
		err := rabbitMQClient.PublishMessage(senderId, m)
		if err != nil {
			s.logger.Error("发布消息失败", zap.Error(err))
			return
		}
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
			users, _ := s.relationDialogClient.GetAllUsersInConversation(ctx, &relationgrpcv1.GetAllUsersInConversationRequest{
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

				//获取未读消息
				msgs, err := s.msgClient.GetUnreadUserMsgs(ctx, &msggrpcv1.GetUnreadUserMsgsRequest{
					UserId:   userID,
					DialogId: v.Id,
				})
				if err != nil {
					return nil, err
				}
				re.DialogId = v.Id
				re.DialogAvatar = info.Avatar
				re.DialogName = info.NickName
				re.DialogType = 0
				re.DialogUnreadCount = len(msgs.UserMessages)
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

			//获取未读消息
			msgs, err := s.msgClient.GetGroupUnreadMessages(ctx, &msggrpcv1.GetGroupUnreadMessagesRequest{
				UserId:   userID,
				DialogId: v.Id,
			})
			if err != nil {
				return nil, err
			}

			re.DialogAvatar = info.Avatar
			re.DialogName = info.Name
			re.DialogType = 1
			re.DialogUnreadCount = len(msgs.GroupMessages)
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

func (s *Service) RecallUserMsg(ctx context.Context, userID string, driverId string, msgID uint32) (interface{}, error) {
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

	//判断是否在对话内
	userIds, err := s.relationDialogClient.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
		DialogId: msginfo.DialogId,
	})
	if err != nil {
		s.logger.Error("获取用户对话信息失败", zap.Error(err))
		return nil, err
	}

	// 调用相应的 gRPC 客户端方法来撤回用户消息
	msg, err := s.msgClient.DeleteUserMessage(ctx, &msggrpcv1.DeleteUserMsgRequest{
		MsgId: msgID,
	})
	if err != nil {
		s.logger.Error("撤回消息失败", zap.Error(err))
		return nil, err
	}

	//查询发送者信息
	info, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}

	wsm := model.WsUserOperatorMsg{
		Id:                     msginfo.Id,
		SenderId:               msginfo.SenderId,
		ReceiverId:             msginfo.ReceiverId,
		Content:                msginfo.Content,
		Type:                   msginfo.Type,
		ReplayId:               msginfo.ReplayId,
		IsRead:                 msginfo.IsRead,
		ReadAt:                 msginfo.ReadAt,
		CreatedAt:              msginfo.CreatedAt,
		DialogId:               msginfo.DialogId,
		IsLabel:                model.LabelMsgType(msginfo.IsLabel),
		IsBurnAfterReadingType: model.BurnAfterReadingType(msginfo.IsBurnAfterReadingType),
		OperatorInfo: model.SenderInfo{
			Avatar: info.Avatar,
			Name:   info.NickName,
			UserId: info.UserId,
		},
	}

	s.SendMsgToUsers(userIds.UserIds, driverId, constants.RecallMsgEvent, wsm, true)

	return msg.Id, nil
}

func (s *Service) EditUserMsg(c *gin.Context, userID string, driverId string, msgID uint32, content string) (interface{}, error) {
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

	//判断是否在对话内
	userIds, err := s.relationDialogClient.GetDialogUsersByDialogID(c, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
		DialogId: msginfo.DialogId,
	})
	if err != nil {
		s.logger.Error("获取用户对话信息失败", zap.Error(err))
		return nil, err
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
	msginfo.Content = content

	s.SendMsgToUsers(userIds.UserIds, driverId, constants.EditMsgEvent, msginfo, true)

	return msgID, nil
}

func (s *Service) ReadUserMsgs(ctx context.Context, userid string, driverId string, dialogId uint32, msgids []uint32) (interface{}, error) {
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

	msgs, err := s.msgClient.GetUserMessagesByIds(ctx, &msggrpcv1.GetUserMessagesByIdsRequest{
		MsgIds: msgids,
		UserId: userid,
	})
	if err != nil {
		return nil, err
	}

	//查询发送者信息
	info, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
		UserId: userid,
	})
	if err != nil {
		return nil, err
	}

	var wsms []model.WsUserOperatorMsg
	for _, msginfo := range msgs.UserMessages {
		wsm := model.WsUserOperatorMsg{
			Id:                     msginfo.Id,
			SenderId:               msginfo.SenderId,
			ReceiverId:             msginfo.ReceiverId,
			Content:                msginfo.Content,
			Type:                   msginfo.Type,
			ReplayId:               msginfo.ReplayId,
			IsRead:                 msginfo.IsRead,
			ReadAt:                 msginfo.ReadAt,
			CreatedAt:              msginfo.CreatedAt,
			DialogId:               msginfo.DialogId,
			IsLabel:                model.LabelMsgType(msginfo.IsLabel),
			IsBurnAfterReadingType: model.BurnAfterReadingType(msginfo.IsBurnAfterReadingType),
		}
		wsms = append(wsms, wsm)
	}

	s.SendMsgToUsers(ids.UserIds, driverId, constants.UserMsgReadEvent, map[string]interface{}{"msgs": wsms, "OperatorInfo": model.SenderInfo{
		Avatar: info.Avatar,
		Name:   info.NickName,
		UserId: info.UserId,
	}}, true)

	return nil, nil
}

// 标注私聊消息
func (s *Service) LabelUserMessage(ctx context.Context, userID string, driverId string, msgID uint32, label model.LabelMsgType) (interface{}, error) {
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
		s.logger.Error("获取用户对话信息失败", zap.Error(err))
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

	msginfo.IsLabel = msggrpcv1.MsgLabel(label)

	//查询发送者信息
	info, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}

	wsm := model.WsUserOperatorMsg{
		Id:                     msginfo.Id,
		SenderId:               msginfo.SenderId,
		ReceiverId:             msginfo.ReceiverId,
		Content:                msginfo.Content,
		Type:                   msginfo.Type,
		ReplayId:               msginfo.ReplayId,
		IsRead:                 msginfo.IsRead,
		ReadAt:                 msginfo.ReadAt,
		CreatedAt:              msginfo.CreatedAt,
		DialogId:               msginfo.DialogId,
		IsLabel:                model.LabelMsgType(msginfo.IsLabel),
		IsBurnAfterReadingType: model.BurnAfterReadingType(msginfo.IsBurnAfterReadingType),
		OperatorInfo: model.SenderInfo{
			Avatar: info.Avatar,
			Name:   info.NickName,
			UserId: info.UserId,
		},
	}

	s.SendMsgToUsers(userIds.UserIds, driverId, constants.LabelMsgEvent, wsm, true)

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

func (s *Service) Ws(conn *websocket.Conn, uid string, driverId string, deviceType, token string) {
	defer conn.Close()
	//用户上线
	wsRid++
	messages := rabbitMQClient.GetChannel()
	if messages.IsClosed() {
		log.Fatal("Channel is Closed")
	}
	cli := &client{
		ClientType:     deviceType,
		Conn:           conn,
		Uid:            uid,
		Rid:            wsRid,
		queue:          messages,
		DriverId:       driverId,
		Rdb:            s.redisClient,
		relationClient: s.relationUserClient,
	}
	if _, ok := pool[uid]; !ok {
		pool[uid] = make(map[string][]*client)
	}

	if s.ac.MultipleDeviceLimit.Enable {
		if _, ok := pool[uid][deviceType]; ok {
			if len(pool[uid][deviceType]) == s.ac.MultipleDeviceLimit.Max {
				return
			}
		}
	}
	//保存到线程池
	cli.wsOnlineClients()

	//todo 加锁
	//更新登录信息
	keys, err := cache.ScanKeys(s.redisClient, cli.Uid+":"+deviceType+":*")
	if err != nil {
		s.logger.Error("获取用户信息失败1", zap.Error(err))
		return
	}

	for _, key := range keys {
		v, err := cache.GetKey(s.redisClient, key)
		if err != nil {
			s.logger.Error("获取用户信息失败", zap.Error(err))
			return
		}
		strKey := v.(string)
		info, err := cache.GetUserInfo(strKey)
		if err != nil {
			s.logger.Error("获取用户信息失败", zap.Error(err))
			return
		}
		if info.Token == token {
			info.Rid = cli.Rid
			resp := cache.GetUserInfoToInterfaces(info)
			err := cache.SetKey(s.redisClient, key, resp, 60*60*24*7*time.Second)
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
			keys, err := cache.ScanKeys(s.redisClient, cli.Uid+":"+deviceType+":*")
			if err != nil {
				s.logger.Error("获取用户信息失败1", zap.Error(err))
				return
			}

			for _, key := range keys {
				v, err := cache.GetKey(s.redisClient, key)
				if err != nil {
					s.logger.Error("获取用户信息失败", zap.Error(err))
					return
				}
				strKey := v.(string)
				info, err := cache.GetUserInfo(strKey)
				if err != nil {
					s.logger.Error("获取用户信息失败", zap.Error(err))
					return
				}
				if info.Token == token {
					info.Rid = 0
					resp := cache.GetUserInfoToInterfaces(info)
					err := cache.SetKey(s.redisClient, key, resp, 60*60*24*7*time.Second)
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
func (s *Service) SendMsg(uid string, driverId string, event constants.WSEventType, data interface{}, pushOffline bool) {
	m := constants.WsMsg{Uid: uid, DriverId: driverId, Event: event, Rid: 0, Data: data, SendAt: pkgtime.Now()}
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
func (s *Service) SendMsgToUsers(uids []string, driverId string, event constants.WSEventType, data interface{}, pushOffline bool) {
	var wg sync.WaitGroup
	for _, uid := range uids {
		wg.Add(1)
		go func(uid string) {
			defer wg.Done()
			s.SendMsg(uid, driverId, event, data, pushOffline)
		}(uid)
	}
	wg.Wait()
}

// 获取对话落后信息
func (s *Service) GetDialogAfterMsg(ctx context.Context, request []model.AfterMsg) ([]*model.GetDialogAfterMsgResponse, error) {
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
	for _, i2 := range infos.Dialogs {
		if i2.Type == uint32(model.GroupConversation) {
			for _, i3 := range request {
				if i2.Id == i3.DialogId {
					infos2 = append(infos2, &msggrpcv1.GetGroupMsgIdAfterMsgRequest{
						DialogId: i2.Id,
						MsgId:    i3.MsgId,
					})
					break
				}
			}

		} else {
			for _, i3 := range request {
				if i2.Id == i3.DialogId {
					infos3 = append(infos3, &msggrpcv1.GetUserMsgIdAfterMsgRequest{
						DialogId: i2.Id,
						MsgId:    i3.MsgId,
					})
					break
				}
			}
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
			info, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
				UserId: i3.UserId,
			})
			if err != nil {
				return nil, err
			}
			msg := model.Message{}
			msg.GroupId = i3.GroupId
			msg.MsgId = uint64(i3.Id)
			msg.MsgType = uint(i3.Type)
			msg.Content = i3.Content
			msg.SenderId = i3.UserId
			msg.SendTime = i3.CreatedAt
			msg.SenderInfo = model.SenderInfo{
				Avatar: info.Avatar,
				Name:   info.NickName,
				UserId: info.UserId,
			}
			msg.AtUsers = i3.AtUsers
			msg.AtAllUser = model.AtAllUserType(i3.AtAllUser)
			msg.IsBurnAfterReadingType = model.BurnAfterReadingType(i3.IsBurnAfterReadingType)
			msg.ReplayId = i3.ReplyId
			msg.IsLabel = model.LabelMsgType(i3.IsLabel)
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
			//查询发送者信息
			info, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
				UserId: i3.SenderId,
			})
			if err != nil {
				return nil, err
			}
			info2, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
				UserId: i3.ReceiverId,
			})
			if err != nil {
				return nil, err
			}
			msg := model.Message{}
			msg.MsgId = uint64(i3.Id)
			msg.MsgType = uint(i3.Type)
			msg.Content = i3.Content
			msg.SenderId = i3.SenderId
			msg.SendTime = i3.CreatedAt
			msg.SenderInfo = model.SenderInfo{
				Avatar: info.Avatar,
				Name:   info.NickName,
				UserId: info.UserId,
			}
			msg.ReceiverInfo = model.SenderInfo{
				Avatar: info2.Avatar,
				Name:   info2.NickName,
				UserId: info2.UserId,
			}
			msg.IsBurnAfterReadingType = model.BurnAfterReadingType(i3.IsBurnAfterReadingType)
			msg.ReplayId = uint32(i3.ReplayId)
			msg.IsLabel = model.LabelMsgType(i3.IsLabel)
			msgs = append(msgs, &msg)
		}
		responses = append(responses, &model.GetDialogAfterMsgResponse{
			DialogId: i2.DialogId,
			Messages: msgs,
		})
	}
	return responses, nil
}
