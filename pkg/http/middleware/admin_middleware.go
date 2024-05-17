package middleware

import (
	"fmt"
	"github.com/cossim/coss-server/internal/admin/app/service/admin"
	"github.com/cossim/coss-server/internal/admin/domain/entity"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/gin-gonic/gin"
)

func AdminAuthMiddleware(ad admin.Service) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取 authorization header
		fmt.Println("admin auth middleware")
		userID, ok := ctx.Get(constants.UserID)
		if !ok {
			ctx.JSON(401, gin.H{
				"code": 401,
				"msg":  "Unauthorized",
			})
			ctx.Abort()
		}

		id, err := ad.GetAdminByUserID(ctx, userID.(string))
		if err != nil {
			ctx.JSON(401, gin.H{
				"code": 401,
				"msg":  "Unauthorized",
			})
			ctx.Abort()
			return
		}

		if id.Status == entity.DisabledStatus {
			ctx.JSON(401, gin.H{
				"code": 401,
				"msg":  "Unauthorized",
			})
			ctx.Abort()
			return
		}

		ctx.Next()

	}
}
