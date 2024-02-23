package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cossim/coss-server/interface/live/api/dto"
	"github.com/cossim/coss-server/interface/live/api/model"
	"github.com/cossim/coss-server/pkg/cache"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/msg_queue"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	usergrpcv1 "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/google/uuid"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	"go.uber.org/zap"
	"strconv"
	"strings"
	"time"
)

func (s *Service) CreateUserCall(ctx context.Context, senderID, recipientID string) (*dto.UserCallResponse, error) {
	inCall, err := s.isEitherUserInCall(ctx, senderID, recipientID)
	if err != nil {
		s.logger.Error("获取通话状态失败", zap.Error(err))
		return nil, code.LiveErrCreateCallFailed
	}
	if inCall {
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

	room, err := s.roomService.CreateRoom(ctx, &livekit.CreateRoomRequest{
		Name:            uuid.New().String(),
		EmptyTimeout:    uint32(s.liveTimeout.Seconds()), // 空闲时间
		MaxParticipants: 2,
	})
	if err != nil {
		s.logger.Error("创建通话失败", zap.Error(err))
		return nil, err
	}

	fmt.Println("CreateUserCall room => ", room)

	roomInfo, err := (&model.UserRoomInfo{
		Room:            room.Name,
		SenderID:        senderID,
		RecipientID:     recipientID,
		MaxParticipants: 2,
		Participants: map[string]*model.ActiveParticipant{
			senderID: {
				Connected: false,
			},
			recipientID: {
				Connected: false,
			},
		},
	}).ToJSONString()
	if err != nil {
		return nil, err
	}
	if err = cache.SetKey(s.redisClient, liveUserPrefix+senderID, roomInfo, s.liveTimeout); err != nil {
		s.logger.Error("保存房间信息失败", zap.Error(err))
		return nil, err
	}
	if err = cache.SetKey(s.redisClient, liveUserPrefix+recipientID, roomInfo, s.liveTimeout); err != nil {
		s.logger.Error("保存房间信息失败", zap.Error(err))
		return nil, err
	}

	msg := constants.WsMsg{
		Uid:   recipientID,
		Event: constants.UserCallReqEvent,
		Data: map[string]interface{}{
			"url":          s.livekitServer,
			"sender_id":    senderID,
			"recipient_id": recipientID,
		}}
	if err = s.publishServiceMessage(ctx, msg); err != nil {
		s.logger.Error("发送消息失败", zap.Error(err))
		return nil, err
	}

	return &dto.UserCallResponse{
		Url: s.livekitServer,
		//Token: senderToken,
	}, nil
}

func (s *Service) checkUserRoom(ctx context.Context, roomInfo *model.UserRoomInfo, uid, roomName string) error {
	if uid != roomInfo.RecipientID && uid != roomInfo.SenderID {
		return code.Unauthorized
	}

	if _, ok := roomInfo.Participants[uid]; ok && roomInfo.Participants[uid].Connected {
		return code.LiveErrAlreadyInCall
	}

	if roomInfo.NumParticipants+1 > roomInfo.MaxParticipants {
		return code.LiveErrMaxParticipantsExceeded
	}

	room, err := s.getLivekitRoom(ctx, roomName, nil)
	if err != nil {
		return err
	}

	if room.NumParticipants+1 > room.MaxParticipants {
		return code.LiveErrMaxParticipantsExceeded
	}

	return nil
}

func (s *Service) UserJoinRoom(ctx context.Context, uid string) (*dto.UserJoinResponse, error) {
	room, err := s.getRedisUserRoom(ctx, uid)
	if err != nil {
		s.logger.Error("获取用户房间信息失败", zap.Error(err))
		return nil, code.LiveErrJoinCallFailed
	}

	if err = s.checkUserRoom(ctx, room, uid, room.Room); err != nil {
		return nil, err
	}

	room.NumParticipants++
	room.Participants[uid] = &model.ActiveParticipant{
		Connected: true,
	}
	ToJSONString, err := room.ToJSONString()
	if err != nil {
		s.logger.Error("更新房间信息失败", zap.Error(err))
		return nil, code.LiveErrJoinCallFailed
	}
	if err = cache.SetKey(s.redisClient, liveUserPrefix+uid, ToJSONString, 365*24*time.Hour); err != nil {
		s.logger.Error("更新房间信息失败", zap.Error(err))
		return nil, err
	}

	user, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: uid})
	if err != nil {
		return nil, err
	}

	token, err := s.GetAdminJoinToken(ctx, room.Room, user.NickName)
	if err != nil {
		return nil, code.LiveErrJoinCallFailed
	}

	for k := range room.Participants {
		if k == uid && uid == room.SenderID {
			continue
		}
		msg := constants.WsMsg{Uid: k, Event: constants.UserCallAcceptEvent, Data: map[string]interface{}{
			"room":         room.Room,
			"sender_id":    room.SenderID,
			"recipient_id": room.RecipientID,
		}}
		if err = s.publishServiceMessage(ctx, msg); err != nil {
			s.logger.Error("发送消息失败", zap.Error(err))
		}
	}

	s.logger.Info("UserJoinRoom", zap.String("room", room.Room), zap.String("SenderID", room.SenderID), zap.String("RecipientID", room.RecipientID))

	return &dto.UserJoinResponse{
		Url:   s.livekitServer,
		Token: token,
	}, nil
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
	//	Connected: true,
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

func (s *Service) getLivekitRoom(ctx context.Context, room string, callback func(ctx context.Context, room string)) (*livekit.Room, error) {
	rooms, err := s.roomService.ListRooms(ctx, &livekit.ListRoomsRequest{Names: []string{room}})
	if err != nil {
		s.logger.Error("获取房间信息失败", zap.Error(err))
		return nil, code.LiveErrGetCallInfoFailed
	}

	if len(rooms.Rooms) == 0 {
		if callback != nil {
			callback(ctx, room)
		}
		return nil, code.LiveErrCallNotFound
	}

	return rooms.Rooms[0], nil
}

func (s *Service) GetUserRoom(ctx context.Context, uid string) (*dto.UserShowResponse, error) {
	room, err := s.getRedisUserRoom(ctx, uid)
	if err != nil {
		return nil, err
	}

	fmt.Println("GetUserRoom room => ", room)

	if uid != room.RecipientID && uid != room.SenderID {
		return nil, code.Unauthorized
	}

	livekitRoom, err := s.getLivekitRoom(ctx, room.Room, func(ctx context.Context, room string) {
		s.deleteUserRoom(ctx, room, uid)
	})
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
		Room: room.Room, // 房间名称
	})
	if err != nil {
		s.logger.Error("获取通话信息失败", zap.Error(err))
		return nil, code.LiveErrGetCallInfoFailed
	}

	for _, p := range res.Participants {
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

func (s *Service) UserRejectRoom(ctx context.Context, uid string) (interface{}, error) {
	room, err := s.getRedisUserRoom(ctx, uid)
	if err != nil {
		return nil, err
	}

	if uid != room.RecipientID && uid != room.SenderID {
		return nil, code.Unauthorized
	}

	rel, err := s.relUserClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: uid, FriendId: room.SenderID})
	if err != nil {
		return nil, err
	}

	if rel.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
		return nil, code.RelationUserErrFriendRelationNotFound
	}

	_, err = s.deleteUserRoom(ctx, room.Room, uid)
	if err != nil {
		return nil, err
	}

	msg := constants.WsMsg{Uid: room.SenderID, Event: constants.UserCallRejectEvent, Data: map[string]interface{}{
		"url":          s.livekitServer,
		"sender_id":    room.SenderID,
		"recipient_id": room.RecipientID,
	}}
	if err = s.publishServiceMessage(ctx, msg); err != nil {
		s.logger.Error("发送消息失败", zap.Error(err))
	}

	s.logger.Info("UserRejectRoom", zap.String("room", room.Room), zap.String("SenderID", room.SenderID), zap.String("RecipientID", room.RecipientID))

	return nil, nil
}

func (s *Service) UserLeaveRoom(ctx context.Context, uid string) (interface{}, error) {
	room, err := s.getRedisUserRoom(ctx, uid)
	if err != nil {
		return nil, err
	}

	if uid != room.RecipientID && uid != room.SenderID {
		return nil, code.Unauthorized
	}

	for k := range room.Participants {
		if k == uid {
			continue
		}
		msg := constants.WsMsg{Uid: k, Event: constants.UserCallEndEvent, Data: map[string]interface{}{
			"room":         room.Room,
			"sender_id":    room.SenderID,
			"recipient_id": room.RecipientID,
		}}
		if err = s.publishServiceMessage(ctx, msg); err != nil {
			s.logger.Error("发送消息失败", zap.Error(err))
		}
	}

	s.logger.Info("UserLeaveRoom", zap.String("room", room.Room), zap.String("SenderID", room.SenderID), zap.String("RecipientID", room.RecipientID))

	return s.deleteUserRoom(ctx, room.Room, uid)
}

func (s *Service) deleteUserRoom(ctx context.Context, room, uid string) (interface{}, error) {
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

	userRoom, err := s.getRedisUserRoom(ctx, uid)
	if err != nil {
		return nil, err
	}

	if err = cache.DelKey(s.redisClient, liveUserPrefix+userRoom.SenderID); err != nil {
		s.logger.Error("删除房间信息失败", zap.Error(err))
	}

	if err = cache.DelKey(s.redisClient, liveUserPrefix+userRoom.RecipientID); err != nil {
		s.logger.Error("删除房间信息失败", zap.Error(err))
	}

	//if err = s.deleteRedisRoomByPrefix(ctx, liveUserPrefix, room); err != nil {
	//	s.logger.Error("删除房间信息失败", zap.Error(err))
	//}

	return nil, nil
}

func (s *Service) deleteRedisRoomByPrefix(ctx context.Context, prefix, roomID string) error {
	pattern := prefix + roomID + ":*"
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

func (s *Service) deleteGroupCallByRoomID(ctx context.Context, roomID string) error {
	pattern := liveGroupPrefix + roomID + ":*"
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
	return is2, nil
}

func (s *Service) isInCall(ctx context.Context, id string) (bool, error) {
	//pattern := liveUserPrefix + "*:" + id + ":*"
	//keys, err := s.redisClient.Keys(ctx, pattern).Result()
	//if err != nil {
	//	return false, err
	//}
	//return len(keys) > 0, nil
	pattern := liveUserPrefix + id
	count, err := s.redisClient.Exists(ctx, pattern).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Service) getRedisUserRoom(ctx context.Context, uid string) (*model.UserRoomInfo, error) {
	room, err := cache.GetKey(s.redisClient, liveUserPrefix+uid)
	if err != nil {
		return nil, code.LiveErrCallNotFound.Reason(err)
	}
	data := &model.UserRoomInfo{}
	if err = json.Unmarshal([]byte(room.(string)), &data); err != nil {
		return nil, err
	}
	return data, nil
	//return s.getRedisRoomWithPrefix(ctx, liveUserPrefix, "*:"+uid+"*", out)
}

// GetRedisObjectWithCondition retrieves an object from Redis cache based on a condition and unmarshals it into the provided output interface.
func (s *Service) getRedisRoomWithPrefix(ctx context.Context, prefix, key string, out interface{}) (string, error) {
	pattern := prefix + key
	s.logger.Info("scanKeysWithPrefixAndContent", zap.String("pattern", pattern))
	keys, err := s.scanKeysWithPrefixAndContent(ctx, pattern)
	if err != nil {
		return "", err
	}
	if len(keys) > 1 {
		return "", code.LiveErrGetCallInfoFailed
	}
	if len(keys) == 0 {
		return "", code.LiveErrCallNotFound
	}
	room, err := cache.GetKey(s.redisClient, keys[0])
	if err != nil {
		return "", code.LiveErrCallNotFound.Reason(err)
	}
	if err = json.Unmarshal([]byte(room.(string)), &out); err != nil {
		return "", err
	}
	return keys[0], nil
}

func (s *Service) scanKeysWithPrefixAndContent(ctx context.Context, pattern string) ([]string, error) {
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
	pattern := liveUserPrefix + "*:" + userID + ":*"
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

func (s *Service) GetUserJoinToken(ctx context.Context, room, userName string) (string, error) {
	at := auth.NewAccessToken(s.liveApiKey, s.liveApiSecret)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     room,
	}
	at.AddGrant(grant).SetIdentity(userName).SetValidFor(s.liveTimeout).SetName(userName)
	jwt, err := at.ToJWT()
	if err != nil {
		s.logger.Error("Failed to generate JWT", zap.Error(err))
		return "", err
	}

	return jwt, nil
}

func (s *Service) GetAdminJoinToken(ctx context.Context, room, userName string) (string, error) {
	at := auth.NewAccessToken(s.liveApiKey, s.liveApiSecret)
	grant := &auth.VideoGrant{
		RoomCreate: true,
		RoomList:   true,
		RoomAdmin:  true,
		RoomJoin:   true,
		Room:       room,
	}
	at.AddGrant(grant).SetIdentity(userName).SetValidFor(s.liveTimeout).SetName(userName)
	jwt, err := at.ToJWT()
	if err != nil {
		s.logger.Error("Failed to generate JWT", zap.Error(err))
		return "", err
	}

	return jwt, nil
}

func (s *Service) publishServiceMessage(ctx context.Context, msg constants.WsMsg) error {
	return s.mqClient.PublishServiceMessage(msg_queue.LiveUserService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
}
