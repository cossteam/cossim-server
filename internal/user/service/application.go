package service

import (
	"context"
	"github.com/cossim/coss-server/internal/user/app"
	"github.com/cossim/coss-server/internal/user/app/command"
	"github.com/cossim/coss-server/internal/user/app/query"
	uc "github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/service"
	"github.com/cossim/coss-server/internal/user/infra/persistence"
	"github.com/cossim/coss-server/internal/user/infra/remote"
	"github.com/cossim/coss-server/internal/user/infra/rpc"
	"github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/cossim/coss-server/pkg/email/smtp"
	"github.com/cossim/coss-server/pkg/storage/minio"
	"go.uber.org/zap"
	"strconv"
)

func NewApplication(ctx context.Context, ac *config.AppConfig, logger *zap.Logger) *app.Application {
	var dtmGrpcServer = ac.Dtm.Addr()

	userCache, err := uc.NewUserCacheRedis(ac.Redis.Addr(), ac.Redis.Password, 0)
	if err != nil {
		panic(err)
	}

	//relationUserCache, err := ruc.NewRelationUserCacheRedis(ac.Redis.Addr(), ac.Redis.Password, 0)
	//if err != nil {
	//	panic(err)
	//}

	mysql, err := db.NewMySQL(ac.MySQL.Address, strconv.Itoa(ac.MySQL.Port), ac.MySQL.Username, ac.MySQL.Password, ac.MySQL.Database, int64(ac.Log.Level), ac.MySQL.Opts)
	if err != nil {
		panic(err)
	}

	dbConn, err := mysql.GetConnection()
	if err != nil {
		panic(err)
	}

	userRepo := persistence.NewMySQLUserRepository(dbConn, userCache)

	userLoginRepo := persistence.NewMySQLUserLoginRepository(dbConn, userCache)

	authDomain := service.NewAuthDomain(ac.SystemConfig.JwtSecret, userRepo, userCache)

	userDomain := service.NewUserDomain(userRepo)

	userLoginDomain := service.NewUserLoginDomain(userRepo, userLoginRepo, userCache, ac.MultipleDeviceLimit.Enable, ac.MultipleDeviceLimit.Max)

	var relationAddr string
	if ac.Discovers["relation"].Direct {
		relationAddr = ac.Discovers["relation"].Addr()
	} else {
		relationAddr = discovery.GetBalanceAddr(ac.Register.Addr(), ac.Discovers["relation"].Name)
	}

	var msgAddr string
	if ac.Discovers["msg"].Direct {
		msgAddr = ac.Discovers["msg"].Addr()
	} else {
		msgAddr = discovery.GetBalanceAddr(ac.Register.Addr(), ac.Discovers["msg"].Name)
	}

	var pushAddr string
	if ac.Discovers["push"].Direct {
		pushAddr = ac.Discovers["push"].Addr()
	} else {
		pushAddr = discovery.GetBalanceAddr(ac.Register.Addr(), ac.Discovers["push"].Name)
	}

	//var userAddr string
	//if ac.Discovers["user"].Direct {
	//	userAddr = ac.Discovers["user"].Addr()
	//} else {
	//	userAddr = discovery.GetBalanceAddr(ac.Register.Addr(), ac.Discovers["user"].Name)
	//}

	relationUserService, err := rpc.NewRelationUserGrpc(relationAddr)
	if err != nil {
		panic(err)
	}

	relationDialogService, err := rpc.NewRelationDialogGrpc(relationAddr)
	if err != nil {
		panic(err)
	}

	msgService, err := rpc.NewMsgServiceGrpc(msgAddr)
	if err != nil {
		panic(err)
	}

	pushService, err := rpc.NewPushService(pushAddr)
	if err != nil {
		panic(err)
	}

	//userService, err := rpc.NewUserGrpc(userAddr)
	//if err != nil {
	//	panic(err)
	//}

	storageProvider, err := minio.NewMinIOStorage(ac.OSS.Addr(), ac.OSS.AccessKey, ac.OSS.SecretKey, ac.OSS.SSL)
	if err != nil {
		panic(err)
	}

	storageService := remote.NewStorageService(storageProvider)

	smtpStorage, err := smtp.NewSmtpStorage(ac.Email.SmtpServer, ac.Email.Port, ac.Email.Username, ac.Email.Password)
	if err != nil {
		panic(err)
	}

	smtpService := remote.NewSmtpService(smtpStorage)

	return &app.Application{
		Commands: app.Commands{
			UserLogin: command.NewUserLoginHandler(
				logger,
				userCache,
				dtmGrpcServer,
				authDomain,
				userDomain,
				userLoginDomain,
				relationUserService,
				relationDialogService,
				msgService,
				pushService,
			),
			UserLogout: command.NewUserLogoutHandler(
				logger,
				userCache,
				dtmGrpcServer,
				authDomain,
				userDomain,
				userLoginDomain,
				pushService,
			),
			UpdatePassword: command.NewUpdatePasswordHandler(logger, userDomain),
			UserActivate:   command.NewUserActivateHandler(logger, userDomain, userCache),
			UserRegister: command.NewUserRegisterHandler(
				logger,
				dtmGrpcServer,
				ac.SystemConfig.GatewayAddress,
				ac.SystemConfig.Ssl,
				ac.Email.Enable,
				userCache,
				userDomain,
				relationUserService,
				smtpService,
				storageService,
			),
			UpdateUserBundle: command.NewUpdateUserBundle(logger, userDomain),
			SetUserPublicKey: command.NewSetUserPublicKeyHandler(logger, userDomain),
			SendUserEmailVerification: command.NewSendUserEmailVerificationHandler(
				logger,
				userDomain,
				userCache,
				smtpService,
			),
			ResetUserPublicKey: command.NewResetUserPublicKeyHandler(
				logger,
				userDomain,
				userCache,
				smtpService,
			),
			UpdateUserAvatarHandler: command.NewUpdateUserAvatarHandler(
				logger,
				ac.SystemConfig.Ssl,
				ac.SystemConfig.GatewayAddress,
				userDomain,
				storageService,
			),
			UpdateUser: command.NewUpdateUserHandler(
				logger,
				userDomain,
				userCache,
			),
		},
		Queries: app.Queries{
			GetUser: query.NewGetUserHandler(
				logger,
				userDomain,
				relationUserService,
			),
			GetUserBundle: query.NewGetUserBundleHandler(
				logger,
				userDomain,
			),
			GetUserLoginClients: query.NewGetUserClientsHandler(
				logger,
				userCache,
				userDomain,
			),
		},
	}
}
