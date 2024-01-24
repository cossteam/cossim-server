package config

import (
	"fmt"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/spf13/viper"
)

var Conf pkgconfig.AppConfig
var ConfigFile string
var RabbitMqConf *rabbitMqConf

func Init() error {
	c, err := pkgconfig.LoadFile(ConfigFile)
	if err != nil {
		return err
	}
	Conf = *c
	if ConfigFile != "" {
		viper.SetConfigFile(ConfigFile)
		if err = viper.ReadInConfig(); err != nil {
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
		mqConf := &rabbitMqConf{
			Port: viper.GetString("message_queue.port"),
			Name: viper.GetString("message_queue.name"),
			User: viper.GetString("message_queue.username"),
			Pass: viper.GetString("message_queue.password"),
			Addr: viper.GetString("message_queue.addr"),
			//Vhost: viper.GetString("rabbitmq.vhost"),
		}
		RabbitMqConf = mqConf
		fmt.Println("RabbitMqConf ", RabbitMqConf)
	}
	return nil
}

type rabbitMqConf struct {
	Port string `mapstructure:"port"`
	Name string `mapstructure:"name"`
	User string `mapstructure:"username"`
	Pass string `mapstructure:"password"`
	Addr string `mapstructure:"addr"`
	//Vhost string `mapstructure:"vhost"`
}
