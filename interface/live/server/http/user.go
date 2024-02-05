package http

import (
	"github.com/cossim/coss-server/interface/live/api/dto"
	"github.com/cossim/coss-server/pkg/code"
	pkghttp "github.com/cossim/coss-server/pkg/http"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UserCreate
// @Summary 创建用户通话
// @Description 创建用户通话
// @Tags liveUser
// @Security Bearer
// @Param Authorization header string true "Bearer JWT"
// @param request body dto.UserCallRequest true "request"
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.Response{data=dto.UserCallResponse}
// @Router /live/user/create [post]
func (h *Handler) UserCreate(c *gin.Context) {
	req := new(dto.UserCallRequest)
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

	resp, err := h.svc.CreateUserCall(c, userID, req.UserID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "创建通话成功", resp)
}

// UserJoin
// @Summary 加入通话
// @Description 加入通话
// @Tags liveUser
// @Security Bearer
// @Param Authorization header string true "Bearer JWT"
// @param request body dto.UserJoinRequest true "request"
// @Accept  json
// @Produce  json
// @Success 200 {object} dto.Response{}
// @Router /live/user/join [post]
func (h *Handler) UserJoin(c *gin.Context) {
	req := new(dto.UserJoinRequest)
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

	resp, err := h.svc.UserJoinRoom(c, userID, req.Room)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "加入通话成功", resp)
}

// UserShow
// @Summary 获取通话房间信息
// @Description 获取通话房间信息
// @Tags liveUser
// @Security Bearer
// @Param Authorization header string true "Bearer JWT"
// @Param room query string true "房间名"
// @Produce  json
// @Success		200 {object} dto.Response{data=dto.UserShowResponse} "participant=通话参与者"
// @Router /live/user/show [get]
func (h *Handler) UserShow(c *gin.Context) {
	rid := c.Query("room")
	uid, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, "token解析失败", nil)
		return
	}

	if uid == "" || rid == "" {
		response.SetFail(c, code.InvalidParameter.Message(), nil)
		return
	}

	resp, err := h.svc.GetUserRoom(c, uid, rid)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "获取通话信息成功", resp)
}

// UserReject
// @Summary 拒绝通话
// @Description 拒绝通话
// @Tags liveUser
// @Security Bearer
// @Param Authorization header string true "Bearer JWT"
// @param request body dto.UserRejectRequest true "request"
// @Accept  json
// @Produce  json
// @Success		200 {object} dto.Response{}
// @Router /live/user/reject [post]
func (h *Handler) UserReject(c *gin.Context) {
	req := new(dto.UserRejectRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	uid, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, "token解析失败", nil)
		return
	}

	resp, err := h.svc.UserRejectRoom(c, uid, req.Room)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetFail(c, "拒绝通话成功", resp)
}

// UserLeave
// @Summary 结束通话
// @Description 结束通话
// @Tags liveUser
// @Security Bearer
// @Param Authorization header string true "Bearer JWT"
// @param request body dto.UserLeaveRequest true "request"
// @Accept  json
// @Produce  json
// @Success		200 {object} dto.Response{}
// @Router /live/user/leave [post]
func (h *Handler) UserLeave(c *gin.Context) {
	req := new(dto.UserLeaveRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	uid, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, "token解析失败", nil)
		return
	}

	resp, err := h.svc.UserLeaveRoom(c, uid, req.Room)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetFail(c, "结束通话成功", resp)
}

func (h *Handler) getJoinToken(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		response.SetFail(c, code.InvalidParameter.Message(), nil)
		return
	}

	resp, err := h.svc.GetJoinToken(c, userID, "")
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取token成功", resp)
}
