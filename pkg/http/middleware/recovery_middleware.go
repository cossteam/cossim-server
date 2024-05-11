package middleware

import (
	"fmt"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime/debug"
)

func RecoveryMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("recover error: %v\n", err)
				// 打印堆栈信息
				debug.PrintStack()
				//if e, ok := err.(error); ok {
				//	response.Fail(ctx, code.Cause(e).Message(), nil)
				//	ctx.Abort()
				//	return
				//}
				response.InternalServerError(ctx)
				ctx.Abort()
			}
		}()
		ctx.Next()
	}
}

func GRPCErrorMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if err := c.Errors.Last(); err != nil {
			if e, ok := err.Err.(error); ok {
				HandleError(c, logger, e)
				c.Abort()
			}
		}
	}
}

func HandleError(c *gin.Context, logger *zap.Logger, err error) {
	// 判断 gRPC 错误码
	//if grpcCode := status.Code(err); grpcCode == codes.Unavailable {
	//	// 连接不可用错误处理
	//	logger.Error("service unavailable", zap.Error(err))
	//} else if grpcCode == codes.Unauthenticated {
	//	// 未认证错误处理
	//	logger.Error("service unauthenticated", zap.Error(err))
	//}

	var ec code.Codes
	if st, ok := status.FromError(err); ok {
		ec = code.Code(int(st.Code()))
	} else {
		ec, ok = err.(code.Codes)
		if !ok {
			logger.Error("service error", zap.Error(err))
			response.InternalServerError(c)
			return
		}
	}
	logger.Error("service error", zap.String("msg", ec.Message()), zap.Error(ec))
	response.Fail(c, ec.Message(), nil)
}

func HandleGRPCErrors(c *gin.Context, logger *zap.Logger, err error) {
	// 判断 gRPC 错误码
	if grpcCode := status.Code(err); grpcCode == codes.Unavailable {
		// 连接不可用错误处理
		logger.Error("user service unavailable", zap.Error(err))
		//response.Fail(c, http.StatusText(http.StatusInternalServerError), nil)
		response.GRPCError(c, err)
		return
	} else if grpcCode == codes.Unauthenticated {
		// 未认证错误处理
		logger.Error("user service unauthenticated", zap.Error(err))
		//response.Fail(c, http.StatusText(http.StatusInternalServerError), nil)
		response.GRPCError(c, err)
		return
	}

	// 其他 gRPC 错误处理
	logger.Error("user service failed", zap.Error(err))
	response.Fail(c, err.Error(), nil)
}
