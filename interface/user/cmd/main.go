package main

import (
	_ "github.com/cossim/coss-server/docs"
	"github.com/cossim/coss-server/interface/user/config"
	_ "github.com/cossim/coss-server/interface/user/config"
	"github.com/cossim/coss-server/interface/user/server/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := config.Init(); err != nil {
		panic(err)
	}

	http.Init(&config.Conf)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	for {
		s := <-c
		switch s {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}
