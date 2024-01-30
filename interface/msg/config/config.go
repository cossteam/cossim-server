package config

import (
	pkgconfig "github.com/cossim/coss-server/pkg/config"
)

var Conf = &pkgconfig.AppConfig{}

func LoadConfigFromFile(file string) error {
	c, err := pkgconfig.LoadFile(file)
	if err != nil {
		return err
	}
	Conf = c
	return nil
}
