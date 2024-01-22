package config

type WSEventType int

const (
	// OnlineEvent 上线事件
	OnlineEvent WSEventType = iota + 1
	// OfflineEvent 下线事件
	OfflineEvent
	// SendUserMessageEvent 发送消息事件
	SendUserMessageEvent
	SendGroupMessageEvent
	// SystemNotificationEvent 推送系统通知事件
	SystemNotificationEvent
	// AddFriendEvent 添加好友事件
	AddFriendEvent
	// ManageFriendEvent 管理好友请求
	ManageFriendEvent
	// PushE2EPublicKeyEvent 推送好友公钥接口
	PushE2EPublicKeyEvent

	// JoinGroupEvent 申请加入群聊
	JoinGroupEvent
	// ApproveJoinGroupEvent 同意加入群聊
	ApproveJoinGroupEvent

	// InviteJoinGroupEvent 邀请加入群聊事件
	InviteJoinGroupEvent
)

type WsMsg struct {
	Uid    string      `json:"uid"`
	Event  WSEventType `json:"event"`
	Rid    int64       `json:"rid"`
	SendAt int64       `json:"send_at"`
	Data   interface{} `json:"data"`
}
