package command

import (
	"context"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/internal/live/domain/live"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils"
	any2 "github.com/golang/protobuf/ptypes/any"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	"go.uber.org/zap"
	"strconv"
)

type JoinRoom struct {
	Room     string
	UserID   string
	DriverID string
	Option   RoomOption
}

type JoinRoomResponse struct {
	Room  string
	Url   string
	Token string
}

func (h *LiveHandler) JoinRoom(ctx context.Context, cmd *JoinRoom) (*JoinRoomResponse, error) {
	var err error
	var room *live.Room

	room1, err := h.liveRepo.GetRoom(ctx, cmd.Room)
	if err != nil {
		return nil, err
	}

	switch room1.Type {
	case live.GroupRoomType:
		room, err = h.joinGroupRoom(ctx, cmd.Room, room1.GroupID, cmd.UserID, cmd.DriverID)
		if err != nil {
			return nil, err
		}
	case live.UserRoomType:
		room, err = h.joinUserRoom(ctx, cmd.Room, cmd.UserID, cmd.DriverID)
		if err != nil {
			return nil, err
		}
	default:
		return nil, errors.New("invalid room type")
	}

	user, err := h.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: cmd.UserID})
	if err != nil {
		return nil, err
	}

	var token string
	if cmd.UserID == room.Creator {
		token, err = h.GetAdminJoinToken(ctx, room.ID, user.NickName, cmd.UserID)
	} else {
		token, err = h.GetUserJoinToken(ctx, room.ID, user.NickName, cmd.UserID)
	}
	if err != nil {
		return nil, err
	}

	return &JoinRoomResponse{
		Room:  room.ID,
		Url:   h.webRtcUrl,
		Token: token,
	}, nil
}

func (h *LiveHandler) joinUserRoom(ctx context.Context, roomID, userID, driverID string) (*live.Room, error) {
	room, err := h.liveRepo.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}

	if _, ok := room.Participants[userID]; !ok {
		return nil, code.Forbidden
	}

	if room.Participants[userID].Connected {
		return nil, code.LiveErrAlreadyInCall
	}

	if room.NumParticipants+1 > room.MaxParticipants {
		return nil, code.LiveErrMaxParticipantsExceeded
	}

	rooms, err := h.roomService.ListRooms(ctx, &livekit.ListRoomsRequest{
		Names: []string{roomID},
	})
	if err != nil {
		return nil, err
	}

	if len(rooms.Rooms) == 0 {
		return nil, code.LiveErrCallNotFound
	}

	if rooms.Rooms[0].NumPublishers+1 > room.MaxParticipants {
		return nil, code.LiveErrMaxParticipantsExceeded
	}

	room.NumParticipants++
	room.Participants[userID].Connected = true
	room.Participants[userID].Status = live.ParticipantInfo_JOINED

	// 人数等于 MaxParticipants(2) 代表双方都加入通话了，更新过期时间为永久直至挂断才删除通话1
	if room.NumParticipants == room.MaxParticipants {
		for i, _ := range room.Participants {
			if err := h.liveRepo.SetUserLivePersist(ctx, i); err != nil {
				h.logger.Error("更新用户过期时间失败", zap.Error(err))
				return nil, err
			}
		}
		//if err := h.liveRepo.SetRoomPersist(ctx, room.ID); err != nil {
		//	h.logger.Error("更新房间过期时间失败", zap.Error(err))
		//	return nil, err
		//}
		if err := h.liveRepo.UpdateRoomWithExpiration(ctx, room, PermanentExpiration); err != nil {
			h.logger.Error("更新房间过期时间失败", zap.Error(err))
			return nil, err
		}
	} else {
		if err := h.liveRepo.UpdateRoomWithExpiration(ctx, room, PreviousExpiration); err != nil {
			h.logger.Error("更新房间过期时间失败", zap.Error(err))
			return nil, err
		}
	}

	for k := range room.Participants {
		if k == userID {
			continue
		}

		data := map[string]interface{}{
			"room":         room.ID,
			"sender_id":    userID,
			"recipient_id": k,
			"driver_id":    driverID,
		}
		bytes, err := utils.StructToBytes(data)
		if err != nil {
			return nil, err
		}

		msg := &pushgrpcv1.WsMsg{Uid: k, Event: pushgrpcv1.WSEventType_UserCallAcceptEvent, Data: &any2.Any{Value: bytes}}
		toBytes, err := utils.StructToBytes(msg)
		if err != nil {
			return nil, err
		}
		_, err = h.pushService.Push(ctx, &pushgrpcv1.PushRequest{
			Type: pushgrpcv1.Type_Ws,
			Data: toBytes,
		})
		if err != nil {
			h.logger.Error("发送消息失败", zap.Error(err))
			continue
		}
		h.logger.Info("推送消息成功", zap.String("uid", k), zap.Any("msg", msg))
	}

	h.logger.Info("用户加入房间", zap.String("uid", userID), zap.Any("room", room), zap.String("webRtcUrl", h.webRtcUrl))

	return room, nil
}

func (h *LiveHandler) joinGroupRoom(ctx context.Context, roomID string, groupID uint32, userID, driverID string) (*live.Room, error) {
	_, err := h.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{UserId: userID, GroupId: groupID})
	if err != nil {
		return nil, err
	}

	room, err := h.liveRepo.GetRoom(ctx, roomID)
	if err != nil {
		return nil, err
	}
	p, ok := room.Participants[userID]
	if ok {
		if p.Connected || p.Status == live.ParticipantInfo_JOINED {
			return nil, code.LiveErrAlreadyInCall
		}
	}
	if room.NumParticipants+1 > room.MaxParticipants {
		return nil, code.LiveErrMaxParticipantsExceeded
	}

	rooms, err := h.liveRepo.GetUserRooms(ctx, userID)
	if err != nil {
		h.logger.Error("获取用户房间失败", zap.Error(err))
		if !(code.IsCode(err, code.LiveErrCallNotFound)) {
			return nil, err
		}
	}
	for _, v := range rooms {
		p, ok = v.Participants[userID]
		if ok && p.Connected {
			return nil, code.LiveErrAlreadyInCall
		}
	}

	livekitRooms, err := h.roomService.ListRooms(ctx, &livekit.ListRoomsRequest{
		Names: []string{roomID},
	})
	if err != nil {
		return nil, err
	}
	if len(livekitRooms.Rooms) == 0 {
		return nil, code.LiveErrCallNotFound
	}
	if livekitRooms.Rooms[0].NumPublishers+1 > room.MaxParticipants {
		return nil, code.LiveErrMaxParticipantsExceeded
	}

	fmt.Println("room => ", room)

	_, ok = room.Participants[userID]
	if ok {
		room.Participants[userID].Connected = true
		room.Participants[userID].Status = live.ParticipantInfo_JOINED
	} else {
		room.Participants[userID] = &live.ActiveParticipant{
			Connected: true,
			Status:    live.ParticipantInfo_JOINED,
		}
	}
	room.NumParticipants++

	// 用户加入了群聊通话，更新过期时间为永久直至挂断才删除通话
	if room.NumParticipants == 1 {
		if err := h.liveRepo.SetGroupLivePersist(ctx, strconv.Itoa(int(groupID))); err != nil {
			return nil, err
		}
	}

	if err := h.liveRepo.UpdateRoomWithExpiration(ctx, room, PermanentExpiration); err != nil {
		h.logger.Error("更新房间过期时间失败", zap.Error(err))
		return nil, err
	}

	if err := h.liveRepo.CreateUsersLive(ctx, roomID, userID); err != nil {
		h.logger.Error("创建用户房间失败", zap.Error(err))
		return nil, err
	}

	if err := h.liveRepo.SetUserLivePersist(ctx, userID); err != nil {
		h.logger.Error("更新用户过期时间失败", zap.Error(err))
		return nil, err
	}

	for k := range room.Participants {
		if k == userID {
			continue
		}
		bytes, err := utils.StructToBytes(map[string]interface{}{
			"room":         room.ID,
			"sender_id":    userID,
			"recipient_id": k,
			"driver_id":    driverID,
		})
		if err != nil {
			h.logger.Error("序列化失败", zap.Error(err))
			return nil, err
		}

		msg := &pushgrpcv1.WsMsg{Uid: k, Event: pushgrpcv1.WSEventType_GroupCallAcceptEvent, Data: &any2.Any{Value: bytes}}
		toBytes, err := utils.StructToBytes(msg)
		if err != nil {
			h.logger.Error("序列化失败", zap.Error(err))
			return nil, err
		}

		_, err = h.pushService.Push(ctx, &pushgrpcv1.PushRequest{
			Type: pushgrpcv1.Type_Ws,
			Data: toBytes,
		})
		if err != nil {
			h.logger.Error("发送消息失败", zap.Error(err))
		}
	}

	h.logger.Info("加入群聊通话", zap.Uint32("gid", groupID), zap.String("uid", userID), zap.Any("room", room), zap.String("webRtcUrl", h.webRtcUrl))
	return room, nil
}

func (h *LiveHandler) GetUserJoinToken(ctx context.Context, room, userName, userID string) (string, error) {
	at := auth.NewAccessToken(h.liveApiKey, h.liveApiSecret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     room,
	}
	at.AddGrant(grant).SetName(userName).SetIdentity(userID).SetValidFor(h.liveTimeout)
	jwt, err := at.ToJWT()
	if err != nil {
		h.logger.Error("Failed to generate JWT", zap.Error(err))
		return "", err
	}

	return jwt, nil
}

func (h *LiveHandler) GetAdminJoinToken(ctx context.Context, room, userName, userID string) (string, error) {
	at := auth.NewAccessToken(h.liveApiKey, h.liveApiSecret)
	grant := &auth.VideoGrant{
		RoomCreate: true,
		RoomList:   true,
		RoomAdmin:  true,
		RoomJoin:   true,
		Room:       room,
	}
	at.AddGrant(grant).SetName(userName).SetIdentity(userID).SetValidFor(h.liveTimeout)
	jwt, err := at.ToJWT()
	if err != nil {
		h.logger.Error("Failed to generate JWT", zap.Error(err))
		return "", err
	}

	return jwt, nil
}
