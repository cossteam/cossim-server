package http

import (
	"github.com/cossim/coss-server/internal/relation/api/http/model"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary 关闭或打开对话(action: 0:关闭对话, 1:打开对话)
// @Description 关闭或打开对话(action: 0:关闭对话, 1:打开对话)
// @Tags Dialog
// @Accept  json
// @Produce  json
// @param request body model.CloseOrOpenDialogRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /relation/dialog/show [post]
func (h *Handler) closeOrOpenDialog(c *gin.Context) {
	req := new(model.CloseOrOpenDialogRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	if !model.IsValidOpenAction(req.Action) {
		response.SetFail(c, "打开或关闭对话失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	_, err := h.svc.OpenOrCloseDialog(c, userID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "操作成功", gin.H{"dialog_id": req.DialogId})
}

// @Summary 是否置顶对话(action: 0:关闭取消置顶对话, 1:置顶对话)
// @Description 是否置顶对话(action: 0:关闭取消置顶对话, 1:置顶对话)
// @Tags Dialog
// @Accept  json
// @Produce  json
// @param request body model.TopOrCancelTopDialogRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /relation/dialog/top [post]
func (h *Handler) topOrCancelTopDialog(c *gin.Context) {
	req := new(model.TopOrCancelTopDialogRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	if !model.IsValidTopAction(req.Action) {
		response.SetFail(c, "置顶或取消置顶对话失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	_, err := h.svc.TopOrCancelTopDialog(c, userID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "操作成功", gin.H{"dialog_id": req.DialogId})
}
