package config

import (
	"fmt"
)

type LogConfig struct {
	Stdout bool   `mapstructure:"stdout"`
	V      int    `mapstructure:"v"`
	Format string `mapstructure:"format"`
}

type MySQLConfig struct {
	DSN          string `mapstructure:"dsn"`
	RootPassword string `mapstructure:"root_password"`
	Database     string `mapstructure:"database"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	Address      string `mapstructure:"address"`
	Port         int    `mapstructure:"port"`
}

func (c MySQLConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

type RedisConfig struct {
	Name     string `mapstructure:"name"`
	Proto    string `mapstructure:"proto"`
	Password string `mapstructure:"password"`
	Address  string `mapstructure:"address"`
	Port     int    `mapstructure:"port"`
}

func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

type HTTPConfig struct {
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
}

func (c HTTPConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

type GRPCConfig struct {
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
}

func (c GRPCConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

type RegisterConfig struct {
	Name    string   `mapstructure:"name"`
	Tags    []string `mapstructure:"tags"`
	Address string   `mapstructure:"address"`
	Port    int      `mapstructure:"port"`
}

func (c RegisterConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

type DiscoversConfig map[string]ServiceConfig

type ServiceConfig struct {
	Name    string `mapstructure:"name"`
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
}

func (c ServiceConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

type MessageQueueConfig struct {
	Name     string `mapstructure:"name"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Address  string `mapstructure:"address"`
	Port     int    `mapstructure:"port"`
}

func (c MessageQueueConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

type DtmConfig struct {
	Name    string `mapstructure:"name"`
	Address string `mapstructure:"address"`
	Port    int    `mapstructure:"port"`
}

func (c DtmConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Address, c.Port)
}

type AppConfig struct {
	Log          LogConfig          `mapstructure:"log"`
	MySQL        MySQLConfig        `mapstructure:"mysql"`
	Redis        RedisConfig        `mapstructure:"redis"`
	HTTP         HTTPConfig         `mapstructure:"http"`
	GRPC         GRPCConfig         `mapstructure:"grpc"`
	Register     RegisterConfig     `mapstructure:"register"`
	Discovers    DiscoversConfig    `mapstructure:"discovers"`
	Encryption   EncryptionConfig   `mapstructure:"encryption"`
	MessageQueue MessageQueueConfig `mapstructure:"message_queue"`
	Dtm          DtmConfig          `mapstructure:"dtm"`
}

type EncryptionConfig struct {
	Enable     bool   `mapstructure:"enable"`
	Name       string `mapstructure:"name"`
	Email      string `mapstructure:"email"`
	RsaBits    int    `mapstructure:"rsaBits"`
	Passphrase string `mapstructure:"passphrase"`
}
