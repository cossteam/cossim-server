package entity

// UserActivateStatus 用户激活状态
type UserActivateStatus uint

const (
	NotActivated UserActivateStatus = iota //未激活
	Activated                              //已经激活
)

type UserLogin struct {
	ID          uint
	CreatedAt   int64
	UserId      string
	LoginCount  uint
	LastAt      int64
	Token       string
	DriverId    string
	DriverToken string
	DriverType  string
	Platform    string
	ClientIP    string
	Rid         string
}
