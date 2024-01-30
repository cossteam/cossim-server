package config

import (
	"fmt"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/spf13/viper"
)

var (
	Conf                    = &pkgconfig.AppConfig{}
	File                    string
	ConfigurationCenterAddr string
)

func Init() error {
	c, err := pkgconfig.LoadFile(File)
	if err != nil {
		return err
	}
	Conf = c
	if File != "" {
		viper.SetConfigFile(File)
		if err = viper.ReadInConfig(); err != nil {
			panic(fmt.Errorf("fatal error config file: %s", err))
		}
	}

	return nil
}

func LoadConfigFromFile(file string) error {
	c, err := pkgconfig.LoadFile(file)
	if err != nil {
		return err
	}
	Conf = c
	return nil
}

func LoadDefaultRemoteConfig(addr string, configName string, token string, ac *pkgconfig.AppConfig) (chan discovery.ConfigUpdate, error) {
	cc, err := discovery.NewDefaultRemoteConfigManager(addr, discovery.WithToken(token))
	if err != nil {
		return nil, err
	}
	c, err := cc.Get(configName)
	if err != nil {
		return nil, err
	}
	*ac = *c

	ch := make(chan discovery.ConfigUpdate)
	if err = cc.Watch(ac, ch, configName); err != nil {
		return nil, err
	}

	return ch, err
}
