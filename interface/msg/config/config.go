package config

import (
	"flag"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
)

var Conf pkgconfig.AppConfig
var configFile string

func Init() error {
	flag.StringVar(&configFile, "config", "", "Path to configuration file")
	flag.Parse()
	c, err := pkgconfig.LoadFile(configFile)
	if err != nil {
		return err
	}
	Conf = *c

	return nil
}
