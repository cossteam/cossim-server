package interfaces

import (
	v1 "github.com/cossim/coss-server/internal/relation/api/http/v1"
	"github.com/cossim/coss-server/internal/relation/app/command"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
)

func (h *HttpServer) ShowDialog(c *gin.Context, id uint32) {
	req := &v1.ShowDialogJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.SetFail(c, "参数错误", nil)
		return
	}

	if err := h.app.Commands.ShowDialog.Handle(c, &command.ShowDialog{
		UserID:   c.Value(constants.UserID).(string),
		DialogID: id,
		Show:     req.Show,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置对话状态成功", nil)

}

func (h *HttpServer) TopDialog(c *gin.Context, id uint32) {
	req := &v1.TopDialogJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.SetFail(c, "参数错误", nil)
		return
	}

	if err := h.app.Commands.TopDialog.Handle(c, &command.TopDialog{
		UserID:   c.Value(constants.UserID).(string),
		DialogID: id,
		Show:     req.Top,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置对话置顶状态成功", nil)
}
