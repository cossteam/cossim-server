package http

import (
	v1 "github.com/cossim/coss-server/internal/msg/api/http/v1"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
)

// SendUserMsg
// @Summary 发送私聊消息
// @Description 发送私聊消息
// @Tags Msg
// @Accept  json
// @Produce  json
// @param request body v1.SendUserMsgRequest true "request"
// @Success		200 {object} v1.Response{}
// @Router /msg/user/send [post]
func (h *Handler) SendUserMsg(c *gin.Context) {
	req := new(v1.SendUserMsgRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	driverID := c.Value(constants.DriverID).(string)
	resp, err := h.svc.SendUserMsg(c, userID, driverID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "发送成功", resp)
}

// SendGroupMsg
// @Summary 发送群聊消息
// @Description 发送群聊消息
// @Tags Msg
// @Accept  json
// @Produce  json
// @param request body v1.SendGroupMsgRequest true "request"
// @Success		200 {object} v1.Response{}
// @Router /msg/group/send [post]
func (h *Handler) SendGroupMsg(c *gin.Context) {
	req := new(v1.SendGroupMsgRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	if req.GroupId == 0 {
		response.SetFail(c, "群聊id不正确", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	driverID := c.Value(constants.DriverID).(string)
	resp, err := h.svc.SendGroupMsg(c, userID, driverID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "发送成功", resp)
}

// GetUserMsgList
// @Summary 获取私聊消息
// @Description 获取私聊消息
// @Tags Msg
// @Accept  json
// @Produce  json
// @Param dialog_id query string true "对话id"
// @Param type query string false "类型"
// @Param user_id query string false "用户id"
// @Param content query string false "消息"
// @Param msg_id query int false "消息id"
// @Param page_num query int true "页码"
// @Param page_size query int true "页大小"
// @Param start_at query int64 false "开始时间"
// @Param end_at query int64 true "结束时间"
// @Success		200 {object} v1.Response{}
// @Router /msg/user/list [get]
func (h *Handler) GetUserMsgList(c *gin.Context, params v1.GetUserMsgListParams) {
	if params.MsgId == nil {
		params.MsgId = new(int)
		*params.MsgId = 0
	}

	if params.Type == nil {
		params.Type = new(int)
		*params.Type = 0
	}

	if params.UserId == nil {
		params.UserId = new(string)
		*params.UserId = ""
	}

	if params.Content == nil {
		params.Content = new(string)
		*params.Content = ""
	}

	if params.StartAt == nil {
		params.StartAt = new(int)
		*params.StartAt = 0
	}

	userID := c.Value(constants.UserID).(string)
	resp, err := h.svc.GetUserMessageList(c, userID, params)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}

// GetGroupMsgList
// @Summary 获取群聊消息
// @Description 获取群聊消息
// @Tags Msg
// @Accept  json
// @Produce  json
// @Param dialog_id query string true "对话id"
// @Param msg_id query int false "消息id"
// @Param user_id query string false "用户id"
// @Param type query string false "类型"
// @Param content query string false "消息"
// @Param page_num query int true "页码"
// @Param page_size query int true "页大小"
// @Success		200 {object} v1.Response{}
// @Router /msg/group/list [get]
func (h *Handler) GetGroupMsgList(c *gin.Context, params v1.GetGroupMsgListParams) {
	if params.MsgId == nil {
		params.MsgId = new(int)
		*params.MsgId = 0
	}

	if params.Type == nil {
		params.Type = new(int)
		*params.Type = 0
	}

	if params.UserId == nil {
		params.UserId = new(string)
		*params.UserId = ""
	}

	if params.Content == nil {
		params.Content = new(string)
		*params.Content = ""
	}

	userID := c.Value(constants.UserID).(string)
	resp, err := h.svc.GetGroupMessageList(c, userID, &params)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}

// GetUserDialogList
// 获取用户对话列表
// @Summary 获取用户对话列表
// @Description 获取用户对话列表
// @Tags Msg
// @Accept  json
// @Produce  json
// @Param page_num query int true "页码"
// @Param page_size query int true "页大小"
// @Success		200 {object} v1.Response{}
// @Router /msg/dialog/list [get]
func (h *Handler) GetUserDialogList(c *gin.Context, params v1.GetUserDialogListParams) {
	var page, pageSize = 1, 10
	if params.PageNum != 0 {
		page = params.PageNum
	}
	if params.PageSize != 0 {
		pageSize = params.PageSize
	}

	userID := c.Value(constants.UserID).(string)
	resp, err := h.svc.GetUserDialogList(c, userID, pageSize, page)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}

// EditUserMsg
// @Summary 编辑用户消息
// @Description 编辑用户消息
// @Tags Msg
// @Accept  json
// @Produce  json
// @Param id path int true "消息ID"
// @param request body v1.EditUserMsgRequest true "request"
// @Success 200 {object} v1.Response{}
// @Router /msg/user/{id} [put]
func (h *Handler) EditUserMsg(c *gin.Context, id int) {
	req := new(v1.EditUserMsgRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	driverID := c.Value(constants.DriverID).(string)
	_, err := h.svc.EditUserMsg(c, userID, driverID, uint32(id), req.Content)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "编辑成功", nil)
}

// EditGroupMsg
// @Summary 编辑群消息
// @Description 编辑群消息
// @Tags Msg
// @Accept  json
// @Produce  json
// @Param id path int true "消息ID"
// @param request body v1.EditGroupMsgRequest true "request"
// @Success 200 {object} v1.Response{}
// @Router /msg/group/{id} [put]
func (h *Handler) EditGroupMsg(c *gin.Context, id int) {
	req := new(v1.EditGroupMsgRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	driverID := c.Value(constants.DriverID).(string)
	_, err := h.svc.EditGroupMsg(c, userID, driverID, uint32(id), req.Content)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "编辑成功", nil)
}

// RecallUserMsg
// @Summary 撤回用户消息
// @Description 撤回用户消息
// @Tags Msg
// @Accept  json
// @Produce  json
// @Param id path int true "消息ID"
// @Success 200 {object} v1.Response{}
// @Router /msg/user/{id} [delete]
func (h *Handler) RecallUserMsg(c *gin.Context, id int) {

	userID := c.Value(constants.UserID).(string)
	driverID := c.Value(constants.DriverID).(string)
	_, err := h.svc.RecallUserMsg(c, userID, driverID, uint32(id))
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "撤回成功", nil)
}

// RecallGroupMsg
// @Summary 撤回群消息
// @Description 撤回群消息
// @Tags Msg
// @Accept  json
// @Produce  json
// @Param id path int true "消息ID"
// @Success 200 {object} v1.Response{}
// @Router /msg/group/{id} [delete]
func (h *Handler) RecallGroupMsg(c *gin.Context, id int) {
	userID := c.Value(constants.UserID).(string)
	driverID := c.Value(constants.DriverID).(string)
	_, err := h.svc.RecallGroupMsg(c, userID, driverID, uint32(id))
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "撤回成功", nil)
}

// LabelUserMsg
// @Summary 标注用户消息状态
// @Description 标注用户消息状态
// @Tags Msg
// @Accept  json
// @Produce  json
// @Param id path int true "消息ID"
// @param request body v1.LabelUserMessageRequest true "request"
// @Success 200 {object} v1.Response{}
// @Router /msg/user/{id}/label [post]
func (h *Handler) LabelUserMsg(c *gin.Context, id int) {
	//id := c.Param("id")
	req := new(v1.LabelUserMessageRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	driverID := c.Value(constants.DriverID).(string)
	err := h.svc.LabelUserMessage(c, userID, driverID, uint32(id), req.IsLabel)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "标注成功", nil)
}

// LabelGroupMsg
// @Summary 标注群聊消息状态
// @Description 标注群聊消息状态
// @Tags Msg
// @Accept  json
// @Produce  json
// @Param id path int true "消息ID"
// @param request body v1.LabelGroupMessageRequest true "request"
// @Success 200 {object} v1.Response{}
// @Router /msg/group/{id}/label [post]
func (h *Handler) LabelGroupMsg(c *gin.Context, id int) {
	req := new(v1.LabelGroupMessageRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	//if !model.IsValidLabelMsgType(req.IsLabel) {
	//	response.SetFail(c, "设置消息标注状态失败", nil)
	//	return
	//}

	userID := c.Value(constants.UserID).(string)
	driverID := c.Value(constants.DriverID).(string)
	_, err := h.svc.LabelGroupMessage(c, userID, driverID, uint32(id), req.IsLabel)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "标注成功", nil)
}

// GetUserLabelMsgList
// 获取私聊标注信息
// @Summary 获取私聊标注信息
// @Description 获取私聊标注信息
// @Tags Msg
// @Accept  json
// @Produce  json
// @Param dialog_id path string true "对话id"
// @Success		200 {object} v1.Response{data=v1.GetUserLabelMsgListResponse}
// @Router /msg/dialog/user/{dialog_id}/label [get]
func (h *Handler) GetUserLabelMsgList(c *gin.Context, id int) {

	userID := c.Value(constants.UserID).(string)
	resp, err := h.svc.GetUserLabelMsgList(c, userID, uint32(id))
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}

// GetGroupLabelMsgList
// 获取群聊标注信息
// @Summary 获取群聊标注信息
// @Description 获取群聊标注信息
// @Tags Msg
// @Accept  json
// @Produce  json
// @Param dialog_id path string true "对话id"
// @Success		200 {object} v1.Response{data=v1.GetGroupLabelMsgListResponse}
// @Router /msg/dialog/group/{dialog_id}/label [get]
func (h *Handler) GetGroupLabelMsgList(c *gin.Context, id int) {
	//var id = c.Query("dialog_id")
	//if id == "" {
	//	response.SetFail(c, "参数验证失败", nil)
	//	return
	//}
	//
	//dialogId, err := strconv.Atoi(id)
	//if err != nil {
	//	response.SetFail(c, "参数验证失败", nil)
	//	return
	//}

	userID := c.Value(constants.UserID).(string)
	resp, err := h.svc.GetGroupLabelMsgList(c, userID, uint32(id))
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}

// ReadUserMsgs
// @Summary 批量设置私聊消息状态为已读
// @Description 批量设置私聊消息状态为已读
// @Tags Msg
// @Accept  json
// @Produce  json
// @param request body v1.ReadUserMsgsRequest true "request"
// @Success 200 {object} v1.Response{}
// @Router /msg/user/read [put]
func (h *Handler) ReadUserMsgs(c *gin.Context) {
	req := new(v1.ReadUserMsgsRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	driverID := c.Value(constants.DriverID).(string)
	_, err := h.svc.ReadUserMsgs(c, userID, driverID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置成功", nil)
}

// GetAfterMsgs
// @Summary 获取指定对话落后消息
// @Description 获取指定对话落后消息
// @Tags Msg
// @Accept  json
// @Produce  json
// @param request body []v1.AfterMsg true "request"
// @Success 200 {object} v1.Response{}
// @Router /msg/dialog/after [post]
func (h *Handler) GetAfterMsgs(c *gin.Context) {
	req := new([]v1.AfterMsg)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	resp, err := h.svc.GetDialogAfterMsg(c, userID, *req)
	if err != nil {
		c.Error(err)
		return
	}
	response.SetSuccess(c, "获取成功", resp)
}

// GroupMessageRead
// @Summary 批量设置群聊消息为已读
// @Description 批量设置群聊消息为已读
// @Tags Msg
// @Accept json
// @Produce json
// @Param request body v1.GroupMessageReadRequest true "请求参数"
// @Success 200 {object} v1.Response{}
// @Router /msg/group/read [put]
func (h *Handler) GroupMessageRead(c *gin.Context) {
	req := new(v1.GroupMessageReadRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	driverID := c.Value(constants.DriverID).(string)
	err := h.svc.SetGroupMessagesRead(c, userID, driverID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置成功", nil)
}

// GetGroupMessageReaders
// @Summary 获取消息已读人员
// @Description 获取消息已读人员
// @Tags Msg
// @Accept json
// @Produce json
// @Param id path uint32 true "消息ID"
// @Param dialog_id query uint32 true "对话ID"
// @Param group_id query uint32 true "群聊ID"
// @Success 200 {object} v1.Response{data=[]v1.GetGroupMessageReadersResponse{}}
// @Router /msg/group/{id}/read [get]
func (h *Handler) GetGroupMessageReaders(c *gin.Context, id int, params v1.GetGroupMessageReadersParams) {
	msgID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	dialogID, err := strconv.ParseUint(c.Query("dialog_id"), 10, 32)
	if err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	groupID, err := strconv.ParseUint(c.Query("group_id"), 10, 32)
	if err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	// 执行获取消息已读人员的逻辑
	resp, err := h.svc.GetGroupMessageReadersResponse(c, userID, uint32(msgID), uint32(dialogID), uint32(groupID))
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}
