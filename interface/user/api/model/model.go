package model

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email       string `json:"email" binding:"required"`
	Password    string `json:"password" binding:"required"`
	ConfirmPass string `json:"confirm_password" binding:"required"`
	Nickname    string `json:"nickname"`
	PublicKey   string `json:"public_key"`
}

type UserInfoResponse struct {
	UserId         string             `json:"user_id"`
	Nickname       string             ` json:"nickname"`
	Email          string             `json:"email"`
	Tel            string             `json:"tel,omitempty"`
	Avatar         string             `json:"avatar"`
	Signature      string             `json:"signature"`
	Status         UserStatus         `json:"status"`
	RelationStatus UserRelationStatus `json:"relation_status"`
}

type UserStatus int

const (
	UserStatusUnknown UserStatus = iota // 未知状态
	UserStatusNormal             = 1    // 正常状态
	UserStatusDisable            = 2    // 被禁用
	UserStatusDeleted            = 3    // 已删除
	UserStatusLock               = 4    // 锁定状态
)

func (s UserStatus) String() string {
	switch s {
	case UserStatusUnknown:
		return "未知状态"
	case UserStatusNormal:
		return "正常状态"
	case UserStatusDisable:
		return "被禁用"
	case UserStatusDeleted:
		return "已删除"
	case UserStatusLock:
		return "锁定状态"
	default:
		return "未知状态"
	}
}

type UserRelationStatus int

const (
	UserRelationStatusUnknown UserRelationStatus = iota // 不是好友
	UserRelationStatusFriend                            // 好友关系
	UserRelationStatusBlacked                           // 黑名单 拉黑对方了
)

func (s UserRelationStatus) String() string {
	switch s {
	case UserRelationStatusUnknown:
		return "不是好友"
	case UserRelationStatusFriend:
		return "好友关系"
	case UserRelationStatusBlacked:
		return "拉黑了"
	default:
		return "不是好友"
	}
}

type GetType int

const (
	EmailType GetType = iota
	UserIdType
)

type SetPublicKeyRequest struct {
	PublicKey string `json:"public_key" binding:"required"`
}

type UserInfoRequest struct {
	NickName  string `json:"nickname"`
	Tel       string `json:"tel"`
	Avatar    string `json:"avatar"`
	Signature string `json:"signature"`
}

type PasswordRequest struct {
	OldPasswprd string `json:"old_password" binding:"required"`
	Password    string `json:"password" binding:"required"`
	ConfirmPass string `json:"confirm_password" binding:"required"`
}

type ModifyUserSecretBundleRequest struct {
	SecretBundle string `json:"secret_bundle" binding:"required"`
}
