package main

import (
	"flag"
	"fmt"
	_ "github.com/cossim/coss-server/docs"
	"github.com/cossim/coss-server/interface/live/config"
	"github.com/cossim/coss-server/interface/live/server/http"
	"github.com/cossim/coss-server/interface/live/service"
	"github.com/cossim/coss-server/pkg/discovery"
	"os"
	"os/signal"
	"syscall"
)

var (
	file              string
	discover          bool
	remoteConfig      bool
	remoteConfigAddr  string
	remoteConfigToken string
)

func init() {
	flag.StringVar(&file, "config", "config/config.yaml", "Path to configuration file")
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
		if err = config.LoadConfigFromFile(file); err != nil {
			panic(err)
		}
	} else {
		ch, err = discovery.LoadDefaultRemoteConfig(remoteConfigAddr, discovery.InterfaceConfigPrefix+"live", remoteConfigToken, config.Conf)
		if err != nil {
			panic(err)
		}
	}

	fmt.Println("config.Conf.Livekit => ", config.Conf.Livekit)

	if config.Conf == nil {
		panic("Config not initialized")
	}

	svc := service.New()
	h := http.NewHandler(svc)
	svc.Start(discover)
	h.Start()

	go func() {
		for {
			select {
			case _ = <-ch:
				h.Stop()
				svc.Stop(discover)
				svc = service.Restart(discover)
				h = http.NewHandler(svc)
				h.Start()
			}
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			h.Stop()
			svc.Stop(discover)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
