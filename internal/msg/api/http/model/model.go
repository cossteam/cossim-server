package model

import "encoding/json"

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type SendUserMsgRequest struct {
	DialogId               uint32               `json:"dialog_id" binding:"required"`
	ReceiverId             string               `json:"receiver_id" binding:"required"`
	Content                string               `json:"content" binding:"required"`
	Type                   UserMessageType      `json:"type" binding:"required"`
	ReplyId                uint                 `json:"reply_id"`
	IsBurnAfterReadingType BurnAfterReadingType `json:"is_burn_after_reading"`
}

type SendGroupMsgRequest struct {
	DialogId               uint32               `json:"dialog_id" binding:"required"`
	GroupId                uint32               `json:"group_id" binding:"required"`
	Content                string               `json:"content" binding:"required"`
	Type                   UserMessageType      `json:"type" binding:"required"`
	ReplyId                uint32               `json:"reply_id"`
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
	GroupId            uint32               `json:"group_id,omitempty"`      //群聊id
	MsgType            uint                 `json:"msg_type"`                // 消息类型
	Content            string               `json:"content"`                 // 消息内容
	SenderId           string               `json:"sender_id"`               // 消息发送者
	SendTime           int64                `json:"send_time"`               // 消息发送时间
	MsgId              uint64               `json:"msg_id"`                  // 消息id
	SenderInfo         SenderInfo           `json:"sender_info"`             // 消息发送者信息
	ReceiverInfo       SenderInfo           `json:"receiver_info,omitempty"` // 消息接受者信息
	AtAllUser          AtAllUserType        `json:"at_all_user,omitempty"`   // @全体用户
	AtUsers            []string             `json:"at_users,omitempty"`      // @用户id
	IsBurnAfterReading BurnAfterReadingType `json:"is_burn_after_reading"`   // 是否阅后即焚
	IsLabel            LabelMsgType         `json:"is_label"`                // 是否标记
	ReplyId            uint32               `json:"reply_id"`                // 回复消息id
}

type UserMessage struct {
	MsgId                   uint32               `json:"msg_id"`
	SenderId                string               `json:"sender_id"`
	ReceiverId              string               `json:"receiver_id"`
	Content                 string               `json:"content"`
	Type                    uint32               `json:"type"`
	ReplyId                 uint64               `json:"reply_id"`
	IsRead                  int32                `json:"is_read"`
	ReadAt                  int64                `json:"read_at"`
	CreatedAt               int64                `json:"created_at"`
	DialogId                uint32               `json:"dialog_id"`
	IsLabel                 LabelMsgType         `json:"is_label"`
	IsBurnAfterReadingType  BurnAfterReadingType `json:"is_burn_after_reading"`
	BurnAfterReadingTimeOut int64                `json:"burn_after_reading_time_out"`
	SenderInfo              SenderInfo           `json:"sender_info"`
	ReceiverInfo            SenderInfo           `json:"receiver_info"`
}

type GetUserMsgListResponse struct {
	UserMessages []*UserMessage `json:"user_messages"`
	Total        int32          `json:"total"`
	CurrentPage  int32          `json:"current_page"`
}

type GroupMessage struct {
	MsgId                  uint32               `protobuf:"varint,1,opt,name=Id,proto3" json:"msg_id"`
	GroupId                uint32               `protobuf:"varint,2,opt,name=Group_id,json=GroupId,proto3" json:"group_id"`
	Type                   uint32               `protobuf:"varint,3,opt,name=Type,proto3" json:"type"`
	ReplyId                uint32               `protobuf:"varint,4,opt,name=Reply_id,json=ReplyId,proto3" json:"reply_id"`
	ReadCount              int32                `protobuf:"varint,5,opt,name=Read_count,json=ReadCount,proto3" json:"read_count"`
	UserId                 string               `protobuf:"bytes,6,opt,name=UserId,proto3" json:"user_id"`
	Content                string               `protobuf:"bytes,7,opt,name=Content,proto3" json:"content"`
	CreatedAt              int64                `protobuf:"varint,8,opt,name=Created_at,json=CreatedAt,proto3" json:"created_at"`
	DialogId               uint32               `protobuf:"varint,9,opt,name=Dialog_id,json=DialogId,proto3" json:"dialog_id"`
	IsLabel                LabelMsgType         `protobuf:"varint,10,opt,name=IsLabel,proto3,enum=v1.MsgLabel" json:"is_label"`
	IsBurnAfterReadingType BurnAfterReadingType `protobuf:"varint,11,opt,name=IsBurnAfterReadingType,proto3,enum=v1.BurnAfterReadingType" json:"is_burn_after_reading"`
	AtUsers                []string             `protobuf:"bytes,12,rep,name=AtUsers,proto3" json:"at_users"`
	AtAllUser              AtAllUserType        `protobuf:"varint,13,opt,name=AtAllUser,proto3,enum=v1.AtAllUserType" json:"at_all_user"`
	ReadAt                 int64                `json:"read_at"`
	IsRead                 int32                `json:"is_read"`
	SenderInfo             SenderInfo           `json:"sender_info"`
}

type GetGroupMsgListResponse struct {
	GroupMessages []*GroupMessage `protobuf:"bytes,1,rep,name=GroupMessages,proto3" json:"group_messages"`
	Total         int32           `protobuf:"varint,2,opt,name=Total,proto3" json:"total"`
	CurrentPage   int32           `protobuf:"varint,3,opt,name=CurrentPage,proto3" json:"current_page"`
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
	MsgIds   []uint32 `json:"msg_ids" binding:"required"`   // 消息ID
}

type GetGroupMessageReadersResponse struct {
	UserId string `json:"user_id"`
	Avatar string `json:"avatar"`
	Name   string `json:"name"`
	//ReadAt int64  `json:"read_at"`
}

func (udlr UserDialogListResponse) MarshalBinary() ([]byte, error) {
	// 将UserDialogListResponse对象转换为二进制数据
	data, err := json.Marshal(udlr)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type UserMessageType uint

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
