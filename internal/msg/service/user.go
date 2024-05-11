package service

//func (s *ServiceImpl) SendUserMsg(ctx context.Context, userID string, driverId string, req *v1.SendUserMsgRequest) (*v1.SendUserMsgResponse, error) {
//
//	//if !model.IsAllowedConversationType(req.IsBurnAfterReading) {
//	//	return nil, code.MsgErrInsertUserMessageFailed
//	//}
//	isBurnAfterReadingType := msggrpcv1.BurnAfterReadingType_IsBurnAfterReading
//	if !req.IsBurnAfterReading {
//		isBurnAfterReadingType = msggrpcv1.BurnAfterReadingType_NotBurnAfterReading
//	}
//	userRelationStatus1, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
//		UserId:   userID,
//		FriendId: req.ReceiverId,
//	})
//	if err != nil {
//		s.logger.Error("获取用户关系失败", zap.Error(err))
//		return nil, err
//	}
//
//	if userRelationStatus1.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
//		return nil, code.RelationUserErrFriendRelationNotFound
//	}
//
//	userRelationStatus2, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
//		UserId:   req.ReceiverId,
//		FriendId: userID,
//	})
//	if err != nil {
//		s.logger.Error("获取用户关系失败", zap.Error(err))
//		return nil, err
//	}
//
//	if userRelationStatus2.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
//		return nil, code.RelationUserErrFriendRelationNotFound
//	}
//
//	dialogs, err := s.relationDialogService.GetDialogById(ctx, &relationgrpcv1.GetDialogByIdRequest{
//		DialogId: uint32(req.DialogId),
//	})
//	if err != nil {
//		s.logger.Error("获取会话失败", zap.Error(err))
//		return nil, err
//	}
//
//	if dialogs.Type != uint32(relationgrpcv1.DialogType_USER_DIALOG) {
//		return nil, code.DialogErrGetDialogByIdFailed
//	}
//
//	dialogUser, err := s.relationDialogService.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
//		DialogId: uint32(req.DialogId),
//		UserId:   userID,
//	})
//	if err != nil {
//		s.logger.Error("获取用户会话失败", zap.Error(err))
//		return nil, err
//	}
//
//	dialogUser2, err := s.relationDialogService.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
//		DialogId: uint32(req.DialogId),
//		UserId:   req.ReceiverId,
//	})
//	if err != nil {
//		s.logger.Error("获取用户会话失败", zap.Error(err))
//		return nil, err
//	}
//
//	var message *msggrpcv1.SendUserMsgResponse
//	workflow.InitGrpc(s.dtmGrpcServer, s.relationServiceAddr, grpc.NewServer())
//	gid := shortuuid.New()
//	wfName := "send_user_msg_workflow_" + gid
//	if err := workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
//		if dialogUser.IsShow != uint32(relationgrpcv1.CloseOrOpenDialogType_OPEN) {
//			_, err := s.relationDialogService.CloseOrOpenDialog(ctx, &relationgrpcv1.CloseOrOpenDialogRequest{
//				DialogId: uint32(req.DialogId),
//				Action:   relationgrpcv1.CloseOrOpenDialogType_OPEN,
//				UserId:   userID,
//			})
//			if err != nil {
//				return status.Error(codes.Aborted, err.Error())
//			}
//
//			wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
//				_, err := s.relationDialogService.CloseOrOpenDialog(ctx, &relationgrpcv1.CloseOrOpenDialogRequest{
//					DialogId: uint32(req.DialogId),
//					Action:   relationgrpcv1.CloseOrOpenDialogType_CLOSE,
//					UserId:   userID,
//				})
//				return err
//			})
//		}
//
//		if dialogUser2.IsShow != uint32(relationgrpcv1.CloseOrOpenDialogType_OPEN) {
//			_, err = s.relationDialogService.CloseOrOpenDialog(ctx, &relationgrpcv1.CloseOrOpenDialogRequest{
//				DialogId: uint32(req.DialogId),
//				Action:   relationgrpcv1.CloseOrOpenDialogType_OPEN,
//				UserId:   dialogUser2.UserId,
//			})
//			if err != nil {
//				return status.Error(codes.Aborted, err.Error())
//			}
//			wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
//				_, err := s.relationDialogService.CloseOrOpenDialog(ctx, &relationgrpcv1.CloseOrOpenDialogRequest{
//					DialogId: uint32(req.DialogId),
//					Action:   relationgrpcv1.CloseOrOpenDialogType_CLOSE,
//					UserId:   dialogUser2.UserId,
//				})
//				return err
//			})
//		}
//
//		message, err = s.msgService.SendUserMessage(ctx, &msggrpcv1.SendUserMsgRequest{
//			DialogId:               uint32(req.DialogId),
//			SenderId:               userID,
//			ReceiverId:             req.ReceiverId,
//			Content:                req.Content,
//			Type:                   int32(req.Type),
//			ReplyId:                uint64(req.ReplyId),
//			IsBurnAfterReadingType: isBurnAfterReadingType,
//		})
//		if err != nil {
//			s.logger.Error("发送消息失败", zap.Error(err))
//			return status.Error(codes.Aborted, err.Error())
//		}
//
//		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
//			_, err := s.msgService.SendUserMessageRevert(wf.Context, &msggrpcv1.MsgIdRequest{MsgId: message.MsgId})
//			return err
//		})
//
//		return err
//	}); err != nil {
//		return nil, err
//	}
//
//	if err := workflow.Execute(wfName, gid, nil); err != nil {
//		return nil, code.MsgErrInsertUserMessageFailed
//	}
//
//	//查询发送者信息
//	info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
//		UserId: userID,
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	resp := &v1.SendUserMsgResponse{
//		MsgId: int(message.MsgId),
//	}
//
//	if req.ReplyId != 0 {
//		msg, err := s.msgService.GetUserMessageById(ctx, &msggrpcv1.GetUserMsgByIDRequest{
//			MsgId: uint32(req.ReplyId),
//		})
//		if err != nil {
//			return nil, err
//		}
//
//		userInfo, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
//			UserId: msg.SenderId,
//		})
//		if err != nil {
//			return nil, err
//		}
//
//		resp.ReplyMsg = &v1.Message{
//			MsgType:  int(msg.Type),
//			Content:  msg.Content,
//			SenderId: msg.SenderId,
//			SendAt:   int(msg.GetCreatedAt()),
//			MsgId:    int(msg.Id),
//			SenderInfo: &v1.SenderInfo{
//				UserId: userInfo.UserId,
//				Name:   userInfo.NickName,
//				Avatar: userInfo.Avatar,
//			},
//			ReplyId: int(msg.ReplyId),
//		}
//
//		if msg.IsBurnAfterReadingType == msggrpcv1.BurnAfterReadingType_IsBurnAfterReading {
//			resp.ReplyMsg.IsBurnAfterReading = true
//		}
//
//		if msg.IsLabel == msggrpcv1.MsgLabel_IsLabel {
//			resp.ReplyMsg.IsLabel = true
//		}
//	}
//	rmsg := &pushv1.MessageInfo{}
//	if resp.ReplyMsg != nil {
//		rmsg = &pushv1.MessageInfo{
//			GroupId:  uint32(resp.ReplyMsg.GroupId),
//			MsgType:  uint32(resp.ReplyMsg.MsgType),
//			Content:  resp.ReplyMsg.Content,
//			SenderId: resp.ReplyMsg.SenderId,
//			SendAt:   int64(resp.ReplyMsg.SendAt),
//			MsgId:    uint64(resp.ReplyMsg.MsgId),
//			SenderInfo: &pushv1.SenderInfo{
//				UserId: resp.ReplyMsg.SenderInfo.UserId,
//				Avatar: resp.ReplyMsg.SenderInfo.Avatar,
//				Name:   resp.ReplyMsg.SenderInfo.Name,
//			},
//			ReceiverInfo: &pushv1.SenderInfo{
//				UserId: resp.ReplyMsg.ReceiverInfo.UserId,
//				Avatar: resp.ReplyMsg.ReceiverInfo.Avatar,
//				Name:   resp.ReplyMsg.ReceiverInfo.Name,
//			},
//			AtAllUser:          resp.ReplyMsg.AtAllUser,
//			AtUsers:            resp.ReplyMsg.AtUsers,
//			IsBurnAfterReading: resp.ReplyMsg.IsBurnAfterReading,
//			IsLabel:            resp.ReplyMsg.IsLabel,
//			ReplyId:            uint32(resp.ReplyMsg.ReplyId),
//			IsRead:             resp.ReplyMsg.IsRead,
//			ReadAt:             int64(resp.ReplyMsg.ReadAt),
//		}
//	}
//
//	//推送
//	s.sendWsUserMsg(userID, req.ReceiverId, driverId, userRelationStatus2.IsSilent, &pushv1.SendWsUserMsg{
//		SenderId:                userID,
//		Content:                 req.Content,
//		MsgType:                 uint32(req.Type),
//		ReplyId:                 uint32(req.ReplyId),
//		MsgId:                   message.MsgId,
//		ReceiverId:              req.ReceiverId,
//		SendAt:                  pkgtime.Now(),
//		DialogId:                uint32(req.DialogId),
//		IsBurnAfterReading:      resp.ReplyMsg.IsBurnAfterReading,
//		BurnAfterReadingTimeOut: userRelationStatus1.OpenBurnAfterReadingTimeOut,
//		SenderInfo: &pushv1.SenderInfo{
//			Avatar: info.Avatar,
//			Name:   info.NickName,
//			UserId: userID,
//		},
//		ReplyMsg: rmsg,
//	})
//
//	return resp, nil
//}
//
//// 推送私聊消息
//func (s *ServiceImpl) sendWsUserMsg(senderId, receiverId, driverId string, silent relationgrpcv1.UserSilentNotificationType, msg *pushv1.SendWsUserMsg) {
//
//	bytes, err := utils.StructToBytes(msg)
//	if err != nil {
//		return
//	}
//	m := &pushv1.WsMsg{Uid: receiverId, DriverId: driverId, Event: pushv1.WSEventType_SendUserMessageEvent, PushOffline: true, SendAt: pkgtime.Now(),
//		Data: &any.Any{Value: bytes},
//	}
//
//	//是否静默通知
//	if silent == relationgrpcv1.UserSilentNotificationType_UserSilent {
//		m.Event = pushv1.WSEventType_SendSilentUserMessageEvent
//	}
//	bytes2, err := utils.StructToBytes(m)
//	if err != nil {
//		return
//	}
//
//	//接受者不为系统则推送
//	if !constants.IsSystemUser(constants.SystemUser(receiverId)) {
//		//遍历该用户所有客户端
//		_, err := s.pushService.Push(context.Background(), &pushv1.PushRequest{
//			Type: pushv1.Type_Ws,
//			Data: bytes2,
//		})
//
//		if err != nil {
//			s.logger.Error("推送失败", zap.Error(err))
//		}
//
//	}
//	m.Uid = senderId
//	m.Event = pushv1.WSEventType_SendUserMessageEvent
//	bytes2, err = utils.StructToBytes(m)
//
//	if err != nil {
//		return
//	}
//	_, err = s.pushService.Push(context.Background(), &pushv1.PushRequest{
//		Type: pushv1.Type_Ws,
//		Data: bytes2,
//	})
//	if err != nil {
//		s.logger.Error("推送失败", zap.Error(err))
//	}
//
//}
//
//func (s *ServiceImpl) GetUserMessageList(ctx context.Context, userID string, req v1.GetUserMsgListParams) (*v1.GetUserMsgListResponse, error) {
//	msg, err := s.msgService.GetUserMessageList(ctx, &msggrpcv1.GetUserMsgListRequest{
//		DialogId: uint32(req.DialogId),
//		UserId:   userID,
//		Content:  req.Content,
//		Type:     int32(req.Type),
//		PageNum:  int32(req.PageNum),
//		PageSize: int32(req.PageSize),
//		MsgId:    uint64(req.MsgId),
//		StartAt:  uint64(req.StartAt),
//		EndAt:    uint64(req.EndAt),
//	})
//	if err != nil {
//		s.logger.Error("获取用户消息列表失败", zap.Error(err))
//		return nil, err
//	}
//
//	//查询对话用户
//	id, err := s.relationDialogService.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{DialogId: uint32(req.DialogId)})
//	if err != nil {
//		return nil, err
//	}
//	id2 := make([]string, 0)
//	for _, v := range id.UserIds {
//		if v != userID {
//			id2 = append(id2, v)
//		}
//	}
//
//	if len(id2) == 0 {
//		return nil, code.DialogErrGetDialogUserByDialogIDAndUserIDFailed
//	}
//
//	//查询关系
//	relation, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: userID, FriendId: id2[0]})
//	if err != nil {
//		s.logger.Error("获取用户关系失败", zap.Error(err))
//		return nil, err
//	}
//
//	resp := v1.GetUserMsgListResponse{}
//	resp.CurrentPage = int(msg.CurrentPage)
//	resp.Total = int(msg.Total)
//
//	msgList := make([]v1.UserMessage, 0)
//	for _, v := range msg.UserMessages {
//		info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
//			UserId: v.SenderId,
//		})
//		if err != nil {
//			return nil, err
//		}
//		info2, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
//			UserId: v.ReceiverId,
//		})
//		if err != nil {
//			return nil, err
//		}
//
//		sendinfo := &v1.SenderInfo{
//			Name:   info.NickName,
//			UserId: info.UserId,
//			Avatar: info.Avatar,
//		}
//
//		receinfo := &v1.SenderInfo{
//			Name:   info2.NickName,
//			UserId: info2.UserId,
//			Avatar: info2.Avatar,
//		}
//		name := relation.Remark
//		if name != "" {
//			if v.SenderId == userID {
//				receinfo.Name = name
//			} else {
//				sendinfo.Name = name
//			}
//		}
//
//		read := false
//		if v.IsRead == int32(msggrpcv1.ReadType_IsRead) {
//			read = true
//		}
//		label := false
//		if v.IsLabel == msggrpcv1.MsgLabel_IsLabel {
//			label = true
//		}
//		isBurnAfterReadingType := false
//		if v.IsBurnAfterReadingType == msggrpcv1.BurnAfterReadingType_IsBurnAfterReading {
//			isBurnAfterReadingType = true
//		}
//		msgList = append(msgList, v1.UserMessage{
//			MsgId:                   int(v.Id),
//			SenderId:                v.SenderId,
//			ReceiverId:              v.ReceiverId,
//			Content:                 v.Content,
//			Type:                    int(v.Type),
//			ReplyId:                 int(v.ReplyId),
//			IsRead:                  read,
//			ReadAt:                  int(v.ReadAt),
//			SendAt:                  int(v.CreatedAt),
//			DialogId:                int(v.DialogId),
//			IsLabel:                 label,
//			IsBurnAfterReadingType:  isBurnAfterReadingType,
//			BurnAfterReadingTimeout: int(relation.OpenBurnAfterReadingTimeOut),
//			SenderInfo:              sendinfo,
//			ReceiverInfo:            receinfo,
//		})
//	}
//
//	resp.UserMessages = &msgList
//
//	return &resp, nil
//}
//
//func (s *ServiceImpl) GetUserDialogList(ctx context.Context, userID string, pageSize int, pageNum int) (*v1.GetUserDialogListResponse, error) {
//	//获取对话id
//	ids, err := s.relationDialogService.GetUserDialogList(ctx, &relationgrpcv1.GetUserDialogListRequest{
//		UserId:   userID,
//		PageNum:  uint32(pageNum),
//		PageSize: uint32(pageSize),
//	})
//	if err != nil {
//		s.logger.Error("获取用户会话id", zap.Error(err))
//		return nil, err
//	}
//
//	//获取对话信息
//	infos, err := s.relationDialogService.GetDialogByIds(ctx, &relationgrpcv1.GetDialogByIdsRequest{
//		DialogIds: ids.DialogIds,
//	})
//	if err != nil {
//		s.logger.Error("获取用户会话信息", zap.Error(err))
//		return nil, err
//	}
//
//	//获取最后一条消息
//	dialogIds, err := s.msgService.GetLastMsgsByDialogIds(ctx, &msggrpcv1.GetLastMsgsByDialogIdsRequest{
//		DialogIds: ids.DialogIds,
//	})
//	if err != nil {
//		s.logger.Error("获取消息失败", zap.Error(err))
//		return nil, err
//	}
//
//	//封装响应数据
//	var responseList = make([]v1.UserDialogListResponse, 0)
//	for _, v := range infos.Dialogs {
//		var re v1.UserDialogListResponse
//		du, err := s.relationDialogService.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
//			DialogId: v.Id,
//			UserId:   userID,
//		})
//		if err != nil {
//			s.logger.Error("获取对话用户信息失败", zap.Error(err))
//			return nil, err
//		}
//		re.TopAt = int(du.TopAt)
//		//用户
//		if v.Type == 0 {
//			users, _ := s.relationDialogService.GetAllUsersInConversation(ctx, &relationgrpcv1.GetAllUsersInConversationRequest{
//				DialogId: v.Id,
//			})
//			if len(users.UserIds) == 0 {
//				continue
//			}
//			for _, id := range users.UserIds {
//				if id == userID {
//					continue
//				}
//				info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
//					UserId: id,
//				})
//				if err != nil {
//					s.logger.Error("获取用户信息失败", zap.Error(err))
//					continue
//				}
//
//				//获取未读消息
//				msgs, err := s.msgService.GetUnreadUserMsgs(ctx, &msggrpcv1.GetUnreadUserMsgsRequest{
//					UserId:   userID,
//					DialogId: v.Id,
//				})
//				if err != nil {
//					s.logger.Error("获取未读消息失败", zap.Error(err))
//				}
//				re.DialogId = int(v.Id)
//				re.DialogAvatar = info.Avatar
//				re.DialogName = info.NickName
//				re.DialogType = 0
//				re.DialogUnreadCount = len(msgs.UserMessages)
//				re.UserId = info.UserId
//				re.DialogCreateAt = int(v.CreateAt)
//
//				relation, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
//					UserId:   userID,
//					FriendId: id,
//				})
//				if err != nil {
//					s.logger.Error("获取用户关系失败", zap.Error(err))
//				}
//
//				if relation.Remark != "" {
//					re.DialogName = relation.Remark
//				}
//				break
//			}
//
//		} else if v.Type == 1 {
//			//群聊
//			info, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{
//				Gid: v.GroupId,
//			})
//			if err != nil {
//				s.logger.Error("获取群聊信息失败", zap.Error(err))
//				continue
//			}
//
//			//获取未读消息
//			msgs, err := s.msgService.GetGroupUnreadMessages(ctx, &msggrpcv1.GetGroupUnreadMessagesRequest{
//				UserId:   userID,
//				DialogId: v.Id,
//			})
//			if err != nil {
//				s.logger.Error("获取群聊关系失败", zap.Error(err))
//				//return nil, err
//			}
//
//			re.DialogAvatar = info.Avatar
//			re.DialogName = info.Name
//			re.DialogType = 1
//			re.DialogUnreadCount = len(msgs.GroupMessages)
//			re.GroupId = int(info.Id)
//			re.DialogId = int(v.Id)
//			re.DialogCreateAt = int(v.CreateAt)
//		}
//		// 匹配最后一条消息
//		for _, msg := range dialogIds.LastMsgs {
//			if msg.DialogId == v.Id {
//				re.LastMessage = &v1.Message{
//					MsgId:    int(msg.Id),
//					Content:  msg.Content,
//					SenderId: msg.SenderId,
//					SendAt:   int(msg.CreatedAt),
//					MsgType:  int(msg.Type),
//					ReadAt:   int(msg.ReadAt),
//					AtUsers:  msg.AtUsers,
//					ReplyId:  int(msg.ReplyId),
//				}
//				if msg.IsRead != 0 {
//					re.LastMessage.IsRead = true
//				}
//				if msg.IsLabel != 0 {
//					re.LastMessage.IsLabel = true
//				}
//				if msg.IsBurnAfterReadingType != 0 {
//					re.LastMessage.IsBurnAfterReading = true
//				}
//				if msg.AtAllUser != 0 {
//					re.LastMessage.AtAllUser = true
//				}
//				if msg.SenderId != "" {
//					info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
//						UserId: msg.SenderId,
//					})
//					if err != nil {
//						s.logger.Error("获取用户信息失败", zap.Error(err))
//						continue
//					}
//					re.LastMessage.SenderInfo = &v1.SenderInfo{
//						UserId: info.UserId,
//						Avatar: info.Avatar,
//						Name:   info.NickName,
//					}
//					if v.Type == 1 {
//						//查询群聊备注
//						relation, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: v.GroupId, UserId: info.UserId})
//						if err != nil {
//							s.logger.Error("查询群聊备注失败", zap.Error(err))
//							//return nil, err
//						}
//						if relation != nil {
//							if relation.Remark != "" {
//								re.LastMessage.SenderInfo.Name = relation.Remark
//							}
//						}
//
//					} else if v.Type == 0 && msg.SenderId != userID {
//						//查询用户备注
//						relation, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: userID, FriendId: msg.SenderId})
//						if err != nil {
//							s.logger.Error("查询用户备注失败", zap.Error(err))
//						}
//
//						if relation != nil {
//							if relation.Remark != "" {
//								re.LastMessage.SenderInfo.Name = relation.Remark
//							}
//						}
//
//					}
//				}
//				break
//			}
//		}
//
//		responseList = append(responseList, re)
//	}
//	//根据发送时间排序
//	sort.Slice(responseList, func(i, j int) bool {
//		return responseList[i].LastMessage.SendAt > responseList[j].LastMessage.SendAt
//	})
//
//	return &v1.GetUserDialogListResponse{
//		DialogList:  &responseList,
//		Total:       int(ids.Total),
//		CurrentPage: pageNum,
//	}, nil
//}
//
//func (s *ServiceImpl) RecallUserMsg(ctx context.Context, userID string, driverId string, msgID uint32) (interface{}, error) {
//	//获取消息
//	msginfo, err := s.msgService.GetUserMessageById(ctx, &msggrpcv1.GetUserMsgByIDRequest{
//		MsgId: msgID,
//	})
//	if err != nil {
//		s.logger.Error("获取消息失败", zap.Error(err))
//		return nil, err
//	}
//
//	if isPromptMessageType(msginfo.Type) {
//		return nil, code.MsgErrDeleteUserMessageFailed
//	}
//
//	if msginfo.SenderId != userID {
//		return nil, code.Unauthorized
//	}
//	//判断发送时间是否超过两分钟
//	if pkgtime.IsTimeDifferenceGreaterThanTwoMinutes(pkgtime.Now(), msginfo.CreatedAt) {
//		return nil, code.MsgErrTimeoutExceededCannotRevoke
//	}
//
//	//判断是否在对话内
//	_, err = s.relationDialogService.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
//		DialogId: msginfo.DialogId,
//	})
//	if err != nil {
//		s.logger.Error("获取用户对话信息失败", zap.Error(err))
//		return nil, err
//	}
//
//	req := &v1.SendUserMsgRequest{
//		ReceiverId: msginfo.ReceiverId,
//		Content:    msginfo.Content,
//		ReplyId:    int(msginfo.Id),
//		Type:       v1.SendUserMsgRequestType(msggrpcv1.MessageType_Delete),
//		DialogId:   int(msginfo.DialogId),
//	}
//
//	_, err = s.SendUserMsg(ctx, userID, driverId, req)
//	if err != nil {
//		return nil, err
//	}
//
//	// 调用相应的 gRPC 客户端方法来撤回用户消息
//	msg, err := s.msgService.DeleteUserMessage(ctx, &msggrpcv1.DeleteUserMsgRequest{
//		MsgId: msgID,
//	})
//	if err != nil {
//		s.logger.Error("撤回消息失败", zap.Error(err))
//		return nil, err
//	}
//
//	return msg.Id, nil
//}
//
//func isPromptMessageType(t uint32) bool {
//	validTypes := map[msggrpcv1.MessageType]struct{}{
//		msggrpcv1.MessageType_Label:       {},
//		msggrpcv1.MessageType_Notice:      {},
//		msggrpcv1.MessageType_Delete:      {},
//		msggrpcv1.MessageType_CancelLabel: {},
//	}
//	_, isValid := validTypes[msggrpcv1.MessageType(t)]
//	return isValid
//}
//
//func (s *ServiceImpl) EditUserMsg(c *gin.Context, userID string, driverId string, msgID uint32, content string) (interface{}, error) {
//	//获取消息
//	msginfo, err := s.msgService.GetUserMessageById(context.Background(), &msggrpcv1.GetUserMsgByIDRequest{
//		MsgId: msgID,
//	})
//	if err != nil {
//		s.logger.Error("获取消息失败", zap.Error(err))
//		return nil, err
//	}
//
//	if msginfo.SenderId != userID {
//		return nil, code.Unauthorized
//	}
//
//	relation, err := s.relationUserService.GetUserRelation(context.Background(), &relationgrpcv1.GetUserRelationRequest{
//		UserId:   msginfo.SenderId,
//		FriendId: msginfo.ReceiverId,
//	})
//	if err != nil {
//		s.logger.Error("获取用户关系失败", zap.Error(err))
//		return nil, err
//	}
//	if relation.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
//		return nil, code.RelationUserErrFriendRelationNotFound
//	}
//
//	relation2, err := s.relationUserService.GetUserRelation(context.Background(), &relationgrpcv1.GetUserRelationRequest{
//		UserId:   msginfo.ReceiverId,
//		FriendId: msginfo.SenderId,
//	})
//	if err != nil {
//		s.logger.Error("获取用户关系失败", zap.Error(err))
//		return nil, err
//	}
//
//	if relation2.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
//		return nil, code.RelationUserErrFriendRelationNotFound
//	}
//	//判断是否在对话内
//	userIds, err := s.relationDialogService.GetDialogUsersByDialogID(c, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
//		DialogId: msginfo.DialogId,
//	})
//	if err != nil {
//		s.logger.Error("获取用户对话信息失败", zap.Error(err))
//		return nil, err
//	}
//
//	// 调用相应的 gRPC 客户端方法来编辑用户消息
//	_, err = s.msgService.EditUserMessage(context.Background(), &msggrpcv1.EditUserMsgRequest{
//		UserMessage: &msggrpcv1.UserMessage{
//			Id:      msgID,
//			Content: content,
//		},
//	})
//	if err != nil {
//		s.logger.Error("编辑用户消息失败", zap.Error(err))
//		return nil, err
//	}
//	msginfo.Content = content
//
//	s.SendMsgToUsers(userIds.UserIds, driverId, pushv1.WSEventType_EditMsgEvent, msginfo, true)
//
//	return msgID, nil
//}
//
//func (s *ServiceImpl) ReadUserMsgs(ctx context.Context, userid string, driverId string, req *v1.ReadUserMsgsRequest) (interface{}, error) {
//	ids, err := s.relationDialogService.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
//		DialogId: uint32(req.DialogId),
//	})
//	if err != nil {
//		s.logger.Error("批量设置私聊消息状态为已读", zap.Error(err))
//		return nil, err
//	}
//
//	found := false
//	index := 0
//	for i, v := range ids.UserIds {
//		if v == userid {
//			index = i
//			found = true
//			break
//		}
//	}
//	if !found {
//		return nil, code.NotFound
//	}
//
//	if len(ids.UserIds) == 1 {
//		return nil, code.SetMsgErrSetUserMsgReadStatusFailed
//	}
//
//	targetId := ""
//	if index == 0 {
//		targetId = ids.UserIds[1]
//	} else {
//		targetId = ids.UserIds[0]
//	}
//
//	relation, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
//		UserId:   userid,
//		FriendId: targetId,
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	if relation.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
//		return nil, code.RelationUserErrFriendRelationNotFound
//	}
//
//	var msgIdList []uint32
//	for _, msgId := range *req.MsgIds {
//		msgIdList = append(msgIdList, uint32(msgId))
//	}
//
//	if req.ReadAll {
//		_, err := s.msgService.ReadAllUserMsg(ctx, &msggrpcv1.ReadAllUserMsgRequest{
//			DialogId: uint32(req.DialogId),
//			UserId:   userid,
//		})
//		if err != nil {
//			s.logger.Error("批量设置私聊消息状态为已读", zap.Error(err))
//			return nil, err
//		}
//		return nil, nil
//
//	} else {
//		_, err = s.msgService.SetUserMsgsReadStatus(ctx, &msggrpcv1.SetUserMsgsReadStatusRequest{
//			MsgIds:                      msgIdList,
//			DialogId:                    uint32(req.DialogId),
//			OpenBurnAfterReadingTimeOut: relation.OpenBurnAfterReadingTimeOut,
//		})
//		if err != nil {
//			s.logger.Error("批量设置私聊消息状态为已读", zap.Error(err))
//			return nil, err
//		}
//	}
//
//	msgs, err := s.msgService.GetUserMessagesByIds(ctx, &msggrpcv1.GetUserMessagesByIdsRequest{
//		MsgIds: msgIdList,
//		UserId: userid,
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	//查询发送者信息
//	info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
//		UserId: userid,
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	var wsms []pushv1.WsUserOperatorMsg
//	for _, msginfo := range msgs.UserMessages {
//		wsm := pushv1.WsUserOperatorMsg{
//			MsgId:      msginfo.Id,
//			SenderId:   msginfo.SenderId,
//			ReceiverId: msginfo.ReceiverId,
//			Content:    msginfo.Content,
//			Type:       msginfo.Type,
//			ReplyId:    uint32(msginfo.ReplyId),
//			ReadAt:     msginfo.ReadAt,
//			CreatedAt:  msginfo.CreatedAt,
//			DialogId:   msginfo.DialogId,
//		}
//		if msginfo.IsRead != 0 {
//			wsm.IsRead = true
//		}
//		if msginfo.IsLabel != 0 {
//			wsm.IsLabel = true
//		}
//		if msginfo.IsBurnAfterReadingType != 0 {
//			wsm.IsBurnAfterReadingType = true
//		}
//		wsms = append(wsms, wsm)
//	}
//
//	s.SendMsgToUsers(ids.UserIds, driverId, pushv1.WSEventType_UserMsgReadEvent, map[string]interface{}{"msgs": wsms, "operator_info": v1.SenderInfo{
//		Avatar: info.Avatar,
//		Name:   info.NickName,
//		UserId: info.UserId,
//	}}, true)
//
//	return nil, nil
//}
//
//// 标注私聊消息
//func (s *ServiceImpl) LabelUserMessage(ctx context.Context, userID string, driverId string, msgID uint32, label bool) (interface{}, error) {
//	// 获取用户消息
//	msginfo, err := s.msgService.GetUserMessageById(ctx, &msggrpcv1.GetUserMsgByIDRequest{
//		MsgId: msgID,
//	})
//	if err != nil {
//		s.logger.Error("获取用户消息失败", zap.Error(err))
//		return nil, err
//	}
//
//	if isPromptMessageType(msginfo.Type) {
//		return nil, code.SetMsgErrSetUserMsgLabelFailed
//	}
//
//	//判断是否在对话内
//	userIds, err := s.relationDialogService.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
//		DialogId: msginfo.DialogId,
//	})
//	if err != nil {
//		s.logger.Error("获取用户对话信息失败", zap.Error(err))
//		return nil, err
//	}
//
//	found := false
//	for _, v := range userIds.UserIds {
//		if v == userID {
//			found = true
//			break
//		}
//	}
//
//	if !found {
//		return nil, code.RelationUserErrFriendRelationNotFound
//	}
//
//	islabel := msggrpcv1.MsgLabel_NotLabel
//	if label {
//		islabel = msggrpcv1.MsgLabel_IsLabel
//	}
//	// 调用 gRPC 客户端方法将用户消息设置为标注状态
//	_, err = s.msgService.SetUserMsgLabel(context.Background(), &msggrpcv1.SetUserMsgLabelRequest{
//		MsgId:   msgID,
//		IsLabel: islabel,
//	})
//	if err != nil {
//		s.logger.Error("设置用户消息标注失败", zap.Error(err))
//		return nil, err
//	}
//
//	msginfo.IsLabel = islabel
//
//	req := &v1.SendUserMsgRequest{
//		ReceiverId: msginfo.ReceiverId,
//		Content:    msginfo.Content,
//		ReplyId:    int(msginfo.Id),
//		Type:       v1.SendUserMsgRequestType(msggrpcv1.MsgLabel_IsLabel),
//		DialogId:   int(msginfo.DialogId),
//	}
//
//	if msginfo.SenderId != userID {
//		req.ReceiverId = msginfo.SenderId
//	}
//
//	if !label {
//		req.Type = v1.SendUserMsgRequestType(msggrpcv1.MessageType_CancelLabel)
//	}
//
//	_, err = s.SendUserMsg(ctx, userID, driverId, req)
//	if err != nil {
//		return nil, err
//	}
//
//	return nil, nil
//}
//
//func (s *ServiceImpl) GetUserLabelMsgList(ctx context.Context, userID string, dialogID uint32) (*v1.GetUserLabelMsgListResponse, error) {
//	_, err := s.relationDialogService.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
//		UserId:   userID,
//		DialogId: dialogID,
//	})
//	if err != nil {
//		s.logger.Error("获取用户对话失败", zap.Error(err))
//		return nil, err
//	}
//
//	msgs, err := s.msgService.GetUserMsgLabelByDialogId(ctx, &msggrpcv1.GetUserMsgLabelByDialogIdRequest{
//		DialogId: dialogID,
//	})
//	if err != nil {
//		s.logger.Error("获取用户标注消息失败", zap.Error(err))
//		return nil, err
//
//	}
//
//	resp := &v1.GetUserLabelMsgListResponse{}
//	for _, i2 := range msgs.MsgList {
//		read := false
//		if i2.ReadAt != 0 {
//			read = true
//		}
//		resp.MsgList = append(resp.MsgList, v1.Message{
//			MsgId:    int(i2.Id),
//			Content:  i2.Content,
//			MsgType:  int(i2.Type),
//			ReplyId:  int(i2.ReplyId),
//			SendAt:   int(i2.CreatedAt),
//			IsLabel:  i2.IsLabel == msggrpcv1.MsgLabel_IsLabel,
//			SenderId: i2.SenderId,
//			ReadAt:   int(i2.ReadAt),
//			IsRead:   read,
//		})
//	}
//	return resp, nil
//}
//
//// SendMsg 推送消息
//func (s *ServiceImpl) SendMsg(uid string, driverId string, event pushv1.WSEventType, data interface{}, pushOffline bool) {
//	bytes, err := utils.StructToBytes(data)
//	if err != nil {
//		return
//	}
//
//	m := &pushv1.WsMsg{Uid: uid, DriverId: driverId, Event: event, Rid: "", Data: &any.Any{Value: bytes}, PushOffline: pushOffline, SendAt: pkgtime.Now()}
//	bytes2, err := utils.StructToBytes(m)
//	if err != nil {
//		return
//	}
//	_, err = s.pushService.Push(context.Background(), &pushv1.PushRequest{
//		Type: pushv1.Type_Ws,
//		Data: bytes2,
//	})
//	if err != nil {
//		s.logger.Error("发送消息失败", zap.Error(err))
//	}
//}
//
//// SendMsgToUsers 推送多个用户消息
//func (s *ServiceImpl) SendMsgToUsers(uids []string, driverId string, event pushv1.WSEventType, data interface{}, pushOffline bool) {
//
//	for _, uid := range uids {
//		s.SendMsg(uid, driverId, event, data, pushOffline)
//	}
//}
//
//// 获取对话落后信息
//func (s *ServiceImpl) GetDialogAfterMsg(ctx context.Context, userID string, request []v1.AfterMsg) ([]*v1.GetDialogAfterMsgResponse, error) {
//	var responses = make([]*v1.GetDialogAfterMsgResponse, 0)
//	dialogIds := make([]uint32, 0)
//	for _, v := range request {
//		dialogIds = append(dialogIds, uint32(v.DialogId))
//	}
//
//	infos, err := s.relationDialogService.GetDialogByIds(ctx, &relationgrpcv1.GetDialogByIdsRequest{
//		DialogIds: dialogIds,
//	})
//	if err != nil {
//		s.logger.Error("获取用户会话信息", zap.Error(err))
//		return nil, err
//	}
//
//	//群聊对话
//	var infos2 = make([]*msggrpcv1.GetGroupMsgIdAfterMsgRequest, 0)
//	//私聊对话
//	var infos3 = make([]*msggrpcv1.GetUserMsgIdAfterMsgRequest, 0)
//
//	addToInfos := func(dialogID uint32, msgID uint32, dialogType uint32) {
//		if dialogType == uint32(relationgrpcv1.DialogType_GROUP_DIALOG) {
//			if msgID == 0 {
//				responses, err = s.getGroupDialogLast20Msg(ctx, userID, dialogID, responses)
//				if err != nil {
//					s.logger.Error("获取群聊落后消息失败", zap.Error(err))
//					return
//				}
//				return
//			}
//			infos2 = append(infos2, &msggrpcv1.GetGroupMsgIdAfterMsgRequest{
//				DialogId: dialogID,
//				MsgId:    msgID,
//			})
//			return
//		}
//
//		if msgID == 0 {
//			responses, err = s.getUserDialogLast20Msg(ctx, dialogID, responses)
//			if err != nil {
//				s.logger.Error("获取群聊落后消息失败", zap.Error(err))
//				return
//			}
//			return
//		}
//		infos3 = append(infos3, &msggrpcv1.GetUserMsgIdAfterMsgRequest{
//			DialogId: dialogID,
//			MsgId:    msgID,
//		})
//	}
//
//	for _, i2 := range infos.Dialogs {
//		for _, i3 := range request {
//			if i2.Id == uint32(i3.DialogId) {
//				addToInfos(i2.Id, uint32(i3.MsgId), i2.Type)
//				break
//			}
//		}
//	}
//
//	//获取群聊消息
//	grouplist, err := s.msgService.GetGroupMsgIdAfterMsgList(ctx, &msggrpcv1.GetGroupMsgIdAfterMsgListRequest{
//		List: infos2,
//	})
//	if err != nil {
//		return nil, err
//	}
//	for _, i2 := range grouplist.Messages {
//		msgs := make([]v1.Message, 0)
//		for _, i3 := range i2.GroupMessages {
//			info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
//				UserId: i3.UserId,
//			})
//			if err != nil {
//				s.logger.Error("获取用户信息", zap.Error(err))
//				continue
//			}
//			readmsg, err := s.msgGroupService.GetGroupMessageReadByMsgIdAndUserId(ctx, &msggrpcv1.GetGroupMessageReadByMsgIdAndUserIdRequest{
//				MsgId:  i3.Id,
//				UserId: userID,
//			})
//			if err != nil {
//				s.logger.Error("获取消息是否已读失败", zap.Error(err))
//				continue
//			}
//			msg := v1.Message{}
//			msg.GroupId = int(i3.GroupId)
//			msg.MsgId = int(i3.Id)
//			msg.MsgType = int(i3.Type)
//			msg.Content = i3.Content
//			msg.SenderId = i3.UserId
//			msg.SendAt = int(i3.CreatedAt)
//			msg.SenderInfo = &v1.SenderInfo{
//				Avatar: info.Avatar,
//				Name:   info.NickName,
//				UserId: info.UserId,
//			}
//			msg.AtUsers = i3.AtUsers
//			msg.ReplyId = int(i3.ReplyId)
//			msg.ReadAt = int(readmsg.ReadAt)
//			if readmsg.ReadAt != 0 {
//				msg.IsRead = true
//			}
//			if i3.AtAllUser != 0 {
//				msg.AtAllUser = true
//			}
//			if i3.IsLabel != 0 {
//				msg.IsLabel = true
//			}
//			if i3.IsBurnAfterReadingType != 0 {
//				msg.IsBurnAfterReading = true
//			}
//			msgs = append(msgs, msg)
//		}
//		responses = append(responses, &v1.GetDialogAfterMsgResponse{
//			DialogId: int(i2.DialogId),
//			Messages: &msgs,
//			Total:    int(i2.Total),
//		})
//	}
//
//	//获取私聊消息
//	userlist, err := s.msgService.GetUserMsgIdAfterMsgList(ctx, &msggrpcv1.GetUserMsgIdAfterMsgListRequest{
//		List: infos3,
//	})
//	if err != nil {
//		return nil, err
//	}
//	for _, i2 := range userlist.Messages {
//		msgs := make([]v1.Message, 0)
//		for _, i3 := range i2.UserMessages {
//			//查询发送者信息
//			info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
//				UserId: i3.SenderId,
//			})
//			if err != nil {
//				return nil, err
//			}
//
//			info2, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
//				UserId: i3.ReceiverId,
//			})
//			if err != nil {
//				return nil, err
//			}
//			msg := v1.Message{}
//			msg.MsgId = int(i3.Id)
//			msg.MsgType = int(i3.Type)
//			msg.Content = i3.Content
//			msg.SenderId = i3.SenderId
//			msg.SendAt = int(i3.CreatedAt)
//			msg.SenderInfo = &v1.SenderInfo{
//				Avatar: info.Avatar,
//				Name:   info.NickName,
//				UserId: info.UserId,
//			}
//			msg.ReceiverInfo = &v1.SenderInfo{
//				Avatar: info2.Avatar,
//				Name:   info2.NickName,
//				UserId: info2.UserId,
//			}
//			msg.ReplyId = int(i3.ReplyId)
//			msg.ReadAt = int(i3.ReadAt)
//			if i3.ReadAt != 0 {
//				msg.IsRead = true
//			}
//			if i3.IsLabel != 0 {
//				msg.IsLabel = true
//			}
//			if i3.IsBurnAfterReadingType != 0 {
//				msg.IsBurnAfterReading = true
//			}
//			msgs = append(msgs, msg)
//		}
//		responses = append(responses, &v1.GetDialogAfterMsgResponse{
//			DialogId: int(i2.DialogId),
//			Messages: &msgs,
//			Total:    int(i2.Total),
//		})
//	}
//
//	return responses, nil
//}
//
//// 获取群聊对话的最后二十条消息
//func (s *ServiceImpl) getGroupDialogLast20Msg(ctx context.Context, thisID string, dialogId uint32, responses []*v1.GetDialogAfterMsgResponse) ([]*v1.GetDialogAfterMsgResponse, error) {
//	list, err := s.msgService.GetGroupLastMessageList(ctx, &msggrpcv1.GetLastMsgListRequest{
//		DialogId: dialogId,
//		PageNum:  1,
//		PageSize: 20,
//	})
//	if err != nil {
//		return responses, err
//	}
//	msgs := make([]v1.Message, 0)
//	for _, gm := range list.GroupMessages {
//		info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
//			UserId: gm.UserId,
//		})
//		if err != nil {
//			return responses, err
//		}
//
//		readmsg, err := s.msgGroupService.GetGroupMessageReadByMsgIdAndUserId(ctx, &msggrpcv1.GetGroupMessageReadByMsgIdAndUserIdRequest{
//			MsgId:  gm.Id,
//			UserId: thisID,
//		})
//		if err != nil {
//			s.logger.Error("获取消息是否已读失败", zap.Error(err))
//			continue
//		}
//		msg := v1.Message{}
//		msg.GroupId = int(gm.GroupId)
//		msg.MsgId = int(gm.Id)
//		msg.MsgType = int(gm.Type)
//		msg.Content = gm.Content
//		msg.SenderId = gm.UserId
//		msg.SendAt = int(gm.CreatedAt)
//		msg.SenderInfo = &v1.SenderInfo{
//			Avatar: info.Avatar,
//			Name:   info.NickName,
//			UserId: info.UserId,
//		}
//		msg.AtUsers = gm.AtUsers
//		msg.ReplyId = int(gm.ReplyId)
//		if readmsg.ReadAt != 0 {
//			msg.IsRead = true
//			msg.ReadAt = int(readmsg.ReadAt)
//		}
//		if gm.IsLabel != 0 {
//			msg.IsLabel = true
//		}
//		if gm.AtAllUser != 0 {
//			msg.AtAllUser = true
//		}
//		msgs = append(msgs, msg)
//	}
//	responses = append(responses, &v1.GetDialogAfterMsgResponse{
//		DialogId: int(dialogId),
//		Messages: &msgs,
//		Total:    int(list.Total),
//	})
//
//	return responses, nil
//}
//
//// 获取私聊对话的最后二十条消息
//func (s *ServiceImpl) getUserDialogLast20Msg(ctx context.Context, dialogId uint32, responses []*v1.GetDialogAfterMsgResponse) ([]*v1.GetDialogAfterMsgResponse, error) {
//	list, err := s.msgService.GetUserLastMessageList(ctx, &msggrpcv1.GetLastMsgListRequest{
//		DialogId: dialogId,
//		PageNum:  1,
//		PageSize: 20,
//	})
//	if err != nil {
//		return responses, err
//	}
//	msgs := make([]v1.Message, 0)
//	for _, um := range list.UserMessages {
//		//查询发送者信息
//		info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
//			UserId: um.SenderId,
//		})
//		if err != nil {
//			return responses, err
//		}
//		info2, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
//			UserId: um.ReceiverId,
//		})
//		if err != nil {
//			return responses, err
//		}
//		msg := v1.Message{}
//		msg.MsgId = int(um.Id)
//		msg.MsgType = int(um.Type)
//		msg.Content = um.Content
//		msg.SenderId = um.SenderId
//		msg.SendAt = int(um.CreatedAt)
//		msg.SenderInfo = &v1.SenderInfo{
//			Avatar: info.Avatar,
//			Name:   info.NickName,
//			UserId: info.UserId,
//		}
//		msg.ReceiverInfo = &v1.SenderInfo{
//			Avatar: info2.Avatar,
//			Name:   info2.NickName,
//			UserId: info2.UserId,
//		}
//		msg.ReadAt = int(um.ReadAt)
//		msg.ReplyId = int(um.ReplyId)
//
//		if msg.ReadAt != 0 {
//			msg.IsRead = true
//		}
//		if um.IsBurnAfterReadingType != 0 {
//			msg.IsBurnAfterReading = true
//		}
//		if um.IsLabel != 0 {
//			msg.IsLabel = true
//		}
//		msgs = append(msgs, msg)
//	}
//	responses = append(responses, &v1.GetDialogAfterMsgResponse{
//		DialogId: int(dialogId),
//		Messages: &msgs,
//		Total:    int(list.Total),
//	})
//	return responses, nil
//}
