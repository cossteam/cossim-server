package entity

// UserActivateStatus 用户激活状态
type UserActivateStatus uint

const (
	NotActivated UserActivateStatus = iota // 未激活
	Activated                              // 已经激活
)

type LoginRequest struct {
	Email       string
	Password    string
	DriverID    string
	DriverToken string
	Platform    string
	ClientIP    string
}

type LoginResponse struct {
	Token string
}

type UserLogin struct {
	ID          uint
	CreatedAt   int64
	UserID      string
	LoginCount  uint
	LastAt      int64
	Token       string
	DriverID    string
	DriverToken string
	DriverType  string
	Platform    string
	ClientIP    string
	Rid         string
}

type UserLoginOpt func(*UserLogin)

type UserLoginOpts []UserLoginOpt

func (opts UserLoginOpts) Apply(u *UserLogin) {
	for _, opt := range opts {
		opt(u)
	}
}

func WithUserIDUserLogin(userID string) UserLoginOpt {
	return func(u *UserLogin) {
		u.UserID = userID
	}
}

func WithDriverIDUserLogin(userID string) UserLoginOpt {
	return func(u *UserLogin) {
		u.UserID = userID
	}
}
