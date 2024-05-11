package middleware

import (
	"context"
	"errors"
	"fmt"
	authv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	omiddleware "github.com/oapi-codegen/gin-middleware"
	"net/http"
	"strings"
)

var (
	ErrNoAuthHeader      = errors.New("Authorization header is missing")
	ErrInvalidAuthHeader = errors.New("Authorization header is malformed")
	ErrClaimsInvalid     = errors.New("Provided claims do not match expected scopes")
)

func HandleOpenAPIError(c *gin.Context, message string, statusCode int) {
	if strings.Contains(message, "security requirements failed: authorization failed") {
		statusCode = http.StatusUnauthorized
		message = code.Unauthorized.Message()
	}
	if strings.Contains(message, "request body has an error: doesn't match schema") {
		index := strings.Index(message, "Error at")
		if index != -1 {
			message = strings.TrimSpace(message[index+len("Error at "):])
		} else {
			message = code.InvalidParameter.Message()
		}
	}
	response.SetResponse(c, statusCode, message, nil)
}

// GetJWSFromRequest extracts a JWS string from an Authorization: Bearer <jws> header
func GetJWSFromRequest(req *http.Request) (string, error) {
	queryToken := req.URL.Query().Get("token")
	if queryToken != "" {
		return queryToken, nil
	}
	authHdr := req.Header.Get("Authorization")
	// Check for the Authorization header.
	if authHdr == "" {
		return "", ErrNoAuthHeader
	}
	// We expect a header value of the form "Bearer <token>", with 1 space after
	// Bearer, per spec.
	prefix := "Bearer "
	if !strings.HasPrefix(authHdr, prefix) {
		return "", ErrInvalidAuthHeader
	}
	return strings.TrimPrefix(authHdr, prefix), nil
}

func NewAuthenticator(authService authv1.UserAuthServiceClient) openapi3filter.AuthenticationFunc {
	return func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
		return Authenticate(ctx, authService, input)
	}
}

func Authenticate(ctx context.Context, authService authv1.UserAuthServiceClient, input *openapi3filter.AuthenticationInput) error {
	// Our security scheme is named BearerAuth, ensure this is the case
	if input.SecuritySchemeName != "BearerAuth" && input.SecuritySchemeName != "bearerAuth" {
		return fmt.Errorf("security scheme %s != 'BearerAuth'", input.SecuritySchemeName)
	}

	jws, err := GetJWSFromRequest(input.RequestValidationInput.Request)
	if err != nil {
		return err
	}

	claims, err := authService.Access(ctx, &authv1.AccessRequest{Token: jws})
	if err != nil {
		return code.Unauthorized
	}

	gctx := omiddleware.GetGinContext(ctx)
	gctx.Set(constants.UserID, claims.UserID)
	gctx.Set(constants.DriverID, claims.DriverID)
	gctx.Set(constants.PublicKey, claims.PublicKey)
	return nil
}

func HandleOpenApiAuthentication(ctx context.Context, authService authv1.UserAuthServiceClient, input *openapi3filter.AuthenticationInput) error {
	if err := Authenticate(ctx, authService, input); err != nil {
		//gx := omiddleware.GetGinContext(ctx)
		//gx.JSON(http.StatusUnauthorized, gin.H{
		//	"code": 401,
		//	"msg":  err.Error(),
		//})
		//gx.Abort()
		return input.NewError(err)
	}

	return nil
}
