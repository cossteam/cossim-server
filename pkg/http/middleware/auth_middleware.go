package middleware

import (
	authv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func AuthMiddleware(authService authv1.UserAuthServiceClient) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		tokenString := ""
		if ctx.GetHeader("Authorization") != "" {
			tokenString = ctx.GetHeader("Authorization")
			if !strings.HasPrefix(tokenString, "Bearer ") {
				ctx.JSON(http.StatusUnauthorized, gin.H{
					"code": 401,
					"msg":  http.StatusText(http.StatusUnauthorized),
				})
				ctx.Abort()
				return
			}
			tokenString = tokenString[7:]
		}
		if ctx.Query("token") != "" {
			tokenString = ctx.Query("token")
		}
		if tokenString == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  http.StatusText(http.StatusUnauthorized),
			})
			ctx.Abort()
			return
		}

		_, err := authService.Access(ctx, &authv1.AccessRequest{Token: tokenString})
		if err != nil {
			var c int
			var m string
			if code.Cause(err).Code() == code.Unauthorized.Code() {
				c = http.StatusUnauthorized
				m = http.StatusText(http.StatusUnauthorized)
			} else {
				c = http.StatusUnauthorized
				m = http.StatusText(http.StatusUnauthorized)
			}
			ctx.JSON(c, gin.H{
				"code": c,
				"msg":  m,
			})
			ctx.Abort()
			return
		}

		claims, err := authService.ParseToken(ctx, &authv1.ParseTokenRequest{Token: tokenString})
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  http.StatusText(http.StatusUnauthorized),
			})
			ctx.Abort()
			return
		}

		ctx.Set(constants.UserID, claims.UserID)
		ctx.Set(constants.DriverID, claims.DriverID)
		ctx.Set(constants.PublicKey, claims.PublicKey)
		ctx.Next()
	}
}
