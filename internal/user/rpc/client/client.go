package client

import (
	authv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"google.golang.org/grpc"
)

const serviceName = "user_service"

func NewAuthClient(addr string) authv1.UserAuthServiceClient {
	//conn, err := discovery.NewGrpcClient(serviceName, addr)
	//if err != nil {
	//	return nil
	//}

	var grpcOptions = []grpc.DialOption{grpc.WithInsecure()}

	//addr = fmt.Sprintf(baseUrl, addr, name)
	grpcOptions = append(grpcOptions, grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy": "round_robin"}`))

	conn, err := grpc.Dial(
		addr,
		grpcOptions...,
	)
	if err != nil {
		panic(err)
	}

	return authv1.NewUserAuthServiceClient(conn)
}
