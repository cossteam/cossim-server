package manager

import (
	"context"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	server2 "github.com/cossim/coss-server/pkg/manager/server"
	"github.com/cossim/coss-server/pkg/utils/os"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"net/http/pprof"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cossim/coss-server/pkg/healthz"
	"github.com/cossim/coss-server/pkg/metrics"
	"github.com/go-logr/logr"
)

const (
	defaultLeaseDuration          = 15 * time.Second
	defaultRenewDeadline          = 10 * time.Second
	defaultRetryPeriod            = 2 * time.Second
	defaultGracefulShutdownPeriod = 30 * time.Second

	defaultReadinessEndpoint      = "/ready"
	defaultLivenessEndpoint       = "/health"
	defaultMetricsEndpoint        = "/metrics"
	defaultHealthProbeBindAddress = ":8081"
)

var _ Runnable = &controllerManager{}

type controllerManager struct {
	sync.Mutex
	started bool

	stopProcedureEngaged *int64
	errChan              chan error
	runnables            *runnables

	// cluster holds a variety of methods to interact with a cluster. Required.
	//cluster cluster.Cluster

	httpServer *server2.HttpService

	grpcServer *server2.GrpcService

	optsHttpServer *HttpServer

	optsGrpcServer *GrpcServer

	// metricsServer is used to serve prometheus metrics
	metricsServer metrics.Server

	healthCheckAddress string

	// httpHealthProbeListener is used to serve liveness probe
	httpHealthProbeListener net.Listener

	// metricsListener is used to serve prometheus metrics
	metricsListener net.Listener

	// metricsExtraHandlers contains extra handlers to register on http server that serves metrics.
	metricsExtraHandlers map[string]http.Handler

	// Readiness probe endpoint name
	readinessEndpointName string

	// Liveness probe endpoint name
	livenessEndpointName string

	// Readyz probe handler
	readyzHandler *healthz.Handler

	// Healthz probe handler
	healthzHandler *healthz.Handler

	// pprofListener is used to serve pprof
	pprofListener net.Listener

	// Logger is the logger that should be used by this manager.
	// If none is set, it defaults to log.Log global logger.
	logger logr.Logger

	// gracefulShutdownTimeout is the duration given to runnable to stop
	// before the manager actually returns on stop.
	gracefulShutdownTimeout time.Duration

	// shutdownCtx is the context that can be used during shutdown. It will be cancelled
	// after the gracefulShutdownTimeout ended. It must not be accessed before internalStop
	// is closed because it will be nil.
	shutdownCtx context.Context

	internalCtx    context.Context
	internalCancel context.CancelFunc

	// internalProceduresStop channel is used internally to the manager when coordinating
	// the proper shutdown of servers. This channel is also used for dependency injection.
	internalProceduresStop chan struct{}

	configUpdateCh chan discovery.ConfigUpdate

	config *config.AppConfig
}

func (cm *controllerManager) Stop(ctx context.Context) error {
	return nil
}

func (cm *controllerManager) GetHTTPClient() *http.Client {
	//TODO implement me
	panic("implement me")
}

func (cm *controllerManager) GetGRPCClient() *grpc.ClientConn {
	//TODO implement me
	panic("implement me")
}

func (cm *controllerManager) Add(runnable Runnable) error {
	//TODO implement me
	panic("implement me")
}

func (cm *controllerManager) AddMetricsExtraHandler(path string, handler http.Handler) error {
	cm.Lock()
	defer cm.Unlock()

	if cm.started {
		return fmt.Errorf("unable to add new metrics handler because metrics endpoint has already been created")
	}

	if path == defaultMetricsEndpoint {
		return fmt.Errorf("overriding builtin %s endpoint is not allowed", defaultMetricsEndpoint)
	}

	if _, found := cm.metricsExtraHandlers[path]; found {
		return fmt.Errorf("can't register extra handler by duplicate path %q on metrics http server", path)
	}

	cm.metricsExtraHandlers[path] = handler
	cm.logger.V(2).Info("Registering metrics http server extra handler", "path", path)
	return nil
}

func (cm *controllerManager) AddHealthzCheck(name string, check healthz.Checker) error {
	cm.Lock()
	defer cm.Unlock()

	if cm.started {
		return fmt.Errorf("unable to add new checker because healthz endpoint has already been created")
	}

	if cm.healthzHandler == nil {
		cm.healthzHandler = &healthz.Handler{Checks: map[string]healthz.Checker{}}
	}

	cm.healthzHandler.Checks[name] = check
	return nil
}

func (cm *controllerManager) AddReadyzCheck(name string, check healthz.Checker) error {
	cm.Lock()
	defer cm.Unlock()

	if cm.started {
		return fmt.Errorf("unable to add new checker because healthz endpoint has already been created")
	}

	if cm.readyzHandler == nil {
		cm.readyzHandler = &healthz.Handler{Checks: map[string]healthz.Checker{}}
	}

	cm.readyzHandler.Checks[name] = check
	return nil
}

func (cm *controllerManager) GetLogger() logr.Logger {
	return cm.logger
}

// runnableGroup manages a group of runnables that are
// meant to be running together until StopAndWait is called.
//
// Runnables can be added to a group after the group has started
// but not after it's stopped or while shutting down.
type runnableGroup struct {
	ctx    context.Context
	cancel context.CancelFunc

	start        sync.Mutex
	startOnce    sync.Once
	started      bool
	startQueue   []*readyRunnable
	startReadyCh chan *readyRunnable

	stop     sync.RWMutex
	stopOnce sync.Once
	stopped  bool
	//stopQueue []*readyRunnable

	// errChan is the error channel passed by the caller
	// when the group is created.
	// All errors are forwarded to this channel once they occur.
	errChan chan error

	// ch is the internal channel where the runnables are read off from.
	ch chan *readyRunnable

	// wg is an internal sync.WaitGroup that allows us to properly stop
	// and wait for all the runnables to finish before returning.
	wg *sync.WaitGroup
}

// readyRunnable encapsulates a runnable with
// a ready check.
type readyRunnable struct {
	Runnable
	Check       runnableCheck
	signalReady bool
}

// runnableCheck can be passed to Add() to let the runnable group determine that a
// runnable is ready. A runnable check should block until a runnable is ready,
// if the returned result is false, the runnable is considered not ready and failed.
type runnableCheck func(ctx context.Context) bool

// Start starts the manager and waits indefinitely.
// There is only two ways to have start return:
// An error has occurred during in one of the internal operations,
// Or, the context is cancelled.
func (cm *controllerManager) Start(ctx context.Context) (err error) {
	cm.Lock()
	if cm.started {
		cm.Unlock()
		return errors.New("manager already started")
	}
	cm.started = true

	var ready bool
	defer func() {
		// Only unlock the manager if we haven't reached
		// the internal readiness condition.
		if !ready {
			cm.Unlock()
		}
	}()

	// Initialize the internal context.
	cm.internalCtx, cm.internalCancel = context.WithCancel(ctx)

	// This chan indicates that stop is complete, in other words all runnables have returned or timeout on stop request
	stopComplete := make(chan struct{})
	defer close(stopComplete)
	// This must be deferred after closing stopComplete, otherwise we deadlock.
	defer func() {
		// https://hips.hearstapps.com/hmg-prod.s3.amazonaws.com/images/gettyimages-459889618-1533579787.jpg
		stopErr := cm.engageStopProcedure(stopComplete)
		if stopErr != nil {
			if err != nil {
				// Utilerrors.Aggregate allows to use errors.Is for all contained errors
				// whereas fmt.Errorf allows wrapping at most one error which means the
				// other one can not be found anymore.
				//err = kerrors.NewAggregate([]error{err, stopErr})
			} else {
				err = stopErr
			}
		}
	}()

	// Add the cluster runnable.
	//if err := cm.add(cm.cluster); err != nil {
	//	return fmt.Errorf("failed to add cluster to runnables: %w", err)
	//}

	// Metrics should be served whether the controller is leader or not.
	// (If we don't serve metrics for non-leaders, prometheus will still scrape
	// the pod but will get a connection refused).
	if cm.metricsListener != nil {
		cm.serveMetrics()
	}
	// Note: We are adding the metrics httpServer directly to HTTPServers here as matching on the
	// metricsserver.Server interface in cm.runnables.Add would be very brittle.
	//if cm.runnables.HTTPServers != nil {
	//	if err := cm.runnables.HTTPServers.Add(cm.httpServer, nil); err != nil {
	//		return fmt.Errorf("failed to add metrics httpServer: %w", err)
	//	}
	//}

	// Serve health probes.
	if cm.httpHealthProbeListener != nil {
		if err := cm.addHealthProbeServer(); err != nil {
			return fmt.Errorf("failed to add health probe httpServer: %w", err)
		}
	}

	// Add pprof httpServer
	if cm.pprofListener != nil {
		if err := cm.addPprofServer(); err != nil {
			return fmt.Errorf("failed to add pprof httpServer: %w", err)
		}
	}

	if cm.runnables.HTTPServers != nil && cm.httpServer != nil {
		if err := cm.runnables.HTTPServers.Add(cm.httpServer, nil); err != nil {
			return fmt.Errorf("failed to add metrics httpServer: %w", err)
		}
	}

	if cm.runnables.GRPCServers != nil && cm.grpcServer != nil {
		if err := cm.runnables.GRPCServers.Add(cm.grpcServer, nil); err != nil {
			return fmt.Errorf("failed to add metrics httpServer: %w", err)
		}
	}

	// First start any internal GRPC servers, which includes health probes, metrics and profiling if enabled.
	if err := cm.runnables.GRPCServers.Start(cm.internalCtx); err != nil {
		if err != nil {
			return fmt.Errorf("failed to start HTTP servers: %w", err)
		}
	}

	// First start any internal HTTP servers, which includes health probes, metrics and profiling if enabled.
	if err := cm.runnables.HTTPServers.Start(cm.internalCtx); err != nil {
		if err != nil {
			return fmt.Errorf("failed to start HTTP servers: %w", err)
		}
	}

	ready = true
	cm.Unlock()
	for {
		select {
		case <-ctx.Done():
			// We are done
			return nil
		case _ = <-cm.configUpdateCh:
			if err := cm.reloadConfiguration(); err != nil {
				return err
			}
		case err := <-cm.errChan:
			// Error starting or running a runnable
			return err
		}
	}
}

// reloadConfiguration reloads the configuration and updates the HTTP and GRPC server.
func (cm *controllerManager) reloadConfiguration() error {
	cm.runnables.HTTPServers.StopAndWait(cm.internalCtx)
	cm.runnables.GRPCServers.StopAndWait(cm.internalCtx)

	//地址
	//grpc地址
	//获取本地地址
	if cm.config.HTTP.Address == "" {
		ip, err := os.GetOutBoundIP()
		if err != nil {
			return fmt.Errorf("failed to get outbound ip: %w", err)
		}
		cm.config.HTTP.Address = ip
	}
	if cm.config.GRPC.Address == "" {
		ip, err := os.GetOutBoundIP()
		if err != nil {
			return fmt.Errorf("failed to get outbound ip: %w", err)
		}
		cm.config.GRPC.Address = ip
	}

	fmt.Println("grpc add", cm.config.GRPC.Address)

	fmt.Println("http add", cm.config.HTTP.Address)

	errChan := make(chan error, 1)
	cm.runnables = newRunnables(context.Background, errChan)
	if cm.httpServer != nil {
		cm.httpServer = server2.NewHttpService(cm.config, cm.optsHttpServer, cm.healthCheckAddress+cm.livenessEndpointName, cm.GetLogger())
		if err := cm.runnables.HTTPServers.Add(cm.httpServer, nil); err != nil {
			return fmt.Errorf("failed to add metrics httpServers: %w", err)
		}
	}

	if cm.grpcServer != nil {
		cm.grpcServer = server2.NewGRPCService(cm.config, cm.optsGrpcServer, cm.GetLogger())
		if err := cm.runnables.GRPCServers.Add(cm.grpcServer, nil); err != nil {
			return fmt.Errorf("failed to add metrics grpcServers: %w", err)
		}
	}

	if err := cm.runnables.GRPCServers.Start(cm.internalCtx); err != nil {
		if err != nil {
			return fmt.Errorf("failed to start grpcServers: %w", err)
		}
	}

	if err := cm.runnables.HTTPServers.Start(cm.internalCtx); err != nil {
		if err != nil {
			return fmt.Errorf("failed to start HTTP servers: %w", err)
		}
	}

	httpHealthProbeListener, err := (&Options{
		newHealthProbeListener: defaultHealthProbeListener,
	}).newHealthProbeListener(cm.healthCheckAddress)
	if err != nil {
		return err
	}
	cm.httpHealthProbeListener = httpHealthProbeListener

	// Serve health probes.
	if err := cm.addHealthProbeServer(); err != nil {
		return fmt.Errorf("failed to add health probe httpServer: %w", err)
	}

	return nil
}

func (cm *controllerManager) serveMetrics() {
	metrics.Registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)
	handler := promhttp.HandlerFor(metrics.Registry, promhttp.HandlerOpts{
		ErrorHandling: promhttp.HTTPErrorOnError,
	})
	// TODO(JoelSpeed): Use existing Kubernetes machinery for serving metrics
	mux := http.NewServeMux()
	mux.Handle(defaultMetricsEndpoint, handler)
	for path, extraHandler := range cm.metricsExtraHandlers {
		mux.Handle(path, extraHandler)
	}

	server := server2.New(mux)
	go cm.httpServe("metrics", cm.logger.WithValues("path", defaultMetricsEndpoint), server, cm.metricsListener)
}

func (cm *controllerManager) httpServe(kind string, log logr.Logger, server *http.Server, ln net.Listener) {
	log = log.WithValues("kind", kind, "addr", ln.Addr())

	go func() {
		log.Info("Starting server")
		if err := server.Serve(ln); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			if atomic.LoadInt64(cm.stopProcedureEngaged) > 0 {
				// There might be cases where connections are still open and we try to shutdown
				// but not having enough time to close the connection causes an error in Serve
				//
				// In that case we want to avoid returning an error to the main error channel.
				log.Error(err, "error on Serve after stop has been engaged")
				return
			}
			cm.errChan <- err
		}
	}()

	// Shutdown the server when stop is closed.
	<-cm.internalProceduresStop
	if err := server.Shutdown(cm.shutdownCtx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			// Avoid logging context related errors.
			return
		}
		if atomic.LoadInt64(cm.stopProcedureEngaged) > 0 {
			cm.logger.Error(err, "error on Shutdown after stop has been engaged")
			return
		}
		cm.errChan <- err
	}
}

// engageStopProcedure signals all runnables to stop, reads potential errors
// from the errChan and waits for them to end. It must not be called more than once.
func (cm *controllerManager) engageStopProcedure(stopComplete <-chan struct{}) error {
	if !atomic.CompareAndSwapInt64(cm.stopProcedureEngaged, 0, 1) {
		return errors.New("stop procedure already engaged")
	}

	// Populate the shutdown context, this operation MUST be done before
	// closing the internalProceduresStop channel.
	//
	// The shutdown context immediately expires if the gracefulShutdownTimeout is not set.
	var shutdownCancel context.CancelFunc
	if cm.gracefulShutdownTimeout < 0 {
		// We want to wait forever for the runnables to stop.
		cm.shutdownCtx, shutdownCancel = context.WithCancel(context.Background())
	} else {
		cm.shutdownCtx, shutdownCancel = context.WithTimeout(context.Background(), cm.gracefulShutdownTimeout)
	}
	defer shutdownCancel()

	// Start draining the errors before acquiring the lock to make sure we don't deadlock
	// if something that has the lock is blocked on trying to write into the unbuffered
	// channel after something else already wrote into it.
	var closeOnce sync.Once
	go func() {
		for {
			// Closing in the for loop is required to avoid race conditions between
			// the closure of all internal procedures and making sure to have a reader off the error channel.
			closeOnce.Do(func() {
				// Cancel the internal stop channel and wait for the procedures to stop and complete.
				close(cm.internalProceduresStop)
				cm.internalCancel()
			})
			select {
			case err, ok := <-cm.errChan:
				if ok {
					cm.logger.Error(err, "error received after stop sequence was engaged")
				}
			case <-stopComplete:
				return
			}
		}
	}()

	go func() {
		cm.logger.Info("Stopping and waiting for HTTP servers")
		cm.runnables.HTTPServers.StopAndWait(cm.shutdownCtx)

		cm.logger.Info("Stopping and waiting for GRPC servers")
		cm.runnables.GRPCServers.StopAndWait(cm.shutdownCtx)

		// Proceed to close the manager and overall shutdown context.
		cm.logger.Info("Wait completed, proceeding to shutdown the manager")
		shutdownCancel()
	}()

	<-cm.shutdownCtx.Done()
	if err := cm.shutdownCtx.Err(); err != nil && !errors.Is(err, context.Canceled) {
		if errors.Is(err, context.DeadlineExceeded) {
			if cm.gracefulShutdownTimeout > 0 {
				return fmt.Errorf("failed waiting for all runnables to end within grace period of %s: %w", cm.gracefulShutdownTimeout, err)
			}
			return nil
		}
		// For any other error, return the error.
		return err
	}

	return nil
}

func (cm *controllerManager) addHealthProbeServer() error {
	mux := http.NewServeMux()
	srv := NewHttp(mux)

	if cm.readyzHandler != nil {
		mux.Handle(cm.readinessEndpointName, http.StripPrefix(cm.readinessEndpointName, cm.readyzHandler))
		// Append '/' suffix to handle subpaths
		mux.Handle(cm.readinessEndpointName+"/", http.StripPrefix(cm.readinessEndpointName, cm.readyzHandler))
	}
	if cm.healthzHandler != nil {
		mux.Handle(cm.livenessEndpointName, http.StripPrefix(cm.livenessEndpointName, cm.healthzHandler))
		// Append '/' suffix to handle subpaths
		mux.Handle(cm.livenessEndpointName+"/", http.StripPrefix(cm.livenessEndpointName, cm.healthzHandler))
	}

	return cm.add(&httpServer{
		Kind:     "health probe",
		Log:      cm.logger,
		Server:   srv,
		Listener: cm.httpHealthProbeListener,
	})
}

func (cm *controllerManager) addPprofServer() error {
	mux := http.NewServeMux()
	srv := NewHttp(mux)

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return cm.add(&httpServer{
		Kind:     "pprof",
		Log:      cm.logger,
		Server:   srv,
		Listener: cm.pprofListener,
	})
}

// NewHttp returns a new server with sane defaults.
func NewHttp(handler http.Handler) *http.Server {
	return &http.Server{
		Handler:           handler,
		MaxHeaderBytes:    1 << 20,
		IdleTimeout:       90 * time.Second, // matches http.DefaultTransport keep-alive timeout
		ReadHeaderTimeout: 32 * time.Second,
	}
}

func (cm *controllerManager) add(r Runnable) error {
	return cm.runnables.Add(r)
}

type hasCache interface {
	Runnable
	// GetCache() cache.Cache
}
