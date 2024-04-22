package query

import (
	"context"
	"github.com/cossim/coss-server/internal/live/domain/live"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/livekit/protocol/livekit"
	lksdk "github.com/livekit/server-sdk-go"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type LiveHandler struct {
	logger   *zap.Logger
	liveRepo live.Repository

	webRtcUrl            string
	liveApiKey           string
	liveApiSecret        string
	liveTimeout          time.Duration
	roomService          *lksdk.RoomServiceClient
	relationGroupService relationgrpcv1.GroupRelationServiceClient
}

func NewLiveHandler(options ...LiveHandlerOption) *LiveHandler {
	h := &LiveHandler{}
	for _, option := range options {
		option(h)
	}
	return h
}

type LiveHandlerOption func(*LiveHandler)

func WithRepo(repo live.Repository) LiveHandlerOption {
	return func(h *LiveHandler) {
		h.liveRepo = repo
	}
}

func WithLogger(logger *zap.Logger) LiveHandlerOption {
	return func(h *LiveHandler) {
		h.logger = logger
	}
}

func WithLiveKit(c config.LivekitConfig) LiveHandlerOption {
	return func(h *LiveHandler) {
		h.webRtcUrl = c.Url
		h.liveApiKey = c.ApiKey
		h.liveApiSecret = c.ApiSecret
		h.liveTimeout = c.Timeout
		h.roomService = lksdk.NewRoomServiceClient(c.Url, c.ApiKey, c.ApiSecret)
	}
}

func WithRelationGroupService(service relationgrpcv1.GroupRelationServiceClient) LiveHandlerOption {
	return func(h *LiveHandler) {
		h.relationGroupService = service
	}
}

type GetRoom struct {
	UserID string
	Room   string
}

type Room struct {
	ID              string
	Type            string
	Creator         string
	Owner           string
	NumParticipants uint32
	MaxParticipants uint32
	StartAt         int64
	Participant     []*ParticipantInfo
}

type ParticipantInfo struct {
	Identity    string `json:"identity"`
	IsPublisher bool   `json:"is_publisher"`
	JoinedAt    int64  `json:"joined_at,omitempty"`
	Name        string `json:"name"`
	//Room        string `json:"room"`
	State int8 `json:"state"`
	//Uid         string `json:"uid"`
}

type GetUserLive struct {
	Type   string
	UserID string
}

type GetGroupLive struct {
	GroupID uint32
	UserID  string
}

func (h *LiveHandler) GetUserLive(ctx context.Context, query *GetUserLive) ([]*Room, error) {
	if query.UserID == "" {
		return nil, code.InvalidParameter
	}

	rooms, err := h.liveRepo.GetUserRooms(ctx, query.UserID)
	if err != nil {
		if code.IsCode(err, code.LiveErrCallNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if len(rooms) == 0 {
		return nil, code.NotFound
	}

	var userRooms []*Room

	for _, room := range rooms {
		livekitRooms, err := h.roomService.ListRooms(ctx, &livekit.ListRoomsRequest{
			Names: []string{room.ID},
		})
		if err != nil {
			return nil, err
		}

		if len(livekitRooms.Rooms) == 0 {
			h.liveRepo.DeleteRoom(ctx, room.ID)
			h.logger.Error("房间不存在", zap.String("RoomID", room.ID))
			continue
		}

		livekitRoom := livekitRooms.Rooms[0]
		userRoom := &Room{
			ID:              room.ID,
			Type:            string(room.Type),
			Owner:           room.Owner,
			NumParticipants: room.NumParticipants,
			MaxParticipants: room.MaxParticipants,
			StartAt:         livekitRoom.CreationTime,
		}

		// 获取当前房间的参与者信息
		res, err := h.roomService.ListParticipants(ctx, &livekit.ListParticipantsRequest{
			Room: room.ID,
		})
		if err != nil {
			h.logger.Error("获取通话信息失败", zap.Error(err))
			return nil, code.LiveErrGetCallInfoFailed
		}

		for _, p := range res.Participants {
			userRoom.Participant = append(userRoom.Participant, &ParticipantInfo{
				Identity:    p.Identity,
				IsPublisher: p.IsPublisher,
				JoinedAt:    p.JoinedAt,
				Name:        p.Name,
				State:       int8(p.State),
			})
		}

		userRooms = append(userRooms, userRoom)
	}

	return userRooms, nil
}

func (h *LiveHandler) GetGroupLive(ctx context.Context, query *GetGroupLive) ([]*Room, error) {
	if query.GroupID == 0 {
		return nil, code.InvalidParameter
	}

	_, err := h.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		GroupId: query.GroupID,
		UserId:  query.UserID,
	})
	if err != nil {
		return nil, err
	}

	// 获取群组通话房间信息
	room, err := h.liveRepo.GetGroupRoom(ctx, strconv.Itoa(int(query.GroupID)))
	if err != nil {
		if code.IsCode(err, code.LiveErrCallNotFound) {
			return nil, nil
		}
		return nil, err
	}

	// 获取 LiveKit 房间信息
	livekitRooms, err := h.roomService.ListRooms(ctx, &livekit.ListRoomsRequest{
		Names: []string{room.ID},
	})
	if err != nil {
		return nil, err
	}

	if len(livekitRooms.Rooms) == 0 {
		h.liveRepo.DeleteRoom(ctx, room.ID)
		h.logger.Error("房间不存在", zap.String("RoomID", room.ID))
		return nil, code.LiveErrCallNotFound
	}

	//livekitRoom := livekitRooms.Rooms[0]
	groupRoom := &Room{
		ID:              room.ID,
		Type:            string(room.Type),
		Owner:           room.Owner,
		NumParticipants: room.NumParticipants,
		MaxParticipants: room.MaxParticipants,
		StartAt:         livekitRooms.Rooms[0].CreationTime,
	}

	// 获取当前房间的参与者信息
	res, err := h.roomService.ListParticipants(ctx, &livekit.ListParticipantsRequest{
		Room: room.ID,
	})
	if err != nil {
		h.logger.Error("获取通话信息失败", zap.Error(err))
		return nil, code.LiveErrGetCallInfoFailed
	}

	for _, p := range res.Participants {
		groupRoom.Participant = append(groupRoom.Participant, &ParticipantInfo{
			Identity:    p.Identity,
			IsPublisher: p.IsPublisher,
			JoinedAt:    p.JoinedAt,
			Name:        p.Name,
			State:       int8(p.State),
		})
	}

	return []*Room{groupRoom}, nil
}

func (h *LiveHandler) GetRoom(ctx context.Context, query *GetRoom) (*Room, error) {
	room, err := h.liveRepo.GetRoom(ctx, query.Room)
	if err != nil {
		return nil, err
	}

	_, ok := room.Participants[query.UserID]
	if !ok {
		return nil, code.Forbidden
	}

	rooms, err := h.roomService.ListRooms(ctx, &livekit.ListRoomsRequest{
		Names: []string{room.ID},
	})
	if err != nil {
		return nil, err
	}

	if len(rooms.Rooms) == 0 {
		h.liveRepo.DeleteRoom(ctx, query.Room)
		return nil, code.LiveErrCallNotFound
	}

	var participant []*ParticipantInfo

	res, err := h.roomService.ListParticipants(ctx, &livekit.ListParticipantsRequest{
		Room: room.ID,
	})
	if err != nil {
		h.logger.Error("获取通话信息失败", zap.Error(err))
		return nil, code.LiveErrGetCallInfoFailed
	}

	for _, p := range res.Participants {
		participant = append(participant, &ParticipantInfo{
			Identity:    p.Sid,
			IsPublisher: p.IsPublisher,
			JoinedAt:    p.JoinedAt,
			Name:        p.Name,
			//Room:        room.ID,
			State: int8(p.State),
		})
	}

	return &Room{
		ID:              room.ID,
		Type:            string(room.Type),
		Creator:         room.Creator,
		Owner:           room.Owner,
		NumParticipants: rooms.Rooms[0].NumParticipants,
		MaxParticipants: rooms.Rooms[0].MaxParticipants,
		StartAt:         rooms.Rooms[0].CreationTime,
		Participant:     participant,
	}, nil
}
