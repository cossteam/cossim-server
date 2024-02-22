package http

import (
	"github.com/cossim/coss-server/admin/api/model"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary 发送全体通知
// @Description 发送全体通知
// @Accept  json
// @Produce  json
// @param request body model.SendAllNotificationRequest{} true "request"
// @Success		200 {object} model.Response{}
// @Router /admin/notification/send_all [post]
func (h *Handler) sendAllNotification(c *gin.Context) {
	req := new(model.SendAllNotificationRequest)
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
