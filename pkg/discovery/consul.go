package discovery

import (
	"errors"
	"fmt"
	"github.com/hashicorp/consul/api"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"
)

// ConsulRegistry Consul服务注册和发现实现
type ConsulRegistry struct {
	client        *api.Client
	keepAliveSync time.Duration // 心跳同步间隔
}

type Option func(*ConsulRegistry)

// WithKeepAliveSync 设置心跳同步间隔
func WithKeepAliveSync(sync time.Duration) Option {
	return func(r *ConsulRegistry) {
		r.keepAliveSync = sync
	}
}

func NewConsulRegistry(addr string, opts ...Option) (Discovery, error) {
	config := api.DefaultConfig()
	if addr != "" {
		config.Address = addr
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	registry := &ConsulRegistry{
		client:        client,
		keepAliveSync: 15 * time.Second, // 默认心跳同步间隔
	}

	for _, opt := range opts {
		opt(registry)
	}
	return registry, nil
}

func (c *ConsulRegistry) Register(serviceName, addr string, serviceID string) error {

	address, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return err
	}

	// 生成对应grpc的检查对象
	check := &api.AgentServiceCheck{
		GRPC:                           addr,
		Timeout:                        "5s",
		Interval:                       "5s",
		DeregisterCriticalServiceAfter: "15s",
	}

	// 创建服务实例
	service := &api.AgentServiceRegistration{
		ID:      serviceID,
		Name:    serviceName,
		Address: address,
		Port:    port,
		Check:   check,
	}

	// 注册服务
	if err = c.client.Agent().ServiceRegister(service); err != nil {
		return err
	}

	// 定期检查Consul的可用性，并在Consul重新启动后重新注册服务
	go c.keepAlive(service)

	return nil
}

func (c *ConsulRegistry) keepAlive(service *api.AgentServiceRegistration) {
	for {
		time.Sleep(c.keepAliveSync) // 定期检查间隔
		// 检查Consul的健康状态
		health := c.client.Health()
		_, _, err := health.Service("consul", "", false, nil)
		if err != nil {
			log.Println("Failed to check Consul health:", err)
			// Consul不可用，重新注册服务
			if err = c.client.Agent().ServiceRegister(service); err != nil {
				log.Println("Failed to register service:", err)
				continue
			}
			log.Printf("Re-registered service: %s at %s\n", service.Name, service.Address+":"+strconv.Itoa(service.Port))
		}
	}
}

func (c *ConsulRegistry) Cancel(serviceID string) error {
	// 注销服务
	err := c.client.Agent().ServiceDeregister(fmt.Sprintf(serviceID))
	if err != nil {
		return err
	}

	return nil
}

func (c *ConsulRegistry) Discover(serviceName string) (string, error) {
	// 查询服务实例
	services, _, err := c.client.Health().Service(serviceName, "", true, &api.QueryOptions{})
	if err != nil {
		return "", err
	}

	// 随机排序服务实例
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(services), func(i, j int) {
		services[i], services[j] = services[j], services[i]
	})

	if len(services) > 0 {
		return services[0].Service.Address + ":" + strconv.Itoa(services[0].Service.Port), nil
	}

	return "", errors.New("no service instance available")
}

func (c *ConsulRegistry) Health(serviceName string) bool {
	// 检查服务健康状态
	checks, _, err := c.client.Health().Checks(serviceName, &api.QueryOptions{})
	if err != nil {
		return false
	}

	for _, check := range checks {
		if check.Status != api.HealthPassing {
			return false
		}
	}

	return true
}
