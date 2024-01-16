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
	//管理好友请求
	ManageFriendEvent
	//推送好友公钥接口
	PushE2EPublicKeyEvent

	//申请加入群聊
	JoinGroupEvent
	//同意加入群聊
	ApproveJoinGroupEvent
)

type WsMsg struct {
	Uid    string      `json:"uid"`
	Event  WSEventType `json:"event"`
	Rid    int64       `json:"rid"`
	SendAt int64       `json:"send_at"`
	Data   interface{} `json:"data"`
}
