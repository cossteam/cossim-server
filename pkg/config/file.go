package config

import (
	"github.com/spf13/viper"
)

func LoadFile(path string) (*AppConfig, error) {
	var Conf AppConfig
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	if err := viper.Unmarshal(&Conf); err != nil {
		return nil, err
	}

	return &Conf, nil
}
