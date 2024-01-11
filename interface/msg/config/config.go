package config

import (
	"github.com/cossim/coss-server/pkg/config"
)

var Conf config.AppConfig
var configFile string

func Init() error {
	c, err := config.LoadFile(configFile)
	if err != nil {
		return err
	}
	Conf = *c

	return nil
}
