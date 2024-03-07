package config

import (
	"flag"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type LogConfig struct {
	Stdout bool   `mapstructure:"stdout" yaml:"stdout"`
	Level  int    `mapstructure:"level" yaml:"level"`
	Format string `mapstructure:"format" yaml:"format"`
}

type MySQLConfig struct {
	DSN      string `mapstructure:"dsn" yaml:"dsn"`
	Address  string `mapstructure:"address" yaml:"address"`
	Port     int    `mapstructure:"port" yaml:"port"`
	Username string `mapstructure:"username" yaml:"username"`
	Password string `mapstructure:"password" yaml:"password"`
	Database string `mapstructure:"database" yaml:"database"`
	//Opts     yaml.MapSlice `mapstructure:"opts" yaml:"opts"`
	Opts map[string]interface{} `mapstructure:"opts" yaml:"opts"`
}

func (c MySQLConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

type RedisConfig struct {
	Proto    string `mapstructure:"proto" yaml:"proto"`
	Password string `mapstructure:"password" yaml:"password"`
	Address  string `mapstructure:"address" yaml:"address"`
	Port     int    `mapstructure:"port" yaml:"port"`
}

func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

type HTTPConfig struct {
	Address string `mapstructure:"address" yaml:"address"`
	Port    int    `mapstructure:"port" yaml:"port"`
}

func (c HTTPConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

type GRPCConfig struct {
	Address string `mapstructure:"address" yaml:"address"`
	Port    int    `mapstructure:"port" yaml:"port"`
}

func (c GRPCConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

type RegistryConfig struct {
	Name     string   `mapstructure:"name" yaml:"name"`
	Tags     []string `mapstructure:"tags" yaml:"tags"`
	Address  string   `mapstructure:"address" yaml:"address"`
	Port     int      `mapstructure:"port" yaml:"port"`
	Discover bool     `mapstructure:"discover" yaml:"discover"`
	Register bool     `mapstructure:"register" yaml:"register"`
}

func (c RegistryConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

type DiscoversConfig map[string]ServiceConfig

type ServiceConfig struct {
	Name    string `mapstructure:"name" yaml:"name"`
	Address string `mapstructure:"address" yaml:"address"`
	Port    int    `mapstructure:"port" yaml:"port"`
	Direct  bool   `mapstructure:"direct" yaml:"direct"`
}

func (c ServiceConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

type MessageQueueConfig struct {
	Name     string `mapstructure:"name" yaml:"name"`
	Username string `mapstructure:"username" yaml:"username"`
	Password string `mapstructure:"password" yaml:"password"`
	Address  string `mapstructure:"address" yaml:"address"`
	Port     int    `mapstructure:"port" yaml:"port"`
}

func (c MessageQueueConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

type DtmConfig struct {
	Name    string `mapstructure:"name" yaml:"name"`
	Address string `mapstructure:"address" yaml:"address"`
	Port    int    `mapstructure:"port" yaml:"port"`
}

func (c DtmConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

type OSSCommonConfig struct {
	Address   string `mapstructure:"address" yaml:"address"`
	Port      int    `mapstructure:"port" yaml:"port"`
	AccessKey string `mapstructure:"accessKey" yaml:"accessKey"`
	SecretKey string `mapstructure:"secretKey" yaml:"secretKey"`
	SSL       bool   `mapstructure:"ssl" yaml:"ssl"`
	//PresignedExpires int    `mapstructure:"presignedExpires"`
}

func (c OSSCommonConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

type OssConfig map[string]OSSCommonConfig

type PushConfig struct {
	Address string `mapstructure:"address" yaml:"address"`
	Port    int    `mapstructure:"port" yaml:"port"`
	// 手机厂商对应的appid 例如ios对应com.hitosea.xxx
	PlatformAppID map[string]string `mapstructure:"platform_appid" yaml:"platform_appid"`
}

func (c PushConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

type AppConfig struct {
	Log                 LogConfig                 `mapstructure:"log" yaml:"log"`
	MySQL               MySQLConfig               `mapstructure:"mysql" yaml:"mysql"`
	Redis               RedisConfig               `mapstructure:"redis" yaml:"redis"`
	HTTP                HTTPConfig                `mapstructure:"http" yaml:"http"`
	GRPC                GRPCConfig                `mapstructure:"grpc" yaml:"grpc"`
	Register            RegistryConfig            `mapstructure:"register" yaml:"register"`
	Discovers           DiscoversConfig           `mapstructure:"discovers" yaml:"discovers"`
	Encryption          EncryptionConfig          `mapstructure:"encryption" yaml:"encryption"`
	MessageQueue        MessageQueueConfig        `mapstructure:"message_queue" yaml:"message_queue"`
	MultipleDeviceLimit MultipleDeviceLimitConfig `mapstructure:"multiple_device_limit" yaml:"multiple_device_limit"`
	SystemConfig        SystemConfig              `mapstructure:"system" yaml:"system"`
	Dtm                 DtmConfig                 `mapstructure:"dtm" yaml:"dtm"`
	OSS                 OssConfig                 `mapstructure:"oss" yaml:"oss"`
	Email               EmailConfig               `mapstructure:"email" yaml:"email"`
	Livekit             LivekitConfig             `mapstructure:"livekit" yaml:"livekit"`
	AdminConfig         AdminConfig               `mapstructure:"admin" yaml:"admin"`
	Push                PushConfig                `mapstructure:"push" yaml:"push"`
}

type EncryptionConfig struct {
	Enable     bool   `mapstructure:"enable" yaml:"enable"`
	Name       string `mapstructure:"name" yaml:"name"`
	Email      string `mapstructure:"email" yaml:"email"`
	RsaBits    int    `mapstructure:"rsaBits" yaml:"rsaBits"`
	Passphrase string `mapstructure:"passphrase" yaml:"passphrase"`
}

type MultipleDeviceLimitConfig struct {
	Enable bool `mapstructure:"enable" yaml:"enable"`
	Max    int  `mapstructure:"max" yaml:"max"`
}

type SystemConfig struct {
	Environment       string `mapstructure:"environment" yaml:"environment"`
	Ssl               bool   `mapstructure:"ssl" yaml:"ssl"`
	AvatarFilePath    string `mapstructure:"avatar_file_path" yaml:"avatar_file_path"`
	AvatarFilePathDev string `mapstructure:"avatar_file_path_dev" yaml:"avatar_file_path_dev"`
	GatewayAddress    string `mapstructure:"gateway_address" yaml:"gateway_address"`
	GatewayPort       string `mapstructure:"gateway_port" yaml:"gateway_port"`
	GatewayAddressDev string `mapstructure:"gateway_address_dev" yaml:"gateway_address_dev"`
	GatewayPortDev    string `mapstructure:"gateway_port_dev" yaml:"gateway_port_dev"`
}

type EmailConfig struct {
	Enable     bool   `mapstructure:"enable" yaml:"enable"`
	SmtpServer string `mapstructure:"smtp_server" yaml:"smtp_server"`
	Port       int    `mapstructure:"port" yaml:"port"`
	Username   string `mapstructure:"username" yaml:"username"`
	Password   string `mapstructure:"password" yaml:"password"`
}

type LivekitConfig struct {
	Address   string        `mapstructure:"address" yaml:"address"`
	Url       string        `mapstructure:"url" yaml:"url"`
	ApiKey    string        `mapstructure:"api_key" yaml:"api_key"`
	ApiSecret string        `mapstructure:"secret_key" yaml:"secret_key"`
	Timeout   time.Duration `mapstructure:"timeout" yaml:"timeout"`
	Port      int           `mapstructure:"port" yaml:"port"`
}

type AdminConfig struct {
	Email    string `mapstructure:"email" yaml:"email"`
	Password string `mapstructure:"password" yaml:"password"`
	NickName string `mapstructure:"nickname" yaml:"nickname"`
	UserId   string `mapstructure:"user_id" yaml:"user_id"`
}

func (c LivekitConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

// ConfigFlagName is the name of the config flag
const ConfigFlagName = "config"
const RecommendedConfigPathEnvVar = "CONFIG"

var configPath string

// init registers the "config" flag to the default command line FlagSet.
func init() {
	RegisterFlags(flag.CommandLine)
}

// RegisterFlags registers flag variables to the given FlagSet if not already registered.
// It uses the default command line FlagSet, if none is provided. Currently, it only registers the config flag.
func RegisterFlags(fs *flag.FlagSet) {
	if fs == nil {
		fs = flag.CommandLine
	}
	if f := fs.Lookup(ConfigFlagName); f != nil {
		configPath = f.Value.String()
	} else {
		fs.StringVar(&configPath, ConfigFlagName, "", "config path")
	}
}

// loadConfig loads the configuration from the specified file path or the default file path.
func loadConfig(filePath string) (*AppConfig, error) {
	//v := viper.New()
	//v.SetConfigType("yaml")
	//
	//// Read configuration from environment variables
	//v.AutomaticEnv()
	//
	//// Read the configuration file
	//if filePath != "" {
	//	v.SetConfigFile(filePath)
	//} else {
	//	//return nil, fmt.Errorf("config file path is empty")
	//	return nil, nil
	//}
	//
	//if err := v.ReadInConfig(); err != nil {
	//	return nil, fmt.Errorf("failed to read config file: %v", err)
	//}
	if filePath == "" {
		log.Printf("config file path is empty")
		return nil, nil
	}
	yamlFile, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("error reading config file: %v", err)
		return nil, nil
	}
	//
	var config AppConfig
	if err := yaml.Unmarshal(yamlFile, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %v", err)
	}

	return &config, nil
}

// LoadConfig loads a REST Config as per the rules specified in GetConfig.
func LoadConfig() (config *AppConfig, configErr error) {
	// If a flag is specified with the config location, use that
	if len(configPath) > 0 {
		return loadConfig(configPath)
	}

	// If the recommended config env variable is not specified,
	configPath = os.Getenv(RecommendedConfigPathEnvVar)
	return loadConfig(configPath)
}

func GetConfigOrDie() *AppConfig {
	config, err := LoadConfig()
	if err != nil {
		log.Printf("failed to load config: %v", err)
		return nil
	}
	return config
}
