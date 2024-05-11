package interfaces

import (
	"context"
	"fmt"
	authv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	v1 "github.com/cossim/coss-server/internal/user/api/http/v1"
	"github.com/cossim/coss-server/internal/user/app"
	"github.com/cossim/coss-server/internal/user/rpc/client"
	pkgconfig "github.com/cossim/coss-server/pkg/config"
	"github.com/cossim/coss-server/pkg/discovery"
	"github.com/cossim/coss-server/pkg/encryption"
	"github.com/cossim/coss-server/pkg/http/middleware"
	"github.com/cossim/coss-server/pkg/version"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	oapimiddleware "github.com/oapi-codegen/gin-middleware"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"os"
)

var _ v1.ServerInterface = &HttpServer{}

func NewHttpServer(logger *zap.Logger, app app.Application) *HttpServer {
	return &HttpServer{logger: logger, app: app}
}

type HttpServer struct {
	logger      *zap.Logger
	app         app.Application
	enc         encryption.Encryptor
	authService authv1.UserAuthServiceClient

	pgpKey string
}

func (h *HttpServer) Init(cfg *pkgconfig.AppConfig) error {
	h.enc = encryption.NewEncryptor([]byte(cfg.Encryption.Passphrase), cfg.Encryption.Name, cfg.Encryption.Email, cfg.Encryption.RsaBits, cfg.Encryption.Enable)
	if cfg.Encryption.Enable {
		if err := h.enc.ReadKeyPair(); err != nil {
			return err
		}
		h.pgpKey = h.enc.GetPublicKey()
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

func (h *HttpServer) Name() string {
	return "user_bff"
}

func (h *HttpServer) Version() string {
	return version.FullVersion()
}

func (h *HttpServer) RegisterRoute(r gin.IRouter) {
	// 添加一些中间件或其他配置
	r.Use(middleware.CORSMiddleware(), middleware.GRPCErrorMiddleware(h.logger), middleware.EncryptionMiddleware(h.enc))
	swagger, err := v1.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how this thing will be run.
	swagger.Servers = nil

	validatorOptions := &oapimiddleware.Options{
		ErrorHandler: middleware.HandleOpenAPIError,
	}
	validatorOptions.Options.AuthenticationFunc = func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		return middleware.HandleOpenApiAuthentication(ctx, h.authService, input)
	}

	// Use our validation middleware to check all requests against the
	// OpenAPI schema.
	r.Use(oapimiddleware.OapiRequestValidatorWithOptions(swagger, validatorOptions))
	v1.RegisterHandlers(r, h)
}

func (h *HttpServer) Health(r gin.IRouter) string {
	return ""
}

func (h *HttpServer) Stop(ctx context.Context) error {
	return nil
}

func (h *HttpServer) DiscoverServices(services map[string]*grpc.ClientConn) error {
	return nil
}
