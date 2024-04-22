package service

import (
	"context"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	"github.com/cossim/coss-server/internal/live/adapters"
	"github.com/cossim/coss-server/internal/live/app"
	"github.com/cossim/coss-server/internal/live/app/command"
	"github.com/cossim/coss-server/internal/live/app/query"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"go.uber.org/zap"
)

func NewApplication(ctx context.Context, ac *config.AppConfig, logger *zap.Logger) *app.Application {
	services, err := discovery.NewBalanceGrpcClient(ac)
	if err != nil {
		panic(err)
	}

	liveRepository, err := adapters.NewRedisLiveRepository(ac.Redis.Addr(), ac.Redis.Password, 0)
	if err != nil {
		panic(err)
	}

	go func() {
		<-ctx.Done()
		for _, conn := range services {
			if err := conn.Close(); err != nil {
				logger.Error("Failed to close gRPC connection", zap.Error(err))
			}
		}
	}()

	return &app.Application{
		Commands: app.Commands{
			LiveHandler: command.NewLiveHandler(
				command.WithRepo(liveRepository),
				command.WithLogger(logger),
				command.WithLiveKit(ac.Livekit),
				command.WithMsgService(msggrpcv1.NewMsgServiceClient(services["msg_service"])),
				command.WithUserService(usergrpcv1.NewUserServiceClient(services["user_service"])),
				command.WithPushService(pushgrpcv1.NewPushServiceClient(services["push_service"])),
				command.WithRelationGroupService(relationgrpcv1.NewGroupRelationServiceClient(services["relation_service"])),
				command.WithRelationUserService(relationgrpcv1.NewUserRelationServiceClient(services["relation_service"])),
				command.WithGroupService(groupgrpcv1.NewGroupServiceClient(services["group_service"])),
			),
		},
		Queries: app.Queries{
			LiveHandler: query.NewLiveHandler(
				query.WithRepo(liveRepository),
				query.WithLogger(logger),
				query.WithLiveKit(ac.Livekit),
				query.WithRelationGroupService(relationgrpcv1.NewGroupRelationServiceClient(services["relation_service"])),
			),
		},
	}
}
