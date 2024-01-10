package config

import (
	"flag"
	"fmt"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/spf13/viper"
)

var Conf config.AppConfig
var MinioConf *MinioConfig
var configFile string

type MinioConfig struct {
	Endpoint         string `mapstructure:"endpoint"`
	AccessKey        string `mapstructure:"accessKey"`
	SecretKey        string `mapstructure:"secretKey"`
	SSL              bool   `mapstructure:"ssl"`
	PresignedExpires int    `mapstructure:"presignedExpires"`
}

func init() {
	flag.StringVar(&configFile, "config", "", "Path to configuration file")
	flag.Parse()
}

func Init() error {
	c, err := config.LoadFile(configFile)
	if err != nil {
		return err
	}
	Conf = *c

	if configFile != "" {
		viper.SetConfigFile(configFile)
		if err = viper.ReadInConfig(); err != nil {
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
		minioConfig := &MinioConfig{
			Endpoint:         viper.GetString("discovers.minio.addr"),
			AccessKey:        viper.GetString("discovers.minio.accessKey"),
			SecretKey:        viper.GetString("discovers.minio.secretKey"),
			SSL:              viper.GetBool("discovers.minio.ssl"),
			PresignedExpires: viper.GetInt("discovers.minio.presignedExpires"),
		}
		MinioConf = minioConfig
	}

	return nil
}
