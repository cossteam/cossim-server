package model

type WsUserMsg struct {
	MsgId              uint32               `json:"msg_id"`
	SenderId           string               `json:"sender_id"`
	ReceiverId         string               `json:"receiver_id"`
	Content            string               `json:"content"`
	MsgType            uint                 `json:"msgType"`
	ReplayId           uint                 `json:"reply_id"`
	SendAt             int64                `json:"send_at"`
	DialogId           uint32               `json:"dialog_id"`
	IsBurnAfterReading BurnAfterReadingType `json:"is_burn_after_reading"`
}

type WsGroupMsg struct {
	MsgId              uint32               `json:"msg_id"`
	GroupId            int64                `json:"group_id"`
	SenderId           string               `json:"sender_id"`
	Content            string               `json:"content"`
	MsgType            uint                 `json:"msgType"`
	ReplayId           uint                 `json:"reply_id"`
	SendAt             int64                `json:"send_at"`
	DialogId           uint32               `json:"dialog_id"`
	AtUsers            []string             `json:"at_users"`
	AtAllUser          AtAllUserType        `json:"at_all_users"`
	IsBurnAfterReading BurnAfterReadingType `json:"is_burn_after_reading"`
	SenderInfo         SenderInfo           `json:"sender_info"`
}

type SenderInfo struct {
	UserId string `json:"user_id"`
	Avatar string `json:"avatar"`
	Name   string `json:"name"`
}

type FriendOnlineStatusMsg struct {
	UserId string `json:"user_id"`
	Status int32  `json:"status"`
}
