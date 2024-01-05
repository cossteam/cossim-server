package http

import (
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/gin-gonic/gin"
)

// ParseTokenReUid 解析请求头中的token返回uid
func ParseTokenReUid(ctx *gin.Context) (string, error) {
	tokenString := ctx.GetHeader("Authorization")
	token := tokenString[7:]
	if token != "" {
		_, c2, err := utils.ParseToken(token)
		if err != nil {
			return "", err
		}
		return c2.UserId, nil
	}
	return "", nil
}
