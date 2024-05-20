package constants

type SystemUser string

const (
	SystemNotification = "10001"
	SystemAdmin        = "10000"
)

var SystemUserList = []SystemUser{
	SystemNotification,
	SystemAdmin,
}

func IsSystemUser(user SystemUser) bool {
	for _, sysUser := range SystemUserList {
		if user == sysUser {
			return true
		}
	}
	return false
}
