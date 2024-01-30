package main

import (
	"flag"
	_ "github.com/cossim/coss-server/docs"
	"github.com/cossim/coss-server/interface/storage/config"
	"github.com/cossim/coss-server/interface/storage/server/http"
	"github.com/cossim/coss-server/pkg/discovery"
	"os"
	"os/signal"
	"syscall"
)

var (
	discover          bool
	remoteConfig      bool
	remoteConfigAddr  string
	remoteConfigToken string
)

func init() {
	flag.StringVar(&config.File, "config", "/config/config.yaml", "Path to configuration file")
	flag.BoolVar(&discover, "discover", false, "Enable service discovery")
	flag.BoolVar(&remoteConfig, "remote-config", false, "Load configuration from remote source")
	flag.StringVar(&remoteConfigAddr, "config-center-addr", "", "Address of the configuration center")
	flag.StringVar(&remoteConfigToken, "config-center-token", "", "Token for accessing the configuration center")
	flag.Parse()
}

func main() {
	ch := make(chan discovery.ConfigUpdate)
	var err error
	if !remoteConfig {
		if err = config.LoadConfigFromFile(config.File); err != nil {
			panic(err)
		}
	} else {
		ch, err = config.LoadDefaultRemoteConfig(config.ConfigurationCenterAddr, discovery.InterfaceConfigPrefix+"storage", remoteConfigToken, config.Conf)
		if err != nil {
			panic(err)
		}
	}

	if config.Conf == nil {
		panic("Config not initialized")
	}

	http.Start(discover)

	go func() {
		for {
			select {
			case _ = <-ch:
				http.Stop(discover)
				http.Restart(discover)
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			http.Stop(discover)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
