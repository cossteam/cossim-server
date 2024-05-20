package service

// 推送群聊消息
//func (s *ServiceImpl) sendWsGroupMsg(ctx context.Context, uIds []string, driverId string, msg *pushv1.SendWsGroupMsg) {
//	bytes, err := utils.StructToBytes(msg)
//	if err != nil {
//		return
//	}
//
//	//发送群聊消息
//	for _, uid := range uIds {
//		m := &pushv1.WsMsg{Uid: uid, DriverId: driverId, Event: pushv1.WSEventType_SendGroupMessageEvent, PushOffline: true, SendAt: pkgtime.Now(), Data: &any.Any{Value: bytes}}
//
//		//查询是否静默通知
//		groupRelation, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
//			GroupID: uint32(msg.GroupID),
//			ID:  uid,
//		})
//		if err != nil {
//			s.logger.Error("获取群聊关系失败", zap.Error(err))
//			continue
//		}
//
//		//判断是否静默通知
//		if groupRelation.IsSilent == relationgrpcv1.GroupSilentNotificationType_GroupSilent {
//			m.Event = pushv1.WSEventType_SendSilentGroupMessageEvent
//		}
//
//		bytes2, err := utils.StructToBytes(m)
//		if err != nil {
//			return
//		}
//
//		_, err = s.pushService.Push(ctx, &pushv1.PushRequest{
//			Type: pushv1.Type_Ws,
//			Data: bytes2,
//		})
//		if err != nil {
//			s.logger.Error("发送消息失败", zap.Error(err))
//		}
//
//	}
//}
//
//func (s *ServiceImpl) SendGroupMsg(ctx context.Context, userID string, driverId string, req *v1.SendGroupMsgRequest) (*v1.SendGroupMsgResponse, error) {
//	group, err := s.groupService.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{
//		Gid: uint32(req.GroupID),
//	})
//	if err != nil {
//		s.logger.Error("获取群聊信息失败", zap.Error(err))
//		return nil, err
//	}
//
//	if group.Status != groupgrpcv1.GroupStatus_GROUP_STATUS_NORMAL {
//		return nil, code.GroupErrGroupStatusNotAvailable
//	}
//
//	if group.SilenceTime != 0 {
//		return nil, code.GroupErrGroupIsSilence
//	}
//
//	dialogID := uint32(req.DialogID)
//
//	groupRelation, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
//		GroupID: uint32(req.GroupID),
//		ID:  userID,
//	})
//	if err != nil {
//		s.logger.Error("获取群聊关系失败", zap.Error(err))
//		return nil, err
//	}
//
//	if groupRelation.MuteEndTime > pkgtime.Now() && groupRelation.MuteEndTime != 0 {
//		return nil, code.GroupErrUserIsMuted
//	}
//
//	dialogs, err := s.relationDialogService.GetDialogByIds(ctx, &relationgrpcv1.GetDialogByIdsRequest{
//		DialogIds: []uint32{dialogID},
//	})
//	if err != nil {
//		s.logger.Error("获取会话失败", zap.Error(err))
//		return nil, err
//	}
//	if len(dialogs.Dialogs) == 0 {
//		return nil, code.DialogErrGetDialogUserByDialogIDAndUserIDFailed
//	}
//
//	_, err = s.relationDialogService.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
//		DialogID: dialogID,
//		ID:   userID,
//	})
//	if err != nil {
//		s.logger.Error("获取用户会话失败", zap.Error(err))
//		return nil, code.DialogErrGetDialogUserByDialogIDAndUserIDFailed
//	}
//
//	//查询群聊所有用户id
//	uids, err := s.relationGroupService.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{
//		GroupID: uint32(req.GroupID),
//	})
//
//	var message *msggrpcv1.SendGroupMsgResponse
//
//	workflow.InitGrpc(s.dtmGrpcServer, s.relationServiceAddr, grpc.NewServer())
//	gid := shortuuid.New()
//	wfName := "send_group_msg_workflow_" + gid
//	if err := workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
//
//		_, err := s.relationDialogService.BatchCloseOrOpenDialog(ctx, &relationgrpcv1.BatchCloseOrOpenDialogRequest{
//			DialogID: dialogID,
//			Action:   relationgrpcv1.CloseOrOpenDialogType_OPEN,
//			UserIds:  uids.UserIds,
//		})
//		if err != nil {
//			return status.Error(codes.Aborted, err.Error())
//		}
//
//		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
//			_, err := s.relationDialogService.BatchCloseOrOpenDialog(ctx, &relationgrpcv1.BatchCloseOrOpenDialogRequest{
//				DialogID: dialogID,
//				Action:   relationgrpcv1.CloseOrOpenDialogType_CLOSE,
//				UserIds:  uids.UserIds,
//			})
//			return err
//		})
//
//		isAtAll := msggrpcv1.AtAllUserType_NotAtAllUser
//		if req.AtAllUser {
//			isAtAll = msggrpcv1.AtAllUserType_AtAllUser
//		}
//		message, err = s.msgService.SendGroupMessage(ctx, &msggrpcv1.SendGroupMsgRequest{
//			DialogID:  dialogID,
//			GroupID:   uint32(req.GroupID),
//			ID:    userID,
//			Content:   req.Content,
//			Type:      uint32(req.Type),
//			ReplyId:   uint32(req.ReplyId),
//			AtUsers:   req.AtUsers,
//			AtAllUser: isAtAll,
//		})
//		// 发送成功进行消息推送
//		if err != nil {
//			s.logger.Error("发送消息失败", zap.Error(err))
//			return status.Error(codes.Aborted, err.Error())
//		}
//
//		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
//			_, err := s.msgService.SendGroupMessageRevert(wf.Context, &msggrpcv1.MsgIdRequest{MsgId: message.MsgId})
//
//			return err
//		})
//		// 发送成功后添加自己的已读记录
//		data2 := &msggrpcv1.SetGroupMessageReadRequest{
//			MsgId:    message.MsgId,
//			GroupID:  message.GroupID,
//			DialogID: uint32(req.DialogID),
//			ID:   userID,
//			ReadAt:   pkgtime.Now(),
//		}
//
//		var list []*msggrpcv1.SetGroupMessageReadRequest
//		list = append(list, data2)
//		_, err = s.msgGroupService.SetGroupMessageRead(context.Background(), &msggrpcv1.SetGroupMessagesReadRequest{
//			List: list,
//		})
//		if err != nil {
//			return status.Error(codes.Aborted, err.Error())
//		}
//		return err
//	}); err != nil {
//		return nil, err
//	}
//
//	if err := workflow.Execute(wfName, gid, nil); err != nil {
//		return nil, code.MsgErrInsertGroupMessageFailed
//	}
//
//	//查询发送者信息
//	info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
//		ID: userID,
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	resp := &v1.SendGroupMsgResponse{
//		MsgId: int(message.MsgId),
//	}
//
//	if req.ReplyId != 0 {
//		msg, err := s.msgService.GetGroupMessageById(ctx, &msggrpcv1.GetGroupMsgByIDRequest{
//			MsgId: uint32(req.ReplyId),
//		})
//		if err != nil {
//			return nil, err
//		}
//
//		userInfo, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
//			ID: msg.ID,
//		})
//		if err != nil {
//			return nil, err
//		}
//
//		resp.ReplyMsg = &v1.Message{
//			MsgType:  int(msg.Type),
//			Content:  msg.Content,
//			SenderId: msg.ID,
//			SendAt:   int(msg.GetCreatedAt()),
//			MsgId:    int(msg.ID),
//			SenderInfo: &v1.SenderInfo{
//				ID: userInfo.ID,
//				Name:   userInfo.NickName,
//				Avatar: userInfo.Avatar,
//			},
//			ReplyId: int(msg.ReplyId),
//		}
//		if msg.IsLabel != 0 {
//			resp.ReplyMsg.IsLabel = true
//		}
//	}
//
//	rmsg := &pushv1.MessageInfo{}
//	if resp.ReplyMsg != nil {
//		rmsg = &pushv1.MessageInfo{
//			GroupID:  uint32(resp.ReplyMsg.GroupID),
//			MsgType:  uint32(resp.ReplyMsg.MsgType),
//			Content:  resp.ReplyMsg.Content,
//			SenderId: resp.ReplyMsg.SenderId,
//			SendAt:   int64(resp.ReplyMsg.SendAt),
//			MsgId:    uint64(resp.ReplyMsg.MsgId),
//			SenderInfo: &pushv1.SenderInfo{
//				ID: resp.ReplyMsg.SenderInfo.ID,
//				Avatar: resp.ReplyMsg.SenderInfo.Avatar,
//				Name:   resp.ReplyMsg.SenderInfo.Name,
//			},
//			RecipientInfo: &pushv1.SenderInfo{
//				ID: resp.ReplyMsg.RecipientInfo.ID,
//				Avatar: resp.ReplyMsg.RecipientInfo.Avatar,
//				Name:   resp.ReplyMsg.RecipientInfo.Name,
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
//	s.sendWsGroupMsg(ctx, uids.UserIds, driverId, &pushv1.SendWsGroupMsg{
//		MsgId:      message.MsgId,
//		GroupID:    int64(req.GroupID),
//		SenderId:   userID,
//		Content:    req.Content,
//		MsgType:    uint32(req.Type),
//		ReplyId:    uint32(req.ReplyId),
//		SendAt:     pkgtime.Now(),
//		DialogID:   uint32(req.DialogID),
//		AtUsers:    req.AtUsers,
//		AtAllUsers: req.AtAllUser,
//		SenderInfo: &pushv1.SenderInfo{
//			Avatar: info.Avatar,
//			Name:   info.NickName,
//			ID: userID,
//		},
//		ReplyMsg: rmsg,
//	})
//
//	return resp, nil
//}
//
//func (s *ServiceImpl) EditGroupMsg(ctx context.Context, userID string, driverId string, msgID uint32, content string) (interface{}, error) {
//	//获取消息
//	msginfo, err := s.msgService.GetGroupMessageById(ctx, &msggrpcv1.GetGroupMsgByIDRequest{
//		MsgId: msgID,
//	})
//	if err != nil {
//		s.logger.Error("获取消息失败", zap.Error(err))
//		return nil, err
//	}
//	if msginfo.ID != userID {
//		return nil, code.Unauthorized
//	}
//
//	//判断是否在对话内
//	userIds, err := s.relationDialogService.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
//		DialogID: msginfo.DialogID,
//	})
//	if err != nil {
//		s.logger.Error("获取用户对话信息失败", zap.Error(err))
//		return nil, err
//	}
//	var found bool
//	for _, user := range userIds.UserIds {
//		if user == userID {
//			found = true
//			break
//		}
//	}
//	if !found {
//		return nil, code.DialogErrGetDialogByIdFailed
//	}
//
//	// 调用相应的 gRPC 客户端方法来编辑群消息
//	_, err = s.msgService.EditGroupMessage(ctx, &msggrpcv1.EditGroupMsgRequest{
//		GroupMessage: &msggrpcv1.GroupMessage{
//			ID:      msgID,
//			Content: content,
//		},
//	})
//	if err != nil {
//		s.logger.Error("编辑群消息失败", zap.Error(err))
//		return nil, err
//	}
//
//	msginfo.Content = content
//	s.SendMsgToUsers(userIds.UserIds, driverId, pushv1.WSEventType_EditMsgEvent, msginfo, true)
//
//	return msgID, nil
//}
//
//func (s *ServiceImpl) RecallGroupMsg(ctx context.Context, userID string, driverId string, msgID uint32) (interface{}, error) {
//	//获取消息
//	msginfo, err := s.msgService.GetGroupMessageById(ctx, &msggrpcv1.GetGroupMsgByIDRequest{
//		MsgId: msgID,
//	})
//	if err != nil {
//		s.logger.Error("获取消息失败", zap.Error(err))
//		return nil, err
//	}
//
//	if isPromptMessageType(msginfo.Type) {
//		return nil, code.MsgErrDeleteGroupMessageFailed
//	}
//
//	if msginfo.ID != userID {
//		return nil, code.Unauthorized
//	}
//
//	//判断发送时间是否超过两分钟
//	if pkgtime.IsTimeDifferenceGreaterThanTwoMinutes(pkgtime.Now(), msginfo.CreatedAt) {
//		return nil, code.MsgErrTimeoutExceededCannotRevoke
//	}
//
//	//判断是否在对话内
//	userIds, err := s.relationDialogService.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
//		DialogID: msginfo.DialogID,
//	})
//	if err != nil {
//		s.logger.Error("获取用户对话信息失败", zap.Error(err))
//		return nil, err
//	}
//	var found bool
//	for _, user := range userIds.UserIds {
//		if user == userID {
//			found = true
//			break
//		}
//	}
//	if !found {
//		return nil, code.DialogErrGetDialogByIdFailed
//	}
//
//	msg2 := &v1.SendGroupMsgRequest{
//		DialogID: int(msginfo.DialogID),
//		GroupID:  int(msginfo.GroupID),
//		Content:  msginfo.Content,
//		ReplyId:  int(msginfo.ID),
//		Type:     int(msggrpcv1.MessageType_Delete),
//	}
//	_, err = s.SendGroupMsg(ctx, userID, driverId, msg2)
//	if err != nil {
//		return nil, err
//	}
//
//	// 调用相应的 gRPC 客户端方法来撤回群消息
//	msg, err := s.msgService.DeleteGroupMessage(ctx, &msggrpcv1.DeleteGroupMsgRequest{
//		MsgId: msgID,
//	})
//	if err != nil {
//		s.logger.Error("撤回群消息失败", zap.Error(err))
//		return nil, err
//	}
//
//	return msg.ID, nil
//}
//
//func (s *ServiceImpl) LabelGroupMessage(ctx context.Context, userID string, driverId string, msgID uint32, label bool) (interface{}, error) {
//	// 获取群聊消息
//	msginfo, err := s.msgService.GetGroupMessageById(ctx, &msggrpcv1.GetGroupMsgByIDRequest{
//		MsgId: msgID,
//	})
//	if err != nil {
//		s.logger.Error("获取群聊消息失败", zap.Error(err))
//		return nil, err
//	}
//
//	if isPromptMessageType(msginfo.Type) {
//		return nil, code.SetMsgErrSetGroupMsgLabelFailed
//	}
//
//	//判断是否在对话内
//	userIds, err := s.relationDialogService.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
//		DialogID: msginfo.DialogID,
//	})
//	if err != nil {
//		s.logger.Error("获取对话用户失败", zap.Error(err))
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
//	if !found {
//		return nil, code.RelationGroupErrNotInGroup
//	}
//
//	isLabel := msggrpcv1.MsgLabel_NotLabel
//	if label {
//		isLabel = msggrpcv1.MsgLabel_IsLabel
//	}
//	// 调用 gRPC 客户端方法将群聊消息设置为标注状态
//	_, err = s.msgService.SetGroupMsgLabel(ctx, &msggrpcv1.SetGroupMsgLabelRequest{
//		MsgId:   msgID,
//		IsLabel: isLabel,
//	})
//	if err != nil {
//		s.logger.Error("设置群聊消息标注失败", zap.Error(err))
//		return nil, err
//	}
//
//	msginfo.IsLabel = isLabel
//	msg2 := &v1.SendGroupMsgRequest{
//		DialogID: int(msginfo.DialogID),
//		GroupID:  int(msginfo.GroupID),
//		Content:  msginfo.Content,
//		ReplyId:  int(msginfo.ID),
//		Type:     int(msggrpcv1.MessageType_Label),
//	}
//
//	if !label {
//		msg2.Type = int(msggrpcv1.MessageType_CancelLabel)
//	}
//
//	_, err = s.SendGroupMsg(ctx, userID, driverId, msg2)
//	if err != nil {
//		return nil, err
//	}
//	//s.SendMsgToUsers(userIds.UserIds, driverId, constants.LabelMsgEvent, msginfo, true)
//	return nil, nil
//}
//
//func (s *ServiceImpl) GetGroupLabelMsgList(ctx context.Context, userID string, dialogId uint32) (interface{}, error) {
//	_, err := s.relationDialogService.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
//		ID:   userID,
//		DialogID: dialogId,
//	})
//	if err != nil {
//		s.logger.Error("获取对话用户失败", zap.Error(err))
//		return nil, err
//	}
//
//	msgs, err := s.msgService.GetGroupMsgLabelByDialogId(ctx, &msggrpcv1.GetGroupMsgLabelByDialogIdRequest{
//		DialogID: dialogId,
//	})
//	if err != nil {
//		s.logger.Error("获取群聊消息标注失败", zap.Error(err))
//		return nil, err
//	}
//
//	return msgs, nil
//}
//
//func (s *ServiceImpl) GetGroupMessageList(c *gin.Context, id string, request v1.GetGroupMsgListParams) (interface{}, error) {
//	//查询对话信息
//	byId, err := s.relationDialogService.GetDialogById(c, &relationgrpcv1.GetDialogByIdRequest{
//		DialogID: uint32(request.DialogID),
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	if byId.GroupID == 0 {
//		return nil, code.MsgErrGetGroupMsgListFailed
//	}
//
//	_, err = s.groupService.GetGroupInfoByGid(c, &groupApi.GetGroupInfoRequest{
//		Gid: byId.GroupID,
//	})
//	if err != nil {
//		s.logger.Error("获取群聊信息失败", zap.Error(err))
//		return nil, err
//	}
//
//	_, err = s.relationGroupService.GetGroupRelation(c, &relationgrpcv1.GetGroupRelationRequest{
//		GroupID: byId.GroupID,
//		ID:  id,
//	})
//	if err != nil {
//		s.logger.Error("获取群聊关系失败", zap.Error(err))
//		return nil, err
//	}
//
//	msg, err := s.msgService.GetGroupMessageList(c, &msggrpcv1.GetGroupMsgListRequest{
//		DialogID: uint32(request.DialogID),
//		ID:   request.ID,
//		Content:  request.Content,
//		Type:     int32(request.Type),
//		PageNum:  int32(request.PageNum),
//		PageSize: int32(request.PageSize),
//		MsgId:    uint64(request.MsgId),
//	})
//	if err != nil {
//		s.logger.Error("获取群聊消息列表失败", zap.Error(err))
//		return nil, err
//	}
//
//	resp := &v1.GetGroupMsgListResponse{}
//	resp.CurrentPage = int(msg.CurrentPage)
//	resp.Total = int(msg.Total)
//
//	msgList := make([]v1.GroupMessage, 0)
//	for _, v := range msg.GroupMessages {
//		ReadAt := 0
//		isRead := 0
//		//查询是否已读
//		msgRead, err := s.msgGroupService.GetGroupMessageReadByMsgIdAndUserId(c, &msggrpcv1.GetGroupMessageReadByMsgIdAndUserIdRequest{
//			MsgId:  v.ID,
//			ID: request.ID,
//		})
//		if err != nil {
//			s.logger.Error("获取群聊消息是否已读失败", zap.Error(err))
//		}
//		if msgRead != nil {
//			ReadAt = int(msgRead.ReadAt)
//			isRead = 1
//		}
//
//		//查询信息
//		info, err := s.userService.UserInfo(c, &usergrpcv1.UserInfoRequest{
//			ID: v.ID,
//		})
//		if err != nil {
//			return nil, err
//		}
//
//		sendRelation, err := s.relationGroupService.GetGroupRelation(c, &relationgrpcv1.GetGroupRelationRequest{
//			GroupID: byId.GroupID,
//			ID:  v.ID,
//		})
//		if err != nil {
//			s.logger.Error("获取群聊关系失败", zap.Error(err))
//			return nil, err
//		}
//
//		name := info.NickName
//		if sendRelation != nil && sendRelation.Remark != "" {
//			name = sendRelation.Remark
//		}
//
//		isLabel := false
//		if v.IsLabel != msggrpcv1.MsgLabel_NotLabel {
//			isLabel = true
//		}
//		isReadFlag := false
//		if isRead == int(msggrpcv1.ReadType_IsRead) {
//			isReadFlag = true
//		}
//		isAtAll := false
//		if v.AtAllUser == msggrpcv1.AtAllUserType_NotAtAllUser {
//			isAtAll = true
//		}
//		msgList = append(msgList, v1.GroupMessage{
//			MsgId:     int(v.ID),
//			Content:   v.Content,
//			GroupID:   int(v.GroupID),
//			Type:      int(v.Type),
//			SendAt:    int(v.CreatedAt),
//			DialogID:  int(v.DialogID),
//			IsLabel:   isLabel,
//			ReadCount: int(v.ReadCount),
//			ReplyId:   int(v.ReplyId),
//			ID:    v.ID,
//			AtUsers:   v.AtUsers,
//			ReadAt:    ReadAt,
//			IsRead:    isReadFlag,
//			AtAllUser: isAtAll,
//			SenderInfo: &v1.SenderInfo{
//				Name:   name,
//				ID: info.ID,
//				Avatar: info.Avatar,
//			},
//		})
//	}
//	resp.GroupMessages = &msgList
//
//	return resp, nil
//}
//
//func (s *ServiceImpl) SetGroupMessagesRead(c context.Context, uid string, driverId string, request *v1.GroupMessageReadRequest) (interface{}, error) {
//	dialog, err := s.relationDialogService.GetDialogById(c, &relationgrpcv1.GetDialogByIdRequest{
//		DialogID: uint32(request.DialogID),
//	})
//	if err != nil {
//		s.logger.Error("获取对话失败", zap.Error(err))
//		return nil, err
//	}
//
//	info, err := s.userService.UserInfo(c, &usergrpcv1.UserInfoRequest{
//		ID: uid,
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	if dialog.Type != uint32(relationgrpcv1.DialogType_GROUP_DIALOG) && dialog.GroupID == 0 {
//		return nil, code.DialogErrTypeNotSupport
//	}
//
//	_, err = s.groupService.GetGroupInfoByGid(c, &groupApi.GetGroupInfoRequest{
//		Gid: dialog.GroupID,
//	})
//	if err != nil {
//		s.logger.Error("获取群聊信息失败", zap.Error(err))
//		return nil, err
//	}
//
//	_, err = s.relationGroupService.GetGroupRelation(c, &relationgrpcv1.GetGroupRelationRequest{
//		GroupID: dialog.GroupID,
//		ID:  uid,
//	})
//	if err != nil {
//		s.logger.Error("获取群聊关系失败", zap.Error(err))
//		return nil, err
//	}
//
//	_, err = s.relationDialogService.GetDialogUserByDialogIDAndUserID(c, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
//		ID:   uid,
//		DialogID: uint32(request.DialogID),
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	if request.ReadAll {
//		resp1, err := s.msgService.ReadAllGroupMsg(c, &msggrpcv1.ReadAllGroupMsgRequest{
//			DialogID: uint32(request.DialogID),
//			ID:   uid,
//		})
//		if err != nil {
//			s.logger.Error("设置群聊消息已读失败", zap.Error(err))
//			return nil, err
//		}
//
//		//给消息发送者推送谁读了消息
//		for _, v := range resp1.UnreadGroupMessage {
//			if v.ID != uid {
//				data := map[string]interface{}{"msg_id": v.MsgId, "operator_info": &v1.SenderInfo{
//					Name:   info.NickName,
//					ID: info.ID,
//					Avatar: info.Avatar,
//				}}
//				bytes, err := utils.StructToBytes(data)
//				if err != nil {
//					s.logger.Error("push err:", zap.Error(err))
//					continue
//				}
//				msg := &pushv1.WsMsg{Uid: v.ID, Event: pushv1.WSEventType_GroupMsgReadEvent, Data: &any.Any{Value: bytes}}
//				bytes2, err := utils.StructToBytes(msg)
//				if err != nil {
//					s.logger.Error("push err:", zap.Error(err))
//					continue
//				}
//				_, err = s.pushService.Push(c, &pushv1.PushRequest{
//					Type: pushv1.Type_Ws,
//					Data: bytes2,
//				})
//				if err != nil {
//					s.logger.Error("push err:", zap.Error(err))
//					continue
//				}
//			}
//		}
//
//		return nil, nil
//	}
//
//	msgList := make([]*msggrpcv1.SetGroupMessageReadRequest, 0)
//	for _, v := range request.MsgIds {
//		userId, _ := s.msgGroupService.GetGroupMessageReadByMsgIdAndUserId(c, &msggrpcv1.GetGroupMessageReadByMsgIdAndUserIdRequest{
//			MsgId:  uint32(v),
//			ID: uid,
//		})
//		if userId != nil {
//			continue
//		}
//		msgList = append(msgList, &msggrpcv1.SetGroupMessageReadRequest{
//			MsgId:    uint32(v),
//			GroupID:  dialog.GroupID,
//			DialogID: uint32(request.DialogID),
//			ID:   uid,
//			ReadAt:   pkgtime.Now(),
//		})
//	}
//	if len(msgList) == 0 {
//		return nil, nil
//	}
//
//	_, err = s.msgGroupService.SetGroupMessageRead(c, &msggrpcv1.SetGroupMessagesReadRequest{
//		List: msgList,
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	ids := make([]uint32, 0)
//	for _, v := range request.MsgIds {
//		ids = append(ids, uint32(v))
//	}
//	msgs, err := s.msgService.GetGroupMessagesByIds(c, &msggrpcv1.GetGroupMessagesByIdsRequest{
//		MsgIds:  ids,
//		GroupID: dialog.GroupID,
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	//给消息发送者推送谁读了消息
//	for _, message := range msgs.GroupMessages {
//		if message.ID != uid {
//			s.SendMsg(message.ID, driverId, pushv1.WSEventType_GroupMsgReadEvent, map[string]interface{}{"msg_id": message.ID, "operator_info": &v1.SenderInfo{
//				Name:   info.NickName,
//				ID: info.ID,
//				Avatar: info.Avatar,
//			}}, false)
//		}
//	}
//
//	return nil, nil
//}
//
//func (s *ServiceImpl) GetGroupMessageReadersResponse(c context.Context, userId string, msgId uint32, dialogId uint32, groupId uint32) (interface{}, error) {
//	_, err := s.groupService.GetGroupInfoByGid(c, &groupApi.GetGroupInfoRequest{
//		Gid: groupId,
//	})
//	if err != nil {
//		s.logger.Error("获取群聊信息失败", zap.Error(err))
//		return nil, err
//	}
//
//	_, err = s.relationGroupService.GetGroupRelation(c, &relationgrpcv1.GetGroupRelationRequest{
//		GroupID: groupId,
//		ID:  userId,
//	})
//	if err != nil {
//		s.logger.Error("获取群聊关系失败", zap.Error(err))
//		return nil, err
//	}
//
//	_, err = s.relationDialogService.GetDialogUserByDialogIDAndUserID(c, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
//		ID:   userId,
//		DialogID: dialogId,
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	us, err := s.msgGroupService.GetGroupMessageReaders(c, &msggrpcv1.GetGroupMessageReadersRequest{
//		MsgId:    msgId,
//		GroupID:  groupId,
//		DialogID: dialogId,
//	})
//	if err != nil {
//		return nil, err
//	}
//
//	info, err := s.userService.GetBatchUserInfo(c, &usergrpcv1.GetBatchUserInfoRequest{
//		UserIds: us.UserIds,
//	})
//	if err != nil {
//		return nil, err
//	}
//	response := make([]v1.GetGroupMessageReadersResponse, 0)
//
//	if len(us.UserIds) == len(info.Users) {
//		for _, v := range us.UserIds {
//			for _, v6 := range info.Users {
//				if v == v6.ID {
//					response = append(response, v1.GetGroupMessageReadersResponse{
//						ID: v6.ID,
//						Avatar: v6.Avatar,
//						Name:   v6.NickName,
//					})
//				}
//			}
//		}
//	}
//
//	return response, nil
//}
