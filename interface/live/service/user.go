package service

import (
	"context"
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
)

func (s *Service) CreateUserCall(ctx context.Context, senderID, recipientID string) (interface{}, error) {
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

	toMap, err := (&model.RoomInfo{
		SenderID:        senderID,
		RecipientID:     recipientID,
		MaxParticipants: 2,
		Participants:    map[string]*model.ActiveParticipant{},
	}).ToMap()
	if err != nil {
		return nil, err
	}
	if err = cache.SetKey(s.redisClient, room.Name, toMap); err != nil {
		s.logger.Error("保存房间信息失败", zap.Error(err))
		return nil, err
	}

	msg := msgconfig.WsMsg{Uid: recipientID, Event: msgconfig.UserCallEvent, Data: map[string]interface{}{
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

	return map[string]string{
		"url":       s.livekitServer,
		"token":     senderToken,
		"room_name": roomName,
		"room_id":   room.Sid,
	}, nil
}

func (s *Service) UserJoinRoom(ctx context.Context, req *dto.UserJoinRequest) (interface{}, error) {
	room, err := s.getRoom(ctx, req.Room)
	if err != nil {
		s.logger.Error("获取房间信息失败", zap.Error(err))
		return nil, err
	}

	if room.NumParticipants+1 > room.MaxParticipants {
		return nil, code.LiveErrMaxParticipantsExceeded
	}

	//user, err := s.userClient.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: req.UserID})
	//if err != nil {
	//	s.logger.Error("获取用户信息失败", zap.Error(err))
	//	return nil, err
	//}

	listRes, err := s.roomService.ListRooms(ctx, &livekit.ListRoomsRequest{
		Names: []string{
			req.Room,
		},
	})
	if err != nil {
		s.logger.Error("error listing rooms", zap.Error(err))
		return nil, code.LiveErrJoinCallFailed
	}

	if len(listRes.Rooms) == 0 {
		return nil, code.LiveErrCallNotFound
	}

	token, err := s.joinRoom(req.UserID, listRes.Rooms[0])
	if err != nil {
		s.logger.Error("error joining room", zap.Error(err))
		return nil, err
	}

	return map[string]interface{}{
		"token": token,
	}, nil
}

func (s *Service) joinRoom(userName string, room *livekit.Room) (string, error) {
	//roomCache, err := cache.GetKey(s.redisClient, room.Sid)
	//if err != nil {
	//	return "", err
	//}
	//data := &model.RoomInfo{}
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

func (s *Service) getRoom(ctx context.Context, room string) (*livekit.Room, error) {
	rooms, err := s.roomService.ListRooms(ctx, &livekit.ListRoomsRequest{Names: []string{room}})
	if err != nil {
		return nil, err
	}

	if len(rooms.Rooms) == 0 {
		return nil, code.LiveErrCallNotFound
	}

	return rooms.Rooms[0], nil
}

func (s *Service) GetUserRoom(ctx context.Context, userID, room string) ([]*dto.UserShowResponse, error) {
	resp := make([]*dto.UserShowResponse, 0)
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
		resp = append(resp, &dto.UserShowResponse{
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

func (s *Service) UserLeaveRoom(ctx context.Context, req *dto.UserLeaveRequest) (interface{}, error) {
	_, err := s.roomService.DeleteRoom(ctx, &livekit.DeleteRoomRequest{
		Room: req.Room, // 房间名称
	})
	if err != nil {
		s.logger.Error("error deleting room", zap.Error(err))
		return nil, code.LiveErrLeaveCallFailed
	}

	return nil, nil
}

func (s *Service) GetJoinToken(ctx context.Context, uid string, room string) (string, error) {
	at := auth.NewAccessToken(s.liveApiKey, s.liveApiSecret)
	grant := &auth.VideoGrant{
		RoomCreate: true,
		RoomList:   true,
		RoomAdmin:  true,
		RoomJoin:   true,
		Room:       room,
	}
	at.AddGrant(grant).SetIdentity(uid).SetValidFor(s.liveTimeout)
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
