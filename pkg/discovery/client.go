package discovery

import (
	"fmt"
	"github.com/cossim/coss-server/pkg/config"
	"google.golang.org/grpc"
)

var (
	baseUrl = "consul://%s/%s?wait=14s&healthy=true"
)

func NewBalanceGrpcClient(ac *config.AppConfig) (map[string]*grpc.ClientConn, error) {
	servers := map[string]*grpc.ClientConn{}
	for _, sc := range ac.Discovers {
		var addr string
		var grpcOptions = []grpc.DialOption{grpc.WithInsecure()}
		if sc.Direct {
			addr = sc.Addr()
		} else {
			addr = fmt.Sprintf(baseUrl, ac.Register.Addr(), sc.Name)
			grpcOptions = append(grpcOptions, grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`))
		}
		conn, err := grpc.Dial(
			addr,
			grpcOptions...)
		if err != nil {
			return nil, err
		}
		servers[sc.Name] = conn
	}
	return servers, nil
}
