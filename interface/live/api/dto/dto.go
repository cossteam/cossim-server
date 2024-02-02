package dto

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type UserCallRequest struct {
	UserID string     `json:"user_id" binding:"required"` // 接收视频通话的用户ID
	Option CallOption `json:"option"`
}

type UserCallResponse struct {
	Url    string `json:"url"` // webrtc server
	Token  string `json:"token"`
	Room   string `json:"room"`    // 房间名称
	RoomID string `json:"room_id"` // 房间id
}

type GroupCallRequest struct {
	SenderID string     `json:"sender_id" binding:"required"` // 发起视频通话的用户ID
	GroupID  string     `json:"group_id" binding:"required"`  // 群组的ID
	Option   CallOption `json:"option"`
}

type CallOption struct { // 通话选项
	VideoEnabled bool   `json:"video_enabled"` // 是否启用视频
	AudioEnabled bool   `json:"audio_enabled"` // 是否启用音频
	Resolution   string `json:"resolution"`    // 分辨率
	FrameRate    int    `json:"frame_rate"`    // 帧率
	Codec        string `json:"codec"`         // 编解码器
}

type UserJoinRequest struct {
	UserID string     `json:"user_id" binding:"required"` // 用户ID
	Room   string     `json:"room" binding:"required"`    // 房间名称
	Option CallOption `json:"option"`
}

type UserLeaveRequest struct {
	UserID string `json:"user_id" binding:"required"` // 用户ID
	Room   string `json:"room" binding:"required"`    // 房间名称
}

type UserShowRequest struct {
	UserID string `json:"user_id" binding:"required"` // 发起视频通话的用户ID
	Room   string `json:"room" binding:"required"`    // 房间名称
}

type UserShowResponse struct {
	Sid         string           `json:"sid,omitempty"`
	Identity    string           `json:"identity,omitempty"`
	State       ParticipantState `json:"state,omitempty"`
	JoinedAt    int64            `json:"joined_at,omitempty"`
	Name        string           `json:"name,omitempty"`
	IsPublisher bool             `json:"is_publisher,omitempty"`
}

type ParticipantInfo struct {
	Sid         string           `json:"sid,omitempty"`
	Identity    string           `json:"identity,omitempty"`
	State       ParticipantState `json:"state,omitempty"`
	JoinedAt    int64            `json:"joined_at,omitempty"`
	Name        string           `json:"name,omitempty"`
	IsPublisher bool             `json:"is_publisher,omitempty"`
}

type ParticipantState int32

const (
	// ParticipantInfo_JOINING websocket' connected, but not offered yet
	ParticipantInfo_JOINING ParticipantState = iota
	// ParticipantInfo_JOINED server received client offer
	ParticipantInfo_JOINED
	// ParticipantInfo_ACTIVE ICE connectivity established
	ParticipantInfo_ACTIVE
	// ParticipantInfo_DISCONNECTED WS disconnected
	ParticipantInfo_DISCONNECTED
)

//type ParticipantInfo struct {
//	Sid      string                `protobuf:"bytes,1,opt,name=sid,proto3" json:"sid,omitempty"`
//	Identity string                `protobuf:"bytes,2,opt,name=identity,proto3" json:"identity,omitempty"`
//	State    ParticipantInfo_State `protobuf:"varint,3,opt,name=state,proto3,enum=livekit.ParticipantInfo_State" json:"state,omitempty"`
//	Metadata string                `protobuf:"bytes,5,opt,name=metadata,proto3" json:"metadata,omitempty"`
//	// timestamp when participant joined room, in seconds
//	JoinedAt   int64                  `protobuf:"varint,6,opt,name=joined_at,json=joinedAt,proto3" json:"joined_at,omitempty"`
//	Name       string                 `protobuf:"bytes,9,opt,name=name,proto3" json:"name,omitempty"`
//	Version    uint32                 `protobuf:"varint,10,opt,name=version,proto3" json:"version,omitempty"`
//	Permission *ParticipantPermission `protobuf:"bytes,11,opt,name=permission,proto3" json:"permission,omitempty"`
//	Region     string                 `protobuf:"bytes,12,opt,name=region,proto3" json:"region,omitempty"`
//	// indicates the participant has an active publisher connection
//	// and can publish to the server
//	IsPublisher bool                 `protobuf:"varint,13,opt,name=is_publisher,json=isPublisher,proto3" json:"is_publisher,omitempty"`
//	Kind        ParticipantInfo_Kind `protobuf:"varint,14,opt,name=kind,proto3,enum=livekit.ParticipantInfo_Kind" json:"kind,omitempty"`
//}
