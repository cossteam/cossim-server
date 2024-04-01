package constants

//type WSEventType int

type JoinGroupEventData struct {
	GroupId uint32 `json:"group_id"`
	UserId  string `json:"user_id"`
	Status  uint32 `json:"status,omitempty"`
}

type SystemNotificationEventData struct {
	UserId  string `json:"user_id"`
	Content string `json:"content"`
	Type    uint32 `json:"type"`
}

//const (
//	// OnlineEvent 上线事件
//	OnlineEvent WSEventType = iota + 1
//
//	// OfflineEvent 下线事件
//	OfflineEvent
//
//	// SendUserMessageEvent 发送消息事件
//	SendUserMessageEvent
//	SendGroupMessageEvent
//
//	// SystemNotificationEvent 推送系统通知事件
//	SystemNotificationEvent
//
//	// AddFriendEvent 添加好友事件
//	AddFriendEvent
//
//	// ManageFriendEvent 管理好友请求
//	ManageFriendEvent
//
//	// PushE2EPublicKeyEvent 推送好友公钥接口
//	PushE2EPublicKeyEvent
//
//	// JoinGroupEvent 申请加入群聊
//	JoinGroupEvent
//
//	// ApproveJoinGroupEvent 同意加入群聊
//	ApproveJoinGroupEvent
//
//	// InviteJoinGroupEvent 邀请加入群聊事件
//	InviteJoinGroupEvent
//
//	//发送静默私聊消息
//	SendSilentUserMessageEvent
//	//发送静默群聊消息
//	SendSilentGroupMessageEvent
//
//	// 用户通话呼叫请求事件
//	UserCallReqEvent
//	// 群聊通话呼叫请求事件
//	GroupCallReqEvent
//	// 用户通话呼叫拒绝事件
//	UserCallRejectEvent
//	// 群聊通话呼叫拒绝事件
//	GroupCallRejectEvent
//	// 用户通话结束事件
//	UserCallEndEvent
//	// 群聊通话结束事件
//	GroupCallEndEvent
//
//	//群聊消息已读事件
//	GroupMsgReadEvent
//
//	//好友在线状态变更事件
//	FriendUpdateOnlineStatusEvent
//
//	//推送所有好友在线状态事件
//	PushAllFriendOnlineStatusEvent
//
//	//标注消息事件
//	LabelMsgEvent
//
//	//编辑消息事件
//	EditMsgEvent
//
//	//撤回消息事件
//	RecallMsgEvent
//
//	//私聊已读事件
//	UserMsgReadEvent
//
//	// 用户通话接受事件
//	UserCallAcceptEvent
//
//	// 群聊通话接受事件
//	GroupCallAcceptEvent
//
//	//创建群公告
//	CreateGroupAnnouncementEvent
//
//	//修改群公告
//	UpdateGroupAnnouncementEvent
//
//	// 用户离开群聊通话
//	UserLeaveGroupCallEvent
//)
//
//type WsMsg struct {
//	Uid      string      `json:"uid"`
//	Event    WSEventType `json:"event"`
//	Rid      int64       `json:"rid"`
//	DriverId string      `json:"driverId"`
//	SendAt   int64       `json:"send_at"`
//	Data     interface{} `json:"data"`
//}

type OfflineEventData struct {
	Rid        int64      `json:"rid"`
	DriverType DriverType `json:"driver_type"`
}

type ManageFriendEventData struct {
	UserId       string      `json:"user_id" binding:"required"`
	Status       uint32      `json:"status" binding:"required"`
	E2EPublicKey string      `json:"e2e_public_key,omitempty"`
	TargetInfo   interface{} `json:"target_info,omitempty"`
}

type AddFriendEventData struct {
	UserId       string `json:"user_id"`
	Msg          string `json:"msg"`
	E2EPublicKey string `json:"e2e_public_key,omitempty"`
}

type BurnAfterReadingType uint

const (
	NotBurnAfterReading BurnAfterReadingType = iota //非阅后即焚
	IsBurnAfterReading                              //阅后即焚消息
)

type SenderInfo struct {
	UserId string `json:"user_id"`
	Avatar string `json:"avatar"`
	Name   string `json:"name"`
}

type AtAllUserType uint

const (
	NotAtAllUser = iota
	AtAllUser
)

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
}
