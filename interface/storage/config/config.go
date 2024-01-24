package config

import (
	"flag"
	"fmt"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/spf13/viper"
)

var Conf pkgconfig.AppConfig
var MinioConf *MinioConfig
var configFile string

type MinioConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"accessKey"`
	SecretKey string `mapstructure:"secretKey"`
	SSL       bool   `mapstructure:"ssl"`
	//PresignedExpires int    `mapstructure:"presignedExpires"`
}

func init() {
	flag.StringVar(&configFile, "config", "", "Path to configuration file")
	flag.Parse()
}

func Init() error {
	c, err := pkgconfig.LoadFile(configFile)
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
			Endpoint:  viper.GetString("oss.minio.addr"),
			AccessKey: viper.GetString("oss.minio.accessKey"),
			SecretKey: viper.GetString("oss.minio.secretKey"),
			SSL:       viper.GetBool("oss.minio.ssl"),
			//PresignedExpires: viper.GetInt("oss.minio.presignedExpires"),
		}
		MinioConf = minioConfig
	}

	return nil
}
