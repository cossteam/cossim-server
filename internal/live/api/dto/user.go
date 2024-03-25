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
	Url string `json:"url"` // webRtc服务器地址
	//Token string `json:"token"` // 加入通话的token
	//Room   string `json:"room"`    // 房间名称
	//RoomID string `json:"room_id"` // 房间id
}

type CallOption struct { // 通话选项
	VideoEnabled bool   `json:"video_enabled"` // 是否启用视频
	AudioEnabled bool   `json:"audio_enabled"` // 是否启用音频
	Resolution   string `json:"resolution"`    // 分辨率
	FrameRate    int    `json:"frame_rate"`    // 帧率
	Codec        string `json:"codec"`         // 编解码器
}

type UserJoinRequest struct {
	//Room   string     `json:"room" binding:"required"` // 房间名称
	Option CallOption `json:"option"`
}

type UserJoinResponse struct {
	Url   string `json:"url"`   // webRtc服务器地址
	Token string `json:"token"` // 加入通话的token
}

type UserLeaveRequest struct {
	Room string `json:"room" binding:"required"` // 房间名称
}

type UserRejectRequest struct {
	Room string `json:"room" binding:"required"` // 房间名称
}

type UserShowRequest struct {
	UserID string `json:"user_id" binding:"required"` // 发起视频通话的用户ID
	Room   string `json:"room" binding:"required"`    // 房间名称
}

type UserShowResponse struct {
	StartAt            int64              `json:"start_at"` // 创建房间时间
	Duration           int64              `json:"duration"` // 房间持续时间
	Room               string             `json:"room"`     // 房间名称
	Type               string             `json:"type"`     // 房间类型 model.RoomType user、group
	Status             string             `json:"status,omitempty"`
	VideoCallRecordURL string             `json:"video_call_record_url,omitempty"`
	Participant        []*ParticipantInfo `json:"participant"`
}

type ParticipantInfo struct {
	//Uid         string           `json:"uid,omitempty"`       // 用户id
	Sid         string           `json:"sid,omitempty"`       // 房间id
	Identity    string           `json:"identity,omitempty"`  // 用户昵称
	State       ParticipantState `json:"state,omitempty"`     // 用户状态
	JoinedAt    int64            `json:"joined_at,omitempty"` // 加入时间
	Name        string           `json:"name,omitempty"`
	IsPublisher bool             `json:"is_publisher,omitempty"` // 创建者
}

type ParticipantState int32

const (
	// ParticipantInfo_JOINING websocket' connected, but not offered yet
	ParticipantInfo_JOINING ParticipantState = iota // websocket已连接，未加入通话
	// ParticipantInfo_JOINED server received client offer
	ParticipantInfo_JOINED // 已加入通话，对方未响应
	// ParticipantInfo_ACTIVE ICE connectivity established
	ParticipantInfo_ACTIVE // 双方都已加入通话
	// ParticipantInfo_DISCONNECTED WS disconnected
	ParticipantInfo_DISCONNECTED // 断开连接
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

type UserMessageType uint

func (u UserMessageType) Uint() uint {
	return uint(u)
}

func (u UserMessageType) Int32() int32 {
	return int32(u)
}

const (
	MessageTypeText        UserMessageType = iota + 1 // 文本消息
	MessageTypeVoice                                  // 语音消息
	MessageTypeImage                                  // 图片消息
	MessageTypeLabel                                  //标注
	MessageTypeNotice                                 //群公告
	MessageTypeFile                                   // 文件消息
	MessageTypeVideo                                  // 视频消息
	MessageTypeEmojiReply                             //emoji回复
	MessageTypeVoiceCall                              // 语音通话
	MessageTypeVideoCall                              // 视频通话
	MessageTypeDelete                                 // 撤回消息
	MessageTypeCancelLabel                            //取消标注
)

// IsValidMessageType 判断是否是有效的消息类型
func IsValidMessageType(msgType UserMessageType) bool {
	validTypes := map[UserMessageType]struct{}{
		MessageTypeText:        {},
		MessageTypeVoice:       {},
		MessageTypeImage:       {},
		MessageTypeFile:        {},
		MessageTypeVideo:       {},
		MessageTypeVoiceCall:   {},
		MessageTypeVideoCall:   {},
		MessageTypeLabel:       {},
		MessageTypeNotice:      {},
		MessageTypeEmojiReply:  {},
		MessageTypeDelete:      {},
		MessageTypeCancelLabel: {},
	}

	_, isValid := validTypes[msgType]
	return isValid
}

// 提示消息类型校验
func IsPromptMessageType(msgType UserMessageType) bool {
	validTypes := map[UserMessageType]struct{}{
		MessageTypeLabel:       {},
		MessageTypeNotice:      {},
		MessageTypeDelete:      {},
		MessageTypeCancelLabel: {},
	}
	_, isValid := validTypes[msgType]
	return isValid
}
