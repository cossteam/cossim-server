package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cossim/coss-server/interface/live/api/dto"
	"github.com/cossim/coss-server/interface/live/api/model"
	msgconfig "github.com/cossim/coss-server/interface/msg/config"
	"github.com/cossim/coss-server/pkg/cache"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/msg_queue"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	usergrpcv1 "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/google/uuid"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

func (s *Service) CreateUserCall(ctx context.Context, senderID, recipientID string) (*dto.UserCallResponse, error) {
	is, err := s.isEitherUserInCall(ctx, senderID, recipientID)
	if err != nil {
		return nil, err
	}
	if is {
		return nil, code.LiveErrAlreadyInCall
	}

	rel, err := s.relUserClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
		UserId:   senderID,
		FriendId: recipientID,
	})
	if err != nil {
		s.logger.Error("获取用户关系失败", zap.Error(err))
		return nil, err
	}

	if rel.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
		return nil, code.RelationUserErrFriendRelationNotFound
	}

	sender, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: senderID})
	if err != nil {
		return nil, err
	}

	recipient, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: senderID})
	if err != nil {
		return nil, err
	}

	roomName := uuid.New().String()
	room, err := s.roomService.CreateRoom(ctx, &livekit.CreateRoomRequest{
		Name:            roomName,
		EmptyTimeout:    uint32(s.liveTimeout.Seconds()), // 空闲时间
		MaxParticipants: 2,
	})
	if err != nil {
		s.logger.Error("创建通话失败", zap.Error(err))
		return nil, err
	}

	fmt.Println("room => ", room)

	senderToken, err := s.GetJoinToken(ctx, sender.NickName, room.Name)
	if err != nil {
		return nil, err
	}

	RecipientToken, err := s.GetJoinToken(ctx, recipient.NickName, room.Name)
	if err != nil {
		return nil, err
	}

	ToJSONString, err := (&model.UserRoomInfo{
		SenderID:        senderID,
		RecipientID:     recipientID,
		MaxParticipants: 2,
		Participants:    make(map[string]*model.ActiveParticipant),
	}).ToJSONString()
	if err != nil {
		return nil, err
	}
	if err = cache.SetKey(s.redisClient, liveUserPrefix+roomName+":"+senderID+":"+recipientID, ToJSONString, s.liveTimeout); err != nil {
		s.logger.Error("保存房间信息失败", zap.Error(err))
		return nil, err
	}

	msg := msgconfig.WsMsg{Uid: recipientID, Event: msgconfig.UserCallReqEvent, Data: map[string]interface{}{
		"url":          s.livekitServer,
		"token":        RecipientToken,
		"room":         room.Name,
		"sender_id":    senderID,
		"recipient_id": recipientID,
	}}
	if err = s.publishServiceMessage(ctx, msg); err != nil {
		s.logger.Error("发送消息失败", zap.Error(err))
		return nil, err
	}

	return &dto.UserCallResponse{
		Url:    s.livekitServer,
		Token:  senderToken,
		Room:   room.Name,
		RoomID: room.Sid,
	}, nil
}

func (s *Service) UserJoinRoom(ctx context.Context, uid, roomName string) (interface{}, error) {
	roomInfo, key, err := s.getRedisUserRoom(ctx, roomName)
	if err != nil {
		return nil, err
	}

	if uid != roomInfo.RecipientID && uid != roomInfo.SenderID {
		return nil, code.Unauthorized
	}

	if roomInfo.NumParticipants+1 > roomInfo.MaxParticipants {
		return nil, code.LiveErrMaxParticipantsExceeded
	}

	if _, ok := roomInfo.Participants[uid]; ok {
		return nil, code.LiveErrAlreadyInCall
	}

	room, err := s.getLivekitRoom(ctx, roomName)
	if err != nil {
		return nil, err
	}

	if room.NumParticipants+1 > room.MaxParticipants {
		return nil, code.LiveErrMaxParticipantsExceeded
	}

	roomInfo.NumParticipants++
	roomInfo.Participants[uid] = &model.ActiveParticipant{
		Connecting: true,
	}
	ToJSONString, err := roomInfo.ToJSONString()
	if err != nil {
		return nil, err
	}
	if err = cache.SetKey(s.redisClient, key, ToJSONString, 0); err != nil {
		s.logger.Error("更新房间信息失败", zap.Error(err))
		return nil, err
	}

	//if err = cache.UpdateKeyExpiration(s.redisClient, key, 365*24*time.Hour); err != nil {
	//	s.logger.Error("更新房间信息失败", zap.Error(err))
	//	return nil, code.LiveErrJoinCallFailed
	//}

	//token, err := s.joinRoom(uid, listRes.Rooms[0])
	//if err != nil {
	//	s.logger.Error("error joining room", zap.Error(err))
	//	return nil, err
	//}

	return nil, nil
}

func (s *Service) joinRoom(userName string, room *livekit.Room) (string, error) {
	//roomCache, err := cache.GetKey(s.redisClient, room.Sid)
	//if err != nil {
	//	return "", err
	//}
	//data := &model.UserRoomInfo{}
	//if err = data.FromMap(roomCache); err != nil {
	//	return "", err
	//}

	//s.lock.Lock()
	//if _, ok := data.Participants[room.Sid]; ok {
	//	s.lock.Unlock()
	//	return "", code.LiveErrUserAlreadyInCall
	//}
	//data.Participants[room.Sid] = &model.ActiveParticipant{
	//	Connecting: true,
	//}
	//s.lock.Unlock()

	token := s.roomService.CreateToken().SetIdentity(userName).AddGrant(&auth.VideoGrant{
		Room:     room.Name,
		RoomJoin: true,
	})
	jwt, err := token.ToJWT()
	if err != nil {
		return "", err
	}

	return jwt, err
}

func (s *Service) getLivekitRoom(ctx context.Context, room string) (*livekit.Room, error) {
	rooms, err := s.roomService.ListRooms(ctx, &livekit.ListRoomsRequest{Names: []string{room}})
	if err != nil {
		s.logger.Error("获取房间信息失败", zap.Error(err))
		return nil, code.LiveErrGetCallInfoFailed
	}

	if len(rooms.Rooms) == 0 {
		s.deleteRoom(ctx, room)
		return nil, code.LiveErrCallNotFound
	}

	return rooms.Rooms[0], nil
}

func (s *Service) GetUserRoom(ctx context.Context, uid, room string) (*dto.UserShowResponse, error) {
	content, err := s.scanKeysWithPrefixAndContent(ctx, uid)
	if err != nil {
		return nil, err
	}
	fmt.Println("content => ", content)

	roomInfo, _, err := s.getRedisUserRoom(ctx, room)
	if err != nil {
		return nil, err
	}

	if uid != roomInfo.RecipientID && uid != roomInfo.SenderID {
		return nil, code.Unauthorized
	}

	livekitRoom, err := s.getLivekitRoom(ctx, room)
	if err != nil {
		return nil, err
	}

	resp := &dto.UserShowResponse{
		StartAt:     livekitRoom.CreationTime,
		Duration:    0,
		Room:        livekitRoom.Name,
		Participant: make([]*dto.ParticipantInfo, 0),
	}
	res, err := s.roomService.ListParticipants(ctx, &livekit.ListParticipantsRequest{
		Room: room, // 房间名称
	})
	if err != nil {
		s.logger.Error("获取通话信息失败", zap.Error(err))
		return nil, code.LiveErrGetCallInfoFailed
	}

	fmt.Println("GetUserRoom res => ", res)

	for _, p := range res.Participants {
		//if p.Identity == userID {
		//	continue
		//}
		resp.Participant = append(resp.Participant, &dto.ParticipantInfo{
			Sid:         p.Sid,
			Identity:    p.Identity,
			State:       dto.ParticipantState(p.State),
			JoinedAt:    p.JoinedAt,
			Name:        p.Name,
			IsPublisher: p.IsPublisher,
		})
	}

	return resp, nil
}

func (s *Service) UserRejectRoom(ctx context.Context, uid string, room string) (interface{}, error) {
	roomInfo, _, err := s.getRedisUserRoom(ctx, room)
	if err != nil {
		return nil, err
	}

	if uid != roomInfo.RecipientID && uid != roomInfo.SenderID {
		return nil, code.Unauthorized
	}

	_, err = s.deleteRoom(ctx, room)
	if err != nil {
		return nil, err
	}

	msg := msgconfig.WsMsg{Uid: roomInfo.SenderID, Event: msgconfig.UserCallRejectEvent, Data: map[string]interface{}{
		"url":          s.livekitServer,
		"room":         room,
		"sender_id":    roomInfo.SenderID,
		"recipient_id": roomInfo.RecipientID,
	}}
	if err = s.publishServiceMessage(ctx, msg); err != nil {
		s.logger.Error("发送消息失败", zap.Error(err))
		return nil, err
	}

	return nil, nil
}

func (s *Service) UserLeaveRoom(ctx context.Context, uid, room string) (interface{}, error) {
	roomInfo, _, err := s.getRedisUserRoom(ctx, room)
	if err != nil {
		return nil, err
	}

	if uid != roomInfo.RecipientID && uid != roomInfo.SenderID {
		return nil, code.Unauthorized
	}
	return s.deleteRoom(ctx, room)
}

func (s *Service) deleteRoom(ctx context.Context, room string) (interface{}, error) {
	_, err := s.roomService.DeleteRoom(ctx, &livekit.DeleteRoomRequest{
		Room: room, // 房间名称
	})
	if err != nil && !strings.Contains(err.Error(), "room not found") {
		//if strings.Contains(err.Error(), "room not found") {
		//	return nil, code.LiveErrCallNotFound
		//}
		s.logger.Error("error deleting room", zap.Error(err))
		return nil, code.LiveErrLeaveCallFailed
	}

	if err = s.DeleteUserCallByRoomID(ctx, room); err != nil {
		s.logger.Error("删除房间信息失败", zap.Error(err))
	}

	return nil, nil
}

func (s *Service) DeleteUserCallByRoomID(ctx context.Context, roomID string) error {
	pattern := "liveUser:" + roomID + ":*"
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) == 0 {
		return nil
	}

	_, err = s.redisClient.Del(ctx, keys...).Result()
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) isEitherUserInCall(ctx context.Context, userID1, userID2 string) (bool, error) {
	is1, err := s.isInCall(ctx, userID1)
	if err != nil {
		return false, err
	}
	if is1 {
		return true, nil
	}
	is2, err := s.isInCall(ctx, userID2)
	if err != nil {
		return false, err
	}
	if is2 {
		return true, nil
	}
	return false, nil
}

func (s *Service) isInCall(ctx context.Context, id string) (bool, error) {
	pattern := "liveUser:*:" + id + ":*"
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return false, err
	}
	return len(keys) > 0, nil
}

func (s *Service) getRedisUserRoom(ctx context.Context, roomName string) (*model.UserRoomInfo, string, error) {
	keys, err := s.scanKeysWithPrefixAndContent(ctx, roomName)
	if err != nil {
		return nil, "", err
	}

	if len(keys) > 1 {
		return nil, "", code.LiveErrGetCallInfoFailed
	}

	if len(keys) == 0 {
		return nil, "", code.LiveErrCallNotFound
	}

	room, err := cache.GetKey(s.redisClient, keys[0])
	if err != nil {
		return nil, "", code.LiveErrCallNotFound.Reason(err)
	}

	data := &model.UserRoomInfo{}
	if err = json.Unmarshal([]byte(room.(string)), &data); err != nil {
		return nil, "", err
	}
	return data, keys[0], nil
}

func (s *Service) scanKeysWithPrefixAndContent(ctx context.Context, content string) ([]string, error) {
	pattern := liveUserPrefix + "*" + content + "*"
	var cursor uint64 = 0
	var keys []string
	for {
		result, _, err := s.redisClient.Scan(ctx, cursor, pattern, 10).Result()
		if err != nil {
			return nil, err
		}
		if len(result) == 0 {
			break
		}
		keys = append(keys, result...)
		cursor, _ = strconv.ParseUint(result[0], 10, 64)
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

func (s *Service) deleteRedisLiveUser(ctx context.Context, userID string) error {
	pattern := "liveUser:*:" + userID + ":*"
	keys, err := s.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) == 0 {
		return nil
	}
	_, err = s.redisClient.Del(ctx, keys...).Result()
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) GetJoinToken(ctx context.Context, userName string, room string) (string, error) {
	at := auth.NewAccessToken(s.liveApiKey, s.liveApiSecret)
	grant := &auth.VideoGrant{
		RoomCreate: true,
		RoomList:   true,
		RoomAdmin:  true,
		RoomJoin:   true,
		Room:       room,
	}
	at.AddGrant(grant).SetIdentity("admin").SetValidFor(s.liveTimeout).SetName(userName)
	jwt, err := at.ToJWT()
	if err != nil {
		s.logger.Error("Failed to generate JWT", zap.Error(err))
		return "", err
	}

	return jwt, nil
}

func (s *Service) publishServiceMessage(ctx context.Context, msg msgconfig.WsMsg) error {
	return s.mqClient.PublishServiceMessage(msg_queue.LiveUserService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
}
