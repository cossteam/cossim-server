package http

import (
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/gin-gonic/gin"
)

// 解析请求头中的token返回uid
func ParseTokenReUid(ctx *gin.Context) (string, error) {
	token := ctx.GetHeader("Authorization")
	if token != "" {
		_, c2, err := utils.ParseToken(token)
		if err != nil {
			return "", err
		}
		return c2.UserId, nil
	}
	return "", nil
}
