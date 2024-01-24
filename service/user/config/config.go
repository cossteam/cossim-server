package config

import (
	"github.com/cossim/coss-server/pkg/config"
)

var Conf config.AppConfig
var ConfigFile string
var Direct bool

func Init() error {
	c, err := config.LoadFile(ConfigFile)
	if err != nil {
		return err
	}
	Conf = *c

	return nil
}
