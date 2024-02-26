package model

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type SendUserMsgRequest struct {
	DialogId               uint32               `json:"dialog_id" binding:"required"`
	ReceiverId             string               `json:"receiver_id" binding:"required"`
	Content                string               `json:"content" binding:"required"`
	Type                   uint                 `json:"type" binding:"required"`
	ReplayId               uint                 `json:"replay_id"`
	IsBurnAfterReadingType BurnAfterReadingType `json:"is_burn_after_reading"`
}

type SendGroupMsgRequest struct {
	DialogId               uint32               `json:"dialog_id" binding:"required"`
	GroupId                uint32               `json:"group_id" binding:"required"`
	Content                string               `json:"content" binding:"required"`
	Type                   uint32               `json:"type" binding:"required"`
	ReplayId               uint32               `json:"replay_id"`
	AtUsers                []string             `json:"at_users"`
	AtAllUser              AtAllUserType        `json:"at_all_user"`
	IsBurnAfterReadingType BurnAfterReadingType `json:"is_burn_after_reading"`
}

type AtAllUserType uint

const (
	NotAtAllUser = iota
	AtAllUser
)

func isValidAtAllUserType(atAllUserType AtAllUserType) bool {
	return atAllUserType == NotAtAllUser || atAllUserType == AtAllUser
}

type MsgListRequest struct {
	UserId   string `json:"user_id" binding:"required"`
	Type     int32  `json:"type"`
	Content  string `json:"content"`
	PageNum  int    `json:"page_num" binding:"required"`
	PageSize int    `json:"page_size" binding:"required"`
}

type GroupMsgListRequest struct {
	GroupId  uint32 `json:"group_id" binding:"required"`
	UserId   string `json:"user_id"`
	Type     int32  `json:"type"`
	Content  string `json:"content"`
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

	DialogCreateAt int64 `json:"dialog_create_at"`
	TopAt          int64 `json:"top_at"`
}

type Message struct {
	GroupId                uint32               `json:"group_id,omitempty"`      //群聊id
	MsgType                uint                 `json:"msg_type"`                // 消息类型
	Content                string               `json:"content"`                 // 消息内容
	SenderId               string               `json:"sender_id"`               // 消息发送者
	SendTime               int64                `json:"send_time"`               // 消息发送时间
	MsgId                  uint64               `json:"msg_id"`                  // 消息id
	SenderInfo             SenderInfo           `json:"sender_info"`             // 消息发送者信息
	ReceiverInfo           SenderInfo           `json:"receiver_info,omitempty"` // 消息接受者信息
	AtAllUser              AtAllUserType        `json:"at_all_user,omitempty"`   // @全体用户
	AtUsers                []string             `json:"at_users,omitempty"`      // @用户id
	IsBurnAfterReadingType BurnAfterReadingType `json:"is_burn_after_reading"`   // 是否阅后即焚
	IsLabel                LabelMsgType         `json:"is_label"`                // 是否标记
	ReplayId               uint32               `json:"replay_id"`               // 回复消息id
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

// MarkUserMessageRequest 用于标注用户消息状态的请求结构体
type LabelUserMessageRequest struct {
	MsgID   uint32       `json:"msg_id" binding:"required"` // 消息ID
	IsLabel LabelMsgType `json:"is_label"`                  // 是否标注
}

// MarkGroupMessageRequest 用于标注群聊消息状态的请求结构体
type LabelGroupMessageRequest struct {
	MsgID   uint32       `json:"msg_id" binding:"required"` // 消息ID
	IsLabel LabelMsgType `json:"is_label"`                  // 是否标注
}

type ReadUserMsgsRequest struct {
	MsgIds   []uint32 `json:"msg_ids" binding:"required"`   //消息id
	DialogId uint32   `json:"dialog_id" binding:"required"` // 会话ID
}

// IsValidLabelMsgType 用于验证消息标注类型是否为正常类型
func IsValidLabelMsgType(label LabelMsgType) bool {
	return label == NotLabel || label == IsLabel
}

type LabelMsgType uint

const (
	NotLabel LabelMsgType = iota //不标注
	IsLabel                      //标注
)

type GetDialogAfterMsgRequest struct {
	AfterMsg `json:"msg_list"`
}
type AfterMsg struct {
	MsgId    uint32 `json:"msg_id"`
	DialogId uint32 `json:"dialog_id"`
}

type GetDialogAfterMsgResponse struct {
	DialogId uint32     `json:"dialog_id"`
	Messages []*Message `json:"msg_list"`
}

type GetDialogAfterMsgListResponse struct {
	MsgList []*GetDialogAfterMsgResponse `json:"msg_list"`
}

type BurnAfterReadingType uint

const (
	NotBurnAfterReading BurnAfterReadingType = iota //非阅后即焚
	IsBurnAfterReading                              //阅后即焚消息
)

func IsAllowedConversationType(isBurnAfterReading BurnAfterReadingType) bool {
	switch isBurnAfterReading {
	case NotBurnAfterReading, IsBurnAfterReading:
		return true
	default:
		return false
	}
}

type GroupMessageReadRequest struct {
	GroupId  uint32   `json:"group_id" binding:"required"`  // 群组ID
	DialogId uint32   `json:"dialog_id" binding:"required"` // 会话ID
	MsgIds   []uint32 `json:"msg_id" binding:"required"`    // 消息ID
}

type GetGroupMessageReadersResponse struct {
	UserId string `json:"user_id"`
	Avatar string `json:"avatar"`
	Name   string `json:"name"`
	//ReadAt int64  `json:"read_at"`
}
