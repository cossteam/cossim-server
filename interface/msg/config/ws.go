package config

type WSEventType int

const (
	// 上线事件
	OnlineEvent WSEventType = iota + 1
	// 下线事件
	OfflineEvent
	// 发送消息事件
	SendUserMessageEvent
	SendGroupMessageEvent
	// 推送系统通知事件
	SystemNotificationEvent
	// 添加好友事件
	AddFriendEvent
	//确认好友
	ConfirmFriendEvent
)
