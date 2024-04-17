package decorator

import "google.golang.org/grpc"

type GrpcClient interface {
	HandlerClient(serviceName string, conn *grpc.ClientConn)
}
