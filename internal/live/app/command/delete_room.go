package command

import (
	"context"
	"fmt"
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
	"time"
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
		return h.handleDeleteUserRoom(ctx, cmd, room)
	case entity.GroupRoomType:
		return h.deleteGroupRoom(ctx, cmd, room)
	default:
		return code.MyCustomErrorCode.CustomMessage("invalid room type")
	}
}

func (h *LiveHandler) cleanUserRoom(ctx context.Context, room *entity.Room) error {
	// 删除房间和用户连接信息
	if err := h.deleteRoom(ctx, room.ID); err != nil {
		h.logger.Error("删除 Redis 房间错误", zap.Error(err))
	}

	for userID := range room.Participants {
		if err := h.liveRepo.DeleteUsersLive(ctx, userID); err != nil {
			h.logger.Error("删除用户连接信息错误", zap.Error(err))
		}
	}

	return nil
}

func (h *LiveHandler) handleDeleteUserRoom(ctx context.Context, cmd *DeleteRoom, room *entity.Room) error {
	// 获取当前用户的参与者信息
	participant, ok := room.Participants[cmd.UserID]
	if !ok {
		return code.LiveErrUserNotInCall
	}

	// 无论什么情况都先删除房间信息
	if err := h.cleanUserRoom(ctx, room); err != nil {
		return err
	}

	// 存储当前用户的参与者信息
	thisActiveParticipant := participant
	var oppositeActiveParticipant *entity.ActiveParticipant

	// 获取对方用户的参与者信息
	for userID, activeParticipant := range room.Participants {
		if userID != cmd.UserID {
			oppositeActiveParticipant = activeParticipant
			break
		}
	}

	// 当前用户为通话创建者，且对方未接听，取消通话
	if cmd.UserID == room.Creator && !oppositeActiveParticipant.Connected {
		fmt.Println("当前用户为通话创建者且对方未连接")
		return h.handleCancelled(ctx, room, cmd.UserID, cmd.DriverID)
	}

	// 当前用户作为被呼叫者拒绝通话
	if !thisActiveParticipant.Connected && oppositeActiveParticipant.Connected && cmd.UserID != room.Creator {
		fmt.Println("被呼叫者拒绝通话")
		return h.handleRejected(ctx, room, cmd.UserID, cmd.DriverID)
	}

	// 表示双方已加入通话，任意一方挂断
	if thisActiveParticipant.Connected && oppositeActiveParticipant.Connected {
		fmt.Println("任意一方挂断")
		return h.handleAnyDisconnected(ctx, room, cmd.UserID, cmd.DriverID)
	}

	return nil
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

func formatDuration(duration time.Duration) string {
	minutes := int(duration.Minutes())
	seconds := int(duration.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d", minutes, seconds)
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

func (h *LiveHandler) getMessageType(room *entity.Room) uint {
	var msgType uint
	if room.Option.AudioEnabled {
		msgType = uint(msggrpcv1.MessageType_VoiceCall)
	}
	if room.Option.VideoEnabled {
		msgType = uint(msggrpcv1.MessageType_VideoCall)
	}
	return msgType
}

func (h *LiveHandler) sendUserMessage(ctx context.Context, msgType int32, subType int32, dialogID uint32, recipientID, senderID, content string, isBurnAfterReading bool) (uint32, error) {
	message, err := h.msgService.SendUserMessage(ctx, &msggrpcv1.SendUserMsgRequest{
		DialogId:               dialogID,
		SenderId:               senderID,
		ReceiverId:             recipientID,
		Content:                content,
		Type:                   msgType,
		SubType:                subType,
		IsBurnAfterReadingType: isBurnAfterReading,
	})
	if err != nil {
		return 0, err
	}
	return message.MsgId, nil
}

func (h *LiveHandler) pushUserMessageEvent(ctx context.Context, room *entity.Room, dialogID uint32, driverID, senderID, recipientID, content string, msgType, msgSubTye uint, msgID uint32, isBurnAfterReading bool, openBurnAfterReadingTimeOut int64) {
	data := &constants.WsUserMsg{
		SenderId:                senderID,
		Content:                 content,
		MsgType:                 msgType,
		MsgSubType:              msgSubTye,
		MsgId:                   msgID,
		ReceiverId:              recipientID,
		SendAt:                  pkgtime.Now(),
		DialogId:                dialogID,
		IsBurnAfterReading:      isBurnAfterReading,
		BurnAfterReadingTimeOut: openBurnAfterReadingTimeOut,
	}

	bytes, err := utils.StructToBytes(data)
	if err != nil {
		h.logger.Error("转换结构体失败", zap.Error(err))
		return
	}

	h.pushEventToParticipants(ctx, driverID, pushgrpcv1.WSEventType_SendUserMessageEvent, bytes, room)
}

func (h *LiveHandler) pushEventToParticipants(ctx context.Context, driverID string, event pushgrpcv1.WSEventType, data []byte, room *entity.Room) {
	for k := range room.Participants {
		msgContent := &pushgrpcv1.WsMsg{
			Uid:      k,
			Event:    event,
			Data:     &any2.Any{Value: data},
			DriverId: driverID,
		}
		toBytes, err := utils.StructToBytes(msgContent)
		if err != nil {
			h.logger.Error("转换结构体失败", zap.Error(err))
			continue
		}

		h.logger.Debug("push event to room participant", zap.Any("participant", k))

		_, err = h.pushService.Push(ctx, &pushgrpcv1.PushRequest{
			Type: pushgrpcv1.Type_Ws,
			Data: toBytes,
		})
		if err != nil {
			h.logger.Error("发送消息失败", zap.Error(err))
			continue
		}
	}
}

// getRoomDuration 获取房间通话时长
func (h *LiveHandler) getRoomDuration(ctx context.Context, room string) time.Duration {
	livekitRoom, err := h.getLivekitRoom(ctx, room)
	if err != nil {
		h.logger.Error("获取房间信息失败", zap.Error(err))
		return 0
	}
	fmt.Println("livekitRoom.CreationTime => ", livekitRoom.CreationTime)
	creationTime := time.Unix(livekitRoom.CreationTime, 0)
	duration := time.Since(creationTime)
	return duration
}

func (h *LiveHandler) getLivekitRoom(ctx context.Context, room string) (*livekit.Room, error) {
	rooms, err := h.roomService.ListRooms(ctx, &livekit.ListRoomsRequest{Names: []string{room}})
	if err != nil {
		h.logger.Error("获取房间信息失败", zap.Error(err))
		return nil, code.LiveErrGetCallInfoFailed
	}

	if len(rooms.Rooms) == 0 {
		return nil, code.LiveErrCallNotFound
	}

	return rooms.Rooms[0], nil
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

func (h *LiveHandler) handleMissed(ctx context.Context, room *entity.Room, userID, driverID string) error {
	content := "无应答"
	subType := int32(msggrpcv1.CallSubType_Missed)
	return h.handleUserMessage(ctx, room, userID, driverID, content, subType)
}

func (h *LiveHandler) handleCancelled(ctx context.Context, room *entity.Room, userID, driverID string) error {
	content := "取消"
	subType := int32(msggrpcv1.CallSubType_Cancelled)
	return h.handleUserMessage(ctx, room, userID, driverID, content, subType)
}

func (h *LiveHandler) handleRejected(ctx context.Context, room *entity.Room, userID, driverID string) error {
	content := "拒绝"
	subType := int32(msggrpcv1.CallSubType_Rejected)
	return h.handleUserMessage(ctx, room, userID, driverID, content, subType)
}

func (h *LiveHandler) handleAnyDisconnected(ctx context.Context, room *entity.Room, userID, driverID string) error {
	content := fmt.Sprintf("通话时长：%s", formatDuration(h.getRoomDuration(ctx, room.ID)))
	subType := int32(msggrpcv1.CallSubType_Normal)
	return h.handleUserMessage(ctx, room, userID, driverID, content, subType)
}

func (h *LiveHandler) handleUserMessage(ctx context.Context, room *entity.Room, userID, driverID string, content string, subType int32) error {
	var senderID string
	var recipientID string

	if room.Creator == userID {
		senderID = room.Creator
		for participant := range room.Participants {
			if participant != senderID {
				recipientID = participant
				break
			}
		}
	} else {
		senderID = userID
		recipientID = room.Creator
	}

	// 获取关系信息
	relation, err := h.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
		UserId:   senderID,
		FriendId: recipientID,
	})
	if err != nil {
		return err
	}

	var dialogID = relation.DialogId
	var openBurnAfterReading = relation.OpenBurnAfterReading
	var openBurnAfterReadingTimeOut = relation.OpenBurnAfterReadingTimeOut

	// 获取消息类型
	msgType := h.getMessageType(room)

	// 查询发送者信息
	senderInfo, err := h.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: senderID})
	if err != nil {
		return err
	}

	// 创建通话消息
	msgID, err := h.sendUserMessage(ctx, int32(msgType), subType, dialogID, senderID, recipientID, content, openBurnAfterReading)
	if err != nil {
		h.logger.Error("发送消息失败", zap.Error(err))
		return err
	}

	// 推送通话消息
	h.pushUserMessageEvent(ctx, room, dialogID, driverID, senderID, recipientID, content, msgType, uint(subType), msgID, openBurnAfterReading, openBurnAfterReadingTimeOut)

	// 为发送者生成ws推送消息
	message, err := buildWsUserMessage(dialogID, msgType, driverID, senderID, recipientID, content, constants.SenderInfo{
		UserId: senderInfo.UserId,
		Name:   senderInfo.NickName,
		Avatar: senderInfo.Avatar,
	}, openBurnAfterReading)
	if err != nil {
		return err
	}

	// 推送通话事件
	eventType := pushgrpcv1.WSEventType_UserCallRejectEvent
	if subType == int32(msggrpcv1.CallSubType_Normal) {
		eventType = pushgrpcv1.WSEventType_UserCallEndEvent
	}
	h.pushEventToParticipants(ctx, driverID, eventType, message, room)

	return nil
}

func buildWsUserMessage(dialogID uint32, msgType uint, driverID, senderID, recipientID, content string, senderInfo constants.SenderInfo, isBurnAfterReading bool) ([]byte, error) {
	data := &constants.WsUserMsg{
		SenderId:           senderID,
		Content:            content,
		MsgType:            msgType,
		DialogId:           dialogID,
		IsBurnAfterReading: isBurnAfterReading,
		SenderInfo: constants.SenderInfo{
			Avatar: senderInfo.Avatar,
			Name:   senderInfo.Name,
			UserId: senderInfo.UserId,
		},
	}

	bytes, err := utils.StructToBytes(data)
	if err != nil {
		return nil, err
	}

	msgContent := &pushgrpcv1.WsMsg{
		Uid:      recipientID,
		Event:    pushgrpcv1.WSEventType_SendUserMessageEvent,
		DriverId: driverID,
		Data:     &any2.Any{Value: bytes},
	}

	toBytes, err := utils.StructToBytes(msgContent)
	if err != nil {
		return nil, err
	}

	return toBytes, nil
}
