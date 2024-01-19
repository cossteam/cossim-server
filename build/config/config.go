package config

import (
	"github.com/cossim/coss-server/pkg/config"
	"github.com/spf13/viper"
)

var Conf config.EncryptionConfig

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
