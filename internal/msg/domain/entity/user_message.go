package entity

import v1 "github.com/cossim/coss-server/internal/msg/api/http/v1"

type UserMessage struct {
	BaseModel
	Type               UserMessageType
	SubType            UserMessageSubType
	DialogId           uint
	IsRead             ReadType
	ReplyId            uint
	ReadAt             int64
	ReceiveID          string
	SendID             string
	Content            string
	IsLabel            bool
	IsBurnAfterReading bool
	ReplyEmoji         string
}

//type BurnAfterReadingType uint
//
//const (
//	NotBurnAfterReading BurnAfterReadingType = iota // 非阅后即焚
//	IsBurnAfterReading                              // 阅后即焚消息
//)

type UserMessageType uint

const (
	MessageTypeText        UserMessageType = iota + 1 // 文本消息
	MessageTypeVoice                                  // 语音消息
	MessageTypeImage                                  // 图片消息
	MessageTypeLabel                                  // 标注
	MessageTypeNotice                                 // 群公告
	MessageTypeFile                                   // 文件消息
	MessageTypeVideo                                  // 视频消息
	MessageTypeEmojiReply                             // emoji回复
	MessageTypeVoiceCall                              // 语音通话
	MessageTypeVideoCall                              // 视频通话
	MessageTypeDelete                                 // 撤回消息
	MessageTypeCancelLabel                            // 取消标注
)

type UserMessageSubType uint

const (
	CallNormal    UserMessageSubType = iota // 正常通话
	CallCancelled                    = 1    // 取消通话
	CallRejected                     = 2    // 拒绝通话
	CallMissed                       = 3    // 未接通话
)

type MessageLabelType uint

const (
	NotLabel MessageLabelType = iota // 不标记消息
	IsLabel                          // 标记消息
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

type ReadType uint

const (
	NotRead ReadType = iota
	IsRead
)

func (um *UserMessage) ToMessage() *v1.Message {
	return &v1.Message{
		Content:            um.Content,
		IsBurnAfterReading: um.IsBurnAfterReading,
		IsLabel:            um.IsLabel,
		IsRead:             um.IsRead == IsRead,
		MsgId:              int(um.ID),
		MsgType:            int(um.Type),
		ReadAt:             int(um.ReadAt),
		ReplyId:            int(um.ReplyId),
		SendAt:             int(um.CreatedAt),
		SenderId:           um.SendID,
		SenderInfo:         nil, // 需要确定如何设置 SenderInfo
		ReceiverInfo:       nil, // 需要确定如何设置 ReceiverInfo
		DialogId:           int(um.DialogId),
	}
}
