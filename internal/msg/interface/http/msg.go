package http

import (
	"github.com/cossim/coss-server/internal/msg/api/http/model"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
)

// @Summary 发送私聊消息
// @Description 发送私聊消息
// @Tags Msg
// @Accept  json
// @Produce  json
// @param request body model.SendUserMsgRequest true "request"
// @Success		200 {object} model.Response{}
// @Router /msg/send/user [post]
func (h *Handler) sendUserMsg(c *gin.Context) {
	req := new(model.SendUserMsgRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	if !model.IsValidMessageType(req.Type) {
		response.SetFail(c, "消息类型错误", nil)
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

// @Summary 发送群聊消息
// @Description 发送群聊消息
// @Tags Msg
// @Accept  json
// @Produce  json
// @param request body model.SendGroupMsgRequest true "request"
// @Success		200 {object} model.Response{}
// @Router /msg/send/group [post]
func (h *Handler) sendGroupMsg(c *gin.Context) {
	req := new(model.SendGroupMsgRequest)
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
// @Success		200 {object} model.Response{}
// @Router /msg/list/user [get]
func (h *Handler) getUserMsgList(c *gin.Context) {
	var num = c.Query("page_num")
	var size = c.Query("page_size")
	var id = c.Query("dialog_id")
	var msgType = c.Query("type")
	var content = c.Query("content")
	var msgId = c.Query("msg_id")
	var uid = c.Query("user_id")

	if num == "" || size == "" || id == "" {
		response.SetFail(c, "参数错误", nil)
		return
	}

	mId, _ := strconv.Atoi(msgId)
	dialogId, _ := strconv.Atoi(id)
	pageNum, _ := strconv.Atoi(num)
	pageSize, _ := strconv.Atoi(size)
	mt, _ := strconv.Atoi(msgType)
	if pageNum == 0 || pageSize == 0 {
		response.SetFail(c, "参数错误", nil)
		return
	}

	var msgListRequest = &model.MsgListRequest{
		UserId:   uid,
		DialogId: uint32(dialogId),
		Type:     int32(mt),
		Content:  content,
		PageNum:  pageNum,
		PageSize: pageSize,
		MsgId:    uint32(mId),
	}

	userID := c.Value(constants.UserID).(string)
	resp, err := h.svc.GetUserMessageList(c, userID, msgListRequest)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}

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
// @Success		200 {object} model.Response{}
// @Router /msg/list/group [get]
func (h *Handler) getGroupMsgList(c *gin.Context) {
	var dialogId = c.Query("dialog_id")
	var num = c.Query("page_num")
	var size = c.Query("page_size")
	var id = c.Query("user_id")
	var msgType = c.Query("type")
	var content = c.Query("content")
	var msgId = c.Query("msg_id")

	if num == "" || size == "" || dialogId == "" {
		response.SetFail(c, "参数错误", nil)
		return
	}

	did, _ := strconv.Atoi(dialogId)
	pageNum, _ := strconv.Atoi(num)
	pageSize, _ := strconv.Atoi(size)
	mId, _ := strconv.Atoi(msgId)
	mt, _ := strconv.Atoi(msgType)
	if pageNum == 0 || pageSize == 0 {
		response.SetFail(c, "参数错误", nil)
		return
	}

	var msgListRequest = &model.MsgListRequest{
		DialogId: uint32(did),
		UserId:   id,
		Type:     int32(mt),
		Content:  content,
		PageNum:  pageNum,
		PageSize: pageSize,
		MsgId:    uint32(mId),
	}

	userID := c.Value(constants.UserID).(string)
	resp, err := h.svc.GetGroupMessageList(c, userID, msgListRequest)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}

// 获取用户对话列表
// @Summary 获取用户对话列表
// @Description 获取用户对话列表
// @Tags Msg
// @Accept  json
// @Produce  json
// @Success		200 {object} model.Response{}
// @Router /msg/dialog/list [get]
func (h *Handler) getUserDialogList(c *gin.Context) {
	userID := c.Value(constants.UserID).(string)
	resp, err := h.svc.GetUserDialogList(c, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}

// @Summary 编辑用户消息
// @Description 编辑用户消息
// @Tags Msg
// @Accept  json
// @Produce  json
// @param request body model.EditUserMsgRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /msg/edit/user [post]
func (h *Handler) editUserMsg(c *gin.Context) {
	req := new(model.EditUserMsgRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	driverID := c.Value(constants.DriverID).(string)
	resp, err := h.svc.EditUserMsg(c, userID, driverID, req.MsgId, req.Content)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "编辑成功", gin.H{"msg_id": resp})
}

// @Summary 编辑群消息
// @Description 编辑群消息
// @Tags Msg
// @Accept  json
// @Produce  json
// @param request body model.EditGroupMsgRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /msg/edit/group [post]
func (h *Handler) editGroupMsg(c *gin.Context) {
	req := new(model.EditGroupMsgRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	driverID := c.Value(constants.DriverID).(string)
	resp, err := h.svc.EditGroupMsg(c, userID, driverID, req.MsgId, req.Content)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "编辑成功", gin.H{"msg_id": resp})
}

// @Summary 撤回用户消息
// @Description 撤回用户消息
// @Tags Msg
// @Accept  json
// @Produce  json
// @param request body model.RecallUserMsgRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /msg/recall/user [post]
func (h *Handler) recallUserMsg(c *gin.Context) {
	req := new(model.RecallUserMsgRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	driverID := c.Value(constants.DriverID).(string)
	resp, err := h.svc.RecallUserMsg(c, userID, driverID, req.MsgId)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "撤回成功", gin.H{"msg_id": resp})
}

// @Summary 撤回群消息
// @Description 撤回群消息
// @Tags Msg
// @Accept  json
// @Produce  json
// @param request body model.RecallGroupMsgRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /msg/recall/group [post]
func (h *Handler) recallGroupMsg(c *gin.Context) {
	req := new(model.RecallGroupMsgRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	driverID := c.Value(constants.DriverID).(string)
	resp, err := h.svc.RecallGroupMsg(c, userID, driverID, req.MsgId)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "撤回成功", resp)
}

// @Summary 标注用户消息状态
// @Description 标注用户消息状态
// @Tags Msg
// @Accept  json
// @Produce  json
// @param request body model.LabelUserMessageRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /msg/label/user [post]
func (h *Handler) labelUserMessage(c *gin.Context) {
	req := new(model.LabelUserMessageRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	if !model.IsValidLabelMsgType(req.IsLabel) {
		response.SetFail(c, "设置消息标注状态失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	driverID := c.Value(constants.DriverID).(string)
	_, err := h.svc.LabelUserMessage(c, userID, driverID, req.MsgID, req.IsLabel)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "用户消息标注成功", nil)
}

// @Summary 标注群聊消息状态
// @Description 标注群聊消息状态
// @Tags Msg
// @Accept  json
// @Produce  json
// @param request body model.LabelGroupMessageRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /msg/label/group [post]
func (h *Handler) labelGroupMessage(c *gin.Context) {
	req := new(model.LabelGroupMessageRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	if !model.IsValidLabelMsgType(req.IsLabel) {
		response.SetFail(c, "设置消息标注状态失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	driverID := c.Value(constants.DriverID).(string)
	_, err := h.svc.LabelGroupMessage(c, userID, driverID, req.MsgID, req.IsLabel)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "群聊消息标注成功", nil)
}

// 获取私聊标注信息
// @Summary 获取私聊标注信息
// @Description 获取私聊标注信息
// @Tags Msg
// @Accept  json
// @Produce  json
// @Param dialog_id query string true "对话id"
// @Success		200 {object} model.Response{}
// @Router /msg/label/user [get]
func (h *Handler) getUserLabelMsgList(c *gin.Context) {
	var id = c.Query("dialog_id")

	if id == "" {
		response.SetFail(c, "参数验证失败", nil)
		return
	}
	dialogId, err := strconv.Atoi(id)
	if err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	resp, err := h.svc.GetUserLabelMsgList(c, userID, uint32(dialogId))
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}

// 获取群聊标注信息
// @Summary 获取群聊标注信息
// @Description 获取群聊标注信息
// @Tags Msg
// @Accept  json
// @Produce  json
// @Param dialog_id query string true "对话id"
// @Success		200 {object} model.Response{}
// @Router /msg/label/group [get]
func (h *Handler) getGroupLabelMsgList(c *gin.Context) {
	var id = c.Query("dialog_id")
	if id == "" {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	dialogId, err := strconv.Atoi(id)
	if err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	resp, err := h.svc.GetGroupLabelMsgList(c, userID, uint32(dialogId))
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}

// @Summary 批量设置私聊消息状态为已读
// @Description 批量设置私聊消息状态为已读
// @Tags Msg
// @Accept  json
// @Produce  json
// @param request body model.ReadUserMsgsRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /msg/read/user [post]
func (h *Handler) readUserMsgs(c *gin.Context) {
	req := new(model.ReadUserMsgsRequest)
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

// @Summary 获取指定对话落后消息
// @Description 获取指定对话落后消息
// @Tags Msg
// @Accept  json
// @Produce  json
// @param request body []model.AfterMsg true "request"
// @Success 200 {object} model.Response{}
// @Router /msg/after/get [post]
func (h *Handler) getDialogAfterMsg(c *gin.Context) {
	req := new([]model.AfterMsg)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	resp, err := h.svc.GetDialogAfterMsg(c, userID, *req)
	if err != nil {
		c.Error(err)
	}
	response.SetSuccess(c, "获取成功", resp)
}

// @Summary 批量设置群聊消息为已读
// @Description 批量设置群聊消息为已读
// @Tags Msg
// @Accept json
// @Produce json
// @Param request body model.GroupMessageReadRequest true "请求参数"
// @Success 200 {object} model.Response{}
// @Router /msg/read/group [post]
func (h *Handler) setGroupMessagesRead(c *gin.Context) {
	req := new(model.GroupMessageReadRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID := c.Value(constants.UserID).(string)
	driverID := c.Value(constants.DriverID).(string)
	_, err := h.svc.SetGroupMessagesRead(c, userID, driverID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置成功", nil)
}

// @Summary 获取消息已读人员
// @Description 获取消息已读人员
// @Tags Msg
// @Accept json
// @Produce json
// @Param msg_id query uint32 true "消息ID"
// @Param dialog_id query uint32 true "对话ID"
// @Param group_id query uint32 true "群聊ID"
// @Success 200 {object} model.Response{data=[]model.GetGroupMessageReadersResponse{}}
// @Router /msg/read/group [get]
func (h *Handler) getGroupMessageReaders(c *gin.Context) {
	msgID, err := strconv.ParseUint(c.Query("msg_id"), 10, 32)
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
