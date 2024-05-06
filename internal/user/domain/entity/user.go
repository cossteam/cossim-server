package entity

type User struct {
	ID           string
	CossID       string
	Email        string
	Tel          string
	NickName     string
	Avatar       string
	PublicKey    string
	Password     string
	LastIp       string
	LineIp       string
	CreatedIp    string
	Signature    string
	LineAt       int64
	LastAt       int64
	Status       UserStatus
	EmailVerity  uint
	Bot          uint
	SecretBundle string
	CreatedAt    int64
	//UpdatedAt    int64
	//DeletedAt    int64
}

type UserStatus uint

const (
	// UserStatusNormal 正常状态
	UserStatusNormal UserStatus = iota + 1
	// UserStatusDisable 禁用状态
	UserStatusDisable
	// UserStatusDeleted 删除状态
	UserStatusDeleted
	// UserStatusLock 锁定状态
	UserStatusLock
)
