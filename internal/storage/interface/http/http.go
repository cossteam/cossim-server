package http

import (
	"context"
	"fmt"
	v1 "github.com/cossim/coss-server/internal/storage/api/http/v1"
	service "github.com/cossim/coss-server/internal/storage/app/service/storage"
	grpcinter "github.com/cossim/coss-server/internal/storage/interface/grpc"
	authv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
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
	StorageClient *grpcinter.Handler
	logger        *zap.Logger
	enc           encryption.Encryptor
	svc           service.Service
	minioAddr     string
	authService   authv1.UserAuthServiceClient
}

func (h *Handler) Init(cfg *pkgconfig.AppConfig) error {
	h.logger = plog.NewDefaultLogger("storage_bff", int8(cfg.Log.Level))
	h.minioAddr = cfg.OSS.Addr()
	h.enc = encryption.NewEncryptor([]byte(cfg.Encryption.Passphrase), cfg.Encryption.Name, cfg.Encryption.Email, cfg.Encryption.RsaBits, cfg.Encryption.Enable)

	if cfg.Encryption.Enable {
		return h.enc.ReadKeyPair()
	}
	h.svc = service.NewService(h.logger)
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
	return nil
}

func (h *Handler) Name() string {
	return "storage_bff"
}

func (h *Handler) Version() string {
	return version.FullVersion()
}

// @title CossApi

func (h *Handler) RegisterRoute(r gin.IRouter) {
	r.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(h.logger), middleware.EncryptionMiddleware(h.enc), middleware.RecoveryMiddleware())

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

	//注册类型
	openapi3filter.RegisterBodyDecoder("multipart/form-data", openapi3filter.FileBodyDecoder)

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
	return nil
}

func (h *Handler) DiscoverServices(services map[string]*grpc.ClientConn) error {
	for k, v := range services {
		if err := h.svc.HandlerGrpcClient(k, v); err != nil {
			h.logger.Error("handler grpc client error", zap.String("name", k), zap.String("addr", v.Target()), zap.Error(err))
		}
	}
	return nil
}

var (
	downloadURL     = "/api/v1/storage/files/download"
	systemEnableSSL bool
	gatewayAddress  string
	gatewayPort     string
)
