package constants

type DriverType string

const (
	//移动客户端
	MobileClient DriverType = "Mobile"
	//桌面客户端
	DesktopClient DriverType = "Desktop"
	// 未定义客户端类型,不带默认是未定义
	UnDefinedClient DriverType = "UnDefined"
)

func DetermineClientType(clientType DriverType) DriverType {
	if clientType == MobileClient || clientType == DesktopClient {
		return clientType
	}
	return UnDefinedClient
}

func GetDriverTypeList() []DriverType {
	return []DriverType{
		MobileClient,
		DesktopClient,
		UnDefinedClient,
	}
}
