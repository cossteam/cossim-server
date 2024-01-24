package main

import (
	"flag"
	_ "github.com/cossim/coss-server/docs"
	"github.com/cossim/coss-server/interface/relation/config"
	_ "github.com/cossim/coss-server/interface/relation/config"
	"github.com/cossim/coss-server/interface/relation/server/http"
	"github.com/cossim/coss-server/interface/relation/service"
	"os"
	"os/signal"
	"syscall"
)

var Direct bool

func init() {
	flag.StringVar(&config.ConfigFile, "config", "/config/config.yaml", "Path to configuration file")
	flag.BoolVar(&Direct, "direct", false, "Enable direct connection")
	flag.Parse()
}

func main() {
	if err := config.Init(); err != nil {
		panic(err)
	}

	svc := service.New(&config.Conf)
	http.Init(&config.Conf, svc)
	svc.Start(Direct)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			svc.Close()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
