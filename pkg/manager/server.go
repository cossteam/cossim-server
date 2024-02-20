package manager

import (
	"context"
	"errors"
	"net"
	"net/http"

	"github.com/go-logr/logr"
)

// httpServer is a general purpose HTTP httpServer Runnable for a manager
// to serve some internal handlers such as health probes, metrics and profiling.
type httpServer struct {
	Kind     string
	Log      logr.Logger
	Server   *http.Server
	Listener net.Listener
}

func (s *httpServer) Start(ctx context.Context) error {
	log := s.Log.WithValues("kind", s.Kind, "addr", s.Listener.Addr())

	serverShutdown := make(chan struct{})
	go func() {
		<-ctx.Done()
		log.Info("Shutting down httpServer")

		if err := s.Server.Shutdown(ctx); err != nil {
			log.Error(err, "error shutting down httpServer")
		}
		close(serverShutdown)
	}()

	log.Info("Starting httpServer")

	if err := s.Server.Serve(s.Listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	<-serverShutdown
	return nil
}

type grpcServer struct {
	Kind     string
	Log      logr.Logger
	Server   *http.Server
	Listener net.Listener
}

func (s *grpcServer) Start(ctx context.Context) error {
	log := s.Log.WithValues("kind", s.Kind, "addr", s.Listener.Addr())

	serverShutdown := make(chan struct{})
	go func() {
		<-ctx.Done()
		log.Info("Shutting down grpcServer")
		if err := s.Server.Shutdown(context.Background()); err != nil {
			log.Error(err, "error shutting down grpcServer")
		}
		close(serverShutdown)
	}()

	log.Info("Starting grpcServer")
	if err := s.Server.Serve(s.Listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	<-serverShutdown
	return nil
}

func (s *grpcServer) Stop(ctx context.Context) error {
	return nil
}
