package service

import (
	"context"
	"github.com/cossim/coss-server/interface/live/api/model"
	msgconfig "github.com/cossim/coss-server/interface/msg/config"
	"github.com/cossim/coss-server/pkg/cache"
	"github.com/cossim/coss-server/pkg/code"
	groupgrpcv1 "github.com/cossim/coss-server/service/group/api/v1"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	usergrpcv1 "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/google/uuid"
	"github.com/livekit/protocol/livekit"
	"go.uber.org/zap"
)

func (s *Service) CreateGroupCall(ctx context.Context, uid string, gid uint32, member []string) (interface{}, error) {
	group, err := s.groupClient.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{Gid: gid})
	if err != nil {
		s.logger.Error("create group call failed", zap.Error(err))
		return nil, err
	}

	if group.Status != groupgrpcv1.GroupStatus_GROUP_STATUS_NORMAL {
		return nil, code.GroupErrGroupStatusNotAvailable
	}

	rel, err := s.relGroupClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: gid, UserId: uid})
	if err != nil {
		return nil, err
	}

	if rel.MuteEndTime > 0 {
		return nil, code.GroupErrUserIsMuted
	}

	type MemberInfo struct {
		UserID   string
		UserName string
		Token    string
	}

	rels, err := s.relGroupClient.GetBatchGroupRelation(ctx, &relationgrpcv1.GetBatchGroupRelationRequest{GroupId: gid, UserIds: member})
	if err != nil {
		return nil, err
	}

	if len(member) != len(rels.GroupRelationResponses) {
		return nil, code.RelationGroupErrNotInGroup
	}

	roomName := uuid.New().String()
	room, err := s.roomService.CreateRoom(ctx, &livekit.CreateRoomRequest{
		Name:            roomName,
		EmptyTimeout:    uint32(s.liveTimeout.Seconds()), // 空闲时间
		MaxParticipants: 256,
	})
	if err != nil {
		s.logger.Error("创建通话失败", zap.Error(err))
		return nil, err
	}

	redisRoom := &model.GroupRoomInfo{
		SenderID:        uid,
		MaxParticipants: 256,
		Participants:    map[string]*model.ActiveParticipant{},
	}

	// 提取未被禁言的成员信息
	participants := make([]MemberInfo, 0)
	participants = append(participants, MemberInfo{UserID: uid})
	for _, v := range rels.GroupRelationResponses {
		//if v.MuteEndTime == 0 {
		participants = append(participants, MemberInfo{UserID: v.UserId})
		//}
	}

	var senderToken string

	// 获取成员的名称和其他信息
	for i := range participants {
		memberUser, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: participants[i].UserID})
		if err != nil {
			s.logger.Error("获取用户信息失败", zap.Error(err))
			continue
		}

		if memberUser.Status != usergrpcv1.UserStatus_USER_STATUS_NORMAL {
			s.logger.Error("用户状态异常", zap.String(memberUser.UserId, memberUser.Status.String()))
			continue
		}

		redisRoom.Participants[participants[i].UserID] = &model.ActiveParticipant{}
		participants[i].UserName = memberUser.NickName
		token, err := s.GetJoinToken(ctx, memberUser.NickName, room.Name)
		if err != nil {
			s.logger.Error("获取token失败", zap.Error(err))
			continue
		}

		if participants[i].UserID == uid {
			senderToken = token
			continue
		}

		msg := msgconfig.WsMsg{Uid: participants[i].UserID, Event: msgconfig.GroupCallReqEvent, Data: map[string]interface{}{
			"url":          s.livekitServer,
			"token":        token,
			"room":         room.Name,
			"sender_id":    uid,
			"recipient_id": participants[i].UserID,
		}}
		if err = s.publishServiceMessage(ctx, msg); err != nil {
			s.logger.Error("发送消息失败", zap.Error(err))
		}
	}

	ToJSONString, err := redisRoom.ToJSONString()
	if err != nil {
		return nil, err
	}
	if err = cache.SetKey(s.redisClient, room.Name, ToJSONString, 0); err != nil {
		s.logger.Error("保存房间信息失败", zap.Error(err))
		return nil, err
	}

	return map[string]string{
		"url":       s.livekitServer,
		"token":     senderToken,
		"room_name": roomName,
		"room_id":   room.Sid,
	}, nil
}
