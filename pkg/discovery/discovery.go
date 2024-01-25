package discovery

// Discovery 定义服务注册发现接口
type Discovery interface {
	Register(serviceName, addr string, serviceID string) error
	RegisterHTTP(serviceName, addr string, serviceID string) error
	Cancel(serviceID string) error
	Discover(serviceName string) (string, error)
	Health(serviceName string) bool
}

// ConfigCenter 定义配置中心接口
type ConfigCenter interface {
	Set(key string, value string)
	Get(key string) string
}
