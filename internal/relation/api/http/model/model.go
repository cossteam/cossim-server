package model

import (
	"encoding/json"
	"errors"
)

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type GroupRequestListResponse struct {
	ID              uint32             `json:"id"`
	GroupId         uint32             `json:"group_id" description:"群组ID"`
	GroupType       uint32             `json:"group_type" description:"群组类型"`
	GroupStatus     uint32             `json:"group_status" description:"群组状态"`
	MaxMembersLimit int32              `json:"max_members_limit,omitempty" description:"最大成员限制"`
	CreatorId       string             `json:"creator_id,omitempty" description:"创建者ID"`
	GroupName       string             `json:"group_name" description:"群组名称"`
	GroupAvatar     string             `json:"group_avatar" description:"群组头像"`
	SenderInfo      *UserInfo          `json:"sender_info" description:"发送者信息"`
	ReceiverInfo    *UserInfo          `json:"receiver_info" description:"接收者信息"`
	Status          GroupRequestStatus `json:"status" description:"请求状态"`
	Remark          string             `json:"remark" description:"申请消息"`
	CreateAt        int64              `json:"create_at"`
	ExpiredAt       int64              `json:"expired_at"`
}

type GroupRequestListResponseList struct {
	List        []*GroupRequestListResponse `json:"list"`
	Total       int64                       `json:"total"`
	CurrentPage int32                       `json:"current_page"`
}

type AddGroupAdminRequest struct {
	GroupID uint32   `json:"group_id" binding:"required"`
	UserIDs []string `json:"user_ids" binding:"required"`
}

type SetGroupUserRemarkRequest struct {
	GroupId uint32 `json:"group_id" binding:"required"`
	Remark  string `json:"remark" description:"备注信息"`
}

type GroupRequestStatus uint32

const (
	Pending            GroupRequestStatus = iota // 等待
	Accepted                                     // 已通过
	Rejected                                     // 已拒绝
	InviteSender                                 // 邀请发送者
	InvitationReceived                           // 邀请接收者
)

type UserRequestListResponse struct {
	ID           uint32    `json:"id"`
	SenderId     string    `json:"sender_id" description:"发送者ID"`
	ReceiverId   string    `json:"receiver_id" description:"接收者ID"`
	Remark       string    `json:"remark" description:"申请消息"`
	Status       uint32    `json:"status" description:"申请状态 (0=申请中, 1=已通过, 2=被拒绝)"`
	SenderInfo   *UserInfo `json:"sender_info,omitempty"`
	ReceiverInfo *UserInfo `json:"receiver_info,omitempty"`
	CreateAt     int64     `json:"create_at"`
	ExpiredAt    int64     `json:"expired_at"`
}

type UserRequestListResponseList struct {
	List        []*UserRequestListResponse `json:"list"`
	Total       int64                      `json:"total"`
	CurrentPage int32                      `json:"current_page"`
}

type UserInfo struct {
	UserID     string `json:"user_id,omitempty" description:"用户ID"`
	UserName   string `json:"user_name,omitempty" description:"用户昵称"`
	UserAvatar string `json:"user_avatar,omitempty" description:"用户头像"`
}

type DeleteFriendRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type DeleteRecordRequest struct {
	ID uint32 `json:"id" binding:"required"` // 申请记录id
}

type DeleteBlacklistRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type AddBlacklistRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type ManageFriendRequest struct {
	RequestID    uint32     `json:"request_id" binding:"required"`
	Action       ActionEnum `json:"action"`
	E2EPublicKey string     `json:"e2e_public_key"`
}

type ActionEnum int

const (
	// ActionRejected 拒绝
	ActionRejected ActionEnum = iota // 拒绝
	// ActionAccepted 同意
	ActionAccepted // 同意
)

func (m *ManageFriendRequest) Validator() error {
	if m.Action != ActionRejected && m.Action != ActionAccepted {
		return errors.New("invalid action")
	}

	return nil
}

type AddFriendRequest struct {
	UserID       string `json:"user_id" binding:"required"`
	Msg          string `json:"msg"`
	E2EPublicKey string `json:"e2e_public_key"`
}

type JoinGroupRequest struct {
	GroupID uint32 `json:"group_id" binding:"required"`
}

type InviteGroupRequest struct {
	GroupID uint32   `json:"group_id" binding:"required"`
	Member  []string `json:"member"  binding:"required"`
}

type ManageJoinGroupRequest struct {
	ID      uint32     `json:"id" binding:"required"`
	GroupID uint32     `json:"group_id" binding:"required"`
	Action  ActionEnum `json:"action"`
}

func (m *ManageJoinGroupRequest) Validator() error {
	if m.Action != ActionRejected && m.Action != ActionAccepted {
		return errors.New("invalid action")
	}
	return nil
}

type RemoveUserFromGroupRequest struct {
	GroupID uint32   `json:"group_id" binding:"required"`
	Member  []string `json:"member" binding:"required"`
}

type QuitGroupRequest struct {
	GroupID uint32 `json:"group_id" binding:"required"`
}

type SwitchUserE2EPublicKeyRequest struct {
	UserId    string `json:"user_id" binding:"required"`
	PublicKey string `json:"public_key" binding:"required"`
}

type SetGroupSilentNotificationRequest struct {
	GroupId  uint32 `json:"group_id" binding:"required"` // 群ID
	IsSilent bool   `json:"is_silent"`                   // 是否开启静默通知
}

type SetUserSilentNotificationRequest struct {
	UserId   string `json:"user_id" binding:"required"` // 用户ID
	IsSilent bool   `json:"is_silent"`                  // 是否开启静默通知
}

//
//type SilentNotificationType uint
//
//const (
//	NotSilent SilentNotificationType = iota //静默通知关闭
//	IsSilent                                //开启静默通知
//)
//
//func IsValidSilentNotificationType(isSilent SilentNotificationType) bool {
//	return isSilent == NotSilent || isSilent == IsSilent
//}

type SendFriendRequest struct {
	UserId       string `json:"user_id" binding:"required"`
	Remark       string `json:"remark"`
	E2EPublicKey string `json:"e2e_public_key"`
}

type CloseOrOpenDialogRequest struct {
	DialogId uint32                  `json:"dialog_id" binding:"required"`
	Action   CloseOrOpenDialogAction `json:"action"`
}

type CloseOrOpenDialogAction uint

const (
	CloseDialog CloseOrOpenDialogAction = iota
	OpenDialog
)

func IsValidOpenAction(isOpen CloseOrOpenDialogAction) bool {
	return isOpen == CloseDialog || isOpen == OpenDialog
}

type TopOrCancelTopDialogRequest struct {
	DialogId uint32               `json:"dialog_id" binding:"required"`
	Action   TopOrCancelTopAction `json:"action"`
}

func IsValidTopAction(isTop TopOrCancelTopAction) bool {
	return isTop == CancelTopDialog || isTop == TopDialog
}

type TopOrCancelTopAction uint

const (
	CancelTopDialog TopOrCancelTopAction = iota
	TopDialog
)

type RequestListResponse struct {
	UserID   string                `json:"user_id"`
	Nickname string                `json:"nickname"`
	Avatar   string                `json:"avatar"`
	Remark   string                `json:"remark"`
	Identity GroupRelationIdentity `json:"identity"`
}

type GroupRelationIdentity uint

const (
	IdentityUser  GroupRelationIdentity = iota //普通用户
	IdentityAdmin                              //管理员
	IdentityOwner                              //群主
)

type OpenUserBurnAfterReadingRequest struct {
	UserId               string `json:"user_id" binding:"required"` // 用户ID
	TimeOut              int64  `json:"timeout"`
	OpenBurnAfterReading bool   `json:"open_burn_after_reading"`
}

//type OpenBurnAfterReadingType uint
//
//const (
//	BurnClose OpenBurnAfterReadingType = iota
//	BurnOpen
//)

type OpenGroupBurnAfterReadingRequest struct {
	GroupId              uint32 `json:"group_id" binding:"required"` // 群组ID
	OpenBurnAfterReading bool   `json:"open_burn_after_reading"`
}

//func IsValidOpenBurnAfterReadingType(input OpenBurnAfterReadingType) bool {
//	return input == BurnClose || input == BurnOpen
//}

type CreateGroupAnnouncementRequest struct {
	GroupId uint32 `json:"group_id" binding:"required"` // 群组ID
	Title   string `json:"title"`                       // 公告标题
	Content string `json:"content" binding:"required"`  // 公告内容
}

type CreateGroupAnnouncementResponse struct {
	Id           uint32     `json:"id" binding:"required"`       // 公告ID
	GroupId      uint32     `json:"group_id" binding:"required"` // 群组ID
	Title        string     `json:"title"`                       // 公告标题
	Content      string     `json:"content" binding:"required"`  // 公告内容
	CreateAt     int64      `json:"create_at"`
	UpdateAt     int64      `json:"update_at"`
	OperatorInfo SenderInfo `json:"operator_info"`
}

type GetGroupAnnouncementListResponse struct {
	Id           uint32                                  `json:"id" binding:"required"`       // 公告ID
	GroupId      uint32                                  `json:"group_id" binding:"required"` // 群组ID
	Title        string                                  `json:"title"`                       // 公告标题
	Content      string                                  `json:"content" binding:"required"`  // 公告内容
	CreateAt     int64                                   `json:"create_at"`
	UpdateAt     int64                                   `json:"update_at"`
	OperatorInfo SenderInfo                              `json:"operator_info"`
	ReadUserList []*GetGroupAnnouncementReadUsersRequest `json:"read_user_list"` // 已读用户列表
}

type DeleteGroupAnnouncementRequest struct {
	Id      uint32 `json:"id" binding:"required"`       // 公告ID
	GroupId uint32 `json:"group_id" binding:"required"` // 群组ID
}

type UpdateGroupAnnouncementRequest struct {
	Id      uint32 `json:"id" binding:"required"`       // 公告ID
	GroupId uint32 `json:"group_id" binding:"required"` // 群组ID
	Title   string `json:"title" binding:"required"`    // 公告标题
	Content string `json:"content" binding:"required"`  // 公告内容
}

type SetUserFriendRemarkRequest struct {
	UserId string `json:"user_id" binding:"required"` // 用户ID
	Remark string `json:"remark"`                     // 备注
}

type WsGroupRelationOperatorMsg struct {
	Id           uint32     `json:"id" binding:"required"`       // 公告ID
	GroupId      uint32     `json:"group_id" binding:"required"` // 群组ID
	Title        string     `json:"title"`                       // 公告标题
	Content      string     `json:"content" binding:"required"`  // 公告内容
	OperatorInfo SenderInfo `json:"operator_info"`
}

type SenderInfo struct {
	UserId string `json:"user_id"`
	Avatar string `json:"avatar"`
	Name   string `json:"name"`
}

type ReadGroupAnnouncementRequest struct {
	GroupId uint32 `json:"group_id" binding:"required"` // 群组ID
	Id      uint32 `json:"id" binding:"required"`       // 公告ID
}

type GetGroupAnnouncementReadUsersRequest struct {
	ID             uint32     `json:"id" binding:"required"`              // 公告ID
	GroupId        uint32     `json:"group_id" binding:"required"`        // 群组ID
	AnnouncementId uint32     `json:"announcement_id" binding:"required"` // 公告ID
	ReadAt         int64      `json:"read_at"`
	UserId         string     `json:"user_id"`
	ReaderInfo     SenderInfo `json:"reader_info"`
}

type ConversationType uint

const (
	UserConversation ConversationType = iota
	GroupConversation
)

type UserDialogListResponse struct {
	DialogId uint32 `json:"dialog_id"`
	UserId   string `json:"user_id,omitempty"`
	GroupId  uint32 `json:"group_id,omitempty"`
	// 会话类型
	DialogType ConversationType `json:"dialog_type"`
	// 会话名称
	DialogName string `json:"dialog_name"`
	// 会话头像
	DialogAvatar string `json:"dialog_avatar"`
	// 会话未读消息数
	DialogUnreadCount int     `json:"dialog_unread_count"`
	LastMessage       Message `json:"last_message"`

	DialogCreateAt int64 `json:"dialog_create_at"`
	TopAt          int64 `json:"top_at"`
}
type Message struct {
	GroupId            uint32               `json:"group_id,omitempty"`      //群聊id
	MsgType            uint                 `json:"msg_type"`                // 消息类型
	Content            string               `json:"content"`                 // 消息内容
	SenderId           string               `json:"sender_id"`               // 消息发送者
	SendAt             int64                `json:"send_at"`                 // 消息发送时间
	MsgId              uint64               `json:"msg_id"`                  // 消息id
	SenderInfo         SenderInfo           `json:"sender_info"`             // 消息发送者信息
	ReceiverInfo       SenderInfo           `json:"receiver_info,omitempty"` // 消息接受者信息
	AtAllUser          AtAllUserType        `json:"at_all_user,omitempty"`   // @全体用户
	AtUsers            []string             `json:"at_users,omitempty"`      // @用户id
	IsBurnAfterReading BurnAfterReadingType `json:"is_burn_after_reading"`   // 是否阅后即焚
	IsLabel            LabelMsgType         `json:"is_label"`                // 是否标记
	ReplyId            uint32               `json:"reply_id"`                // 回复消息id
}

type BurnAfterReadingType uint

const (
	NotBurnAfterReading BurnAfterReadingType = iota //非阅后即焚
	IsBurnAfterReading                              //阅后即焚消息
)

type LabelMsgType uint

const (
	NotLabel LabelMsgType = iota //不标注
	IsLabel                      //标注
)

type AtAllUserType uint

const (
	NotAtAllUser = iota
	AtAllUser
)

func (udlr UserDialogListResponse) MarshalBinary() ([]byte, error) {
	// 将UserDialogListResponse对象转换为二进制数据
	data, err := json.Marshal(udlr)
	if err != nil {
		return nil, err
	}
	return data, nil
}

type SetGroupOpenBurnAfterReadingTimeOutRequest struct {
	GroupId                     uint32 `json:"group_id" binding:"required"`
	OpenBurnAfterReadingTimeOut int64  `json:"open_burn_after_reading_time_out" binding:"required"`
}

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

type SendGroupMsgRequest struct {
	DialogId               uint32               `json:"dialog_id" binding:"required"`
	GroupId                uint32               `json:"group_id" binding:"required"`
	Content                string               `json:"content" binding:"required"`
	Type                   UserMessageType      `json:"type" binding:"required"`
	ReplyId                uint32               `json:"reply_id"`
	AtUsers                []string             `json:"at_users"`
	AtAllUser              AtAllUserType        `json:"at_all_user"`
	IsBurnAfterReadingType BurnAfterReadingType `json:"is_burn_after_reading"`
}
