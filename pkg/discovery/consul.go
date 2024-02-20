package discovery

import (
	"errors"
	"fmt"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/hashicorp/consul/api"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ConsulRegistry Consul服务注册和发现实现
type ConsulRegistry struct {
	client        *api.Client
	token         string
	keepAliveSync time.Duration // 心跳同步间隔
}

func (c *ConsulRegistry) RegisterHTTP(serviceName, addr, serviceID, healthAddr string) error {
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
		HTTP:                           healthAddr,
		Method:                         http.MethodGet,
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

type Option func(*ConsulRegistry)

// WithKeepAliveSync 设置心跳同步间隔
func WithKeepAliveSync(sync time.Duration) Option {
	return func(r *ConsulRegistry) {
		r.keepAliveSync = sync
	}
}

func NewConsulRegistry(addr string, opts ...Option) (Registry, error) {
	config := api.DefaultConfig()
	if addr != "" {
		config.Address = addr
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	if token := os.Getenv("CONSUL_HTTP_TOKEN"); token != "" {
		config.Token = token
	}

	registry := &ConsulRegistry{
		client:        client,
		token:         config.Token,
		keepAliveSync: 15 * time.Second, // 默认心跳同步间隔
	}

	for _, opt := range opts {
		opt(registry)
	}
	return registry, nil
}

func (c *ConsulRegistry) RegisterGRPC(serviceName, addr string, serviceID string) error {

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
	services, _, err := c.client.Health().Service(serviceName, "", true, &api.QueryOptions{Token: c.token})
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
	checks, _, err := c.client.Health().Checks(serviceName, &api.QueryOptions{Token: c.token})
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

type ConsulConfigCenter struct {
	token  string
	client *api.Client
}

func NewConsulConfigCenter(addr string, token string) (ConfigCenter, error) {
	config := api.DefaultConfig()
	config.Address = addr
	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	if token == "" {
		token = os.Getenv("CONSUL_HTTP_TOKEN")
	}

	cc := &ConsulConfigCenter{
		token:  token,
		client: client,
	}

	return cc, nil
}

func (c *ConsulConfigCenter) Get(key string) (string, error) {
	// 从 Consul 获取配置项的值
	kvPair, _, err := c.client.KV().Get(key, &api.QueryOptions{Token: c.token})
	if err != nil {
		return "", err
	}

	if kvPair == nil {
		return "", fmt.Errorf("config key '%s' not found", key)
	}

	return string(kvPair.Value), nil
}

func (c *ConsulConfigCenter) Set(key, value string) error {
	// 更新配置项的值
	p := &api.KVPair{
		Key:   key,
		Value: []byte(value),
	}

	_, err := c.client.KV().Put(p, &api.WriteOptions{Token: c.token})
	if err != nil {
		return err
	}

	return nil
}

func (c *ConsulConfigCenter) Watch(key string) (<-chan string, error) {
	configCh := make(chan string, 1)
	go func() {
		params := &api.QueryOptions{
			Token:     c.token,
			WaitIndex: 0,
		}
		for {
			time.Sleep(15 * time.Second)
			p, _, err := c.client.KV().Get(key, params)
			if err != nil {
				fmt.Println("Failed to get Consul KV: ", err)
				continue
			}
			//fmt.Printf("key %s params.WaitIndex => %d p.ModifyIndex => %d\n", key, params.WaitIndex, p.ModifyIndex)
			if p != nil && params.WaitIndex > 0 && params.WaitIndex < p.ModifyIndex {
				configCh <- string(p.Value)
			}
			params.WaitIndex = p.ModifyIndex
		}
	}()
	return configCh, nil
}

func (c *ConsulConfigCenter) Close() error {
	// 关闭 Consul 客户端连接
	c.client = nil
	return nil
}

const (
	CommonConfigPrefix   = "common/"
	CommonMySQLogKey     = CommonConfigPrefix + "log"           // MySQL配置项的键名
	CommonMySQLConfigKey = CommonConfigPrefix + "mysql"         // MySQL配置项的键名
	CommonRedisConfigKey = CommonConfigPrefix + "redis"         // Redis配置项的键名
	CommonMQConfigKey    = CommonConfigPrefix + "message_queue" // 消息队列（MQ）配置项的键名
	CommonDtmConfigKey   = CommonConfigPrefix + "dtm"           // 分布式事务
	CommonOssConfigKey   = CommonConfigPrefix + "oss"           // 对象存储

	InterfaceConfigPrefix = "interface/"

	ServiceConfigPrefix = "service/"
)

var (
	DefaultKeys = []string{CommonMySQLConfigKey, CommonRedisConfigKey, CommonMQConfigKey}
)

type RemoteConfigManager struct {
	cc           ConfigCenter
	keys         []string
	persistence  bool       // 持久化选项
	configFile   string     // 配置文件路径
	token        string     //
	configLocker sync.Mutex // 用于保证并发访问的互斥锁

	once sync.Once
}

type RemoteConfigOption func(*RemoteConfigManager)

func WithMysql() RemoteConfigOption {
	return func(m *RemoteConfigManager) {
		m.keys = append(m.keys, CommonMySQLConfigKey)
	}
}

func WithRedis() RemoteConfigOption {
	return func(m *RemoteConfigManager) {
		m.keys = append(m.keys, CommonRedisConfigKey)
	}
}

func WithMQ() RemoteConfigOption {
	return func(m *RemoteConfigManager) {
		m.keys = append(m.keys, CommonMQConfigKey)
	}
}

// WithPersistence 是设置持久化选项和配置文件路径的选项函数
func WithPersistence(configFile string) RemoteConfigOption {
	return func(m *RemoteConfigManager) {
		m.persistence = true
		m.configFile = configFile
		m.configLocker = sync.Mutex{}
	}
}

func WithToken(token string) RemoteConfigOption {
	return func(m *RemoteConfigManager) {
		m.token = token
	}
}

func NewDefaultRemoteConfigManager(configCenterURL string, opts ...RemoteConfigOption) (*RemoteConfigManager, error) {
	rcm := &RemoteConfigManager{
		keys: []string{},
	}

	for _, opt := range opts {
		opt(rcm)
	}

	cc, err := NewConsulConfigCenter(configCenterURL, rcm.token)
	if err != nil {
		log.Fatal("Failed to create ConsulConfigCenter: ", err)
	}
	rcm.cc = cc
	return rcm, nil
}

func NewRemoteConfigManager(configCenterURL string, opts ...RemoteConfigOption) (*RemoteConfigManager, error) {
	rcm := &RemoteConfigManager{}

	for _, opt := range opts {
		opt(rcm)
	}

	cc, err := NewConsulConfigCenter(configCenterURL, rcm.token)
	if err != nil {
		log.Fatal("Failed to create ConsulConfigCenter: ", err)
	}
	rcm.cc = cc
	return rcm, nil
}

func (m *RemoteConfigManager) Get(key string, keys ...string) (*config.AppConfig, error) {
	ac := &config.AppConfig{}

	newValue, err := m.cc.Get(key)
	if err != nil {
		return nil, err
	}

	v := viper.New()
	v.SetConfigType("yaml")
	if err = v.ReadConfig(strings.NewReader(newValue)); err != nil {
		return nil, err
	}
	if err = v.Unmarshal(ac); err != nil {
		return nil, err
	}

	for _, v := range DefaultKeys {
		newValue, err = m.cc.Get(v)
		if err != nil {
			return nil, err
		}
		if err = m.handlerConfig(v, ac, newValue); err != nil {
			return nil, err
		}
	}
	return ac, nil
}

func (m *RemoteConfigManager) Set(key, value string) error {
	return m.cc.Set(key, value)
}

// ConfigUpdate 包含配置更新的信息
type ConfigUpdate struct {
	Key   string
	Value interface{}
}

// Watch 监听配置项的变化并更新配置
func (m *RemoteConfigManager) Watch(ac *config.AppConfig, updateCh chan<- ConfigUpdate, key string, keys ...string) error {
	//m.once.Do(func() {
	go func(k string) {
		log.Printf("开始监听配置 => %s", k)
		configCh, err := m.cc.Watch(k)
		if err != nil {
			log.Printf("Failed to watch config for key %s: %v", k, err)
			return
		}
		for {
			select {
			case newValue := <-configCh:
				if err = m.handlerAppConfig(ac, newValue); err != nil {
					log.Printf("Failed to update config for key %s: %v", k, err)
					continue
				}
				//log.Printf("监听到配置更新 key %s, value %s", k, newValue)
				updateCh <- ConfigUpdate{Key: k, Value: newValue}

			}
		}
	}(key)

	m.keys = append(m.keys, keys...)
	log.Printf("开始监听common配置 => %s", m.keys)
	for _, k := range m.keys {
		go func(k string) {
			configCh, err := m.cc.Watch(k)
			if err != nil {
				log.Printf("Failed to watch config for key %s: %v", k, err)
				return
			}
			for {
				select {
				case newValue := <-configCh:
					if err = m.handlerConfig(k, ac, newValue); err != nil {
						log.Printf("Failed to update config for key %s: %v", k, err)
						continue
					}
					//log.Printf("监听到配置更新 key %s, value %s", k, newValue)
					updateCh <- ConfigUpdate{Key: k, Value: newValue}
				}
			}
		}(k)
	}
	//})

	return nil
}

func (m *RemoteConfigManager) handlerAppConfig(ac *config.AppConfig, newValue string) error {
	// 解析新的配置值
	var newConfig *config.AppConfig

	if err := yaml.Unmarshal([]byte(newValue), &newConfig); err != nil {
		return err
	}

	// 检查新的配置是否与旧的配置相同
	if reflect.DeepEqual(ac, newConfig) {
		return nil // 配置相同，不触发更新
	}

	ac.HTTP.Address = newConfig.HTTP.Address
	ac.HTTP.Port = newConfig.HTTP.Port

	ac.Register.Address = newConfig.Register.Address
	ac.Register.Name = newConfig.Register.Name
	ac.Register.Port = newConfig.Register.Port

	ac.GRPC.Address = newConfig.GRPC.Address
	ac.GRPC.Port = newConfig.GRPC.Port

	ac.Register.Tags = newConfig.Register.Tags
	ac.Discovers = newConfig.Discovers

	return nil
}

func (m *RemoteConfigManager) handlerConfig(key string, ac *config.AppConfig, newValue string) error {
	var trimmedKey string
	if strings.HasPrefix(key, CommonConfigPrefix) {
		trimmedKey = strings.TrimPrefix(key, CommonConfigPrefix)
	}

	if strings.HasPrefix(key, InterfaceConfigPrefix) {
		trimmedKey = strings.TrimPrefix(key, InterfaceConfigPrefix)
	}

	if strings.HasPrefix(key, ServiceConfigPrefix) {
		trimmedKey = strings.TrimPrefix(key, ServiceConfigPrefix)
	}

	fieldName := ""
	tags := reflect.TypeOf(*ac)

	for i := 0; i < tags.NumField(); i++ {
		tag := tags.Field(i).Tag.Get("mapstructure")
		if tag == trimmedKey {
			fieldName = tags.Field(i).Name
			break
		}
	}

	if fieldName == "" {
		return fmt.Errorf("unknown config key: %s", key)
	}

	// 解析新的配置值
	var newConfig map[string]interface{}
	if err := yaml.Unmarshal([]byte(newValue), &newConfig); err != nil {
		return err
	}

	// 根据配置键更新对应的字段
	if err := updateConfigField(fieldName, ac, newConfig); err != nil {
		return err
	}

	// 持久化到文件
	if m.persistence && m.configFile != "" {
		err := m.saveConfigToFile(ac)
		if err != nil {
			return fmt.Errorf("failed to save config to file: %w", err)
		}
	}

	return nil
}

func updateConfigField(fieldName string, ac *config.AppConfig, newConfig map[string]interface{}) error {
	acValue := reflect.ValueOf(ac).Elem()
	acType := acValue.Type()
	for i := 0; i < acType.NumField(); i++ {
		field := acType.Field(i)
		// 获取字段的 mapstructure 标签
		tag := field.Tag.Get("mapstructure")
		// 将字段名转换为与 mapstructure 标签一致的大小写形式进行比较
		if strings.EqualFold(field.Name, fieldName) {
			// 进行相应的操作，例如比较新旧配置值等
			fieldValue := acValue.Field(i)
			newValue, err := getObjectValue(newConfig, strings.Split(tag, "."))
			if err != nil {
				return err
			}
			if reflect.DeepEqual(fieldValue.Interface(), newValue) {
				//fmt.Printf("字段 %s 没有改变\n", fieldName)
			} else {
				// 将 fieldValue 转换为 map[string]interface{} 类型
				oldValue := make(map[string]interface{})
				if err = mapstructure.Decode(fieldValue.Interface(), &oldValue); err != nil {
					return err
				}

				if reflect.DeepEqual(oldValue, newValue) {
					//fmt.Printf("字段 %s 没有改变\n", fieldName)
					continue
				}
				//fmt.Printf("字段 %s 更新前的值: %v              更新后的值: %v\n", fieldName, oldValue, newValue)
				if err = setFieldValue(fieldValue, newValue); err != nil {
					return err
				}

				//fmt.Println("newConfig => ", newConfig)

				//// 遍历新配置的键值对
				//for key, newValue := range newConfig {
				//	// 检查字段是否存在于旧配置中
				//	if oldValue[key] != nil {
				//		// 比较旧值和新值
				//		if !reflect.DeepEqual(oldValue[key], newValue) {
				//			fieldChanged = true
				//			break
				//		}
				//	} else {
				//		// 如果字段不存在于旧配置中，则字段已经改变
				//		fieldChanged = true
				//		break
				//	}
				//}

				/*				fmt.Printf("字段 %s 更新前的值: %v              更新后的值: %v\n", fieldName, fieldValue.Interface(), newValue)
								if err = setFieldValue(fieldValue, newValue); err != nil {
									return false, err
								}
								fieldChanged = true // 将字段变化标志设置为 true*/
			}
		}
	}
	return nil
}

func setFieldValue(field reflect.Value, newValue interface{}) error {
	// 判断字段类型是否为整数类型
	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// 将字符串转换为整数类型
		intValue, err := strconv.Atoi(newValue.(string))
		if err != nil {
			return fmt.Errorf("failed to convert new value to int: %w", err)
		}
		// 将整数值设置给字段
		field.SetInt(int64(intValue))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		// 将字符串转换为无符号整数类型
		uintValue, err := strconv.ParseUint(newValue.(string), 10, 64)
		if err != nil {
			return fmt.Errorf("failed to convert new value to uint: %w", err)
		}
		// 将无符号整数值设置给字段
		field.SetUint(uintValue)
	default:
		// 其他字段类型的处理
		if err := mapstructure.Decode(newValue, field.Addr().Interface()); err != nil {
			return fmt.Errorf("failed to convert new value to field type: %w", err)
		}
	}
	return nil
}

func getObjectValue(obj map[string]interface{}, path []string) (interface{}, error) {
	for i, key := range path {
		value, ok := obj[key]
		if !ok {
			return nil, fmt.Errorf("failed to find value for key: %s", key)
		}
		if i == len(path)-1 {
			return value, nil
		}
		obj, ok = value.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("value for key %s is not a map", key)
		}
	}
	return nil, fmt.Errorf("empty path")
}

// saveConfigToFile 将配置保存到文件
func (m *RemoteConfigManager) saveConfigToFile(ac *config.AppConfig) error {
	m.configLocker.Lock()
	defer m.configLocker.Unlock()

	data, err := yaml.Marshal(ac)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	err = ioutil.WriteFile(m.configFile, data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func (m *RemoteConfigManager) Close() error {
	return m.cc.Close()
}

func LoadDefaultRemoteConfig(addr string, configName string, tokens string, ac *config.AppConfig) (chan ConfigUpdate, error) {
	cc, err := NewDefaultRemoteConfigManager(addr, WithToken(tokens))
	if err != nil {
		return nil, err
	}
	c, err := cc.Get(configName)
	if err != nil {
		return nil, err
	}
	*ac = *c

	ch := make(chan ConfigUpdate)
	if err = cc.Watch(ac, ch, configName); err != nil {
		return nil, err
	}

	return ch, err
}
