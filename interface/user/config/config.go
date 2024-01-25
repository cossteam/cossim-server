package config

import (
	pkgconfig "github.com/cossim/coss-server/pkg/config"
)

var Conf pkgconfig.AppConfig
var ConfigFile string

func Init() error {
	c, err := pkgconfig.LoadFile(ConfigFile)
	if err != nil {
		return err
	}
	Conf = *c

	return nil
}
