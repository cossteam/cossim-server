package middleware

import (
	"fmt"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				response.Fail(ctx, fmt.Sprint(err), nil)
			}
		}()

		ctx.Next()
	}
}

func GRPCErrorMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if err := c.Errors.Last(); err != nil {
			if grpcErr, ok := err.Err.(error); ok {
				HandleGRPCErrors(c, logger, grpcErr)
				c.Abort()
			}
		}
	}
}

func HandleGRPCErrors(c *gin.Context, logger *zap.Logger, err error) {
	// 判断 gRPC 错误码
	if grpcCode := status.Code(err); grpcCode == codes.Unavailable {
		// 连接不可用错误处理
		logger.Error("user service unavailable", zap.Error(err))
		response.Fail(c, http.StatusText(http.StatusInternalServerError), nil)
		return
	} else if grpcCode == codes.Unauthenticated {
		// 未认证错误处理
		logger.Error("user service unauthenticated", zap.Error(err))
		response.Fail(c, http.StatusText(http.StatusInternalServerError), nil)
		return
	}

	// 其他 gRPC 错误处理
	logger.Error("user service failed", zap.Error(err))
	response.Fail(c, err.Error(), nil)
}
