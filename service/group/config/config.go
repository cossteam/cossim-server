package config

import (
	"flag"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
)

var Conf pkgconfig.AppConfig
var configFile string

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

	return nil
}
