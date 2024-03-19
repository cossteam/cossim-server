package service

import (
	"context"
	"encoding/json"
	"github.com/cossim/coss-server/interface/live/api/dto"
	"github.com/cossim/coss-server/interface/live/api/model"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/google/uuid"
	"github.com/livekit/protocol/livekit"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

func (s *Service) CreateGroupCall(ctx context.Context, uid string, gid uint32, member []string) (*dto.GroupCallResponse, error) {
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

	group, err := s.groupClient.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{Gid: gid})
	if err != nil {
		s.logger.Error("create group call failed", zap.Error(err))
		return nil, err
	}

	if group.Status != groupgrpcv1.GroupStatus_GROUP_STATUS_NORMAL {
		return nil, code.GroupErrGroupStatusNotAvailable
	}

	//rel, err := s.relGroupClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: gid, UserId: uid})
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

	rels, err := s.relGroupClient.GetBatchGroupRelation(ctx, &relationgrpcv1.GetBatchGroupRelationRequest{GroupId: gid, UserIds: member})
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
	if err = s.redisClient.SetKey(liveGroupPrefix+groupID, groupRoom, s.liveTimeout); err != nil {
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
		memberUser, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: participants[i].UserID})
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

	ToJSONString, err := roomInfo.ToJSONString()
	if err != nil {
		return nil, err
	}
	if err = s.redisClient.SetKey(liveRoomPrefix+roomName, ToJSONString, s.liveTimeout); err != nil {
		s.logger.Error("保存房间信息失败", zap.Error(err))
		return nil, err
	}

	for i := range participants {
		//userRoom, err := (&model.Room{
		//	Room: roomName,
		//}).ToJSONString()
		//if err != nil {
		//	return nil, err
		//}
		//if err = s.redisClient.SetKey(liveUserPrefix+participants[i].UserID, userRoom, s.liveTimeout); err != nil {
		//	s.logger.Error("保存用户通话记录失败", zap.Error(err), zap.String("uid", participants[i].UserID))
		//}

		if participants[i].UserID == uid {
			continue
		}
		msg := constants.WsMsg{Uid: participants[i].UserID, Event: constants.GroupCallReqEvent, Data: map[string]interface{}{
			"url":          s.livekitServer,
			"group_id":     gid,
			"room":         roomName,
			"sender_id":    uid,
			"recipient_id": participants[i].UserID,
		}}
		if err = s.publishServiceMessage(ctx, msg); err != nil {
			s.logger.Error("发送消息失败", zap.Error(err), zap.String("room", roomName), zap.String("uid", participants[i].UserID))
		}
	}

	s.logger.Info("创建群聊通话", zap.String("uid", uid), zap.String("room", roomName), zap.String("url", s.livekitServer))

	return &dto.GroupCallResponse{
		Url: s.livekitServer,
	}, nil
}

func (s *Service) GroupJoinRoom(ctx context.Context, gid uint32, uid string) (*dto.GroupJoinResponse, error) {
	_, err := s.relGroupClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: gid, UserId: uid})
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
	ToJSONString, err := room.ToJSONString()
	if err != nil {
		return nil, code.LiveErrJoinCallFailed
	}
	if err = s.redisClient.SetKey(key, ToJSONString, 0); err != nil {
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

	user, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: uid})
	if err != nil {
		return nil, err
	}

	var token string
	if uid == room.SenderID {
		token, err = s.GetAdminJoinToken(ctx, room.Room, user.NickName)
	} else {
		token, err = s.GetUserJoinToken(ctx, room.Room, user.NickName)
	}
	if err != nil {
		s.logger.Error("获取用户加入房间token失败", zap.Error(err))
		return nil, code.LiveErrJoinCallFailed
	}

	s.logger.Info("加入群聊通话", zap.String("uid", uid), zap.String("room", room.Room), zap.String("url", s.livekitServer))

	return &dto.GroupJoinResponse{
		Url:   s.livekitServer,
		Token: token,
	}, nil
}

func (s *Service) GroupShowRoom(ctx context.Context, gid uint32, uid string) (*dto.GroupShowResponse, error) {
	_, err := s.relGroupClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: gid, UserId: uid})
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
	_, err := s.relGroupClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: gid, UserId: uid})
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
		msg := constants.WsMsg{
			Uid:   id,
			Event: constants.GroupCallRejectEvent,
			Data: map[string]interface{}{
				"sender_id":    roomInfo.SenderID,
				"recipient_id": id,
			},
		}
		if err = s.publishServiceMessage(ctx, msg); err != nil {
			s.logger.Error("Failed to send rejection message", zap.Error(err))
		}
	}

	return nil, nil
}

func (s *Service) GroupLeaveRoom(ctx context.Context, gid uint32, uid string, force bool) (interface{}, error) {
	_, err := s.relGroupClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: gid, UserId: uid})
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
		_, err = s.deleteGroupRoom(ctx, roomInfo.Room, key)
		if err != nil {
			return nil, err
		}
	} else {
		delete(roomInfo.Participants, uid)
		if roomInfo.NumParticipants > 0 {
			roomInfo.NumParticipants--
		}
		ToJSONString, err := roomInfo.ToJSONString()
		if err != nil {
			return nil, err
		}
		if err = s.redisClient.SetKey(key, ToJSONString, 0); err != nil {
			s.logger.Error("更新房间信息失败", zap.Error(err))
			return nil, err
		}
	}

	for k := range roomInfo.Participants {
		if k == uid {
			continue
		}
		msg := constants.WsMsg{Uid: k, Event: constants.GroupCallEndEvent, Data: map[string]interface{}{
			"sender_id":    roomInfo.SenderID,
			"recipient_id": k,
		}}
		if err = s.publishServiceMessage(ctx, msg); err != nil {
			s.logger.Error("发送消息失败", zap.Error(err))
			return nil, err
		}
	}

	s.logger.Info("GroupLeaveRoom", zap.String("room", roomInfo.Room), zap.String("SenderID", roomInfo.SenderID), zap.String("uid", uid))

	return nil, nil
}

func (s *Service) getGroupRedisRoom(ctx context.Context, key string) (*model.RoomInfo, error) {
	//k1 := "*:" + strconv.FormatUint(uint64(gid), 10)
	//s.logger.Info("getGroupRedisRoom", zap.String("key", k1))
	//resp := &model.GroupRoomInfo{}
	//k2, err := s.getRedisRoomWithPrefix(ctx, liveGroupPrefix, k1, resp)
	//if err != nil {
	//	return nil, "", err
	//}

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
