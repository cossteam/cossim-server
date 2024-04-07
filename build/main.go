package main

import (
	conf "github.com/cossim/coss-server/build/config"
	"github.com/cossim/coss-server/pkg/encryption"
)

func main() {
	err := conf.Init()
	if err != nil {
		panic(err)
	}
	enc := encryption.NewEncryptor([]byte(conf.Conf.Passphrase), conf.Conf.Name, conf.Conf.Email, conf.Conf.RsaBits, conf.Conf.Enable, nil)
	err = enc.GenerateKeyPair()
	if err != nil {
		panic(err)
	}
}
