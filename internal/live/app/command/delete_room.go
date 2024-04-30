package command

import (
	"context"
	"github.com/cossim/coss-server/internal/live/domain/entity"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/utils"
	pkgtime "github.com/cossim/coss-server/pkg/utils/time"
	any2 "github.com/golang/protobuf/ptypes/any"
	"github.com/livekit/protocol/livekit"
	"go.uber.org/zap"
	"strconv"
)

type DeleteRoom struct {
	Room     string
	UserID   string
	DriverID string
}

func (h *LiveHandler) DeleteRoom(ctx context.Context, cmd *DeleteRoom) error {
	h.logger.Debug("received deleteRoom request", zap.Any("cmd", cmd))

	room, err := h.liveRepo.GetRoom(ctx, cmd.Room)
	if err != nil {
		return err
	}

	switch room.Type {
	case entity.UserRoomType:
		return h.deleteUserRoom(ctx, cmd, room)
	case entity.GroupRoomType:
		return h.deleteGroupRoom(ctx, cmd, room)
	default:
		return code.MyCustomErrorCode.CustomMessage("invalid room type")
	}
}

func (h *LiveHandler) deleteUserRoom(ctx context.Context, cmd *DeleteRoom, room *entity.Room) error {
	_, ok := room.Participants[cmd.UserID]
	if !ok {
		return code.LiveErrUserNotInCall
	}

	for participant, p := range room.Participants {
		if err := h.liveRepo.DeleteUsersLive(ctx, participant); err != nil {
			h.logger.Error("delete user live error", zap.Error(err))
		}
		if participant == cmd.UserID {
			continue
		}

		if !p.Connected {
			if err := h.handleUserReject(ctx, participant, cmd.UserID, cmd.DriverID, room); err != nil {
				return err
			}
		} else {
			h.handleUserLeave(ctx, participant, cmd.UserID, cmd.DriverID)
		}
	}

	return h.deleteRoom(ctx, room.ID)
}

func (h *LiveHandler) deleteGroupRoom(ctx context.Context, cmd *DeleteRoom, room *entity.Room) error {
	if err := h.liveRepo.DeleteUsersLive(ctx, cmd.UserID); err != nil {
		h.logger.Error("delete user live error", zap.Error(err), zap.String("user_id", cmd.UserID), zap.String("room", room.ID))
		return err
	}

	ap, ok := room.Participants[cmd.UserID]
	if !ok {
		return code.LiveErrUserNotInCall
	}

	// 如果是最后一个退出的用户，则删除整个房间
	if room.NumParticipants-1 == 0 {
		if err := h.liveRepo.DeleteGroupLive(ctx, strconv.Itoa(int(room.GroupID))); err != nil {
			h.logger.Error("delete group live error", zap.Error(err))
			return err
		}
		return h.deleteRoom(ctx, room.ID)
	}

	_, err := h.roomService.RemoveParticipant(context.Background(), &livekit.RoomParticipantIdentity{
		Room:     room.ID,
		Identity: cmd.UserID,
	})
	if err != nil {
		h.logger.Error("remove participant error", zap.Error(err))
		return err
	}

	room.NumParticipants--
	delete(room.Participants, cmd.UserID)
	if err := h.liveRepo.UpdateRoom(ctx, room); err != nil {
		h.logger.Error("update room error", zap.Error(err))
		return err
	}

	h.logger.Info("退出群聊通话", zap.String("uid", cmd.UserID), zap.Any("room", room))

	h.notifyGroupCallEnd(ctx, room, cmd.UserID, ap.Connected)
	return nil
}

func (h *LiveHandler) deleteEntireGroupRoom(ctx context.Context, room *entity.Room) error {
	for userID := range room.Participants {
		if err := h.liveRepo.DeleteUsersLive(ctx, userID); err != nil {
			h.logger.Error("delete user live error", zap.Error(err))
		}
	}
	if err := h.liveRepo.DeleteGroupLive(ctx, strconv.Itoa(int(room.GroupID))); err != nil {
		h.logger.Error("delete group live error", zap.Error(err))
	}
	if err := h.liveRepo.DeleteRoom(ctx, room.ID); err != nil {
		h.logger.Error("delete redis room error", zap.Error(err))
	}
	if _, err := h.roomService.DeleteRoom(ctx, &livekit.DeleteRoomRequest{
		Room: room.ID,
	}); err != nil {
		return err
	}
	return nil
}

func (h *LiveHandler) handleUserLeave(ctx context.Context, recipientID, senderID, driverID string) {
	// TODO 需要添加用户挂断事件
	event := pushgrpcv1.WSEventType_UserCallEndEvent
	data := map[string]interface{}{
		"sender_id":    senderID,
		"recipient_id": recipientID,
		"driver_id":    driverID,
	}
	h.sendPushMessage(ctx, senderID, recipientID, event, data)
}

func (h *LiveHandler) notifyGroupCallEnd(ctx context.Context, room *entity.Room, senderID string, connected bool) {
	var event pushgrpcv1.WSEventType
	if !connected {
		event = pushgrpcv1.WSEventType_GroupCallRejectEvent
	} else {
		event = pushgrpcv1.WSEventType_UserLeaveGroupCallEvent
	}

	for participant := range room.Participants {
		data := map[string]interface{}{
			"sender_id":    senderID,
			"recipient_id": participant,
		}
		h.sendPushMessage(ctx, senderID, participant, event, data)
	}
}

func (h *LiveHandler) handleUserReject(ctx context.Context, participantID, senderID, driverID string, room *entity.Room) error {
	rel, err := h.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: participantID, FriendId: senderID})
	if err != nil {
		return err
	}
	if rel.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
		return code.RelationUserErrFriendRelationNotFound
	}

	var msgType int32
	if room.Option.AudioEnabled {
		msgType = int32(msggrpcv1.MessageType_VoiceCall)
	}
	if room.Option.VideoEnabled {
		msgType = int32(msggrpcv1.MessageType_VideoCall)
	}

	var receiverId string
	for k, _ := range room.Participants {
		if k != room.Creator {
			receiverId = k
			break
		}
	}

	isBurnAfterReading := rel.OpenBurnAfterReading
	openBurnAfterReadingTimeout := rel.OpenBurnAfterReadingTimeOut
	message, err := h.msgService.SendUserMessage(ctx, &msggrpcv1.SendUserMsgRequest{
		DialogId:               rel.DialogId,
		SenderId:               room.Creator,
		ReceiverId:             receiverId,
		Content:                "已拒绝",
		Type:                   msgType,
		IsBurnAfterReadingType: msggrpcv1.BurnAfterReadingType(isBurnAfterReading),
	})
	if err != nil {
		h.logger.Error("发送消息失败", zap.Error(err))
		return code.LiveErrRejectCallFailed
	}

	// 获取发送者信息
	info, err := h.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
		UserId: senderID,
	})
	if err != nil {
		h.logger.Error("获取用户信息失败", zap.Error(err))
		return code.LiveErrLeaveCallFailed
	}

	data := &constants.WsUserMsg{
		SenderId:                senderID,
		Content:                 "已拒绝",
		MsgType:                 uint(msgType),
		MsgId:                   message.MsgId,
		ReceiverId:              receiverId,
		SendAt:                  pkgtime.Now(),
		DialogId:                rel.DialogId,
		IsBurnAfterReading:      constants.BurnAfterReadingType(isBurnAfterReading),
		BurnAfterReadingTimeOut: openBurnAfterReadingTimeout,
		SenderInfo: constants.SenderInfo{
			Avatar: info.Avatar,
			Name:   info.NickName,
			UserId: senderID,
		},
	}
	bytes, err := utils.StructToBytes(data)
	if err != nil {
		return err
	}

	var msgs []*pushgrpcv1.WsMsg
	msgContent := func(uid string) *pushgrpcv1.WsMsg {
		return &pushgrpcv1.WsMsg{
			Uid:      uid,
			Event:    pushgrpcv1.WSEventType_SendUserMessageEvent,
			DriverId: driverID,
			Data:     &any2.Any{Value: bytes},
		}
	}
	for k := range room.Participants {
		if k == senderID {
			continue
		}
		msgs = append(msgs, msgContent(k))
	}

	toBytes, err := utils.StructToBytes(msgs)
	if err != nil {
		return err
	}
	_, err = h.pushService.Push(ctx, &pushgrpcv1.PushRequest{
		Type: pushgrpcv1.Type_Ws_Batch_User,
		Data: toBytes,
	})
	if err != nil {
		h.logger.Error("发送消息失败", zap.Error(err))
	}

	//h.sendPushMessage(ctx, participantID, pushgrpcv1.WSEventType_SendUserMessageEvent, msgs)
	return nil
}

func (h *LiveHandler) sendPushMessage(ctx context.Context, senderID, recipientID string, event pushgrpcv1.WSEventType, data map[string]interface{}) {
	messageBytes, err := buildPushMessageBytes(recipientID, event, data)
	if err != nil {
		h.logger.Error("failed to convert message to bytes", zap.Error(err))
		return
	}

	h.logger.Info("push websocket message", zap.String("senderID", senderID), zap.String("recipientID", recipientID), zap.Any("event", event), zap.Any("data", data))

	_, err = h.pushService.Push(ctx, &pushgrpcv1.PushRequest{
		Type: pushgrpcv1.Type_Ws,
		Data: messageBytes,
	})
	if err != nil {
		h.logger.Error("failed to push websocket message", zap.Error(err))
	}
}

func (h *LiveHandler) deleteRoom(ctx context.Context, roomID string) error {
	if err := h.liveRepo.DeleteRoom(ctx, roomID); err != nil {
		h.logger.Error("delete redis room error", zap.Error(err))
		return err
	}

	if _, err := h.roomService.DeleteRoom(ctx, &livekit.DeleteRoomRequest{
		Room: roomID,
	}); err != nil {
		return err
	}
	return nil
}

// buildPushMessageBytes 创建推送消息并转换为字节切片
func buildPushMessageBytes(recipientID string, event pushgrpcv1.WSEventType, data map[string]interface{}) ([]byte, error) {
	structToBytes, err := utils.StructToBytes(data)
	if err != nil {
		return nil, err
	}

	msg := &pushgrpcv1.WsMsg{Uid: recipientID, Event: event, Data: &any2.Any{Value: structToBytes}}
	toBytes, err := utils.StructToBytes(msg)
	if err != nil {
		return nil, err
	}

	return toBytes, nil
}
