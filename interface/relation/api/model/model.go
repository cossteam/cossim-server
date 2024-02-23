package model

import "errors"

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
	RequestAt    uint64    `json:"request_at" description:"申请时间"`
	Status       uint32    `json:"status" description:"申请状态 (0=申请中, 1=已通过, 2=被拒绝)"`
	SenderInfo   *UserInfo `json:"sender_info,omitempty"`
	ReceiverInfo *UserInfo `json:"receiver_info,omitempty"`
}

type UserInfo struct {
	UserID     string `json:"user_id,omitempty" description:"用户ID"`
	UserName   string `json:"user_name,omitempty" description:"用户昵称"`
	UserAvatar string `json:"user_avatar,omitempty" description:"用户头像"`
}

type DeleteFriendRequest struct {
	UserID string `json:"user_id" binding:"required"`
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
	GroupId  uint32                 `json:"group_id" binding:"required"`  // 群ID
	IsSilent SilentNotificationType `json:"is_silent" binding:"required"` // 是否开启静默通知
}

type SetUserSilentNotificationRequest struct {
	UserId   string                 `json:"user_id" binding:"required"`   // 用户ID
	IsSilent SilentNotificationType `json:"is_silent" binding:"required"` // 是否开启静默通知
}

type SilentNotificationType uint

const (
	NotSilent SilentNotificationType = iota //静默通知关闭
	IsSilent                                //开启静默通知
)

func IsValidSilentNotificationType(isSilent SilentNotificationType) bool {
	return isSilent == NotSilent || isSilent == IsSilent
}

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
	Identity GroupRelationIdentity `json:"identity"`
}

type GroupRelationIdentity uint

const (
	IdentityUser  GroupRelationIdentity = iota //普通用户
	IdentityAdmin                              //管理员
	IdentityOwner                              //群主
)

type OpenUserBurnAfterReadingRequest struct {
	UserId string                   `json:"user_id" binding:"required"` // 用户ID
	Action OpenBurnAfterReadingType `json:"action"`
}

type OpenBurnAfterReadingType uint

const (
	Close = iota
	Open
)

type OpenGroupBurnAfterReadingRequest struct {
	GroupId uint32                   `json:"group_id" binding:"required"` // 群组ID
	Action  OpenBurnAfterReadingType `json:"action"`
}

func IsValidOpenBurnAfterReadingType(input OpenBurnAfterReadingType) bool {
	return input == Close || input == Open
}

type CreateGroupAnnouncementRequest struct {
	GroupId uint32 `json:"group_id" binding:"required"` // 群组ID
	Title   string `json:"title" binding:"required"`    // 公告标题
	Content string `json:"content" binding:"required"`  // 公告内容
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
	Remark string `json:"remark" binding:"required"`  // 备注
}
