package service

import (
	"context"
	"encoding/json"
	"errors"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	"github.com/cossim/coss-server/internal/live/api/dto"
	"github.com/cossim/coss-server/internal/live/api/model"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/utils"
	any2 "github.com/golang/protobuf/ptypes/any"
	"github.com/google/uuid"
	"github.com/livekit/protocol/livekit"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

func (s *Service) CreateGroupCall(ctx context.Context, uid string, gid uint32, member []string, option dto.CallOption) (*dto.GroupCallResponse, error) {
	groupID := strconv.FormatUint(uint64(gid), 10)

	// 检查群组和用户是否已经在通话中
	if isInCall, err := s.isEitherGroupInCall(ctx, groupID); err != nil {
		return nil, err
	} else if isInCall {
		return nil, code.LiveErrAlreadyInCall
	}
	if isInCall, err := s.isUserInCall(ctx, uid); err != nil {
		return nil, err
	} else if isInCall {
		return nil, code.LiveErrAlreadyInCall
	}

	group, err := s.groupService.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{Gid: gid})
	if err != nil {
		s.logger.Error("create group call failed", zap.Error(err))
		return nil, err
	}

	if group.Status != groupgrpcv1.GroupStatus_GROUP_STATUS_NORMAL {
		return nil, code.GroupErrGroupStatusNotAvailable
	}

	//rel, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: gid, UserId: uid})
	//if err != nil {
	//	return nil, err
	//}
	//
	//if rel.MuteEndTime > 0 {
	//	return nil, code.GroupErrUserIsMuted
	//}

	type MemberInfo struct {
		UserID   string
		UserName string
	}

	rels, err := s.relationGroupService.GetBatchGroupRelation(ctx, &relationgrpcv1.GetBatchGroupRelationRequest{GroupId: gid, UserIds: member})
	if err != nil {
		return nil, err
	}

	if len(member) != len(rels.GroupRelationResponses) {
		return nil, code.RelationGroupErrNotInGroup
	}

	roomName := uuid.New().String()
	_, err = s.roomService.CreateRoom(ctx, &livekit.CreateRoomRequest{
		Name:            roomName,
		EmptyTimeout:    uint32(s.liveTimeout.Seconds()), // 空闲时间
		MaxParticipants: 256,
	})
	if err != nil {
		s.logger.Error("创建通话失败", zap.Error(err))
		return nil, err
	}

	groupRoom, err := (&model.Room{
		Room: roomName,
	}).ToJSONString()
	if err != nil {
		return nil, err
	}
	if err = s.redisClient.SetKey(liveGroupPrefix+groupID, groupRoom, 0); err != nil {
		s.logger.Error("保存群聊通话记录失败", zap.Error(err), zap.Int("gid", int(gid)))
		return nil, err
	}

	roomInfo := &model.RoomInfo{
		Room:            roomName,
		Type:            model.GroupRoomType,
		SenderID:        uid,
		MaxParticipants: 256,
		Participants:    map[string]*model.ActiveParticipant{},
	}

	participants := make([]MemberInfo, 0)
	participants = append(participants, MemberInfo{UserID: uid})
	for _, v := range rels.GroupRelationResponses {
		// 提取未被禁言的成员信息
		//if v.MuteEndTime == 0 {
		participants = append(participants, MemberInfo{UserID: v.UserId})
		//}
	}

	// 获取成员的名称和其他信息
	for i := range participants {
		memberUser, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: participants[i].UserID})
		if err != nil {
			s.logger.Error("获取用户信息失败", zap.Error(err), zap.String("uid", participants[i].UserID))
			continue
		}
		if memberUser.Status != usergrpcv1.UserStatus_USER_STATUS_NORMAL {
			s.logger.Error("用户状态异常", zap.String(memberUser.UserId, memberUser.Status.String()))
			continue
		}
		roomInfo.Participants[participants[i].UserID] = &model.ActiveParticipant{}
		participants[i].UserName = memberUser.NickName
	}

	roomStr, err := roomInfo.ToJSONString()
	if err != nil {
		return nil, err
	}
	if err = s.redisClient.SetKey(liveRoomPrefix+roomName, roomStr, 0); err != nil {
		s.logger.Error("保存房间信息失败", zap.Error(err))
		return nil, err
	}

	for i := range participants {
		//userRoom, err := (&model.Room{
		//	Room: roomName,
		//}).roomStr()
		//if err != nil {
		//	return nil, err
		//}
		//if err = s.redisClient.SetKey(liveUserPrefix+participants[i].UserID, userRoom, s.liveTimeout); err != nil {
		//	s.logger.Error("保存用户通话记录失败", zap.Error(err), zap.String("uid", participants[i].UserID))
		//}

		if participants[i].UserID == uid {
			continue
		}
		data := map[string]interface{}{
			"url":          s.livekitServer,
			"group_id":     gid,
			"room":         roomName,
			"sender_id":    uid,
			"recipient_id": participants[i].UserID,
			"option":       option,
		}

		bytes, err := utils.StructToBytes(data)
		if err != nil {
			return nil, err
		}

		msg := &pushgrpcv1.WsMsg{Uid: participants[i].UserID, Event: pushgrpcv1.WSEventType_GroupCallReqEvent, Data: &any2.Any{Value: bytes}}
		//if err = s.publishServiceMessage(ctx, msg); err != nil {
		//	s.logger.Error("发送消息失败", zap.Error(err), zap.String("room", roomName), zap.String("uid", participants[i].UserID))
		//}
		toBytes, err := utils.StructToBytes(msg)
		if err != nil {
			return nil, err
		}
		_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{
			Type: pushgrpcv1.Type_Ws,
			Data: toBytes,
		})
		if err != nil {
			s.logger.Error("发送消息失败", zap.Error(err))
		}
	}

	s.logger.Info("创建群聊通话", zap.String("uid", uid), zap.String("room", roomName), zap.String("server", s.livekitServer))

	return &dto.GroupCallResponse{
		Url: s.livekitServer,
	}, nil
}

func (s *Service) GroupJoinRoom(ctx context.Context, gid uint32, uid, driverId string) (*dto.GroupJoinResponse, error) {
	_, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: gid, UserId: uid})
	if err != nil {
		s.logger.Error("获取用户群组关系失败", zap.Error(err))
		return nil, err
	}

	is, err := s.isUserInCall(ctx, uid)
	if err != nil {
		return nil, err
	}
	if is {
		return nil, code.LiveErrAlreadyInCall
	}

	key := liveGroupPrefix + strconv.FormatUint(uint64(gid), 10)
	room, err := s.getGroupRedisRoom(ctx, key)
	if err != nil {
		s.logger.Error("获取群聊房间信息失败", zap.Error(err))
		return nil, code.LiveErrGetCallInfoFailed
	}

	if err = s.checkGroupRoom(ctx, room, uid, room.Room); err != nil {
		return nil, err
	}

	room.NumParticipants++
	room.Participants[uid] = &model.ActiveParticipant{
		Connected: true,
	}
	roomStr, err := room.ToJSONString()
	if err != nil {
		return nil, code.LiveErrJoinCallFailed
	}
	if err = s.redisClient.SetKey(liveRoomPrefix+room.Room, roomStr, 0); err != nil {
		s.logger.Error("更新房间信息失败", zap.Error(err))
		return nil, err
	}

	userRoom, err := (&model.Room{
		Room: room.Room,
	}).ToJSONString()
	if err != nil {
		return nil, err
	}
	if err = s.redisClient.SetKey(liveUserPrefix+uid, userRoom, 0); err != nil {
		s.logger.Error("保存用户通话记录失败", zap.Error(err), zap.String("uid", uid))
	}

	user, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: uid})
	if err != nil {
		return nil, err
	}

	var token string
	if uid == room.SenderID {
		token, err = s.GetAdminJoinToken(ctx, room.Room, user.NickName, uid)
	} else {
		token, err = s.GetUserJoinToken(ctx, room.Room, user.NickName, uid)
	}
	if err != nil {
		return nil, code.LiveErrJoinCallFailed
	}

	for k := range room.Participants {
		if k == uid {
			continue
		}
		bytes, err := utils.StructToBytes(map[string]interface{}{
			"room":         room.Room,
			"sender_id":    room.SenderID,
			"recipient_id": uid,
			"driver_id":    driverId,
		})
		if err != nil {
			return nil, err
		}

		msg := &pushgrpcv1.WsMsg{Uid: k, Event: pushgrpcv1.WSEventType_GroupCallAcceptEvent, Data: &any2.Any{Value: bytes}}
		toBytes, err := utils.StructToBytes(msg)
		if err != nil {
			return nil, err
		}

		_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{
			Type: pushgrpcv1.Type_Ws,
			Data: toBytes,
		})
		if err != nil {
			s.logger.Error("发送消息失败", zap.Error(err))
		}
	}

	s.logger.Info("加入群聊通话", zap.Int("gid", int(gid)), zap.String("uid", uid), zap.String("room", roomStr), zap.String("livekit", s.livekitServer))

	return &dto.GroupJoinResponse{
		Url:   s.livekitServer,
		Token: token,
	}, nil
}

func (s *Service) GroupShowRoom(ctx context.Context, gid uint32, uid string) (*dto.GroupShowResponse, error) {
	_, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: gid, UserId: uid})
	if err != nil {
		s.logger.Error("获取用户群组关系失败", zap.Error(err))
		return nil, err
	}

	key := liveGroupPrefix + strconv.FormatUint(uint64(gid), 10)
	room, err := s.getGroupRedisRoom(ctx, key)
	if err != nil {
		s.logger.Error("获取群聊房间信息失败", zap.Error(err))
		return nil, code.LiveErrGetCallInfoFailed
	}

	livekitRoom, err := s.getLivekitRoom(ctx, room.Room)
	if err != nil || livekitRoom == nil {
		if _, err := s.deleteGroupRoom(ctx, room.Room, key); err != nil {
			s.logger.Error("删除群聊房间信息失败", zap.Error(err))
			return nil, code.LiveErrGetCallInfoFailed
		}
	}

	if err != nil {
		return nil, err
	}

	resp := &dto.GroupShowResponse{
		GroupID:         gid,
		NumParticipants: livekitRoom.NumParticipants,
		MaxParticipants: livekitRoom.MaxParticipants,
		StartAt:         livekitRoom.CreationTime,
		Duration:        0,
		Room:            livekitRoom.Name,
		Participant:     make([]*dto.ParticipantInfo, 0),
	}
	res, err := s.roomService.ListParticipants(ctx, &livekit.ListParticipantsRequest{
		Room: room.Room, // 房间名称
	})
	if err != nil {
		s.logger.Error("获取通话信息失败", zap.Error(err))
		return nil, code.LiveErrGetCallInfoFailed
	}

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

// GroupRejectRoom handles the rejection of a group call by a user.
// It checks the user's group relation, ensures the user is not the sender of the call,
// and that the user is a participant in the call and not already connected.
// Then it sends a rejection message to all participants in the call.
func (s *Service) GroupRejectRoom(ctx context.Context, gid uint32, uid string) (interface{}, error) {
	// Check the user's group relation
	_, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: gid, UserId: uid})
	if err != nil {
		s.logger.Error("Failed to retrieve user's group relation", zap.Error(err))
		return nil, err
	}

	key := liveGroupPrefix + strconv.FormatUint(uint64(gid), 10)
	// Retrieve room information
	roomInfo, err := s.getGroupRedisRoom(ctx, key)
	if err != nil {
		return nil, err
	}

	// Check if the user is the sender of the call
	if roomInfo.SenderID == uid {
		s.logger.Warn("Cannot reject a call initiated by the user", zap.String("uid", uid), zap.Uint32("gid", gid))
		return nil, code.LiveErrRejectCallFailed
	}

	// Check if the user is a participant in the call
	pp, ok := roomInfo.Participants[uid]
	if !ok {
		s.logger.Warn("User is not a participant in the call", zap.String("uid", uid), zap.Uint32("gid", gid))
		return nil, code.LiveErrRejectCallFailed
	}

	// Check if the user is already connected to the call
	if pp.Connected {
		s.logger.Warn("User is already connected to the call", zap.String("uid", uid), zap.Uint32("gid", gid))
		return nil, code.LiveErrRejectCallFailed
	}

	// Send rejection message to all participants in the call

	for id := range roomInfo.Participants {
		data := map[string]interface{}{
			"sender_id":    roomInfo.SenderID,
			"recipient_id": id,
		}
		bytes, err := utils.StructToBytes(data)
		if err != nil {
			return nil, err
		}
		msg := &pushgrpcv1.WsMsg{
			Uid:   id,
			Event: pushgrpcv1.WSEventType_GroupCallRejectEvent,
			Data:  &any2.Any{Value: bytes},
		}

		toBytes, err := utils.StructToBytes(msg)
		if err != nil {
			return nil, err
		}

		_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{
			Type: pushgrpcv1.Type_Ws,
			Data: toBytes,
		})
		if err != nil {
			s.logger.Error("发送消息失败", zap.Error(err))
		}

		//if err = s.publishServiceMessage(ctx, msg); err != nil {
		//	s.logger.Error("Failed to send rejection message", zap.Error(err))
		//}
	}

	return nil, nil
}

func (s *Service) GroupLeaveRoom(ctx context.Context, gid uint32, uid string, force bool) (interface{}, error) {
	s.logger.Info("用户请求离开群聊通话", zap.Int("gid", int(gid)), zap.String("uid", uid))
	_, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: gid, UserId: uid})
	if err != nil {
		s.logger.Error("获取用户群组关系失败", zap.Error(err))
		return nil, err
	}

	key := liveGroupPrefix + strconv.FormatUint(uint64(gid), 10)
	roomInfo, err := s.getGroupRedisRoom(ctx, key)
	if err != nil {
		return nil, err
	}

	if roomInfo.SenderID != uid && force {
		return nil, code.Forbidden
	}

	pp, ok := roomInfo.Participants[uid]
	if !ok {
		return nil, code.Forbidden
	}

	if !pp.Connected {
		return nil, code.LiveErrUserNotInCall
	}

	// 如果是最后一个离开的用户，那么删除房间
	if force || roomInfo.NumParticipants-1 == 0 || roomInfo.NumParticipants == 0 {
		if err := s.redisClient.DelKey(liveUserPrefix + uid); err != nil {
			s.logger.Error("删除用户通话记录失败", zap.String("uid", uid), zap.Error(err))
			return nil, err
		}
		_, err = s.deleteGroupRoom(ctx, roomInfo.Room, key)
		if err != nil {
			return nil, err
		}
	} else {
		r1, err := roomInfo.ToJSONString()
		if err != nil {
			return nil, err
		}
		s.logger.Info("删除用户前信息", zap.String("info", r1))
		delete(roomInfo.Participants, uid)
		if roomInfo.NumParticipants > 0 {
			roomInfo.NumParticipants--
		}
		roomStr, err := roomInfo.ToJSONString()
		if err != nil {
			return nil, err
		}

		if err := s.redisClient.DelKey(liveUserPrefix + uid); err != nil {
			s.logger.Error("删除用户通话记录失败", zap.String("uid", uid), zap.Error(err))
			return nil, err
		}

		if err = s.redisClient.SetKey(liveRoomPrefix+roomInfo.Room, roomStr, 0); err != nil {
			s.logger.Error("更新房间信息失败", zap.Error(err))
			return nil, err
		}
		s.logger.Info("删除后用户信息", zap.String("info", roomStr))
	}

	for k := range roomInfo.Participants {
		if k == uid {
			continue
		}
		data := map[string]interface{}{
			"room": roomInfo.Room,
			"uid":  uid,
			"gid":  gid,
		}
		bytes, err := utils.StructToBytes(data)
		if err != nil {
			return nil, err
		}
		msg := pushgrpcv1.WsMsg{Uid: k, Event: pushgrpcv1.WSEventType_UserLeaveGroupCallEvent, Data: &any2.Any{Value: bytes}}
		//if err = s.publishServiceMessage(ctx, msg); err != nil {
		//	s.logger.Error("发送消息失败", zap.Error(err))
		//	return nil, err
		//}
		toBytes, err := utils.StructToBytes(&msg)
		if err != nil {
			return nil, err
		}
		_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{
			Type: pushgrpcv1.Type_Ws,
			Data: toBytes,
		})
		if err != nil {
			s.logger.Error("发送消息失败", zap.Error(err))
		}
	}

	s.logger.Info("离开群聊通话", zap.Int("gid", int(gid)), zap.String("room", roomInfo.Room), zap.String("房间创建者", roomInfo.SenderID), zap.String("离开用户", uid))

	return nil, nil
}

func (s *Service) getGroupRedisRoom(ctx context.Context, key string) (*model.RoomInfo, error) {
	room, err := s.redisClient.GetKey(key)
	if err != nil {
		return nil, err
	}
	d1 := &model.Room{}
	if err = json.Unmarshal([]byte(room.(string)), &d1); err != nil {
		return nil, err
	}

	roomInfo, err := s.redisClient.GetKey(liveRoomPrefix + d1.Room)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			if err := s.redisClient.DelKey(key); err != nil {
				s.logger.Error("获取群聊通话信息失败后删除房间失败", zap.Error(err))
			}
			return nil, code.LiveErrCallNotFound
		}
		return nil, err
	}
	d2 := &model.RoomInfo{}
	if err = json.Unmarshal([]byte(roomInfo.(string)), &d2); err != nil {
		return nil, err
	}

	return d2, nil
}

func (s *Service) isEitherGroupInCall(ctx context.Context, gid string) (bool, error) {
	key := liveGroupPrefix + gid
	exists, err := s.redisClient.ExistsKey(key)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (s *Service) deleteGroupRoom(ctx context.Context, room string, key string) (interface{}, error) {
	_, err := s.roomService.DeleteRoom(ctx, &livekit.DeleteRoomRequest{
		Room: room,
	})
	if err != nil && !strings.Contains(err.Error(), "room not found") {
		s.logger.Error("error deleting room", zap.Error(err))
		return nil, code.LiveErrLeaveCallFailed
	}

	_, err = s.redisClient.Client.Del(ctx, key).Result()
	if err != nil {
		s.logger.Error("删除房间信息失败", zap.Error(err))
	}

	if err = s.redisClient.DelKey(liveRoomPrefix + room); err != nil {
		s.logger.Error("删除房间失败", zap.Error(err), zap.String("room", room))
	}

	return nil, nil
}

func (s *Service) checkGroupRoom(ctx context.Context, roomInfo *model.RoomInfo, uid, roomName string) error {
	//if uid != roomInfo.RecipientID && uid != roomInfo.SenderID {
	//	return code.Unauthorized
	//}

	if roomInfo.NumParticipants+1 > roomInfo.MaxParticipants {
		return code.LiveErrMaxParticipantsExceeded
	}

	if ps, ok := roomInfo.Participants[uid]; ok && ps.Connected {
		return code.LiveErrAlreadyInCall
	}

	room, err := s.getLivekitRoom(ctx, roomName)
	if err != nil {
		return err
	}

	if room.NumParticipants+1 > room.MaxParticipants {
		return code.LiveErrMaxParticipantsExceeded
	}

	return nil
}
