package http

import (
	"context"
	"fmt"
	v1 "github.com/cossim/coss-server/internal/msg/api/http/v1"
	appservice "github.com/cossim/coss-server/internal/msg/app/service/msg"
	grpcHandler "github.com/cossim/coss-server/internal/msg/interfaces/grpc"
	authv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/rpc/client"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/db"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	"github.com/cossim/coss-server/pkg/http/response"
	plog "github.com/cossim/coss-server/pkg/log"
	"github.com/cossim/coss-server/pkg/manager/server"
	"github.com/cossim/coss-server/pkg/version"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	oapimiddleware "github.com/oapi-codegen/gin-middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"os"
	"strconv"
)

var (
	_ server.HTTPService = &Handler{}
)

type Handler struct {
	logger      *zap.Logger
	enc         encryption.Encryptor
	MsgClient   *grpcHandler.Handler
	authService authv1.UserAuthServiceClient
	svc         appservice.Service
	userCache   cache.UserCache
}

func (h *Handler) Init(cfg *pkgconfig.AppConfig) error {
	h.logger = plog.NewDefaultLogger("msg_bff", int8(cfg.Log.Level))
	if cfg.Encryption.Enable {
		if err := h.enc.ReadKeyPair(); err != nil {
			return err
		}
	}

	h.enc = encryption.NewEncryptor([]byte(cfg.Encryption.Passphrase), cfg.Encryption.Name, cfg.Encryption.Email, cfg.Encryption.RsaBits, cfg.Encryption.Enable)
	var userAddr string
	if cfg.Discovers["user"].Direct {
		userAddr = cfg.Discovers["user"].Addr()
	} else {
		userAddr = discovery.GetBalanceAddr(cfg.Register.Addr(), cfg.Discovers["user"].Name)
	}
	authClient, err := client.NewAuthClient(userAddr)
	if err != nil {
		return err
	}
	h.authService = authClient

	h.svc = appservice.NewService(cfg.Dtm.Addr(), h.logger)

	mysql, err := db.NewMySQL(cfg.MySQL.Address, strconv.Itoa(cfg.MySQL.Port), cfg.MySQL.Username, cfg.MySQL.Password, cfg.MySQL.Database, int64(cfg.Log.Level), cfg.MySQL.Opts)
	if err != nil {
		return err
	}
	dbConn, err := mysql.GetConnection()
	if err != nil {
		return err
	}
	err = h.svc.Init(dbConn, cfg)
	if err != nil {
		return err
	}

	return nil
}

func (h *Handler) Name() string {
	return "msg_bff"
}

func (h *Handler) Version() string {
	return version.FullVersion()
}

// @title CossApi

func (h *Handler) RegisterRoute(r gin.IRouter) {
	//u := r.Group("/api/v1/msg")
	r.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(h.logger), middleware.EncryptionMiddleware(h.enc), middleware.RecoveryMiddleware())
	r.Use(middleware.AuthMiddleware(h.authService))

	swagger, err := v1.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}
	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	validatorOptions := &oapimiddleware.Options{
		ErrorHandler: func(c *gin.Context, message string, statusCode int) {
			fmt.Println("statusCode => ", statusCode)
			response.SetFail(c, message, nil)
		},
	}
	validatorOptions.Options.AuthenticationFunc = func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		return middleware.HandleOpenApiAuthentication(ctx, h.authService, input)
	}

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	r.Use(oapimiddleware.OapiRequestValidatorWithOptions(swagger, validatorOptions))

	v1.RegisterHandlers(r, h)
}

func (h *Handler) Health(r gin.IRouter) string {
	//TODO implement me
	panic("implement me")
}

func (h *Handler) Stop(ctx context.Context) error {
	return h.svc.Stop(ctx)
}

func (h *Handler) DiscoverServices(services map[string]*grpc.ClientConn) error {
	for k, v := range services {
		if err := h.svc.HandlerGrpcClient(k, v); err != nil {
			h.logger.Error("handler grpc client error", zap.String("name", k), zap.String("addr", v.Target()), zap.Error(err))
		}
	}
	return nil
}
