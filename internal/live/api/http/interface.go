package http

import "github.com/gin-gonic/gin"

type UserLiveService interface {
	Create(c *gin.Context)
	Join(c *gin.Context)
	Show(ctx *gin.Context)
	Leave(c *gin.Context)
}
