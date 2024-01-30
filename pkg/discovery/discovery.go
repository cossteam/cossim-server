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
	Get(key string) (string, error)          // 获取配置项的值
	Set(key, value string) error             // 更新配置项的值
	Watch(key string) (<-chan string, error) // 监听配置项的变化并返回一个通道
	Close() error                            // 关闭配置中心连接
}
