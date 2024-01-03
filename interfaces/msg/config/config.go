package config

import (
	"github.com/spf13/viper"
	"im/pkg/config"
)

var Conf config.AppConfig

func Init() error {
	viper.SetConfigFile("./config/config.yaml")
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if err := viper.Unmarshal(&Conf); err != nil {
		return err
	}

	return nil
}
