package command

import (
	"context"
	"fmt"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	"github.com/cossim/coss-server/internal/live/domain/entity"
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
	"strings"
	"time"
)

type CreateRoom struct {
	DriverID     string
	Creator      string
	Type         string
	Participants []string
	GroupID      uint32
	Option       RoomOption
}

type RoomOption struct { // 通话选项
	VideoEnabled bool   `json:"video_enabled"` // 是否启用视频
	AudioEnabled bool   `json:"audio_enabled"` // 是否启用音频
	Resolution   string `json:"resolution"`    // 分辨率
	FrameRate    int    `json:"frame_rate"`    // 帧率
	Codec        string `json:"codec"`         // 编解码器
}

type CreateRoomResponse struct {
	Url     string
	Room    string
	Timeout int
}

func (h *LiveHandler) CreateRoom(ctx context.Context, cmd *CreateRoom) (*CreateRoomResponse, error) {
	h.logger.Debug("received createRoom request",
		zap.String("type", cmd.Type),
		zap.Uint32("group_id", cmd.GroupID),
		zap.String("creator", cmd.Creator),
		zap.Strings("participants", cmd.Participants),
		zap.Any("option", cmd.Option),
	)

	if cmd.Type == entity.GroupRoomType && cmd.GroupID == 0 {
		return nil, code.InvalidParameter.CustomMessage("group_id is required")
	}

	rooms, err := h.liveRepo.GetUserRooms(ctx, cmd.Creator)
	if err != nil {
		if !(code.Cause(err).Code() == code.LiveErrCallNotFound.Code()) && !strings.Contains(err.Error(), redis.Nil.Error()) {
			h.logger.Error("get user rooms", zap.Error(err))
			return nil, err
		}
	}
	if len(rooms) > 0 {
		return nil, code.LiveErrAlreadyInCall
	}

	roomType := entity.RoomType(cmd.Type)
	if !roomType.IsValid() {
		return nil, code.InvalidParameter
	}

	roomName := generateRoomName()
	var maxParticipants uint32
	switch roomType {
	case entity.UserRoomType:
		maxParticipants = entity.MaxParticipantsUser
	case entity.GroupRoomType:
		maxParticipants = entity.MaxParticipantsGroup
	}

	participants := append([]string{cmd.Creator}, cmd.Participants...)

	if len(participants) > int(maxParticipants) {
		return nil, code.LiveErrMaxParticipantsExceeded
	}

	if cmd.Type == entity.UserRoomType {
		err = h.createUserLive(ctx, roomName, participants, cmd.Option)
	}
	if cmd.Type == entity.GroupRoomType {
		err = h.createGroupLive(ctx, roomName, cmd.GroupID, participants, cmd.Option)
	}
	if err != nil {
		h.logger.Error("create live", zap.Error(err))
		return nil, err
	}

	_, err = h.roomService.CreateRoom(ctx, &livekit.CreateRoomRequest{
		Name:            roomName,
		EmptyTimeout:    uint32(h.liveTimeout.Seconds()), // 空闲时间
		MaxParticipants: maxParticipants,
	})
	if err != nil {
		h.logger.Error("创建通话失败", zap.Error(err))
		return nil, err
	}

	er := &entity.Room{
		ID:              roomName,
		Type:            roomType,
		Creator:         cmd.Creator,
		Owner:           cmd.Creator,
		GroupID:         cmd.GroupID,
		NumParticipants: 0,
		MaxParticipants: maxParticipants,
		Participants:    genRoomParticipants(participants),
		Option: entity.RoomOption{
			VideoEnabled: cmd.Option.VideoEnabled,
			AudioEnabled: cmd.Option.AudioEnabled,
			Resolution:   cmd.Option.Resolution,
			FrameRate:    cmd.Option.FrameRate,
			Codec:        cmd.Option.Codec,
		},
	}

	if err := h.liveRepo.CreateRoom(ctx, er); err != nil {
		h.logger.Error("Failed to create room", zap.Error(err))
		return nil, err
	}

	h.logger.Debug("room created success",
		zap.String("room", roomName),
		zap.String("creator", cmd.Creator),
	)

	// 通话超时处理
	h.scheduleLiveTimeout(er, int(h.liveTimeout.Seconds()), cmd.DriverID)

	return &CreateRoomResponse{
		Url:     h.webRtcUrl,
		Room:    roomName,
		Timeout: int(h.liveTimeout.Seconds()),
	}, nil
}

func (h *LiveHandler) scheduleLiveTimeout(room *entity.Room, timeoutSeconds int, driverID string) {
	time.AfterFunc(time.Duration(timeoutSeconds)*time.Second, func() {
		// 通话超时，发送 WebSocket 事件
		h.logger.Info("推送通话超时事件", zap.Duration("timeout", time.Duration(timeoutSeconds)*time.Second), zap.Any("room", room))
		if err := h.handleMissed(context.Background(), room, room.Creator, driverID); err != nil {
			h.logger.Error("Failed to handle missed", zap.Error(err))
		}
	})
}

// areUsersFriends checks if two users are friends
func (h *LiveHandler) areUsersFriends(ctx context.Context, userID, friendID string) (bool, error) {
	rel, err := h.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
		UserId:   userID,
		FriendId: friendID,
	})
	if err != nil {
		h.logger.Error("Failed to get user relation", zap.Error(err))
		return false, err
	}

	return rel.Status == relationgrpcv1.RelationStatus_RELATION_NORMAL, nil
}

func (h *LiveHandler) checkUserRelations(ctx context.Context, creator string, participants []string) error {
	friends, err := h.relationUserService.GetUserRelationByUserIds(ctx, &relationgrpcv1.GetUserRelationByUserIdsRequest{
		UserId:    creator,
		FriendIds: participants,
	})
	if err != nil {
		return err
	}

	if len(friends.Users) == 0 {
		return code.RelationUserErrFriendRelationNotFound
	}

	for _, friend := range friends.Users {
		relation, err := h.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
			UserId:   friend.FriendId,
			FriendId: creator,
		})
		if err != nil {
			return err
		}
		if relation.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
			r1, err := h.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: friend.FriendId})
			if err != nil {
				h.logger.Error("get user info error", zap.Error(err))
				return err
			}
			return code.StatusNotAvailable.CustomMessage(fmt.Sprintf("你不是%s好友", r1.NickName))
		}
		if friend.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
			name := friend.Remark
			if name == "" {
				r2, err := h.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: friend.FriendId})
				if err != nil {
					h.logger.Error("get user info error", zap.Error(err))
					return err
				}
				name = r2.NickName
			}
			return code.StatusNotAvailable.CustomMessage(fmt.Sprintf("%s不是你的好友", name))
		}
	}
	return nil
}

func (h *LiveHandler) createUserLive(ctx context.Context, roomName string, participants []string, option RoomOption) error {
	if err := h.checkUserRelations(ctx, participants[0], participants[1:]); err != nil {
		return err
	}

	if err := h.liveRepo.CreateUsersLive(ctx, roomName, participants...); err != nil {
		h.logger.Error("Failed to create users live", zap.Error(err))
		return err
	}

	data := map[string]interface{}{
		"url":          h.webRtcUrl,
		"room":         roomName,
		"sender_id":    participants[0],
		"recipient_id": participants[1],
		"option":       option,
	}

	bytes, err := utils.StructToBytes(data)
	if err != nil {
		return err
	}

	msg := &pushgrpcv1.WsMsg{
		Uid:   participants[1],
		Event: pushgrpcv1.WSEventType_UserCallReqEvent,
		Data:  &any2.Any{Value: bytes}}

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

	h.logger.Info("创建用户通话", zap.String("sender", participants[0]), zap.String("recipient", participants[1]), zap.String("room", roomName))
	return nil
}

func (h *LiveHandler) createGroupLive(ctx context.Context, roomName string, gid uint32, participants []string, option RoomOption) error {
	creator := participants[0]
	member := participants[1:]

	room, err := h.liveRepo.GetGroupRoom(ctx, fmt.Sprintf("%d", gid))
	if err != nil {
		if !(code.IsCode(err, code.LiveErrCallNotFound)) {
			return err
		}
	}
	if room != nil {
		return code.LiveErrAlreadyInCall
	}

	group, err := h.groupService.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{Gid: gid})
	if err != nil {
		h.logger.Error("create group call failed", zap.Error(err))
		return err
	}
	if group.Status != groupgrpcv1.GroupStatus_GROUP_STATUS_NORMAL {
		return code.GroupErrGroupStatusNotAvailable
	}

	if err := h.checkGroupRelations(ctx, gid, member); err != nil {
		return err
	}

	ps := make([]string, 0)
	for i := range member {
		memberUser, err := h.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: member[i]})
		if err != nil {
			h.logger.Error("获取用户信息失败", zap.Error(err), zap.String("uid", member[i]))
			continue
		}
		if memberUser.Status != usergrpcv1.UserStatus_USER_STATUS_NORMAL {
			h.logger.Error("用户状态异常", zap.String(memberUser.UserId, memberUser.Status.String()))
			continue
		}
		ps = append(ps, member[i])
	}

	for i := range ps {
		data := map[string]interface{}{
			"url":          h.webRtcUrl,
			"group_id":     gid,
			"room":         roomName,
			"sender_id":    creator,
			"recipient_id": i,
			"option":       option,
		}

		bytes, err := utils.StructToBytes(data)
		if err != nil {
			return err
		}

		msg := &pushgrpcv1.WsMsg{Uid: member[i], Event: pushgrpcv1.WSEventType_GroupCallReqEvent, Data: &any2.Any{Value: bytes}}
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

	if err := h.liveRepo.CreateGroupLive(ctx, roomName, fmt.Sprintf("%d", gid)); err != nil {
		return err
	}

	if err := h.liveRepo.CreateUsersLive(ctx, roomName, participants...); err != nil {
		return err
	}

	h.logger.Info("create group live success", zap.Uint32("gid", gid), zap.String("room", roomName), zap.Strings("participants", participants))
	return nil
}

func (h *LiveHandler) checkGroupRelations(ctx context.Context, gid uint32, participants []string) error {
	rels, err := h.relationGroupService.GetBatchGroupRelation(ctx, &relationgrpcv1.GetBatchGroupRelationRequest{GroupId: gid, UserIds: participants})
	if err != nil {
		return err
	}
	notInGroupUsers := make(map[string]string)
	filteredParticipants := participants[:0]
	for _, p := range participants {
		found := false
		for _, r := range rels.GroupRelationResponses {
			if r.UserId == p {
				// TODO 从participants删除这个用户，因为开启了免打扰不需要推送
				if r.IsSilent != relationgrpcv1.GroupSilentNotificationType_GroupSilent {
					filteredParticipants = append(filteredParticipants, p)
				}
				found = true
				break
			}
		}
		if !found {
			notInGroupUsers[p] = ""
		}
	}

	participants = filteredParticipants

	infos, err := h.userService.GetBatchUserInfo(ctx, &usergrpcv1.GetBatchUserInfoRequest{UserIds: keys(notInGroupUsers)})
	if err != nil {
		return err
	}

	// 更新 notInGroupUsers 中的昵称
	for _, u := range infos.Users {
		notInGroupUsers[u.UserId] = u.NickName
	}

	// 构造不在群聊中的用户昵称列表
	var notInGroupNicknames []string
	for _, nickname := range notInGroupUsers {
		notInGroupNicknames = append(notInGroupNicknames, nickname)
	}

	if len(notInGroupNicknames) == 0 {
		return nil
	}

	msg := fmt.Sprintf("%s 没有在群聊中", strings.Join(notInGroupNicknames, ", "))
	return code.RelationGroupErrNotInGroup.CustomMessage(msg)
}

// 辅助函数,用于从 map 中获取所有 key
func keys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func genRoomParticipants(participants []string) map[string]*entity.ActiveParticipant {
	if len(participants) == 0 {
		return nil
	}
	var participantsMap = make(map[string]*entity.ActiveParticipant)
	for _, participant := range participants {
		participantsMap[participant] = &entity.ActiveParticipant{
			Connected: false,
			Status:    entity.ParticipantInfo_WAITING,
		}
	}
	return participantsMap
}

func generateRoomName() string {
	return uuid.New().String()
}
