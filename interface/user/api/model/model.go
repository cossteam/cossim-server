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
	PublicKey   string `json:"public_key" binding:"required"`
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
