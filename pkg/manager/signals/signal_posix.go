//go:build !windows
// +build !windows

package signals

import (
	"os"
	"syscall"
)

//var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGTERM}

var shutdownSignals = []os.Signal{os.Interrupt, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT}
