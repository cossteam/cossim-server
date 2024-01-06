package http

import (
	"errors"
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/gin-gonic/gin"
)

// ParseTokenReUid 解析请求头中的 token 返回 uid
func ParseTokenReUid(ctx *gin.Context) (string, error) {
	tokenString := ctx.GetHeader("Authorization")

	if tokenString == "" {
		return "", errors.New("authorization header is empty")
	}

	token := tokenString[7:]
	if token != "" {
		_, c2, err := utils.ParseToken(token)
		if err != nil {
			return "", err
		}
		return c2.UserId, nil
	}

	return "", errors.New("token is empty")
}
