package http

import (
	"github.com/cossim/coss-server/interface/live/api/dto"
	pkghttp "github.com/cossim/coss-server/pkg/http"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
)

// GroupCreate
// @Summary 创建群聊通话
// @Description 创建群聊通话
// @Tags GroupUser
// @Security Bearer
// @Param Authorization header string true "Bearer JWT"
// @param request body dto.GroupCallRequest true "request"
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.Response{data=dto.GroupCallResponse}
// @Router /live/group/create [post]
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

// GroupJoin
// @Summary 加入群聊通话
// @Description 加入群聊通话
// @Tags GroupUser
// @Security Bearer
// @Param Authorization header string true "Bearer JWT"
// @param request body dto.GroupJoinRequest true "request"
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.Response{data=dto.GroupJoinRequest}
// @Router /live/group/join [post]
func (h *Handler) GroupJoin(c *gin.Context) {
	req := new(dto.GroupJoinRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		h.logger.Error("token解析失败", zap.Error(err))
		response.Fail(c, "token解析失败", nil)
		return
	}

	resp, err := h.svc.GroupJoinRoom(c, req.GroupID, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "加入群聊通话成功", resp)
}

// GroupShow
// @Summary 获取群聊通话信息
// @Description 获取群聊通话信息
// @Tags GroupUser
// @Security Bearer
// @Param Authorization header string true "Bearer JWT"
// @Param group_id query uint32 true "群聊id"
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.Response{data=dto.GroupJoinRequest}
// @Router /live/group/show [get]
func (h *Handler) GroupShow(c *gin.Context) {
	gid, err := strconv.Atoi(c.Query("group_id"))
	if err != nil {
		response.Fail(c, "参数验证失败", nil)
		return
	}

	userID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		h.logger.Error("token解析失败", zap.Error(err))
		response.Fail(c, "token解析失败", nil)
		return
	}

	resp, err := h.svc.GroupShowRoom(c, uint32(gid), userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "获取群聊通话信息成功", resp)
}

// GroupReject
// @Summary 拒绝群聊通话
// @Description 拒绝群聊通话
// @Tags GroupUser
// @Security Bearer
// @Param Authorization header string true "Bearer JWT"
// @param request body dto.GroupShowRequest true "request"
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.Response{data=dto.GroupJoinRequest}
// @Router /live/group/reject [post]
func (h *Handler) GroupReject(c *gin.Context) {
	req := new(dto.GroupRejectRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		h.logger.Error("token解析失败", zap.Error(err))
		response.Fail(c, "token解析失败", nil)
		return
	}

	resp, err := h.svc.GroupRejectRoom(c, req.GroupID, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "拒绝群聊通话成功", resp)
}

// GroupLeave
// @Summary 离开群聊通话
// @Description 离开群聊通话
// @Tags GroupUser
// @Security Bearer
// @Param Authorization header string true "Bearer JWT"
// @param request body dto.GroupLeaveRequest true "request"
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.Response{data=dto.GroupJoinRequest}
// @Router /live/group/leave [post]
func (h *Handler) GroupLeave(c *gin.Context) {
	req := new(dto.GroupLeaveRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		h.logger.Error("token解析失败", zap.Error(err))
		response.Fail(c, "token解析失败", nil)
		return
	}

	resp, err := h.svc.GroupLeaveRoom(c, req.GroupID, userID, req.Force)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "离开群聊通话成功", resp)
}
