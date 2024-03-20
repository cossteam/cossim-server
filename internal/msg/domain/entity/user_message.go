package entity

type UserMessage struct {
	BaseModel
	Type               UserMessageType      `gorm:";comment:消息类型" json:"type"`
	DialogId           uint                 `gorm:"default:0;comment:对话ID" json:"dialog_id"`
	IsRead             ReadType             `gorm:"default:0;comment:是否已读" json:"is_read"`
	ReplyId            uint                 `gorm:"default:0;comment:回复ID" json:"reply_id"`
	ReadAt             int64                `gorm:"comment:阅读时间" json:"read_at"`
	ReceiveID          string               `gorm:"default:0;comment:接收用户id" json:"receive_id"`
	SendID             string               `gorm:"default:0;comment:发送用户id" json:"send_id"`
	Content            string               `gorm:"longtext;comment:详细消息" json:"content"`
	IsLabel            uint                 `gorm:"default:0;comment:是否标注" json:"is_label"`
	IsBurnAfterReading BurnAfterReadingType `gorm:"default:0;comment:是否阅后即焚消息" json:"is_burn_after_reading"`
	ReplyEmoji         string               `gorm:"comment:回复时使用的 Emoji" json:"reply_emoji"`
}

type BurnAfterReadingType uint

const (
	NotBurnAfterReading BurnAfterReadingType = iota //非阅后即焚
	IsBurnAfterReading                              //阅后即焚消息
)

type UserMessageType uint

const (
	MessageTypeText       UserMessageType = iota + 1 // 文本消息
	MessageTypeVoice                                 // 语音消息
	MessageTypeImage                                 // 图片消息
	MessageTypeLabel                                 //标注
	MessageTypeNotice                                //群公告
	MessageTypeFile                                  // 文件消息
	MessageTypeVideo                                 // 视频消息
	MessageTypeEmojiReply                            //emoji回复
	MessageTypeVoiceCall                             // 语音通话
	MessageTypeVideoCall                             // 视频通话
	MessageTypeDelete                                // 撤回消息
	MessageTypeCancelLabel
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
