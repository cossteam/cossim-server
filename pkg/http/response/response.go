package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
)

const (
	InternalServerErrorCode = http.StatusInternalServerError
)

func Response(ctx *gin.Context, httpStatus int, code int, msg string, data gin.H) {
	ctx.JSON(httpStatus, gin.H{
		"code": code,
		"msg":  extractErrorMessage(msg),
		"data": data,
	})
}

func Success(ctx *gin.Context, msg string, data gin.H) {
	Response(ctx, http.StatusOK, 200, msg, data)
}

func Fail(ctx *gin.Context, msg string, data gin.H) {
	Response(ctx, http.StatusOK, 400, msg, data)
}

func InternalServerError(ctx *gin.Context) {
	Response(ctx, InternalServerErrorCode, InternalServerErrorCode, http.StatusText(InternalServerErrorCode), nil)
}

func NetworkErrorRetry(ctx *gin.Context) {
	Response(ctx, 400, 400, "Network error. Please retry.", nil)
}

func GRPCError(ctx *gin.Context, err error) {
	InternalServerError(ctx)
}

func extractErrorMessage(input string) string {
	re := regexp.MustCompile(`rpc error: code = \w+ desc = (.+)`)
	match := re.FindStringSubmatch(input)

	if len(match) < 2 {
		return input
	}

	return match[1]
}
