package service

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/group/adapters"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	"github.com/cossim/coss-server/internal/group/app"
	"github.com/cossim/coss-server/internal/group/app/command"
	"github.com/cossim/coss-server/internal/group/app/query"
	"github.com/cossim/coss-server/internal/group/cache"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	"github.com/cossim/coss-server/pkg/discovery"
	"go.uber.org/zap"
	"strconv"
)

func NewApplication(ctx context.Context, ac *config.AppConfig, logger *zap.Logger) app.Application {

	services, err := discovery.NewBalanceGrpcClient(ac)
	if err != nil {
		panic(err)
	}

	var (
		userService           = &adapters.UserGrpc{}
		pushService           = &adapters.PushServiceGrpc{}
		groupService          = &adapters.GroupGrpc{}
		relationUserService   = &adapters.RelationUserGrpc{}
		relationGroupService  = &adapters.RelationGroupGrpc{}
		relationDialogService = &adapters.RelationDialogGrpc{}
	)

	for serviceName, conn := range services {
		fmt.Println(serviceName, conn.Target())
		//addr := conn.Target()
		switch serviceName {
		case app.UserServiceName:
			userService = adapters.NewUserGrpc(usergrpcv1.NewUserServiceClient(conn))
		case app.RelationUserServiceName:
			relationUserService = adapters.NewRelationUserGrpc(relationgrpcv1.NewUserRelationServiceClient(conn))
			relationGroupService = adapters.NewRelationGroupGrpc(relationgrpcv1.NewGroupRelationServiceClient(conn))
			relationDialogService = adapters.NewRelationDialogGrpc(relationgrpcv1.NewDialogServiceClient(conn))
		case app.PushServiceName:
			pushService = adapters.NewPushService(pushgrpcv1.NewPushServiceClient(conn))
		case app.GroupServiceName:
			groupService = adapters.NewGroupGrpc(groupgrpcv1.NewGroupServiceClient(conn))
		default:
		}
	}

	dtmGrpcServer := ac.Dtm.Addr()

	mysql, err := db.NewMySQL(ac.MySQL.Address, strconv.Itoa(ac.MySQL.Port), ac.MySQL.Username, ac.MySQL.Password, ac.MySQL.Database, int64(ac.Log.Level), ac.MySQL.Opts)
	if err != nil {
		panic(err)
	}

	dbConn, err := mysql.GetConnection()
	if err != nil {
		panic(err)
	}

	groupCache, err := cache.NewGroupCacheRedis(ac.Redis.Addr(), ac.Redis.Password, 0)
	if err != nil {
		panic(err)
	}

	groupRepository := adapters.NewMySQLGroupRepository(dbConn, groupCache)

	return app.Application{
		Commands: app.Commands{
			CreateGroup: command.NewCreateGroupHandler(
				groupRepository,
				logger,
				dtmGrpcServer,
				userService,
				pushService,
				relationUserService,
				relationGroupService,
			),
			DeleteGroup: command.NewDeleteGroupHandler(
				groupRepository,
				logger,
				dtmGrpcServer,
				relationGroupService,
				relationDialogService,
			),
			UpdateGroup: command.NewUpdateGroupHandler(
				groupRepository,
				logger,
				dtmGrpcServer,
				relationGroupService,
				groupService,
				ac.Cache.Enable,
				groupCache,
			),
		},
		Queries: app.Queries{
			GetGroup: query.NewGetGroupHandler(
				groupRepository,
				logger,
				dtmGrpcServer,
				relationGroupService,
				relationDialogService,
			),
		},
	}
}
