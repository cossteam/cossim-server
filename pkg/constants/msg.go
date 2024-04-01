package constants

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
