package main

import (
	"flag"
	_ "github.com/cossim/coss-server/docs"
	"github.com/cossim/coss-server/interface/group/config"
	"github.com/cossim/coss-server/interface/group/server/http"
	"github.com/cossim/coss-server/interface/group/service"
	"github.com/cossim/coss-server/pkg/version"
	"os"
	"os/signal"
	"syscall"
)

var discover bool

func init() {
	flag.StringVar(&config.ConfigFile, "config", "/config/config.yaml", "Path to configuration file")
	flag.BoolVar(&discover, "discover", false, "Enable service discovery")
	flag.Parse()
}

func main() {
	version.FullVersion()
	if err := config.Init(); err != nil {
		panic(err)
	}

	svc := service.New(&config.Conf)
	svc.Start(discover)
	http.Init(&config.Conf, svc)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			svc.Close(discover)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
