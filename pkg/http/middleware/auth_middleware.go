package middleware

import (
	"fmt"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func AuthMiddleware(userCache cache.UserCache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//头像请求跳过验证
		if strings.HasPrefix(ctx.FullPath(), "/api/v1/storage/files/download/") {
			fileType := ctx.Param("type")
			if fileType == minio.PublicBucket {
				ctx.Next()
				return
			}
		}

		// 获取 authorization header
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

		_, claims, err := utils.ParseToken(tokenString)
		if err != nil {
			fmt.Printf("token解析失败: %v", err)
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  http.StatusText(http.StatusUnauthorized),
			})
			ctx.Abort()
			return
		}

		infos, err := userCache.GetUserLoginInfos(ctx, claims.UserId)
		if err != nil {
			return
		}

		var found bool
		for _, v := range infos {
			if v.Token == tokenString {
				found = true
				break
			}
		}

		if !found {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  http.StatusText(http.StatusUnauthorized),
			})
			ctx.Abort()
			return
		}
		ctx.Set(constants.UserID, claims.UserId)
		ctx.Set(constants.DriverID, claims.DriverId)
		ctx.Set(constants.PublicKey, claims.PublicKey)
		ctx.Next()
	}
}
