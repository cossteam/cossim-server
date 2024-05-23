package msg

import (
	"context"
	"fmt"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	v1 "github.com/cossim/coss-server/internal/msg/api/http/v1"
	"github.com/cossim/coss-server/internal/msg/domain/entity"
	pushv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils"
	pkgtime "github.com/cossim/coss-server/pkg/utils/time"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/workflow"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/lithammer/shortuuid/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GroupService interface {
	SendGroupMsg(ctx context.Context, userID string, driverId string, req *v1.SendGroupMsgRequest) (*v1.SendGroupMsgResponse, error)
	EditGroupMsg(ctx context.Context, userID string, driverId string, msgID uint32, content string) (interface{}, error)
	LabelGroupMessage(ctx context.Context, userID string, driverId string, msgID uint32, label bool) (interface{}, error)
	RecallGroupMsg(ctx context.Context, userID string, driverId string, msgID uint32) (interface{}, error)
	GetGroupLabelMsgList(ctx context.Context, userID string, dialogId uint32) (*v1.GetGroupLabelMsgListResponse, error)
	GetGroupMessageList(c context.Context, id string, request *v1.GetGroupMsgListParams) (*v1.GetGroupMsgListResponse, error)
	SetGroupMessagesRead(c context.Context, uid string, driverId string, req *v1.GroupMessageReadRequest) error
	GetGroupMessageReadersResponse(c context.Context, userId string, msgId uint32, dialogId uint32, groupId uint32) (interface{}, error)
}

func (s *ServiceImpl) sendWsGroupMsg(ctx context.Context, uIds []string, driverId string, msg *pushv1.SendWsGroupMsg) {
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

func (s *ServiceImpl) SendGroupMsg(ctx context.Context, userID string, driverId string, req *v1.SendGroupMsgRequest) (*v1.SendGroupMsgResponse, error) {
	group, err := s.groupService.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{
		Gid: uint32(req.GroupId),
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

	dialogID := uint32(req.DialogId)

	groupRelation, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		GroupId: uint32(req.GroupId),
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
		DialogIds: []uint32{dialogID},
	})
	if err != nil {
		s.logger.Error("获取会话失败", zap.Error(err))
		return nil, err
	}
	if len(dialogs.Dialogs) == 0 {
		return nil, code.DialogErrGetDialogUserByDialogIDAndUserIDFailed
	}

	_, err = s.relationDialogService.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		DialogId: dialogID,
		UserId:   userID,
	})
	if err != nil {
		s.logger.Error("获取用户会话失败", zap.Error(err))
		return nil, code.DialogErrGetDialogUserByDialogIDAndUserIDFailed
	}

	//查询群聊所有用户id
	uids, err := s.relationGroupService.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{
		GroupId: uint32(req.GroupId),
	})

	var msgID uint32
	var groupID uint32
	workflow.InitGrpc(s.dtmGrpcServer, "", grpc.NewServer())
	gid := shortuuid.New()
	wfName := "send_group_msg_workflow_" + gid
	if err := workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {

		_, err := s.relationDialogService.BatchCloseOrOpenDialog(ctx, &relationgrpcv1.BatchCloseOrOpenDialogRequest{
			DialogId: dialogID,
			Action:   relationgrpcv1.CloseOrOpenDialogType_OPEN,
			UserIds:  uids.UserIds,
		})
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
		}

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := s.relationDialogService.BatchCloseOrOpenDialog(ctx, &relationgrpcv1.BatchCloseOrOpenDialogRequest{
				DialogId: dialogID,
				Action:   relationgrpcv1.CloseOrOpenDialogType_CLOSE,
				UserIds:  uids.UserIds,
			})
			return err
		})

		isAtAll := entity.NotAtAllUser
		if req.AtAllUser {
			isAtAll = entity.AtAllUser
		}
		mg, err := s.gmd.SendGroupMessage(ctx, &entity.GroupMessage{
			DialogID:  uint(dialogID),
			GroupID:   uint(req.GroupId),
			UserID:    userID,
			Content:   req.Content,
			Type:      entity.UserMessageType(req.Type),
			ReplyId:   uint(req.ReplyId),
			AtUsers:   req.AtUsers,
			AtAllUser: entity.AtAllUserType(isAtAll),
		})
		// 发送成功进行消息推送
		if err != nil {
			s.logger.Error("发送消息失败", zap.Error(err))
			return status.Error(codes.Aborted, err.Error())
		}
		fmt.Println("发送消息成功", mg.ID)

		msgID = uint32(mg.ID)
		groupID = uint32(mg.GroupID)
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			err := s.gmd.SendGroupMessageRevert(wf.Context, mg.ID)

			return err
		})
		// 发送成功后添加自己的已读记录
		data2 := &entity.GroupMessageRead{
			MsgID:    uint(msgID),
			GroupID:  uint(groupID),
			DialogID: uint(req.DialogId),
			UserID:   userID,
			ReadAt:   pkgtime.Now(),
		}

		var list []*entity.GroupMessageRead
		list = append(list, data2)
		err = s.gmrd.SetGroupMessageRead(context.Background(), list)
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
		}
		return err
	}); err != nil {
		return nil, err
	}

	if err := workflow.Execute(wfName, gid, nil); err != nil {
		return nil, code.MsgErrInsertGroupMessageFailed
	}

	//查询发送者信息
	info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		return nil, err
	}

	resp := &v1.SendGroupMsgResponse{
		MsgId: int(msgID),
	}

	if req.ReplyId != 0 {
		msg, err := s.gmd.GetGroupMessageById(ctx, uint(req.ReplyId))
		if err != nil {
			return nil, err
		}

		userInfo, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
			UserId: msg.UserID,
		})
		if err != nil {
			return nil, err
		}

		resp.ReplyMsg = &v1.Message{
			MsgType:  int(msg.Type),
			Content:  msg.Content,
			SenderId: msg.UserID,
			SendAt:   int(msg.CreatedAt),
			MsgId:    int(msg.ID),
			SenderInfo: &v1.SenderInfo{
				UserId: userInfo.UserId,
				Name:   userInfo.NickName,
				Avatar: userInfo.Avatar,
			},
			ReplyId: int(msg.ReplyId),
		}
		if msg.IsLabel != 0 {
			resp.ReplyMsg.IsLabel = true
		}
	}

	rmsg := &pushv1.MessageInfo{}
	if resp.ReplyMsg != nil {
		rmsg = &pushv1.MessageInfo{
			GroupId:  uint32(resp.ReplyMsg.GroupId),
			MsgType:  uint32(resp.ReplyMsg.MsgType),
			Content:  resp.ReplyMsg.Content,
			SenderId: resp.ReplyMsg.SenderId,
			SendAt:   int64(resp.ReplyMsg.SendAt),
			MsgId:    uint64(resp.ReplyMsg.MsgId),
			SenderInfo: &pushv1.SenderInfo{
				UserId: resp.ReplyMsg.SenderInfo.UserId,
				Avatar: resp.ReplyMsg.SenderInfo.Avatar,
				Name:   resp.ReplyMsg.SenderInfo.Name,
			},
			AtAllUser:          resp.ReplyMsg.AtAllUser,
			AtUsers:            resp.ReplyMsg.AtUsers,
			IsBurnAfterReading: resp.ReplyMsg.IsBurnAfterReading,
			IsLabel:            resp.ReplyMsg.IsLabel,
			ReplyId:            uint32(resp.ReplyMsg.ReplyId),
			IsRead:             resp.ReplyMsg.IsRead,
			ReadAt:             int64(resp.ReplyMsg.ReadAt),
		}
	}
	s.sendWsGroupMsg(ctx, uids.UserIds, driverId, &pushv1.SendWsGroupMsg{
		MsgId:      msgID,
		GroupId:    int64(req.GroupId),
		SenderId:   userID,
		Content:    req.Content,
		MsgType:    uint32(req.Type),
		ReplyId:    uint32(req.ReplyId),
		SendAt:     pkgtime.Now(),
		DialogId:   uint32(req.DialogId),
		AtUsers:    req.AtUsers,
		AtAllUsers: req.AtAllUser,
		SenderInfo: &pushv1.SenderInfo{
			Avatar: info.Avatar,
			Name:   info.NickName,
			UserId: userID,
		},
		ReplyMsg: rmsg,
	})

	return resp, nil
}

func (s *ServiceImpl) EditGroupMsg(ctx context.Context, userID string, driverId string, msgID uint32, content string) (interface{}, error) {
	//获取消息
	msginfo, err := s.gmd.GetGroupMessageById(ctx, uint(msgID))
	if err != nil {
		s.logger.Error("获取消息失败", zap.Error(err))
		return nil, err
	}
	if msginfo.UserID != userID {
		return nil, code.Unauthorized
	}

	//判断是否在对话内
	userIds, err := s.relationDialogService.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
		DialogId: uint32(msginfo.DialogID),
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
	//err = s.gmd.EditGroupMessage(ctx, &msggrpcv1.EditGroupMsgRequest{
	//	GroupMessage: &msggrpcv1.GroupMessage{
	//		ID:      msgID,
	//		Content: content,
	//	},
	//})
	err = s.gmd.EditGroupMessage(ctx, &entity.GroupMessage{BaseModel: entity.BaseModel{ID: uint(msgID)}, Content: content})
	if err != nil {
		s.logger.Error("编辑群消息失败", zap.Error(err))
		return nil, err
	}

	msginfo.Content = content
	s.SendMsgToUsers(userIds.UserIds, driverId, pushv1.WSEventType_EditMsgEvent, msginfo, true)

	return msgID, nil
}

func (s *ServiceImpl) RecallGroupMsg(ctx context.Context, userID string, driverId string, msgID uint32) (interface{}, error) {
	//获取消息
	msginfo, err := s.gmd.GetGroupMessageById(ctx, uint(msgID))
	if err != nil {
		s.logger.Error("获取消息失败", zap.Error(err))
		return nil, err
	}

	if isPromptMessageType(uint32(msginfo.Type)) {
		return nil, code.MsgErrDeleteGroupMessageFailed
	}

	if msginfo.UserID != userID {
		return nil, code.Unauthorized
	}

	//判断发送时间是否超过两分钟
	if pkgtime.IsTimeDifferenceGreaterThanTwoMinutes(pkgtime.Now(), msginfo.CreatedAt) {
		return nil, code.MsgErrTimeoutExceededCannotRevoke
	}

	//判断是否在对话内
	userIds, err := s.relationDialogService.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
		DialogId: uint32(msginfo.DialogID),
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

	msg2 := &v1.SendGroupMsgRequest{
		DialogId: int(msginfo.DialogID),
		GroupId:  int(msginfo.GroupID),
		Content:  msginfo.Content,
		ReplyId:  int(msginfo.ID),
		Type:     int(entity.MessageTypeDelete),
	}
	_, err = s.SendGroupMsg(ctx, userID, driverId, msg2)
	if err != nil {
		return nil, err
	}

	// 调用相应的 gRPC 客户端方法来撤回群消息
	err = s.gmd.DeleteGroupMessage(ctx, uint(msgID), false)
	if err != nil {
		s.logger.Error("撤回群消息失败", zap.Error(err))
		return nil, err
	}

	return nil, nil
}

func (s *ServiceImpl) LabelGroupMessage(ctx context.Context, userID string, driverId string, msgID uint32, label bool) (interface{}, error) {
	// 获取群聊消息
	msginfo, err := s.gmd.GetGroupMessageById(ctx, uint(msgID))
	if err != nil {
		s.logger.Error("获取群聊消息失败", zap.Error(err))
		return nil, err
	}

	if isPromptMessageType(uint32(msginfo.Type)) {
		return nil, code.SetMsgErrSetGroupMsgLabelFailed
	}

	//判断是否在对话内
	userIds, err := s.relationDialogService.GetDialogUsersByDialogID(ctx, &relationgrpcv1.GetDialogUsersByDialogIDRequest{
		DialogId: uint32(msginfo.DialogID),
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

	//isLabel := msggrpcv1.MsgLabel_NotLabel
	//if label {
	//	isLabel = msggrpcv1.MsgLabel_IsLabel
	//}
	// 调用 gRPC 客户端方法将群聊消息设置为标注状态
	err = s.gmd.SetGroupMsgLabel(ctx, uint(msgID), label)
	if err != nil {
		s.logger.Error("设置群聊消息标注失败", zap.Error(err))
		return nil, err
	}

	//msginfo.IsLabel = uint(isLabel)
	msg2 := &v1.SendGroupMsgRequest{
		DialogId: int(msginfo.DialogID),
		GroupId:  int(msginfo.GroupID),
		Content:  msginfo.Content,
		ReplyId:  int(msginfo.ID),
		Type:     int(entity.IsLabel),
	}

	if !label {
		msg2.Type = int(entity.MessageTypeCancelLabel)
	}

	fmt.Println("msg2.ReplyId", msg2.ReplyId)
	_, err = s.SendGroupMsg(ctx, userID, driverId, msg2)
	if err != nil {
		return nil, err
	}
	//s.SendMsgToUsers(userIds.UserIds, driverId, constants.LabelMsgEvent, msginfo, true)
	return nil, nil
}

func (s *ServiceImpl) GetGroupLabelMsgList(ctx context.Context, userID string, dialogId uint32) (*v1.GetGroupLabelMsgListResponse, error) {
	_, err := s.relationDialogService.GetDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		UserId:   userID,
		DialogId: dialogId,
	})
	if err != nil {
		s.logger.Error("获取对话用户失败", zap.Error(err))
		return nil, err
	}

	msgs, err := s.gmd.GetGroupMsgLabelByDialogId(ctx, uint(dialogId))
	if err != nil {
		s.logger.Error("获取群聊消息标注失败", zap.Error(err))
		return nil, err
	}

	resp := &v1.GetGroupLabelMsgListResponse{}
	for _, i2 := range msgs {
		//read := false
		//if i2.ReadAt != 0 {
		//	read = true
		//}
		resp.List = append(resp.List, v1.Message{
			DialogId: int(i2.DialogID),
			GroupId:  int(i2.GroupID),
			MsgId:    int(i2.ID),
			Content:  i2.Content,
			MsgType:  int(i2.Type),
			ReplyId:  int(i2.ReplyId),
			SendAt:   int(i2.CreatedAt),
			IsLabel:  i2.IsLabel == uint(entity.IsLabel),
			SenderId: i2.UserID,
		})
	}

	return resp, nil
}

func (s *ServiceImpl) GetGroupMessageList(c context.Context, id string, request *v1.GetGroupMsgListParams) (*v1.GetGroupMsgListResponse, error) {
	//查询对话信息
	byId, err := s.relationDialogService.GetDialogById(c, &relationgrpcv1.GetDialogByIdRequest{
		DialogId: uint32(request.DialogId),
	})
	if err != nil {
		return nil, err
	}

	if byId.GroupId == 0 {
		return nil, code.MsgErrGetGroupMsgListFailed
	}

	_, err = s.groupService.GetGroupInfoByGid(c, &groupgrpcv1.GetGroupInfoRequest{
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

	msg, total, err := s.gmd.GetGroupMessageList(c, &entity.GroupMessage{
		DialogID:  uint(request.DialogId),
		UserID:    *request.UserId,
		Content:   *request.Content,
		Type:      entity.UserMessageType(*request.Type),
		BaseModel: entity.BaseModel{ID: uint(*request.MsgId)},
	}, request.PageSize, request.PageNum)
	if err != nil {
		s.logger.Error("获取群聊消息列表失败", zap.Error(err))
		return nil, err
	}

	resp := &v1.GetGroupMsgListResponse{}
	resp.CurrentPage = request.PageNum
	resp.Total = int(total)

	msgList := make([]v1.GroupMessage, 0)
	for _, v := range msg {
		ReadAt := 0
		isRead := 0
		//查询是否已读
		//msgRead, err := s.gmrd.GetGroupMessageReadByMsgIdAndUserId(c, &msggrpcv1.GetGroupMessageReadByMsgIdAndUserIdRequest{
		//	MsgId:  v.ID,
		//	ID: request.ID,
		//})
		msgRead, err := s.gmrd.GetGroupMessageReadByMsgIdAndUserId(c, v.ID, *request.UserId)
		if err != nil {
			s.logger.Error("获取群聊消息是否已读失败", zap.Error(err))
		}
		if msgRead != nil {
			ReadAt = int(msgRead.ReadAt)
			isRead = 1
		}

		//查询信息
		info, err := s.userService.UserInfo(c, &usergrpcv1.UserInfoRequest{
			UserId: v.UserID,
		})
		if err != nil {
			return nil, err
		}

		sendRelation, err := s.relationGroupService.GetGroupRelation(c, &relationgrpcv1.GetGroupRelationRequest{
			GroupId: byId.GroupId,
			UserId:  v.UserID,
		})
		if err != nil {
			s.logger.Error("获取群聊关系失败", zap.Error(err))
			return nil, err
		}

		name := info.NickName
		if sendRelation != nil && sendRelation.Remark != "" {
			name = sendRelation.Remark
		}

		isLabel := false
		if v.IsLabel != uint(entity.NotLabel) {
			isLabel = true
		}
		isReadFlag := false
		if isRead == int(entity.IsRead) {
			isReadFlag = true
		}
		isAtAll := false
		if v.AtAllUser == entity.AtAllUserType(entity.NotAtAllUser) {
			isAtAll = true
		}
		msgList = append(msgList, v1.GroupMessage{
			MsgId:     int(v.ID),
			Content:   v.Content,
			GroupId:   int(v.GroupID),
			Type:      int(v.Type),
			SendAt:    int(v.CreatedAt),
			DialogId:  int(v.DialogID),
			IsLabel:   isLabel,
			ReadCount: v.ReadCount,
			ReplyId:   int(v.ReplyId),
			UserId:    v.UserID,
			AtUsers:   v.AtUsers,
			ReadAt:    ReadAt,
			IsRead:    isReadFlag,
			AtAllUser: isAtAll,
			SenderInfo: &v1.SenderInfo{
				Name:   name,
				UserId: info.UserId,
				Avatar: info.Avatar,
			},
		})
	}
	resp.GroupMessages = &msgList

	return resp, nil
}

func (s *ServiceImpl) SetGroupMessagesRead(c context.Context, uid string, driverId string, req *v1.GroupMessageReadRequest) error {
	dialog, err := s.relationDialogService.GetDialogById(c, &relationgrpcv1.GetDialogByIdRequest{
		DialogId: uint32(req.DialogId),
	})
	if err != nil {
		s.logger.Error("获取对话失败", zap.Error(err))
		return err
	}

	info, err := s.userService.UserInfo(c, &usergrpcv1.UserInfoRequest{
		UserId: uid,
	})
	if err != nil {
		return err
	}

	if dialog.Type != uint32(relationgrpcv1.DialogType_GROUP_DIALOG) && dialog.GroupId == 0 {
		return code.DialogErrTypeNotSupport
	}

	_, err = s.groupService.GetGroupInfoByGid(c, &groupgrpcv1.GetGroupInfoRequest{
		Gid: dialog.GroupId,
	})
	if err != nil {
		s.logger.Error("获取群聊信息失败", zap.Error(err))
		return err
	}

	_, err = s.relationGroupService.GetGroupRelation(c, &relationgrpcv1.GetGroupRelationRequest{
		GroupId: dialog.GroupId,
		UserId:  uid,
	})
	if err != nil {
		s.logger.Error("获取群聊关系失败", zap.Error(err))
		return err
	}

	_, err = s.relationDialogService.GetDialogUserByDialogIDAndUserID(c, &relationgrpcv1.GetDialogUserByDialogIDAndUserIdRequest{
		UserId:   uid,
		DialogId: uint32(req.DialogId),
	})
	if err != nil {
		return err
	}

	if req.ReadAll {
		//resp1, err := s.gmrd.ReadAllGroupMsg(c, &msggrpcv1.ReadAllGroupMsgRequest{
		//	DialogID: uint32(request.DialogID),
		//	ID:   uid,
		//})
		resp1, err := s.gmrd.ReadAllGroupMsg(c, uint(req.DialogId), uid)
		if err != nil {
			s.logger.Error("设置群聊消息已读失败", zap.Error(err))
			return err
		}

		//给消息发送者推送谁读了消息
		for _, v := range resp1 {
			if v.UserID != uid {
				data := map[string]interface{}{"msg_id": v.MsgID, "operator_info": &v1.SenderInfo{
					Name:   info.NickName,
					UserId: info.UserId,
					Avatar: info.Avatar,
				}}
				bytes, err := utils.StructToBytes(data)
				if err != nil {
					s.logger.Error("push err:", zap.Error(err))
					continue
				}
				msg := &pushv1.WsMsg{Uid: v.UserID, Event: pushv1.WSEventType_GroupMsgReadEvent, Data: &any.Any{Value: bytes}}
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

		return nil
	}

	msgList := make([]*entity.GroupMessageRead, 0)
	for _, v := range req.MsgIds {
		//userId, _ := s.gmrd.GetGroupMessageReadByMsgIdAndUserId(c, &msggrpcv1.GetGroupMessageReadByMsgIdAndUserIdRequest{
		//	MsgId:  uint32(v),
		//	ID: uid,
		//})
		userId, _ := s.gmrd.GetGroupMessageReadByMsgIdAndUserId(c, uint(v), uid)
		if userId != nil {
			continue
		}
		msgList = append(msgList, &entity.GroupMessageRead{
			MsgID:    uint(v),
			GroupID:  uint(dialog.GroupId),
			DialogID: uint(req.DialogId),
			UserID:   uid,
			ReadAt:   pkgtime.Now(),
		})
	}
	if len(msgList) == 0 {
		return nil
	}

	//_, err = s.gmrd.SetGroupMessageRead(c, &msggrpcv1.SetGroupMessagesReadRequest{
	//	List: msgList,
	//})
	err = s.gmrd.SetGroupMessageRead(c, msgList)
	if err != nil {
		return err
	}

	ids := make([]uint, 0)
	for _, v := range req.MsgIds {
		ids = append(ids, uint(v))
	}
	//msgs, err := s.gmd.GetGroupMessagesByIds(c, &msggrpcv1.GetGroupMessagesByIdsRequest{
	//	MsgIds:  ids,
	//	GroupID: dialog.GroupID,
	//})
	msgs, err := s.gmd.GetGroupMessagesByIds(c, ids)
	if err != nil {
		return err
	}

	//给消息发送者推送谁读了消息
	for _, message := range msgs {
		if message.UserID != uid {
			s.SendMsg(message.UserID, driverId, pushv1.WSEventType_GroupMsgReadEvent, map[string]interface{}{"msg_id": message.ID, "operator_info": &v1.SenderInfo{
				Name:   info.NickName,
				UserId: info.UserId,
				Avatar: info.Avatar,
			}}, false)
		}
	}

	return nil
}

func (s *ServiceImpl) GetGroupMessageReadersResponse(c context.Context, userId string, msgId uint32, dialogId uint32, groupId uint32) (interface{}, error) {
	_, err := s.groupService.GetGroupInfoByGid(c, &groupgrpcv1.GetGroupInfoRequest{
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

	//us, err := s.gmrd.GetGroupMessageReaders(c, &msggrpcv1.GetGroupMessageReadersRequest{
	//	MsgId:    msgId,
	//	GroupID:  groupId,
	//	DialogID: dialogId,
	//})
	us, err := s.gmrd.GetGroupMessageReaders(c, uint(msgId), uint(groupId), uint(dialogId))
	if err != nil {
		return nil, err
	}

	info, err := s.userService.GetBatchUserInfo(c, &usergrpcv1.GetBatchUserInfoRequest{
		UserIds: us,
	})
	if err != nil {
		return nil, err
	}
	response := make([]v1.GetGroupMessageReadersResponse, 0)

	if len(us) == len(info.Users) {
		for _, v := range us {
			for _, v6 := range info.Users {
				if v == v6.UserId {
					response = append(response, v1.GetGroupMessageReadersResponse{
						UserId: v6.UserId,
						Avatar: v6.Avatar,
						Name:   v6.NickName,
					})
				}
			}
		}
	}

	return response, nil
}
