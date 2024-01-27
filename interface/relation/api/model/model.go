package model

import "errors"

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type GroupRequestListResponse struct {
	GroupId         uint32 `json:"group_id" description:"群组ID"`
	GroupType       uint32 `json:"group_type" description:"群组类型"`
	GroupStatus     uint32 `json:"group_status" description:"群组状态"`
	MaxMembersLimit int32  `json:"max_members_limit,omitempty" description:"最大成员限制"`
	CreatorId       string `json:"creator_id,omitempty" description:"创建者ID"`
	GroupName       string `json:"group_name" description:"群组名称"`
	GroupAvatar     string `json:"group_avatar" description:"群组头像"`
	UserID          string `json:"user_id" description:"用户ID"`
	UserName        string `json:"user_name" description:"用户昵称"`
	UserAvatar      string `json:"user_avatar" description:"用户头像"`
	Msg             string `json:"msg" description:"申请消息"`
}

type UserRequestListResponse struct {
	ID           uint32    `json:"id"`
	SenderId     string    `json:"sender_id" description:"发送者ID"`
	ReceiverId   string    `json:"receiver_id" description:"接收者ID"`
	Remark       string    `json:"remark" description:"申请消息"`
	RequestAt    uint64    `json:"request_at" description:"申请时间"`
	Status       uint32    `json:"user_status" description:"申请状态 (0=申请中, 1=已通过, 2=被拒绝)"`
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

	// 添加其他验证逻辑...
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
	GroupID   uint32     `json:"group_id" binding:"required"`
	InviterId string     `json:"inviter_id"`
	Action    ActionEnum `json:"action"`
}

func (m *ManageJoinGroupRequest) Validator() error {
	if m.Action != ActionRejected && m.Action != ActionAccepted {
		return errors.New("invalid action")
	}
	return nil
}

type AdminManageJoinGroupRequest struct {
	GroupID uint32     `json:"group_id" binding:"required"`
	UserID  string     `json:"user_id" binding:"required"`
	Action  ActionEnum `json:"action"`
}

func (a *AdminManageJoinGroupRequest) Validator() error {
	if a.Action != ActionRejected && a.Action != ActionAccepted {
		return errors.New("invalid action")
	}
	return nil
}

type RemoveUserFromGroupRequest struct {
	GroupID uint32 `json:"group_id" binding:"required"`
	UserID  string `json:"user_id" binding:"required"`
}

type QuitGroupRequest struct {
	GroupID uint32 `json:"group_id" binding:"required"`
}

type SwitchUserE2EPublicKeyRequest struct {
	UserId    string `json:"user_id" binding:"required"`
	PublicKey string `json:"public_key" binding:"required"`
}

type SetGroupSilentNotificationRequest struct {
	GroupId  uint32                 `json:"group_id" binding:"required"` // 群ID
	IsSilent SilentNotificationType `json:"is_silent" `                  // 是否开启静默通知
}

type SetUserSilentNotificationRequest struct {
	UserId   string                 `json:"user_id" binding:"required"` // 用户ID
	IsSilent SilentNotificationType `json:"is_silent" `                 // 是否开启静默通知
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
	UserId string `json:"user_id" binding:"required"`
	Remark string `json:"remark"`
}
