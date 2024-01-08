package config

import (
	"flag"
	"github.com/cossim/coss-server/pkg/config"
)

var Conf config.AppConfig
var configFile string

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

	return nil
}
