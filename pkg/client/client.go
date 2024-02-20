package client

import (
	"google.golang.org/grpc"
	"net/http"
)

// Options are creation options for a Client.
type Options struct {
	// HTTPClient is the HTTP client to use for requests.
	HTTPClient *http.Client

	GRPCClient *grpc.ClientConn

	// DryRun instructs the client to only perform dry run requests.
	DryRun *bool
}
