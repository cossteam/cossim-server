package command

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/internal/live/domain/live"
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

type RejectLive struct {
	Room     string
	UserID   string
	DriverID string
	Option   RoomOption
}

type RejectLiveResponse struct {
	Room  string
	Url   string
	Token string
}

func (h *LiveHandler) RejectLive(ctx context.Context, cmd *RejectLive) (*RejectLiveResponse, error) {
	room, err := h.liveRepo.GetRoom(ctx, cmd.Room)
	if err != nil {
		return nil, err
	}
	ap, ok := room.Participants[cmd.UserID]
	if !ok {
		return nil, code.Forbidden
	}
	if ap.Connected {
		return nil, code.LiveErrRejectCallFailed.CustomMessage("已加入通话")
	}

	switch room.Type {
	case live.GroupRoomType:
		err = h.rejectGroup(ctx, cmd.Room, room.GroupID, cmd.UserID, cmd.DriverID, room)
	case live.UserRoomType:
		err = h.rejectUser(ctx, cmd.Room, cmd.UserID, cmd.DriverID, room)
	default:
		return nil, errors.New("invalid room type")
	}

	return nil, err
}

func (h *LiveHandler) rejectUser(ctx context.Context, roomID string, userID string, driverID string, room *live.Room) error {
	// Check if the user is the sender of the call
	if room.Creator == userID {
		h.logger.Warn("Cannot reject a call initiated by the user", zap.String("uid", userID), zap.Any("room", room))
		return code.LiveErrRejectCallFailed.CustomMessage("不能拒绝自己发起的通话")
	}

	rel, err := h.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: userID, FriendId: room.Creator})
	if err != nil {
		return err
	}

	if rel.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
		return code.RelationUserErrFriendRelationNotFound
	}

	var creator = room.Creator
	var receiverId string
	for k, _ := range room.Participants {
		if k != userID {
			receiverId = k
			break
		}
	}

	var msgType int32
	if room.Option.AudioEnabled {
		msgType = int32(msggrpcv1.MessageType_VoiceCall)
	}
	if room.Option.VideoEnabled {
		msgType = int32(msggrpcv1.MessageType_VideoCall)
	}

	isBurnAfterReading := rel.OpenBurnAfterReading
	OpenBurnAfterReadingTimeOut := rel.OpenBurnAfterReadingTimeOut
	message, err := h.msgService.SendUserMessage(ctx, &msggrpcv1.SendUserMsgRequest{
		DialogId:               rel.DialogId,
		SenderId:               room.Creator,
		ReceiverId:             userID,
		Content:                "已拒绝",
		Type:                   msgType,
		IsBurnAfterReadingType: msggrpcv1.BurnAfterReadingType(isBurnAfterReading),
	})
	if err != nil {
		h.logger.Error("发送消息失败", zap.Error(err))
		return code.LiveErrRejectCallFailed
	}

	info, err := h.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
		UserId: creator,
	})
	if err != nil {
		h.logger.Error("获取用户信息失败", zap.Error(err))
		return code.LiveErrLeaveCallFailed
	}
	data := &constants.WsUserMsg{
		SenderId:                creator,
		Content:                 "已拒绝",
		MsgType:                 uint(msgType),
		MsgId:                   message.MsgId,
		ReceiverId:              receiverId,
		SendAt:                  pkgtime.Now(),
		DialogId:                rel.DialogId,
		IsBurnAfterReading:      constants.BurnAfterReadingType(isBurnAfterReading),
		BurnAfterReadingTimeOut: OpenBurnAfterReadingTimeOut,
		SenderInfo: constants.SenderInfo{
			Avatar: info.Avatar,
			Name:   info.NickName,
			UserId: creator,
		},
	}
	bytes, err := utils.StructToBytes(data)
	if err != nil {
		return err
	}
	var msgs []*pushgrpcv1.WsMsg
	for k, _ := range room.Participants {
		msgContent := &pushgrpcv1.WsMsg{
			Uid:      k,
			Event:    pushgrpcv1.WSEventType_SendUserMessageEvent,
			DriverId: driverID,
			Data:     &any2.Any{Value: bytes},
		}
		msgs = append(msgs, msgContent)
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

	if err := h.deleteUserLive(ctx, room); err != nil {
		return err
	}

	data2 := map[string]interface{}{
		"url":          h.webRtcUrl,
		"sender_id":    userID,
		"recipient_id": receiverId,
	}
	structToBytes, err := utils.StructToBytes(data2)
	if err != nil {
		return err
	}
	msg := &pushgrpcv1.WsMsg{Uid: receiverId, Event: pushgrpcv1.WSEventType_UserCallRejectEvent, Data: &any2.Any{Value: structToBytes}}
	toBytes3, err := utils.StructToBytes(msg)
	if err != nil {
		return err
	}
	_, err = h.pushService.Push(ctx, &pushgrpcv1.PushRequest{
		Type: pushgrpcv1.Type_Ws,
		Data: toBytes3,
	})
	if err != nil {
		h.logger.Error("发送消息失败", zap.Error(err))
	}

	h.logger.Info("UserRejectRoom", zap.Any("room", room), zap.String("SenderID", userID), zap.String("RecipientID", receiverId))

	return nil
}

func (h *LiveHandler) rejectGroup(ctx context.Context, roomID string, groupID uint32, userID, driverID string, room *live.Room) error {
	if groupID == 0 {
		return code.LiveErrRejectCallFailed.CustomMessage("群组不存在")
	}

	_, err := h.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		GroupId: groupID,
		UserId:  userID,
	})
	if err != nil {
		return err
	}

	// Check if the user is the sender of the call
	if room.Creator == userID {
		h.logger.Warn("Cannot reject a call initiated by the user", zap.String("uid", userID), zap.Uint32("gid", groupID))
		return code.LiveErrRejectCallFailed.CustomMessage("不能拒绝自己发起的通话")
	}

	// Check if the user is a participant in the call
	pp, ok := room.Participants[userID]
	if !ok {
		h.logger.Warn("User is not a participant in the call", zap.String("uid", userID), zap.Uint32("gid", groupID))
		return code.Forbidden
	}

	// Check if the user is already connected to the call
	if pp.Connected {
		h.logger.Warn("User is already connected to the call", zap.String("uid", userID), zap.Uint32("gid", groupID))
		return code.LiveErrRejectCallFailed.CustomMessage("已加入通话")
	}

	if err := h.deleteGroupLive(ctx, groupID, room); err != nil {
		h.logger.Error("DeleteGroupRoom failed", zap.Error(err))
		return err
	}

	// Send rejection message to all participants in the call

	for id := range room.Participants {
		data := map[string]interface{}{
			"sender_id":    room.Creator,
			"recipient_id": id,
		}
		bytes, err := utils.StructToBytes(data)
		if err != nil {
			return err
		}
		msg := &pushgrpcv1.WsMsg{
			Uid:   id,
			Event: pushgrpcv1.WSEventType_GroupCallRejectEvent,
			Data:  &any2.Any{Value: bytes},
		}

		toBytes, err := utils.StructToBytes(msg)
		if err != nil {
			return err
		}

		_, err = h.pushService.Push(ctx, &pushgrpcv1.PushRequest{
			Type: pushgrpcv1.Type_Ws,
			Data: toBytes,
		})
		if err != nil {
			h.logger.Error("发送消息失败", zap.Error(err))
		}
	}
	return nil
}

func (h *LiveHandler) deleteUserLive(ctx context.Context, room *live.Room) error {
	var users []string
	for k, _ := range room.Participants {
		users = append(users, k)
	}
	if err := h.liveRepo.DeleteUsersLive(ctx, users...); err != nil {
		return err
	}
	if err := h.liveRepo.DeleteRoom(ctx, room.ID); err != nil {
		return err
	}
	_, err := h.roomService.DeleteRoom(ctx, &livekit.DeleteRoomRequest{Room: room.ID})
	if err != nil {
		return err
	}
	return nil
}

func (h *LiveHandler) deleteGroupLive(ctx context.Context, groupID uint32, room *live.Room) error {
	var users []string
	for k, _ := range room.Participants {
		users = append(users, k)
	}
	if err := h.liveRepo.DeleteUsersLive(ctx, users...); err != nil {
		return err
	}
	if err := h.liveRepo.DeleteGroupLive(ctx, strconv.Itoa(int(groupID))); err != nil {
		return err
	}
	if err := h.liveRepo.DeleteRoom(ctx, room.ID); err != nil {
		return err
	}
	_, err := h.roomService.DeleteRoom(ctx, &livekit.DeleteRoomRequest{Room: room.ID})
	if err != nil {
		return err
	}
	return nil
}
