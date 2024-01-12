package main

import (
	_ "github.com/cossim/coss-server/docs"
	"github.com/cossim/coss-server/interface/msg/config"
	_ "github.com/cossim/coss-server/interface/msg/config"
	"github.com/cossim/coss-server/interface/msg/server/http"
	"github.com/cossim/coss-server/interface/msg/service"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := config.Init(); err != nil {
		panic(err)
	}

	http.Init(&config.Conf)
	service := service.New(&config.Conf)
	service.Start()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			service.Stop()
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
