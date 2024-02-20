package server

import (
	"context"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

// HTTPService HTTP 服务的实例想要使用 manager 进行生命周期管理，必须实现该接口
type HTTPService interface {
	InternalDependency
	// RegisterRoute 注册服务的路由
	RegisterRoute(r gin.IRouter)

	// Health 实现一个健康检查并返回服务健康检查的路径
	Health(r gin.IRouter) string

	Stop(ctx context.Context) error

	// DiscoverServices 根据提供的服务名称和对应的 gRPC 客户端连接进行服务发现
	// 参数 services 是一个映射，键是服务名称，值是对应的 gRPC 客户端连接
	DiscoverServices(services map[string]*grpc.ClientConn) error
}

// GRPCService GRPC 服务的实例想要使用 manager 进行生命周期管理，必须实现该接口
type GRPCService interface {
	InternalDependency
	Registry(s *grpc.Server)
}

// Registry 是一个服务发现接口，定义了Manager服务注册和发现的方法
type Registry interface {
	// RegisterGRPC 注册 GRPC 服务
	// serviceName 是服务名称
	// addr 是服务地址
	// serviceID 是服务唯一标识符
	RegisterGRPC(serviceName, addr string, serviceID string) error

	// RegisterHTTP 注册 HTTP 服务
	// serviceName 是服务名称
	// addr 是服务地址
	// serviceID 是服务唯一标识符
	RegisterHTTP(serviceName, addr string, serviceID string) error

	Discover() error
}

type InternalDependency interface {
	// Init 根据提供的配置进行自定义的初始化操作
	Init(cfg *config.AppConfig) error
	// Name 服务名称
	Name() string
	// Version 服务版本信息
	Version() string
}
