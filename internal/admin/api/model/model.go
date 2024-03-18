package model

type SendAllNotificationRequest struct {
	Content string `json:"content"`
}

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}
