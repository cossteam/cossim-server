package dto

type GroupCallRequest struct {
	GroupID uint32     `json:"group_id" binding:"required"` // 群组的ID
	Member  []string   `json:"member"`                      // 成员ids
	Option  CallOption `json:"option"`
}

type GroupCallResponse struct {
	Url    string `json:"url"`     // webRtc服务器地址
	Token  string `json:"token"`   // 加入通话的token
	Room   string `json:"room"`    // 房间名称
	RoomID string `json:"room_id"` // 房间id
}

type GroupJoinRequest struct {
	GroupID uint32 `json:"group_id" binding:"required"` // 群组的ID
	//Room    string     `json:"room" binding:"required"`     // 房间名称
	Option CallOption `json:"option"`
}

type GroupJoinResponse struct {
	Url   string `json:"url"`   // webRtc服务器地址
	Token string `json:"token"` // 加入通话的token
}

type GroupShowRequest struct {
	GroupID uint32 `json:"group_id" binding:"required"` // 群组的ID
	//Room    string     `json:"room" binding:"required"`     // 房间名称
	Option CallOption `json:"option"`
}

type GroupShowResponse struct {
	StartAt         int64              `json:"start_at"`         // 创建房间时间
	Duration        int64              `json:"duration"`         // 房间持续时间
	GroupID         uint32             `json:"group_id"`         // 群组ID
	NumParticipants uint32             `json:"num_participants"` // 当前房间人数
	MaxParticipants uint32             `json:"max_participants"` // 最大人数
	SenderID        string             `json:"sender_id"`        // 创建者
	Room            string             `json:"room"`             // 房间名称
	Status          string             `json:"status,omitempty"`
	RecordURL       string             `json:"video_call_record_url,omitempty"`
	Participant     []*ParticipantInfo `json:"participant"`
}

type GroupRejectRequest struct {
	GroupID uint32     `json:"group_id" binding:"required"` // 群组的ID
	Room    string     `json:"room" binding:"required"`     // 房间名称
	Option  CallOption `json:"option"`
}

type GroupLeaveRequest struct {
	GroupID uint32 `json:"group_id" binding:"required"` // 群组的ID
	//Room    string     `json:"room" binding:"required"`     // 房间名称
	Force  bool       `json:"force"` // 是否要结束整个通话
	Option CallOption `json:"option"`
}
