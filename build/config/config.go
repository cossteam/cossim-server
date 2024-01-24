package config

import (
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/spf13/viper"
)

var Conf pkgconfig.EncryptionConfig

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
