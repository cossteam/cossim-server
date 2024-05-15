package entity

type UpdateUser struct {
	Email     *string
	Tel       *string
	NickName  *string
	Avatar    *string
	PublicKey *string
	Password  *string
}

type UserRegister struct {
	Email     string
	NickName  string
	Password  string
	Avatar    string
	PublicKey string
}

type UpdateUserStatus struct {
	Status      *UserStatus
	EmailVerity *bool
}

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
	EmailVerity  bool
	Bot          uint
	SecretBundle string
	CreatedAt    int64
	//UpdatedAt    int64
	//DeletedAt    int64
}

type UserInfo struct {
	UserID         string
	CossID         string
	Nickname       string
	Email          string
	Tel            string
	Avatar         string
	Signature      string
	Status         UserStatus
	RelationStatus UserRelationStatus
	LoginNumber    uint
	Preferences    *Preferences
	NewDeviceLogin bool
	LastLoginTime  int64
	PublicKey      string
}

type Preferences struct {
	SilentNotification          bool
	Remark                      string
	OpenBurnAfterReading        bool
	OpenBurnAfterReadingTimeOut int64
}

type UserRelationStatus int

const (
	UserRelationStatusNone    UserRelationStatus = iota // 没有用户关系
	UserRelationStatusFriend                            // 正常
	UserRelationStatusBlack                             // 拉黑
	UserRelationStatusBlacked                           // 删除
)

func (s UserRelationStatus) Int() int {
	return int(s)
}

func (s UserRelationStatus) String() string {
	switch s {
	case UserRelationStatusNone:
		return "没有用户关系"
	case UserRelationStatusBlack:
		return "拉黑"
	case UserRelationStatusFriend:
		return "正常"
	case UserRelationStatusBlacked:
		return "删除"
	default:
		return "状态不正常"
	}
}

type UserStatus int

const (
	UserStatusUnknown UserStatus = iota
	// UserStatusNormal 正常状态
	UserStatusNormal
	// UserStatusDisable 禁用状态
	UserStatusDisable
	// UserStatusDeleted 删除状态
	UserStatusDeleted
	// UserStatusLock 锁定状态
	UserStatusLock
)

func (s UserStatus) String() string {
	switch s {
	case UserStatusNormal:
		return "正常状态"
	case UserStatusDisable:
		return "禁用状态"
	case UserStatusDeleted:
		return "删除状态"
	case UserStatusLock:
		return "锁定状态"
	default:
		return "未知状态"
	}
}

type ListUserOptions struct {
	UserID []string
}

// User Optional
type UserOpt func(*User)
type UserOpts []UserOpt

func (opts UserOpts) Apply(u *User) {
	for _, opt := range opts {
		opt(u)
	}
}

func WithUserID(id string) UserOpt {
	return func(u *User) {
		u.ID = id
	}
}

func WithEmail(email string) UserOpt {
	return func(u *User) {
		u.Email = email
	}
}

func WithNickName(nickname string) UserOpt {
	return func(u *User) {
		u.NickName = nickname
	}
}

func WithPassword(password string) UserOpt {
	return func(u *User) {
		u.Password = password
	}
}
