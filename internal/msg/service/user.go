package service

import (
	"context"
	"fmt"
	groupApi "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	"github.com/cossim/coss-server/internal/msg/api/http/model"
	pushv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
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
	"sort"
)

func (s *Service) SendUserMsg(ctx context.Context, userID string, driverId string, req *model.SendUserMsgRequest) (interface{}, error) {
	if !model.IsAllowedConversationType(req.IsBurnAfterReadingType) {
		return nil, code.MsgErrInsertUserMessageFailed
	}
	userRelationStatus1, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
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

	userRelationStatus2, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
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

	dialogs, err := s.relationDialogService.GetDialogById(ctx, &relationgrpcv1.GetDialogByIdRequest{
		DialogId: req.DialogId,
	})
	if err != nil {
		s.logger.Error("获取会话失败", zap.Error(err))
		return nil, err
	}

	if dialogs.Type != uint32(relationgrpcv1.DialogType_USER_DIALOG) {
		return nil, code.DialogErrGetDialogByIdFailed
	}

	dialogUser, err := s.relationDialogService.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		DialogId: req.DialogId,
		UserId:   userID,
	})
	if err != nil {
		s.logger.Error("获取用户会话失败", zap.Error(err))
		return nil, err
	}

	dialogUser2, err := s.relationDialogService.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		DialogId: req.DialogId,
		UserId:   req.ReceiverId,
	})
	if err != nil {
		s.logger.Error("获取用户会话失败", zap.Error(err))
		return nil, err
	}

	var message *msggrpcv1.SendUserMsgResponse
	workflow.InitGrpc(s.dtmGrpcServer, s.relationServiceAddr, grpc.NewServer())
	gid := shortuuid.New()
	wfName := "send_user_msg_workflow_" + gid
	if err := workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		if dialogUser.IsShow != uint32(relationgrpcv1.CloseOrOpenDialogType_OPEN) {
			_, err := s.relationDialogService.CloseOrOpenDialog(ctx, &relationgrpcv1.CloseOrOpenDialogRequest{
				DialogId: req.DialogId,
				Action:   relationgrpcv1.CloseOrOpenDialogType_OPEN,
				UserId:   userID,
			})
			if err != nil {
				return status.Error(codes.Aborted, err.Error())
			}

			wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
				_, err := s.relationDialogService.CloseOrOpenDialog(ctx, &relationgrpcv1.CloseOrOpenDialogRequest{
					DialogId: req.DialogId,
					Action:   relationgrpcv1.CloseOrOpenDialogType_CLOSE,
					UserId:   userID,
				})
				return err
			})
		}

		if dialogUser2.IsShow != uint32(relationgrpcv1.CloseOrOpenDialogType_OPEN) {
			_, err = s.relationDialogService.CloseOrOpenDialog(ctx, &relationgrpcv1.CloseOrOpenDialogRequest{
				DialogId: req.DialogId,
				Action:   relationgrpcv1.CloseOrOpenDialogType_OPEN,
				UserId:   dialogUser2.UserId,
			})
			if err != nil {
				return status.Error(codes.Aborted, err.Error())
			}
			wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
				_, err := s.relationDialogService.CloseOrOpenDialog(ctx, &relationgrpcv1.CloseOrOpenDialogRequest{
					DialogId: req.DialogId,
					Action:   relationgrpcv1.CloseOrOpenDialogType_CLOSE,
					UserId:   dialogUser2.UserId,
				})
				return err
			})
		}

		message, err = s.msgService.SendUserMessage(ctx, &msggrpcv1.SendUserMsgRequest{
			DialogId:               req.DialogId,
			SenderId:               userID,
			ReceiverId:             req.ReceiverId,
			Content:                req.Content,
			Type:                   int32(req.Type),
			ReplyId:                uint64(req.ReplyId),
			IsBurnAfterReadingType: msggrpcv1.BurnAfterReadingType(req.IsBurnAfterReadingType),
		})
		if err != nil {
			s.logger.Error("发送消息失败", zap.Error(err))
			return status.Error(codes.Aborted, err.Error())
		}

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := s.msgService.SendUserMessageRevert(wf.Context, &msggrpcv1.MsgIdRequest{MsgId: message.MsgId})
			return err
		})

		return err
	}); err != nil {
		return "", err
	}
	if err := workflow.Execute(wfName, gid, nil); err != nil {
		return "", code.MsgErrInsertUserMessageFailed
	}

	//查询发送者信息
	info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}

	resp := &model.SendUserMsgResponse{
		MsgId: message.MsgId,
	}

	if req.ReplyId != 0 {
		msg, err := s.msgService.GetUserMessageById(ctx, &msggrpcv1.GetUserMsgByIDRequest{
			MsgId: uint32(req.ReplyId),
		})
		if err != nil {
			return nil, err
		}

		userInfo, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
			UserId: msg.SenderId,
		})
		if err != nil {
			return nil, err
		}

		resp.ReplyMsg = &model.Message{
			MsgType:  uint(msg.Type),
			Content:  msg.Content,
			SenderId: msg.SenderId,
			SendAt:   msg.GetCreatedAt(),
			MsgId:    uint64(msg.Id),
			SenderInfo: model.SenderInfo{
				UserId: userInfo.UserId,
				Name:   userInfo.NickName,
				Avatar: userInfo.Avatar,
			},
			IsBurnAfterReading: model.BurnAfterReadingType(msg.IsBurnAfterReadingType),
			IsLabel:            model.LabelMsgType(msg.IsLabel),
			ReplyId:            uint32(msg.ReplyId),
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

	//推送
	s.sendWsUserMsg(userID, req.ReceiverId, driverId, userRelationStatus2.IsSilent, &pushv1.SendWsUserMsg{
		SenderId:                userID,
		Content:                 req.Content,
		MsgType:                 uint32(req.Type),
		ReplyId:                 uint32(req.ReplyId),
		MsgId:                   message.MsgId,
		ReceiverId:              req.ReceiverId,
		SendAt:                  pkgtime.Now(),
		DialogId:                req.DialogId,
		IsBurnAfterReading:      uint32(req.IsBurnAfterReadingType),
		BurnAfterReadingTimeOut: userRelationStatus1.OpenBurnAfterReadingTimeOut,
		SenderInfo: &pushv1.SenderInfo{
			Avatar: info.Avatar,
			Name:   info.NickName,
			UserId: userID,
		},
		ReplyMsg: rmsg,
	})

	return resp, nil
}

// 推送私聊消息
func (s *Service) sendWsUserMsg(senderId, receiverId, driverId string, silent relationgrpcv1.UserSilentNotificationType, msg *pushv1.SendWsUserMsg) {

	bytes, err := utils.StructToBytes(msg)
	if err != nil {
		return
	}
	m := &pushv1.WsMsg{Uid: receiverId, DriverId: driverId, Event: pushv1.WSEventType_SendUserMessageEvent, PushOffline: true, SendAt: pkgtime.Now(),
		Data: &any.Any{Value: bytes},
	}

	//是否静默通知
	if silent == relationgrpcv1.UserSilentNotificationType_UserSilent {
		m.Event = pushv1.WSEventType_SendSilentUserMessageEvent
	}
	bytes2, err := utils.StructToBytes(m)
	if err != nil {
		return
	}

	//接受者不为系统则推送
	if !constants.IsSystemUser(constants.SystemUser(receiverId)) {
		fmt.Println("给接受者推送消息")
		//遍历该用户所有客户端
		_, err := s.pushService.Push(context.Background(), &pushv1.PushRequest{
			Type: pushv1.Type_Ws,
			Data: bytes2,
		})

		if err != nil {
			s.logger.Error("推送失败", zap.Error(err))
		}

	}
	m.Uid = senderId
	m.Event = pushv1.WSEventType_SendUserMessageEvent
	bytes2, err = utils.StructToBytes(m)
	fmt.Println("给发送者推送消息")

	if err != nil {
		return
	}
	_, err = s.pushService.Push(context.Background(), &pushv1.PushRequest{
		Type: pushv1.Type_Ws,
		Data: bytes2,
	})
	if err != nil {
		s.logger.Error("推送失败", zap.Error(err))
	}

}

func (s *Service) GetUserMessageList(ctx context.Context, userID string, req *model.MsgListRequest) (interface{}, error) {
	msg, err := s.msgService.GetUserMessageList(ctx, &msggrpcv1.GetUserMsgListRequest{
		DialogId: req.DialogId,
		UserId:   userID,
		Content:  req.Content,
		Type:     req.Type,
		PageNum:  int32(req.PageNum),
		PageSize: int32(req.PageSize),
		MsgId:    uint64(req.MsgId),
	})
	if err != nil {
		s.logger.Error("获取用户消息列表失败", zap.Error(err))
		return nil, err
	}

	//查询对话用户
	id, err := s.relationDialogService.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{DialogId: req.DialogId})
	if err != nil {
		return nil, err
	}
	id2 := make([]string, 0)
	for _, v := range id.UserIds {
		if v != userID {
			id2 = append(id2, v)
		}
	}

	if len(id2) == 0 {
		return nil, code.DialogErrGetDialogUserByDialogIDAndUserIDFailed
	}

	//查询关系
	relation, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: userID, FriendId: id2[0]})
	if err != nil {
		s.logger.Error("获取用户关系失败", zap.Error(err))
		return nil, err
	}

	resp := &model.GetUserMsgListResponse{}
	resp.CurrentPage = msg.CurrentPage
	resp.Total = msg.Total

	msgList := make([]*model.UserMessage, 0)
	for _, v := range msg.UserMessages {
		info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
			UserId: v.SenderId,
		})
		if err != nil {
			return nil, err
		}
		info2, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
			UserId: v.ReceiverId,
		})
		if err != nil {
			return nil, err
		}

		sendinfo := model.SenderInfo{
			Name:   info.NickName,
			UserId: info.UserId,
			Avatar: info.Avatar,
		}

		receinfo := model.SenderInfo{
			Name:   info2.NickName,
			UserId: info2.UserId,
			Avatar: info2.Avatar,
		}
		name := relation.Remark
		if name != "" {
			if v.SenderId == userID {
				receinfo.Name = name
			} else {
				sendinfo.Name = name
			}
		}

		msgList = append(msgList, &model.UserMessage{
			MsgId:                   v.Id,
			SenderId:                v.SenderId,
			ReceiverId:              v.ReceiverId,
			Content:                 v.Content,
			Type:                    v.Type,
			ReplyId:                 v.ReplyId,
			IsRead:                  v.IsRead,
			ReadAt:                  v.ReadAt,
			SendAt:                  v.CreatedAt,
			DialogId:                v.DialogId,
			IsLabel:                 model.LabelMsgType(v.IsLabel),
			IsBurnAfterReadingType:  model.BurnAfterReadingType(v.IsBurnAfterReadingType),
			BurnAfterReadingTimeOut: relation.OpenBurnAfterReadingTimeOut,
			SenderInfo:              sendinfo,
			ReceiverInfo:            receinfo,
		})
	}
	resp.UserMessages = msgList

	return resp, nil
}

func (s *Service) GetUserDialogList(ctx context.Context, userID string, pageSize int, pageNum int) (interface{}, error) {
	//获取对话id
	ids, err := s.relationDialogService.GetUserDialogList(ctx, &relationgrpcv1.GetUserDialogListRequest{
		UserId:   userID,
		PageNum:  uint32(pageNum),
		PageSize: uint32(pageSize),
	})
	if err != nil {
		s.logger.Error("获取用户会话id", zap.Error(err))
		return nil, err
	}

	fmt.Println("ids => ", ids)

	//获取对话信息
	infos, err := s.relationDialogService.GetDialogByIds(ctx, &relationgrpcv1.GetDialogByIdsRequest{
		DialogIds: ids.DialogIds,
	})
	if err != nil {
		s.logger.Error("获取用户会话信息", zap.Error(err))
		return nil, err
	}

	//获取最后一条消息
	dialogIds, err := s.msgService.GetLastMsgsByDialogIds(ctx, &msggrpcv1.GetLastMsgsByDialogIdsRequest{
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
		du, err := s.relationDialogService.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
			DialogId: v.Id,
			UserId:   userID,
		})
		fmt.Println("du => ", du)
		if err != nil {
			s.logger.Error("获取对话用户信息失败", zap.Error(err))
			return nil, err
		}
		re.TopAt = int64(du.TopAt)
		//用户
		if v.Type == 0 {
			users, _ := s.relationDialogService.GetAllUsersInConversation(ctx, &relationgrpcv1.GetAllUsersInConversationRequest{
				DialogId: v.Id,
			})
			if len(users.UserIds) == 0 {
				continue
			}
			for _, id := range users.UserIds {
				if id == userID {
					continue
				}
				info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
					UserId: id,
				})
				if err != nil {
					s.logger.Error("获取用户信息失败", zap.Error(err))
					continue
				}

				//获取未读消息
				msgs, err := s.msgService.GetUnreadUserMsgs(ctx, &msggrpcv1.GetUnreadUserMsgsRequest{
					UserId:   userID,
					DialogId: v.Id,
				})
				if err != nil {
					s.logger.Error("获取未读消息失败", zap.Error(err))
				}
				re.DialogId = v.Id
				re.DialogAvatar = info.Avatar
				re.DialogName = info.NickName
				re.DialogType = 0
				re.DialogUnreadCount = len(msgs.UserMessages)
				re.UserId = info.UserId
				re.DialogCreateAt = v.CreateAt

				relation, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
					UserId:   userID,
					FriendId: id,
				})
				if err != nil {
					s.logger.Error("获取用户关系失败", zap.Error(err))
				}

				if relation.Remark != "" {
					re.DialogName = relation.Remark
				}
				break
			}

		} else if v.Type == 1 {
			//群聊
			info, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{
				Gid: v.GroupId,
			})
			if err != nil {
				s.logger.Error("获取群聊信息失败", zap.Error(err))
				continue
			}

			//获取未读消息
			msgs, err := s.msgService.GetGroupUnreadMessages(ctx, &msggrpcv1.GetGroupUnreadMessagesRequest{
				UserId:   userID,
				DialogId: v.Id,
			})
			if err != nil {
				s.logger.Error("获取群聊关系失败", zap.Error(err))
				//return nil, err
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
					MsgId:              uint64(msg.Id),
					Content:            msg.Content,
					SenderId:           msg.SenderId,
					SendAt:             msg.CreatedAt,
					MsgType:            uint(msg.Type),
					IsRead:             msg.IsRead,
					ReadAt:             msg.ReadAt,
					AtUsers:            msg.AtUsers,
					AtAllUser:          model.AtAllUserType(msg.AtAllUser),
					IsLabel:            model.LabelMsgType(msg.IsLabel),
					IsBurnAfterReading: model.BurnAfterReadingType(msg.IsBurnAfterReadingType),
					ReplyId:            msg.ReplyId,
				}
				if msg.SenderId != "" {
					info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
						UserId: msg.SenderId,
					})
					if err != nil {
						s.logger.Error("获取用户信息失败", zap.Error(err))
						continue
					}
					re.LastMessage.SenderInfo = model.SenderInfo{
						UserId: info.UserId,
						Avatar: info.Avatar,
						Name:   info.NickName,
					}
					if v.Type == 1 {
						//查询群聊备注
						relation, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: v.GroupId, UserId: info.UserId})
						if err != nil {
							s.logger.Error("查询群聊备注失败", zap.Error(err))
							//return nil, err
						}
						if relation != nil {
							if relation.Remark != "" {
								re.LastMessage.SenderInfo.Name = relation.Remark
							}
						}

					} else if v.Type == 0 && msg.SenderId != userID {
						//查询用户备注
						relation, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: userID, FriendId: msg.SenderId})
						if err != nil {
							s.logger.Error("查询用户备注失败", zap.Error(err))
						}

						if relation != nil {
							if relation.Remark != "" {
								re.LastMessage.SenderInfo.Name = relation.Remark
							}
						}

					}
				}
				break
			}
		}

		responseList = append(responseList, re)
	}
	//根据发送时间排序
	sort.Slice(responseList, func(i, j int) bool {
		return responseList[i].LastMessage.SendAt > responseList[j].LastMessage.SendAt
	})

	return model.GetUserDialogListResponse{
		DialogList:  responseList,
		Total:       int64(ids.Total),
		CurrentPage: int32(pageNum),
	}, nil
}

func (s *Service) RecallUserMsg(ctx context.Context, userID string, driverId string, msgID uint32) (interface{}, error) {
	//获取消息
	msginfo, err := s.msgService.GetUserMessageById(ctx, &msggrpcv1.GetUserMsgByIDRequest{
		MsgId: msgID,
	})
	if err != nil {
		s.logger.Error("获取消息失败", zap.Error(err))
		return nil, err
	}

	if model.IsPromptMessageType(model.UserMessageType(msginfo.Type)) {
		return nil, code.MsgErrDeleteUserMessageFailed
	}

	if msginfo.SenderId != userID {
		return nil, code.Unauthorized
	}
	//判断发送时间是否超过两分钟
	if pkgtime.IsTimeDifferenceGreaterThanTwoMinutes(pkgtime.Now(), msginfo.CreatedAt) {
		return nil, code.MsgErrTimeoutExceededCannotRevoke
	}

	//判断是否在对话内
	_, err = s.relationDialogService.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
		DialogId: msginfo.DialogId,
	})
	if err != nil {
		s.logger.Error("获取用户对话信息失败", zap.Error(err))
		return nil, err
	}

	req := &model.SendUserMsgRequest{
		ReceiverId: msginfo.ReceiverId,
		Content:    msginfo.Content,
		ReplyId:    uint(msginfo.Id),
		Type:       model.MessageTypeDelete,
		DialogId:   msginfo.DialogId,
	}

	_, err = s.SendUserMsg(ctx, userID, driverId, req)
	if err != nil {
		return nil, err
	}

	// 调用相应的 gRPC 客户端方法来撤回用户消息
	msg, err := s.msgService.DeleteUserMessage(ctx, &msggrpcv1.DeleteUserMsgRequest{
		MsgId: msgID,
	})
	if err != nil {
		s.logger.Error("撤回消息失败", zap.Error(err))
		return nil, err
	}

	return msg.Id, nil
}

func (s *Service) EditUserMsg(c *gin.Context, userID string, driverId string, msgID uint32, content string) (interface{}, error) {
	//获取消息
	msginfo, err := s.msgService.GetUserMessageById(context.Background(), &msggrpcv1.GetUserMsgByIDRequest{
		MsgId: msgID,
	})
	if err != nil {
		s.logger.Error("获取消息失败", zap.Error(err))
		return nil, err
	}

	if msginfo.SenderId != userID {
		return nil, code.Unauthorized
	}

	relation, err := s.relationUserService.GetUserRelation(context.Background(), &relationgrpcv1.GetUserRelationRequest{
		UserId:   msginfo.SenderId,
		FriendId: msginfo.ReceiverId,
	})
	if err != nil {
		s.logger.Error("获取用户关系失败", zap.Error(err))
		return nil, err
	}
	if relation.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
		return nil, code.RelationUserErrFriendRelationNotFound
	}

	relation2, err := s.relationUserService.GetUserRelation(context.Background(), &relationgrpcv1.GetUserRelationRequest{
		UserId:   msginfo.ReceiverId,
		FriendId: msginfo.SenderId,
	})
	if err != nil {
		s.logger.Error("获取用户关系失败", zap.Error(err))
		return nil, err
	}

	if relation2.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
		return nil, code.RelationUserErrFriendRelationNotFound
	}
	//判断是否在对话内
	userIds, err := s.relationDialogService.GetDialogUsersByDialogID(c, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
		DialogId: msginfo.DialogId,
	})
	if err != nil {
		s.logger.Error("获取用户对话信息失败", zap.Error(err))
		return nil, err
	}

	// 调用相应的 gRPC 客户端方法来编辑用户消息
	_, err = s.msgService.EditUserMessage(context.Background(), &msggrpcv1.EditUserMsgRequest{
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

	s.SendMsgToUsers(userIds.UserIds, driverId, pushv1.WSEventType_EditMsgEvent, msginfo, true)

	return msgID, nil
}

func (s *Service) ReadUserMsgs(ctx context.Context, userid string, driverId string, req *model.ReadUserMsgsRequest) (interface{}, error) {
	ids, err := s.relationDialogService.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
		DialogId: req.DialogId,
	})
	if err != nil {
		s.logger.Error("批量设置私聊消息状态为已读", zap.Error(err))
		return nil, err
	}

	found := false
	index := 0
	for i, v := range ids.UserIds {
		if v == userid {
			index = i
			found = true
			break
		}
	}
	if !found {
		return nil, code.NotFound
	}

	if len(ids.UserIds) == 1 {
		return nil, code.SetMsgErrSetUserMsgReadStatusFailed
	}

	targetId := ""
	if index == 0 {
		targetId = ids.UserIds[1]
	} else {
		targetId = ids.UserIds[0]
	}

	relation, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
		UserId:   userid,
		FriendId: targetId,
	})
	if err != nil {
		return nil, err
	}

	if relation.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
		return nil, code.RelationUserErrFriendRelationNotFound
	}

	if req.ReadAll {
		_, err := s.msgService.ReadAllUserMsg(ctx, &msggrpcv1.ReadAllUserMsgRequest{
			DialogId: req.DialogId,
			UserId:   userid,
		})
		if err != nil {
			s.logger.Error("批量设置私聊消息状态为已读", zap.Error(err))
			return nil, err
		}
		return nil, nil

	} else {
		_, err = s.msgService.SetUserMsgsReadStatus(ctx, &msggrpcv1.SetUserMsgsReadStatusRequest{
			MsgIds:                      req.MsgIds,
			DialogId:                    req.DialogId,
			OpenBurnAfterReadingTimeOut: relation.OpenBurnAfterReadingTimeOut,
		})
		if err != nil {
			s.logger.Error("批量设置私聊消息状态为已读", zap.Error(err))
			return nil, err
		}
	}

	msgs, err := s.msgService.GetUserMessagesByIds(ctx, &msggrpcv1.GetUserMessagesByIdsRequest{
		MsgIds: req.MsgIds,
		UserId: userid,
	})
	if err != nil {
		return nil, err
	}

	//查询发送者信息
	info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
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
			ReplyId:                msginfo.ReplyId,
			IsRead:                 msginfo.IsRead,
			ReadAt:                 msginfo.ReadAt,
			CreatedAt:              msginfo.CreatedAt,
			DialogId:               msginfo.DialogId,
			IsLabel:                model.LabelMsgType(msginfo.IsLabel),
			IsBurnAfterReadingType: model.BurnAfterReadingType(msginfo.IsBurnAfterReadingType),
		}
		wsms = append(wsms, wsm)
	}

	s.SendMsgToUsers(ids.UserIds, driverId, pushv1.WSEventType_UserMsgReadEvent, map[string]interface{}{"msgs": wsms, "operator_info": model.SenderInfo{
		Avatar: info.Avatar,
		Name:   info.NickName,
		UserId: info.UserId,
	}}, true)

	return nil, nil
}

// 标注私聊消息
func (s *Service) LabelUserMessage(ctx context.Context, userID string, driverId string, msgID uint32, label model.LabelMsgType) (interface{}, error) {
	// 获取用户消息
	msginfo, err := s.msgService.GetUserMessageById(ctx, &msggrpcv1.GetUserMsgByIDRequest{
		MsgId: msgID,
	})
	if err != nil {
		s.logger.Error("获取用户消息失败", zap.Error(err))
		return nil, err
	}

	if model.IsPromptMessageType(model.UserMessageType(msginfo.Type)) {
		return nil, code.SetMsgErrSetUserMsgLabelFailed
	}

	//判断是否在对话内
	userIds, err := s.relationDialogService.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
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
	_, err = s.msgService.SetUserMsgLabel(context.Background(), &msggrpcv1.SetUserMsgLabelRequest{
		MsgId:   msgID,
		IsLabel: msggrpcv1.MsgLabel(label),
	})
	if err != nil {
		s.logger.Error("设置用户消息标注失败", zap.Error(err))
		return nil, err
	}

	msginfo.IsLabel = msggrpcv1.MsgLabel(label)

	req := &model.SendUserMsgRequest{
		ReceiverId: msginfo.ReceiverId,
		Content:    msginfo.Content,
		ReplyId:    uint(msginfo.Id),
		Type:       model.MessageTypeLabel,
		DialogId:   msginfo.DialogId,
	}

	if msginfo.SenderId != userID {
		req.ReceiverId = msginfo.SenderId
	}

	if label == model.NotLabel {
		req.Type = model.MessageTypeCancelLabel
	}

	_, err = s.SendUserMsg(ctx, userID, driverId, req)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (s *Service) GetUserLabelMsgList(ctx context.Context, userID string, dialogID uint32) (interface{}, error) {
	_, err := s.relationDialogService.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		UserId:   userID,
		DialogId: dialogID,
	})
	if err != nil {
		s.logger.Error("获取用户对话失败", zap.Error(err))
		return nil, err
	}

	msgs, err := s.msgService.GetUserMsgLabelByDialogId(ctx, &msggrpcv1.GetUserMsgLabelByDialogIdRequest{
		DialogId: dialogID,
	})
	if err != nil {
		s.logger.Error("获取用户标注消息失败", zap.Error(err))
		return nil, err

	}

	return msgs, nil
}

// SendMsg 推送消息
func (s *Service) SendMsg(uid string, driverId string, event pushv1.WSEventType, data interface{}, pushOffline bool) {
	bytes, err := utils.StructToBytes(data)
	if err != nil {
		return
	}

	m := &pushv1.WsMsg{Uid: uid, DriverId: driverId, Event: event, Rid: "", Data: &any.Any{Value: bytes}, PushOffline: pushOffline, SendAt: pkgtime.Now()}
	bytes2, err := utils.StructToBytes(m)
	if err != nil {
		return
	}
	_, err = s.pushService.Push(context.Background(), &pushv1.PushRequest{
		Type: pushv1.Type_Ws,
		Data: bytes2,
	})
	if err != nil {
		s.logger.Error("发送消息失败", zap.Error(err))
	}
}

// SendMsgToUsers 推送多个用户消息
func (s *Service) SendMsgToUsers(uids []string, driverId string, event pushv1.WSEventType, data interface{}, pushOffline bool) {

	for _, uid := range uids {
		s.SendMsg(uid, driverId, event, data, pushOffline)
	}
}

// 获取对话落后信息
func (s *Service) GetDialogAfterMsg(ctx context.Context, userID string, request []model.AfterMsg) ([]*model.GetDialogAfterMsgResponse, error) {
	var responses = make([]*model.GetDialogAfterMsgResponse, 0)
	dialogIds := make([]uint32, 0)
	for _, v := range request {
		dialogIds = append(dialogIds, v.DialogId)
	}

	infos, err := s.relationDialogService.GetDialogByIds(ctx, &relationgrpcv1.GetDialogByIdsRequest{
		DialogIds: dialogIds,
	})
	if err != nil {
		s.logger.Error("获取用户会话信息", zap.Error(err))
		return nil, err
	}

	//群聊对话
	var infos2 = make([]*msggrpcv1.GetGroupMsgIdAfterMsgRequest, 0)
	//私聊对话
	var infos3 = make([]*msggrpcv1.GetUserMsgIdAfterMsgRequest, 0)

	addToInfos := func(dialogID uint32, msgID uint32, dialogType uint32) {
		if dialogType == uint32(model.GroupConversation) {
			if msgID == 0 {
				responses, err = s.getGroupDialogLast20Msg(ctx, userID, dialogID, responses)
				if err != nil {
					s.logger.Error("获取群聊落后消息失败", zap.Error(err))
					return
				}
				return
			}
			infos2 = append(infos2, &msggrpcv1.GetGroupMsgIdAfterMsgRequest{
				DialogId: dialogID,
				MsgId:    msgID,
			})
			return
		}

		if msgID == 0 {
			responses, err = s.getUserDialogLast20Msg(ctx, dialogID, responses)
			if err != nil {
				s.logger.Error("获取群聊落后消息失败", zap.Error(err))
				return
			}
			return
		}
		infos3 = append(infos3, &msggrpcv1.GetUserMsgIdAfterMsgRequest{
			DialogId: dialogID,
			MsgId:    msgID,
		})
	}

	for _, i2 := range infos.Dialogs {
		for _, i3 := range request {
			if i2.Id == i3.DialogId {
				addToInfos(i2.Id, i3.MsgId, i2.Type)
				break
			}
		}
	}

	//获取群聊消息
	grouplist, err := s.msgService.GetGroupMsgIdAfterMsgList(ctx, &msggrpcv1.GetGroupMsgIdAfterMsgListRequest{
		List: infos2,
	})
	if err != nil {
		return nil, err
	}
	for _, i2 := range grouplist.Messages {
		msgs := make([]*model.Message, 0)
		for _, i3 := range i2.GroupMessages {
			info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
				UserId: i3.UserId,
			})
			if err != nil {
				s.logger.Error("获取用户信息", zap.Error(err))
				continue
			}
			readmsg, err := s.msgGroupService.GetGroupMessageReadByMsgIdAndUserId(ctx, &msggrpcv1.GetGroupMessageReadByMsgIdAndUserIdRequest{
				MsgId:  i3.Id,
				UserId: userID,
			})
			if err != nil {
				s.logger.Error("获取消息是否已读失败", zap.Error(err))
				continue
			}
			msg := model.Message{}
			msg.GroupId = i3.GroupId
			msg.MsgId = uint64(i3.Id)
			msg.MsgType = uint(i3.Type)
			msg.Content = i3.Content
			msg.SenderId = i3.UserId
			msg.SendAt = i3.CreatedAt
			msg.SenderInfo = model.SenderInfo{
				Avatar: info.Avatar,
				Name:   info.NickName,
				UserId: info.UserId,
			}
			msg.AtUsers = i3.AtUsers
			msg.AtAllUser = model.AtAllUserType(i3.AtAllUser)
			msg.IsBurnAfterReading = model.BurnAfterReadingType(i3.IsBurnAfterReadingType)
			msg.ReplyId = i3.ReplyId
			msg.IsLabel = model.LabelMsgType(i3.IsLabel)
			msg.ReadAt = readmsg.ReadAt
			if msg.ReadAt != 0 {
				msg.IsRead = int32(msggrpcv1.ReadType_IsRead)
			}
			msgs = append(msgs, &msg)
		}
		responses = append(responses, &model.GetDialogAfterMsgResponse{
			DialogId: i2.DialogId,
			Messages: msgs,
			Total:    int64(i2.Total),
		})
	}

	//获取私聊消息
	userlist, err := s.msgService.GetUserMsgIdAfterMsgList(ctx, &msggrpcv1.GetUserMsgIdAfterMsgListRequest{
		List: infos3,
	})
	if err != nil {
		return nil, err
	}
	for _, i2 := range userlist.Messages {
		msgs := make([]*model.Message, 0)
		for _, i3 := range i2.UserMessages {
			//查询发送者信息
			info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
				UserId: i3.SenderId,
			})
			if err != nil {
				return nil, err
			}

			info2, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
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
			msg.SendAt = i3.CreatedAt
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
			msg.IsBurnAfterReading = model.BurnAfterReadingType(i3.IsBurnAfterReadingType)
			msg.ReplyId = uint32(i3.ReplyId)
			msg.ReadAt = i3.ReadAt
			msg.IsRead = i3.IsRead
			msg.IsLabel = model.LabelMsgType(i3.IsLabel)
			msgs = append(msgs, &msg)
		}
		responses = append(responses, &model.GetDialogAfterMsgResponse{
			DialogId: i2.DialogId,
			Messages: msgs,
			Total:    int64(i2.Total),
		})
	}

	return responses, nil
}

// 获取群聊对话的最后二十条消息
func (s *Service) getGroupDialogLast20Msg(ctx context.Context, thisID string, dialogId uint32, responses []*model.GetDialogAfterMsgResponse) ([]*model.GetDialogAfterMsgResponse, error) {
	list, err := s.msgService.GetGroupLastMessageList(ctx, &msggrpcv1.GetLastMsgListRequest{
		DialogId: dialogId,
		PageNum:  1,
		PageSize: 20,
	})
	if err != nil {
		return responses, err
	}
	msgs := make([]*model.Message, 0)
	for _, gm := range list.GroupMessages {
		info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
			UserId: gm.UserId,
		})
		if err != nil {
			return responses, err
		}

		readmsg, err := s.msgGroupService.GetGroupMessageReadByMsgIdAndUserId(ctx, &msggrpcv1.GetGroupMessageReadByMsgIdAndUserIdRequest{
			MsgId:  gm.Id,
			UserId: thisID,
		})
		if err != nil {
			s.logger.Error("获取消息是否已读失败", zap.Error(err))
			continue
		}
		msg := model.Message{}
		msg.GroupId = gm.GroupId
		msg.MsgId = uint64(gm.Id)
		msg.MsgType = uint(gm.Type)
		msg.Content = gm.Content
		msg.SenderId = gm.UserId
		msg.SendAt = gm.CreatedAt
		msg.SenderInfo = model.SenderInfo{
			Avatar: info.Avatar,
			Name:   info.NickName,
			UserId: info.UserId,
		}
		msg.AtUsers = gm.AtUsers
		msg.AtAllUser = model.AtAllUserType(gm.AtAllUser)
		msg.IsBurnAfterReading = model.BurnAfterReadingType(gm.IsBurnAfterReadingType)
		msg.ReplyId = gm.ReplyId
		msg.IsLabel = model.LabelMsgType(gm.IsLabel)
		msg.ReadAt = readmsg.ReadAt
		if msg.ReadAt != 0 {
			msg.IsRead = int32(msggrpcv1.ReadType_IsRead)
		}
		msgs = append(msgs, &msg)
	}
	responses = append(responses, &model.GetDialogAfterMsgResponse{
		DialogId: dialogId,
		Messages: msgs,
		Total:    int64(list.Total),
	})

	return responses, nil
}

// 获取私聊对话的最后二十条消息
func (s *Service) getUserDialogLast20Msg(ctx context.Context, dialogId uint32, responses []*model.GetDialogAfterMsgResponse) ([]*model.GetDialogAfterMsgResponse, error) {
	list, err := s.msgService.GetUserLastMessageList(ctx, &msggrpcv1.GetLastMsgListRequest{
		DialogId: dialogId,
		PageNum:  1,
		PageSize: 20,
	})
	if err != nil {
		return responses, err
	}
	msgs := make([]*model.Message, 0)
	for _, um := range list.UserMessages {
		//查询发送者信息
		info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
			UserId: um.SenderId,
		})
		if err != nil {
			return responses, err
		}
		info2, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
			UserId: um.ReceiverId,
		})
		if err != nil {
			return responses, err
		}
		msg := model.Message{}
		msg.MsgId = uint64(um.Id)
		msg.MsgType = uint(um.Type)
		msg.Content = um.Content
		msg.SenderId = um.SenderId
		msg.SendAt = um.CreatedAt
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
		msg.ReadAt = um.ReadAt
		msg.IsRead = um.IsRead
		msg.IsBurnAfterReading = model.BurnAfterReadingType(um.IsBurnAfterReadingType)
		msg.ReplyId = uint32(um.ReplyId)
		msg.IsLabel = model.LabelMsgType(um.IsLabel)
		msgs = append(msgs, &msg)
	}
	responses = append(responses, &model.GetDialogAfterMsgResponse{
		DialogId: dialogId,
		Messages: msgs,
		Total:    int64(list.Total),
	})
	return responses, nil
}
