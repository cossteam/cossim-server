package model

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	DriverId string `json:"driver_id" binding:"required"`
}

type LogoutRequest struct {
	LoginNumber uint `json:"login_number"`
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
	LoginNumber    uint               `json:"login_number"`
	Preferences    *Preferences       `json:"preferences,omitempty"`
}

type Preferences struct {
	SilentNotification   SilentNotification       `json:"silent_notification"`
	Remark               string                   ` json:"remark"`
	OpenBurnAfterReading OpenBurnAfterReadingType `json:"open_burn_after_reading"`
}

type OpenBurnAfterReadingType uint

const (
	CloseBurnAfterReading OpenBurnAfterReadingType = iota //关闭阅后即焚
	OpenBurnAfterReading                                  //开启阅后即焚消息
)

type SilentNotification uint

const (
	NotSilentNotification SilentNotification = iota //不开启静默通知
	IsSilentNotification                            //开启静默通知
)

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
	UserRelationStatusUnknown UserRelationStatus = iota // 拉黑
	UserRelationStatusFriend                            // 正常
	UserRelationStatusBlacked                           // 删除
)

func (s UserRelationStatus) String() string {
	switch s {
	case UserRelationStatusUnknown:
		return "拉黑"
	case UserRelationStatusFriend:
		return "正常"
	case UserRelationStatusBlacked:
		return "删除"
	default:
		return "状态不正常"
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

type UserSecretBundleResponse struct {
	UserId       string `json:"user_id"`
	SecretBundle string `json:"secret_bundle"`
}

type GetUserLoginClientsResponse struct {
	ClientIP    string `json:"client_ip"`
	DriverType  string `json:"driver_type"`
	LoginNumber uint   `json:"login_number"`
	LoginAt     int64  `json:"login_at"`
}

type ResetPublicKeyRequest struct {
	PublicKey string `json:"public_key" binding:"required"`
	Email     string `json:"email" binding:"required"`
	Code      string `json:"code" binding:"required"`
}

type SendEmailCodeRequest struct {
	Email string `json:"email" binding:"required"`
}
