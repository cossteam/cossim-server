package entity

type UserMessage struct {
	BaseModel
	Type      UserMessageType `gorm:";comment:消息类型" json:"type"`
	DialogId  uint            `gorm:"default:0;comment:对话ID" json:"dialog_id"`
	IsRead    uint            `gorm:"default:0;comment:是否已读" json:"is_read"`
	ReplyId   uint            `gorm:"default:0;comment:回复ID" json:"reply_id"`
	ReadAt    int64           `gorm:"comment:阅读时间" json:"read_at"`
	ReceiveID string          `gorm:"default:0;comment:接收用户id" json:"receive_id"`
	SendID    string          `gorm:"default:0;comment:发送用户id" json:"send_id"`
	Content   string          `gorm:"longtext;comment:详细消息" json:"content"`
	IsLabel   uint            `gorm:"default:0;comment:是否标注" json:"is_label"`
}

type UserMessageType uint

const (
	MessageTypeText      UserMessageType = iota + 1 // 文本消息
	MessageTypeVoice                                // 语音消息
	MessageTypeImage                                // 图片消息
	MessageTypeFile                                 // 文件消息
	MessageTypeVideo                                // 视频消息
	MessageTypeEmoji                                // Emoji表情
	MessageTypeSticker                              // 表情包
	MessageTypeVoiceCall                            // 语音通话
	MessageTypeVideoCall                            // 视频通话
	MessageTypeDelete                               //阅后即焚消息

)

type MessageLabelType uint

const (
	NotLabel MessageLabelType = iota // 不标记消息
	IsLabel                          // 标记消息
)

// IsValidMessageType 判断是否是有效的消息类型
func IsValidMessageType(msgType UserMessageType) bool {
	validTypes := map[UserMessageType]struct{}{
		MessageTypeText:      {},
		MessageTypeVoice:     {},
		MessageTypeImage:     {},
		MessageTypeFile:      {},
		MessageTypeVideo:     {},
		MessageTypeEmoji:     {},
		MessageTypeSticker:   {},
		MessageTypeVoiceCall: {},
		MessageTypeVideoCall: {},
	}

	_, isValid := validTypes[msgType]
	return isValid
}
