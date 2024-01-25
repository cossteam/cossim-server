package constants

type ClientType string

const (
	//移动客户端
	MobileClient = "Mobile"
	//桌面客户端
	DesktopClient = "Desktop"
	// 未定义客户端类型,不带默认是未定义
	UnDefinedClient = "UnDefined"
)

func DetermineClientType(clientType string) string {
	if clientType == "Mobile" || clientType == "Desktop" {
		return clientType
	}
	return "UnDefined"
}
