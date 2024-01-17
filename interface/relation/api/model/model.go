package model

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type RequestListResponse struct {
	UserID    string `json:"user_id" description:"用户ID"`
	Nickname  string `json:"nickname" description:"用户昵称"`
	Avatar    string `json:"avatar" description:"用户头像"`
	Msg       string `json:"msg" description:"申请消息"`
	RequestAt string `json:"request_at" description:"申请时间"`
	Status    uint32 `json:"status" description:"申请状态 (0=申请中, 1=已加入, 2=被拒绝, 3=被封禁)"`
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
