package http

import (
	v1 "github.com/cossim/coss-server/internal/admin/api/http/v1"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// SendAllNotification
// @Summary 发送全体通知
// @Description 发送全体通知
// @Tags Admin
// @Accept  json
// @Produce  json
// @param request body v1.SendAllNotificationRequest{} true "request"
// @Success		200 {object} v1.Response{}
// @Router /admin/notification/send_all [post]
func (h *Handler) SendAllNotification(c *gin.Context) {
	req := new(v1.SendAllNotificationRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	_, err := h.svc.SendAllNotification(c, req.Content)
	if err != nil {
		response.SetFail(c, "发送失败", err)
		return
	}

	response.SetSuccess(c, "发送成功", nil)
}
