package metrics

import "context"

// Server is a server that serves metrics.
type Server interface {

	// Start runs the server.
	// It will install the metrics related resources depending on the server configuration.
	Start(ctx context.Context) error
}
