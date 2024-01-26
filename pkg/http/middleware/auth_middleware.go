package middleware

import (
	"fmt"
	"github.com/cossim/coss-server/pkg/auth"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/db"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"net/http"
	"strings"
)

func AuthMiddleware(rdb *redis.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
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
		fmt.Println(tokenString)
		//validate token formate
		if tokenString == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  http.StatusText(http.StatusUnauthorized),
			})
			ctx.Abort()
			return
		}
		fmt.Println(tokenString)
		conn, err := db.NewDefaultMysqlConn().GetConnection()
		if err != nil {
			fmt.Printf("获取数据库连接失败: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{
				"code": 500,
				"msg":  http.StatusText(http.StatusInternalServerError),
			})
			ctx.Abort()
			return
		}

		a := auth.NewAuthenticator(conn, rdb)

		drive := ctx.GetHeader("X-Device-Type")
		drive = constants.DetermineClientType(drive)

		is, err := a.ValidateToken(tokenString, drive)
		if err != nil || !is {
			fmt.Printf("token验证失败: %v", err)
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  http.StatusText(http.StatusUnauthorized),
			})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}
