package main

import (
	"flag"
	"github.com/cossim/coss-server/internal/gateway/server/http"
	ctrl "github.com/cossim/coss-server/pkg/alias"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/healthz"
	"github.com/cossim/coss-server/pkg/manager/signals"
)

var (
	discover          bool
	register          bool
	remoteConfig      bool
	remoteConfigAddr  string
	remoteConfigToken string
	pprofAddr         string
	metricsAddr       string
	httpProbeAddr     string
	grpcProbeAddr     string
	hotReload         bool
	configKey         string
)

func init() {
	flag.BoolVar(&discover, "discover", false, "Enable service discovery")
	flag.BoolVar(&register, "register", false, "Enable service register")
	flag.BoolVar(&remoteConfig, "remote-config", false, "Load config from remote source")
	flag.StringVar(&remoteConfigAddr, "config-center-addr", "", "Address of the config center")
	flag.StringVar(&remoteConfigToken, "config-center-token", "", "Token for accessing the config center")
	flag.BoolVar(&hotReload, "hot-reload", false, "Enable hot reloading")
	flag.StringVar(&configKey, "config-key", "service/gateway", "Service configuration path in the configuration center")
	flag.StringVar(&pprofAddr, "pprof-bind-address", "0", "The address the pprof endpoint binds to")
	flag.StringVar(&metricsAddr, "metrics-bind-address", "0", "The address the metric endpoint binds to")
	flag.StringVar(&httpProbeAddr, "http-health-probe-bind-address", "0", "The address to bind the http health probe endpoint")
	flag.StringVar(&grpcProbeAddr, "grpc-health-probe-bind-address", "0", "The address to bind the grpc health probe endpoint")
	flag.Parse()
}

func main() {
	mgr, err := ctrl.NewManager(config.GetConfigOrDie(), ctrl.Options{
		Http: ctrl.HTTPServer{
			HTTPService:        &http.Handler{},
			HealthCheckAddress: httpProbeAddr,
		},
		Config: ctrl.Config{
			LoadFromConfigCenter: remoteConfig,
			RemoteConfigAddr:     remoteConfigAddr,
			RemoteConfigToken:    remoteConfigToken,
			Hot:                  hotReload,
			Key:                  configKey,
			//Keys:                 discovery.DefaultKeys,
			Registry: ctrl.Registry{
				Discover: discover,
				Register: register,
			},
		},
		PprofBindAddress:   pprofAddr,
		MetricsBindAddress: metricsAddr,
	})
	if err != nil {
		panic(err)
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		panic(err)
	}

	if err = mgr.Start(signals.SetupSignalHandler()); err != nil {
		panic(err)
	}
}
