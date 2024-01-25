package main

import (
	"flag"
	_ "github.com/cossim/coss-server/docs"
	"github.com/cossim/coss-server/interface/storage/config"
	"github.com/cossim/coss-server/interface/storage/server/http"
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
	if err := config.Init(); err != nil {
		panic(err)
	}

	http.Init(&config.Conf, discover)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			http.Close(discover)
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
