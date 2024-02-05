package http

import (
	"github.com/cossim/coss-server/interface/live/api/dto"
	pkghttp "github.com/cossim/coss-server/pkg/http"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GroupCreate
// @Summary 创建群聊通话
// @Description 创建群聊通话
// @Tags liveUser
// @Security Bearer
// @Param Authorization header string true "Bearer JWT"
// @param request body dto.UserCallRequest true "request"
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.Response{data=map[string]string} "url=webRtc服务器地址 token=加入通话的token room_name=房间名称 room_id=房间id"
// @Router /live/user/create [post]
func (h *Handler) GroupCreate(c *gin.Context) {
	req := new(dto.GroupCallRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数验证失败", nil)
		return
	}

	userID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		h.logger.Error("token解析失败", zap.Error(err))
		response.Fail(c, "token解析失败", nil)
		return
	}

	resp, err := h.svc.CreateGroupCall(c, userID, req.GroupID, req.Member)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "创建群聊通话成功", resp)
}

func (h *Handler) GroupJoin(c *gin.Context) {
	req := new(dto.UserJoinRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}
}
