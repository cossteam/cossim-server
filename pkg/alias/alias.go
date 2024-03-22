package alias

import "github.com/cossim/coss-server/pkg/manager"

// Manager initializes shared dependencies, and provides them to Runnables.
// A Manager is required to create Controllers.
type Manager = manager.Manager

// Options are the arguments for creating a new Manager.
type Options = manager.Options

type HTTPServer = manager.HttpServer

type GRPCServer = manager.GrpcServer

type Config = manager.Config

type Registry = manager.Registry

var (
	// NewManager returns a new Manager for creating Controllers.
	NewManager = manager.New
)
