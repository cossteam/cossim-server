package middleware

import (
	"fmt"
	"github.com/cossim/coss-server/pkg/auth"
	"github.com/cossim/coss-server/pkg/db"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取 authorization header
		tokenString := ctx.GetHeader("Authorization")

		//validate token formate
		if tokenString == "" || !strings.HasPrefix(tokenString, "Bearer ") {
			ctx.JSON(http.StatusUnauthorized, gin.H{
				"code": 401,
				"msg":  http.StatusText(http.StatusUnauthorized),
			})
			ctx.Abort()
			return
		}

		tokenString = tokenString[7:] //截取字符

		//token, claims, err := utils.ParseToken(tokenString)
		//
		//if err != nil || !token.Valid {
		//	ctx.JSON(http.StatusUnauthorized, gin.H{
		//		"code": 401,
		//		"msg":  http.StatusText(http.StatusUnauthorized),
		//	})
		//	ctx.Abort()
		//	return
		//}
		//
		//conn, err := db.GetConnection()
		//if err != nil {
		//	fmt.Printf("获取数据库连接失败: %v", err)
		//	ctx.JSON(http.StatusInternalServerError, gin.H{
		//		"code": 500,
		//		"msg":  http.StatusText(http.StatusInternalServerError),
		//	})
		//	ctx.Abort()
		//	return
		//}

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

		a := auth.NewAuthenticator(conn)

		is, err := a.ValidateToken(tokenString)
		if err != nil || !is {
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
