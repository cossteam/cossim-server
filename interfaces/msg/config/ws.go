package config

type WSEventType int

const (
	// 上线事件
	OnlineEvent WSEventType = iota
	// 下线事件
	OfflineEvent
	// 发送消息事件
	SendMessageEvent
	// 推送系统通知事件
	SystemNotificationEvent
)
