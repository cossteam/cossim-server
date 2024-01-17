package model

type WsUserMsg struct {
	SenderId string `json:"sender_id"`
	Content  string `json:"content"`
	MsgType  uint   `json:"msgType"`
	ReplayId uint   `json:"reply_id"`
	SendAt   int64  `json:"send_at"`
	DialogId uint32 `json:"dialog_id"`
}

type WsGroupMsg struct {
	GroupId  int64  `json:"group_id"`
	UserId   string `json:"uid"`
	Content  string `json:"content"`
	MsgType  uint   `json:"msgType"`
	ReplayId uint   `json:"reply_id"`
	SendAt   int64  `json:"send_at"`
	DialogId uint32 `json:"dialog_id"`
}
