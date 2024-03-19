package http

import (
	"fmt"
	"net/http"
	"strconv"
	"sync"

	"github.com/cossim/coss-server/internal/msg/api/http/model"
	"github.com/cossim/coss-server/pkg/constants"
	pkghttp "github.com/cossim/coss-server/pkg/http"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

var (
	wsRid    int64 = 0            //全局客户端id
	wsMutex        = sync.Mutex{} //锁
	upgrader       = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	Pool = make(map[string]map[string][]*client)
)

type client struct {
	Conn       *websocket.Conn
	Uid        string //客户端所有者
	Rid        int64  //客户端id
	ClientType string //客户端类型
	queue      *amqp091.Channel
}

// @Summary websocket请求
// @Description websocket请求
// @Router /msg/ws [get]
func (h *Handler) ws(c *gin.Context) {
	var uid string
	var driverId string
	token := c.Query("token")
	//判断设备类型
	deviceType := c.Request.Header.Get("X-Device-Type")
	deviceType = string(constants.DetermineClientType(constants.DriverType(deviceType)))

	if token == "" {
		//id, err := pkghttp.ParseTokenReUid(c)
		//if err != nil {
		//	h.logger.Error("token解析失败", zap.Error(err))
		//	return
		//}
		//uid = id
		return
	} else {
		_, c2, err := utils.ParseToken(token)
		if err != nil {
			h.logger.Error("token解析失败", zap.Error(err))
			return
		}
		uid = c2.UserId
		driverId = c2.DriverId
	}
	if uid == "" {
		return
	}

	//升级http请求为websocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.Error(err)
		return
	}
	h.svc.Ws(conn, uid, driverId, deviceType, token)

}

// @Summary 发送私聊消息
// @Description 发送私聊消息
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
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	driverId, err := pkghttp.ParseTokenReDriverId(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	resp, err := h.svc.SendUserMsg(c, thisId, driverId, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "发送成功", resp)
}

// @Summary 发送群聊消息
// @Description 发送群聊消息
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
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		fmt.Println(err)
		return
	}
	if req.GroupId == 0 {
		response.SetFail(c, "群聊id不正确", nil)
		return
	}

	driverId, err := pkghttp.ParseTokenReDriverId(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	resp, err := h.svc.SendGroupMsg(c, thisId, driverId, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "发送成功", gin.H{"msg_id": resp})
}

// @Summary 获取私聊消息
// @Description 获取私聊消息
// @Accept  json
// @Produce  json
// @Param user_id query string true "用户id"
// @Param type query string false "类型"
// @Param content query string false "消息"
// @Param page_num query int true "页码"
// @Param page_size query int true "页大小"
// @Success		200 {object} model.Response{}
// @Router /msg/list/user [get]
func (h *Handler) getUserMsgList(c *gin.Context) {
	var num = c.Query("page_num")
	var size = c.Query("page_size")
	var id = c.Query("user_id")
	var msgType = c.Query("type")
	var content = c.Query("content")

	if num == "" || size == "" || id == "" {
		response.SetFail(c, "参数错误", nil)
		return
	}
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}
	pageNum, _ := strconv.Atoi(num)
	pageSize, _ := strconv.Atoi(size)
	mt, _ := strconv.Atoi(msgType)
	if pageNum == 0 || pageSize == 0 {
		response.SetFail(c, "参数错误", nil)
		return
	}

	var msgListRequest = &model.MsgListRequest{
		UserId:   id,
		Type:     int32(mt),
		Content:  content,
		PageNum:  pageNum,
		PageSize: pageSize,
	}

	resp, err := h.svc.GetUserMessageList(c, thisId, msgListRequest)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}

// @Summary 获取群聊消息
// @Description 获取群聊消息
// @Accept  json
// @Produce  json
// @Param group_id query string true "群聊id"
// @Param user_id query string false "用户id"
// @Param type query string false "类型"
// @Param content query string false "消息"
// @Param page_num query int true "页码"
// @Param page_size query int true "页大小"
// @Success		200 {object} model.Response{}
// @Router /msg/list/group [get]
func (h *Handler) getGroupMsgList(c *gin.Context) {
	var gid = c.Query("group_id")
	var num = c.Query("page_num")
	var size = c.Query("page_size")
	var id = c.Query("user_id")
	var msgType = c.Query("type")
	var content = c.Query("content")

	if num == "" || size == "" || gid == "" {
		response.SetFail(c, "参数错误", nil)
		return
	}
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	gidInt, _ := strconv.Atoi(gid)
	pageNum, _ := strconv.Atoi(num)
	pageSize, _ := strconv.Atoi(size)
	mt, _ := strconv.Atoi(msgType)
	if pageNum == 0 || pageSize == 0 {
		response.SetFail(c, "参数错误", nil)
		return
	}

	var msgListRequest = &model.GroupMsgListRequest{
		GroupId:  uint32(gidInt),
		UserId:   id,
		Type:     int32(mt),
		Content:  content,
		PageNum:  pageNum,
		PageSize: pageSize,
	}

	resp, err := h.svc.GetGroupMessageList(c, thisId, msgListRequest)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}

// 获取用户对话列表
// @Summary 获取用户对话列表
// @Description 获取用户对话列表
// @Accept  json
// @Produce  json
// @Success		200 {object} model.Response{}
// @Router /msg/dialog/list [get]
func (h *Handler) getUserDialogList(c *gin.Context) {
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	resp, err := h.svc.GetUserDialogList(c, thisId)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}

// @Summary 编辑用户消息
// @Description 编辑用户消息
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

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	driverId, err := pkghttp.ParseTokenReDriverId(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	resp, err := h.svc.EditUserMsg(c, thisId, driverId, req.MsgId, req.Content)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "编辑成功", gin.H{"msg_id": resp})
}

// @Summary 编辑群消息
// @Description 编辑群消息
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

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	driverId, err := pkghttp.ParseTokenReDriverId(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	resp, err := h.svc.EditGroupMsg(c, thisId, driverId, req.MsgId, req.Content)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "编辑成功", gin.H{"msg_id": resp})
}

// @Summary 撤回用户消息
// @Description 撤回用户消息
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

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	driverId, err := pkghttp.ParseTokenReDriverId(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	resp, err := h.svc.RecallUserMsg(c, thisId, driverId, req.MsgId)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "撤回成功", gin.H{"msg_id": resp})
}

// @Summary 撤回群消息
// @Description 撤回群消息
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

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	driverId, err := pkghttp.ParseTokenReDriverId(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	resp, err := h.svc.RecallGroupMsg(c, thisId, driverId, req.MsgId)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "撤回成功", gin.H{"msg_id": resp})
}

// @Summary 标注用户消息状态
// @Description 标注用户消息状态
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

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	if !model.IsValidLabelMsgType(req.IsLabel) {
		response.SetFail(c, "设置消息标注状态失败", nil)
		return
	}

	driverId, err := pkghttp.ParseTokenReDriverId(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	_, err = h.svc.LabelUserMessage(c, thisId, driverId, req.MsgID, req.IsLabel)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "用户消息标注成功", nil)
}

// @Summary 标注群聊消息状态
// @Description 标注群聊消息状态
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

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	if !model.IsValidLabelMsgType(req.IsLabel) {
		response.SetFail(c, "设置消息标注状态失败", nil)
		return
	}

	driverId, err := pkghttp.ParseTokenReDriverId(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	_, err = h.svc.LabelGroupMessage(c, thisId, driverId, req.MsgID, req.IsLabel)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "群聊消息标注成功", nil)
}

// 获取私聊标注信息
// @Summary 获取私聊标注信息
// @Description 获取私聊标注信息
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

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	resp, err := h.svc.GetUserLabelMsgList(c, thisId, uint32(dialogId))
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}

// 获取群聊标注信息
// @Summary 获取群聊标注信息
// @Description 获取群聊标注信息
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

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	resp, err := h.svc.GetGroupLabelMsgList(c, thisId, uint32(dialogId))
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}

// @Summary 批量设置私聊消息状态为已读
// @Description 批量设置私聊消息状态为已读
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

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	driverId, err := pkghttp.ParseTokenReDriverId(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	_, err = h.svc.ReadUserMsgs(c, thisId, driverId, req.DialogId, req.MsgIds)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置成功", nil)
}

// @Summary 获取指定对话落后消息
// @Description 获取指定对话落后消息
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

	_, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}
	resp, err := h.svc.GetDialogAfterMsg(c, *req)
	if err != nil {
		c.Error(err)
	}
	response.SetSuccess(c, "获取成功", resp)
}

// @Summary 批量设置群聊消息为已读
// @Description 批量设置群聊消息为已读
// @Accept json
// @Produce json
// @Param request body model.GroupMessageReadRequest true "请求参数"
// @Success 200 {object} model.Response{}
// @Router /msg/group/read/set [post]
func (h *Handler) setGroupMessagesRead(c *gin.Context) {
	req := new(model.GroupMessageReadRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	driverId, err := pkghttp.ParseTokenReDriverId(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	_, err = h.svc.SetGroupMessagesRead(c, thisId, driverId, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置成功", nil)
}

// @Summary 获取消息已读人员
// @Description 获取消息已读人员
// @Accept json
// @Produce json
// @Param msg_id query uint32 true "消息ID"
// @Param dialog_id query uint32 true "对话ID"
// @Param group_id query uint32 true "群聊ID"
// @Success 200 {object} model.Response{data=[]model.GetGroupMessageReadersResponse{}}
// @Router /msg/group/read/get [get]
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

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	// 执行获取消息已读人员的逻辑
	resp, err := h.svc.GetGroupMessageReadersResponse(c, thisId, uint32(msgID), uint32(dialogID), uint32(groupID))
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}
