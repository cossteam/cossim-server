package entity

import (
	"encoding/json"
)

type Room struct {
	ID              string                        `json:"id"`
	Type            RoomType                      `json:"type"`
	Creator         string                        `json:"creator"`
	Owner           string                        `json:"owner"`
	GroupID         uint32                        `json:"group_id"`
	NumParticipants uint32                        `json:"num_participants"`
	MaxParticipants uint32                        `json:"max_participants"`
	Participants    map[string]*ActiveParticipant `json:"participants"`
	Option          RoomOption                    `json:"option"`
}

func (r *Room) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func (r *Room) Unmarshal(bytes []byte) error {
	return json.Unmarshal(bytes, r)
}

func (r *Room) String() string {
	bytes, err := json.Marshal(r)
	if err != nil {
		return ""
	}
	return string(bytes)
}

type RoomType string

const (
	GroupRoomType = "group"
	UserRoomType  = "user"

	MaxParticipantsGroup = 100
	MaxParticipantsUser  = 2
)

// IsValid checks if the room type is valid.
func (rt RoomType) IsValid() bool {
	return rt == GroupRoomType || rt == UserRoomType
}

type ActiveParticipant struct {
	Connected bool
	Status    ParticipantState
	DriverID  string
}

type RoomOption struct { // 通话选项
	VideoEnabled bool   `json:"video_enabled"` // 是否启用视频
	AudioEnabled bool   `json:"audio_enabled"` // 是否启用音频
	Resolution   string `json:"resolution"`    // 分辨率
	FrameRate    int    `json:"frame_rate"`    // 帧率
	Codec        string `json:"codec"`         // 编解码器
}

type ParticipantState int32

const (
	// ParticipantInfo_WAITING indicates the participant is waiting for connection
	ParticipantInfo_WAITING ParticipantState = iota // 等待用户连接
	// ParticipantInfo_JOINING websocket' connected, but not offered yet
	ParticipantInfo_JOINING // websocket已连接，未加入通话
	// ParticipantInfo_JOINED server received client offer
	ParticipantInfo_JOINED // 已加入通话，对方未响应
	// ParticipantInfo_ACTIVE ICE connectivity established
	ParticipantInfo_ACTIVE // 双方都已加入通话
	// ParticipantInfo_DISCONNECTED WS disconnected
	ParticipantInfo_DISCONNECTED // 断开连接
)

type ParticipantInfo struct {
	Room     string
	Identity string
	State    ParticipantState
	// timestamp when participant joined room, in seconds
	JoinedAt int64
	Name     string
	// indicates the participant has an active publisher connection
	// and can publish to the server
	IsPublisher bool
}
