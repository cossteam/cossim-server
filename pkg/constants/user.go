package constants

type SystemUser string

const (
	SystemNotification = "10001"
)

var SystemUserList = []SystemUser{
	SystemNotification,
}

func IsSystemUser(user SystemUser) bool {
	for _, sysUser := range SystemUserList {
		if user == sysUser {
			return true
		}
	}
	return false
}
