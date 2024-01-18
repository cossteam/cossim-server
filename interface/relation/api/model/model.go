package model

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type RequestListResponse struct {
	GroupId         uint32 `json:"group_id,omitempty" description:"群组ID"`
	GroupType       uint32 `json:"group_type,omitempty" description:"群组类型"`
	GroupStatus     uint32 `json:"group_status,omitempty" description:"群组状态"`
	MaxMembersLimit int32  `json:"max_members_limit,omitempty" description:"最大成员限制"`
	CreatorId       string `json:"creator_id,omitempty" description:"创建者ID"`
	GroupName       string `json:"group_name,omitempty" description:"群组名称"`
	GroupAvatar     string `json:"group_avatar,omitempty" description:"群组头像"`
	UserID          string `json:"user_id,omitempty" description:"用户ID"`
	UserName        string `json:"user_name,omitempty" description:"用户昵称"`
	UserAvatar      string `json:"user_avatar,omitempty" description:"用户头像"`
	Msg             string `json:"msg,omitempty" description:"申请消息"`
	RequestAt       string `json:"request_at,omitempty" description:"申请时间"`
	UserStatus      uint32 `json:"user_status,omitempty" description:"申请状态 (0=申请中, 1=已加入, 2=被拒绝, 3=被封禁)"`
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
	UserID       string `json:"user_id" binding:"required"`
	Status       int32  `json:"status"`
	E2EPublicKey string `json:"e2e_public_key"`
}

type AddFriendRequest struct {
	UserID       string `json:"user_id" binding:"required"`
	Msg          string `json:"msg"`
	E2EPublicKey string `json:"e2e_public_key"`
}

type JoinGroupRequest struct {
	GroupID uint32 `json:"group_id" binding:"required"`
}

type ManageJoinGroupRequest struct {
	GroupID uint32 `json:"group_id" binding:"required"`
	UserID  string `json:"user_id" binding:"required"`
	Status  uint32 `json:"status"`
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
