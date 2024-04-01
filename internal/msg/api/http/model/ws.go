package model

type WsUserMsg struct {
	MsgId                   uint32               `json:"msg_id"`
	SenderId                string               `json:"sender_id"`
	ReceiverId              string               `json:"receiver_id"`
	Content                 string               `json:"content"`
	MsgType                 uint                 `json:"msg_type"`
	ReplyId                 uint                 `json:"reply_id"`
	SendAt                  int64                `json:"send_at"`
	DialogId                uint32               `json:"dialog_id"`
	IsBurnAfterReading      BurnAfterReadingType `json:"is_burn_after_reading"`
	BurnAfterReadingTimeOut int64                `json:"burn_after_reading_time_out"`
	SenderInfo              SenderInfo           `json:"sender_info"`
	ReplyMsg                *Message             `json:"reply_msg,omitempty"`
}

type WsGroupMsg struct {
	MsgId              uint32               `json:"msg_id"`
	GroupId            int64                `json:"group_id"`
	SenderId           string               `json:"sender_id"`
	Content            string               `json:"content"`
	MsgType            uint                 `json:"msg_type"`
	ReplyId            uint                 `json:"reply_id"`
	SendAt             int64                `json:"send_at"`
	DialogId           uint32               `json:"dialog_id"`
	AtUsers            []string             `json:"at_users"`
	AtAllUser          AtAllUserType        `json:"at_all_users"`
	IsBurnAfterReading BurnAfterReadingType `json:"is_burn_after_reading"`
	SenderInfo         SenderInfo           `json:"sender_info"`
	ReplyMsg           *Message             `json:"reply_msg,omitempty"`
}

type SenderInfo struct {
	UserId string `json:"user_id"`
	Avatar string `json:"avatar"`
	Name   string `json:"name"`
}

type WsUserOperatorMsg struct {
	Id                     uint32               `json:"id"`
	SenderId               string               `json:"sender_id"`
	ReceiverId             string               `json:"receiver_id"`
	Content                string               `json:"content"`
	Type                   uint32               `json:"type"`
	ReplyId                uint64               `json:"reply_id"`
	IsRead                 int32                `json:"is_read"`
	ReadAt                 int64                `json:"read_at"`
	CreatedAt              int64                `json:"created_at"`
	DialogId               uint32               `json:"dialog_id"`
	IsLabel                LabelMsgType         `json:"is_label"`
	IsBurnAfterReadingType BurnAfterReadingType `json:"is_burn_after_reading"`
	OperatorInfo           SenderInfo           `json:"operator_info"`
}
