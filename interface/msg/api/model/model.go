package model

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type SendUserMsgRequest struct {
	DialogId   uint32 `json:"dialog_id" binding:"required"`
	ReceiverId string `json:"receiver_id" binding:"required"`
	Content    string `json:"content" binding:"required"`
	Type       uint   `json:"type" binding:"required"`
	ReplayId   uint   `json:"replay_id" `
}

type SendGroupMsgRequest struct {
	DialogId uint32 `json:"dialog_id" binding:"required"`
	GroupId  uint32 `json:"group_id" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Type     uint32 `json:"type" binding:"required"`
	ReplayId uint32 `json:"replay_id"`
	//AtUserIds []string `json:"at_user_ids"`
}

type MsgListRequest struct {
	//GroupId    int64  `json:"group_id" binding:"required"`
	UserId   string `json:"user_id" binding:"required"`
	Type     int32  `json:"type" binding:"required"`
	Content  string `json:"content" binding:"required"`
	PageNum  int    `json:"page_num" binding:"required"`
	PageSize int    `json:"page_size" binding:"required"`
}

type ConversationType uint

const (
	UserConversation ConversationType = iota
	GroupConversation
)

type UserDialogListResponse struct {
	DialogId uint32 `json:"dialog_id"`
	UserId   string `json:"user_id,omitempty"`
	GroupId  uint32 `json:"group_id,omitempty"`
	// 会话类型
	DialogType ConversationType `json:"dialog_type"`
	// 会话名称
	DialogName string `json:"dialog_name"`
	// 会话头像
	DialogAvatar string `json:"dialog_avatar"`
	// 会话未读消息数
	DialogUnreadCount int     `json:"dialog_unread_count"`
	LastMessage       Message `json:"last_message"`
}

type Message struct {
	// 消息类型
	MsgType uint `json:"msg_type"`
	// 消息内容
	Content string `json:"content"`
	// 消息发送者
	SenderId string `json:"sender_id"`
	// 消息发送时间
	SendTime int64 `json:"send_time"`
	// 消息id
	MsgId uint64 `json:"msg_id"`
}

// EditUserMsgRequest represents the request structure for editing user message.
type EditUserMsgRequest struct {
	MsgId uint32 `json:"msg_id" binding:"required"` // Message ID
	// 消息类型
	MsgType uint `json:"msg_type" binding:"required"`
	// 消息内容
	Content string `json:"content" binding:"required"`
}

// EditGroupMsgRequest represents the request structure for editing group message.
type EditGroupMsgRequest struct {
	MsgId   uint32 `json:"msg_id" binding:"required"` // Message ID
	MsgType uint   `json:"msg_type" binding:"required"`
	Content string `json:"content" binding:"required"` // New content of the message
}

// RecallUserMsgRequest represents the request structure for recalling user message.
type RecallUserMsgRequest struct {
	MsgId uint32 `json:"msg_id" binding:"required"` // Message ID
}

// RecallGroupMsgRequest represents the request structure for recalling group message.
type RecallGroupMsgRequest struct {
	MsgId uint32 `json:"msg_id" binding:"required"` // Message ID
}
