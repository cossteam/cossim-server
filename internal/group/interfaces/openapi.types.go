// Package interfaces provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen/v2 version v2.1.0 DO NOT EDIT.
package interfaces

// Defines values for CreateGroupRequestType.
const (
	CreateGroupRequestTypeN0 CreateGroupRequestType = 0
	CreateGroupRequestTypeN1 CreateGroupRequestType = 1
)

// Defines values for UpdateGroupRequestType.
const (
	UpdateGroupRequestTypeN0 UpdateGroupRequestType = 0
	UpdateGroupRequestTypeN1 UpdateGroupRequestType = 1
)

// CreateGroupRequest defines model for CreateGroupRequest.
type CreateGroupRequest struct {
	// Avatar 群组头像
	Avatar string `json:"avatar,omitempty"`

	// Encrypt 是否开启加密，只有当群聊为私密群才能开启
	Encrypt bool `json:"encrypt,omitempty"`

	// JoinApprove 入群审批
	JoinApprove bool `json:"join_approve,omitempty"`

	// Member 群组成员列表
	Member []string `json:"member,omitempty"`

	// Name 群组名称
	Name string `json:"name"`

	// Type 群组类型 0(私密群) 1(公开群)
	Type CreateGroupRequestType `json:"type"`
}

// CreateGroupRequestType 群组类型 0(私密群) 1(公开群)
type CreateGroupRequestType uint

// CreateGroupResponse defines model for CreateGroupResponse.
type CreateGroupResponse struct {
	// Avatar 群组头像
	Avatar string `json:"avatar,omitempty"`

	// CreatorId 创建者ID
	CreatorId string `json:"creator_id,omitempty"`

	// DialogId 对话ID
	DialogId uint32 `json:"dialog_id,omitempty"`

	// Id 群聊ID
	Id uint32 `json:"id,omitempty"`

	// MaxMembersLimit 群组成员上限
	MaxMembersLimit int `json:"max_members_limit,omitempty"`

	// Name 群聊名称
	Name string `json:"name,omitempty"`

	// Status 群组状态
	Status int `json:"status,omitempty"`

	// Type 群组类型
	Type uint32 `json:"type,omitempty"`
}

// Group defines model for Group.
type Group struct {
	// Avatar 群组头像
	Avatar string `json:"avatar,omitempty"`

	// Id 群聊ID
	Id uint32 `json:"id"`

	// MaxMembersLimit 群组成员上限
	MaxMembersLimit int `json:"max_members_limit"`

	// Member 群聊成员数量
	Member int `json:"member"`

	// Name 群聊名称
	Name string `json:"name"`
}

// GroupInfo defines model for GroupInfo.
type GroupInfo struct {
	// Avatar 群组头像
	Avatar string `json:"avatar,omitempty"`

	// CreatorId 创建者ID
	CreatorId string `json:"creator_id,omitempty"`

	// DialogId 对话ID
	DialogId uint32 `json:"dialog_id,omitempty"`

	// Encrypt 是否开启加密，只有当群聊为私密群才能开启
	Encrypt bool `json:"encrypt,omitempty"`

	// Id 群聊ID
	Id uint32 `json:"id,omitempty"`

	// JoinApprove 入群审批
	JoinApprove bool `json:"join_approve,omitempty"`

	// MaxMembersLimit 群组成员上限
	MaxMembersLimit int `json:"max_members_limit,omitempty"`

	// Name 群聊名称
	Name        string       `json:"name,omitempty"`
	Preferences *Preferences `json:"preferences,omitempty"`

	// SilenceTime 群禁言结束时间，不为0表示开启群聊全员禁言
	SilenceTime int64 `json:"silence_time,omitempty"`

	// Status 群组状态
	Status int `json:"status,omitempty"`

	// Type 群组类型
	Type uint8 `json:"type,omitempty"`
}

// Preferences defines model for Preferences.
type Preferences struct {
	// EntryMethod 进入方式
	EntryMethod uint `json:"entry_method"`

	// Identity 身份
	Identity uint `json:"identity"`

	// Inviter 邀请者
	Inviter string `json:"inviter"`

	// JoinedAt 加入时间
	JoinedAt int64 `json:"joined_at"`

	// MuteEndTime 静音结束时间
	MuteEndTime int64 `json:"mute_end_time"`

	// OpenBurnAfterReading 阅后即焚开启
	OpenBurnAfterReading uint `json:"open_burn_after_reading"`

	// Remark 备注
	Remark string `json:"remark"`

	// SilentNotification 静默通知
	SilentNotification uint `json:"silent_notification"`
}

// UpdateGroupRequest defines model for UpdateGroupRequest.
type UpdateGroupRequest struct {
	// Avatar 新的群聊头像
	Avatar *string `json:"avatar,omitempty"`

	// Encrypt 是否开启加密，只有当群聊为私密群才能开启
	Encrypt *bool `json:"encrypt,omitempty"`

	// JoinApprove 入群审批
	JoinApprove *bool `json:"join_approve,omitempty"`

	// Name 新的群聊名称
	Name *string `json:"name,omitempty"`

	// SilenceTime 群禁言结束时间，不为0表示开启群聊全员禁言
	SilenceTime *int64 `json:"silence_time,omitempty"`

	// Type 群组类型 0(私密群) 1(公开群)
	Type *UpdateGroupRequestType `json:"type,omitempty"`
}

// UpdateGroupRequestType 群组类型 0(私密群) 1(公开群)
type UpdateGroupRequestType uint

// SearchGroupParams defines parameters for SearchGroup.
type SearchGroupParams struct {
	// Keyword 搜索关键字 群聊名称或id
	Keyword string `form:"keyword" json:"keyword"`

	// Page 页码
	Page *int32 `form:"page,omitempty" json:"page,omitempty"`

	// PageSize 每页大小
	PageSize *int32 `form:"page_size,omitempty" json:"page_size,omitempty"`
}

// CreateGroupJSONRequestBody defines body for CreateGroup for application/json ContentType.
type CreateGroupJSONRequestBody = CreateGroupRequest

// UpdateGroupJSONRequestBody defines body for UpdateGroup for application/json ContentType.
type UpdateGroupJSONRequestBody = UpdateGroupRequest
