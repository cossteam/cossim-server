package http

import (
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/api/http/model"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/http"
	pkghttp "github.com/cossim/coss-server/pkg/http"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
)

// 设置自己在群聊内的名称
// @Summary 设置自己在群聊内的名称
// @Description 设置自己在群聊内的名称
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @param request body model.SetGroupUserRemarkRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /relation/group/remark/set [post]
func (h *Handler) setGroupUserRemark(c *gin.Context) {
	req := new(model.SetGroupUserRemarkRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}
	thisID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	err = h.svc.SetGroupUserRemark(c, thisID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置成功", nil)
}

// 设置群聊阅后即焚消息销毁时间
// @Summary 设置群聊阅后即焚消息销毁时间
// @Description 设置群聊阅后即焚消息销毁时间
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @param request body model.SetGroupOpenBurnAfterReadingTimeOutRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /relation/group/burn/timeout/set [post]
func (h *Handler) setGroupOpenBurnAfterReadingTimeOut(c *gin.Context) {
	req := new(model.SetGroupOpenBurnAfterReadingTimeOutRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}
	thisID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	err = h.svc.SetGroupOpenBurnAfterReadingTimeOut(c, thisID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置成功", nil)
}

// 设置群聊公告为已读
// @Summary 设置群聊公告为已读
// @Description 设置群聊公告为已读
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @param request body model.ReadGroupAnnouncementRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /relation/group/announcement/read [post]
func (h *Handler) readGroupAnnouncement(c *gin.Context) {
	req := new(model.ReadGroupAnnouncementRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}
	thisID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	_, err = h.svc.ReadGroupAnnouncement(c, thisID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置成功", nil)
}

// 获取已读公告用户
// @Summary 获取已读公告用户
// @Description 获取已读公告用户
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @param id query string true "id"
// @param group_id query string true "group_id"
// @Success 200 {object} model.Response{}
// @Router /relation/group/announcement/read/list [get]
func (h *Handler) getReadGroupAnnouncementList(c *gin.Context) {
	var id = c.Query("id")
	var gid = c.Query("group_id")

	if id == "" || gid == "" {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	aId, err := strconv.Atoi(id)
	if err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}
	groupId, err := strconv.Atoi(gid)
	if err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	thisID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	resp, err := h.svc.GetReadGroupAnnouncementUserList(c, thisID, uint32(aId), uint32(groupId))
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}

// @Summary 创建群公告(管理员)
// @Description 创建群公告(管理员)
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @param request body model.CreateGroupAnnouncementRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /relation/group/admin/announcement [post]
func (h *Handler) createGroupAnnouncement(c *gin.Context) {
	req := new(model.CreateGroupAnnouncementRequest)
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

	resp, err := h.svc.CreateGroupAnnouncement(c, thisId, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "创建成功", resp)
}

// 获取群公告
// @Summary 获取群公告
// @Description 获取群公告
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @Param group_id query string true "群聊id"
// @Success		200 {object} model.Response{}
// @Router /relation/group/announcement/list [get]
func (h *Handler) getGroupAnnouncementList(c *gin.Context) {
	var id = c.Query("group_id")
	if id == "" {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	groupId, err := strconv.Atoi(id)
	if err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	resp, err := h.svc.GetGroupAnnouncementList(c, thisId, uint32(groupId))
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}

// 查询群公告详情
// @Summary 查询群公告详情
// @Description 查询群公告详情
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @Param id query string true "群公告id"
// @Param group_id query string true "群id"
// @Success		200 {object} model.Response{}
// @Router /relation/group/announcement/detail [get]
func (h *Handler) getGroupAnnouncementDetail(c *gin.Context) {
	var id = c.Query("id")
	var gid = c.Query("group_id")

	if id == "" || gid == "" {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	aId, err := strconv.Atoi(id)
	if err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}
	groupId, err := strconv.Atoi(gid)
	if err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	resp, err := h.svc.GetGroupAnnouncementDetail(c, thisId, uint32(aId), uint32(groupId))
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取成功", resp)
}

// 更新群公告
// @Summary 更新群公告
// @Description 更新群公告
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @param request body model.UpdateGroupAnnouncementRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /relation/group/admin/announcement/update [post]
func (h *Handler) updateGroupAnnouncement(c *gin.Context) {
	req := new(model.UpdateGroupAnnouncementRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	thisID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	resp, err := h.svc.UpdateGroupAnnouncement(c, thisID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "更新成功", resp)
}

// 删除群公告
// @Summary 删除群公告
// @Description 删除群公告
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @param request body model.DeleteGroupAnnouncementRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /relation/group/admin/announcement/delete [post]
func (h *Handler) deleteGroupAnnouncement(c *gin.Context) {
	req := new(model.DeleteGroupAnnouncementRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}
	thisID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	resp, err := h.svc.DeleteGroupAnnouncement(c, thisID, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "删除成功", resp)
}

// @Summary 是否打开群聊阅后即焚消息(action: 0:关闭, 1:打开)
// @Description 是否打开群聊阅后即焚消息(action: 0:关闭, 1:打开)
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @param request body model.OpenGroupBurnAfterReadingRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /relation/group/burn/open [post]
func (h *Handler) openGroupBurnAfterReading(c *gin.Context) {
	req := new(model.OpenGroupBurnAfterReadingRequest)
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

	if !model.IsValidOpenBurnAfterReadingType(req.Action) {
		response.SetFail(c, "参数验证失败", nil)
		return
	}
	_, err = h.svc.SetGroupBurnAfterReading(c, thisId, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "操作成功", nil)
}

// @Summary 设置群聊静默通知
// @Description 设置群聊静默通知
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @param request body model.SetGroupSilentNotificationRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /relation/group/silent [post]
func (h *Handler) setGroupSilentNotification(c *gin.Context) {
	req := new(model.SetGroupSilentNotificationRequest)
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

	if !model.IsValidSilentNotificationType(req.IsSilent) {
		response.SetFail(c, "设置消息标注状态失败", nil)
		return
	}

	_, err = h.svc.SetGroupSilentNotification(c, req.GroupId, thisId, req.IsSilent)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置成功", gin.H{"group_id": req.GroupId})
}

// @Summary 群聊成员列表
// @Description 群聊成员列表
// @Tags GroupRelation
// @Param group_id query integer true "群聊ID"
// @Produce  json
// @Success		200 {object} model.Response{}
// @Router /relation/group/member [get]
func (h *Handler) getGroupMember(c *gin.Context) {
	// 从请求中获取群聊ID
	groupID := c.Query("group_id")
	if groupID == "" {
		response.SetFail(c, "群聊ID不能为空", nil)
		return
	}

	gid, err := strconv.ParseUint(groupID, 10, 32)
	if err != nil {
		response.SetFail(c, "群聊ID格式错误", nil)
		return
	}

	userID, err := http.ParseTokenReUid(c)
	if err != nil {
		h.logger.Error("token解析失败", zap.Error(err))
		response.SetFail(c, "token解析失败", nil)
		return
	}

	resp, err := h.svc.GetGroupMember(c, uint32(gid), userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取群聊成员成功", resp)
}

// groupRequestList 获取群聊申请列表
// @Summary 获取群聊申请列表
// @Description 获取用户的群聊申请列表
// @Tags GroupRelation
// @Accept json
// @Produce json
// @Security Bearer
// @Param Authorization header string true "Bearer JWT"
// @Success		200 {object} model.Response{data=model.GroupRequestListResponse} "status (0=等待, 1=已通过, 2=已拒绝, 3=邀请发送者, 4=邀请接收者)"
// @Router /relation/group/request_list [get]
func (h *Handler) groupRequestList(c *gin.Context) {
	userID, err := http.ParseTokenReUid(c)
	if err != nil {
		h.logger.Error("token解析失败", zap.Error(err))
		response.SetFail(c, "token解析失败", nil)
		return
	}

	resp, err := h.svc.GroupRequestList(c, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取群聊申请列表成功", resp)
}

// @Summary 邀请加入群聊
// @Description 邀请加入群聊
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @param request body model.InviteGroupRequest true "request"
// @Success		200 {object} model.Response{}
// @Router /relation/group/invite [post]
func (h *Handler) inviteGroup(c *gin.Context) {
	req := new(model.InviteGroupRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	uid, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	if err = h.svc.InviteGroup(c, uid, req); err != nil {
		h.logger.Error("邀请好友加入群聊失败", zap.Error(err))
		response.SetFail(c, code.Cause(err).Message(), nil)
		return
	}

	response.SetSuccess(c, "邀请好友加入群聊成功", nil)
}

// @Summary 加入群聊
// @Description 加入群聊
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @param request body model.JoinGroupRequest true "request"
// @Success		200 {object} model.Response{}
// @Router /relation/group/join [post]
func (h *Handler) joinGroup(c *gin.Context) {
	req := new(model.JoinGroupRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	uid, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	_, err = h.svc.JoinGroup(c, uid, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "发送加入群聊请求成功", nil)
}

// @Summary 用户管理加入群聊
// @Description 用户管理加入群聊 action (0=拒绝, 1=同意)
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @param request body model.ManageJoinGroupRequest true "Action (0: rejected, 1: joined)"
// @Success		200 {object} model.Response{}
// @Router /relation/group/manage_join [post]
func (h *Handler) manageJoinGroup(c *gin.Context) {
	req := new(model.ManageJoinGroupRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	if err := req.Validator(); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, err.Error(), nil)
		return
	}

	userID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	var status relationgrpcv1.GroupRequestStatus
	var msg string
	if req.Action == model.ActionAccepted {
		status = relationgrpcv1.GroupRequestStatus_Accepted
		msg = "同意加入群聊"
	} else {
		status = relationgrpcv1.GroupRequestStatus_Rejected
		msg = "拒绝加入群聊"
	}

	if err = h.svc.ManageJoinGroup(c, req.GroupID, req.ID, userID, status); err != nil {
		h.logger.Error("用户管理群聊申请", zap.Error(err))
		response.SetFail(c, code.Cause(err).Message(), nil)
		return
	}

	response.Success(c, msg+"成功", nil)
}

// @Summary 管理员管理加入群聊
// @Description 管理员管理加入群聊 action (0=拒绝, 1=同意)
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @param request body model.ManageJoinGroupRequest true "Action (0: rejected, 1: joined)"
// @Success		200 {object} model.Response{}
// @Router /relation/group/admin/manage/join [post]
func (h *Handler) adminManageJoinGroup(c *gin.Context) {
	req := new(model.ManageJoinGroupRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	if err := req.Validator(); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, err.Error(), nil)
		return
	}

	adminID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	var status relationgrpcv1.GroupRequestStatus
	var msg string
	if req.Action == model.ActionAccepted {
		status = relationgrpcv1.GroupRequestStatus_Accepted
		msg = "同意加入群聊"
	} else {
		status = relationgrpcv1.GroupRequestStatus_Rejected
		msg = "拒绝加入群聊"
	}

	if err = h.svc.AdminManageJoinGroup(c, req.ID, req.GroupID, adminID, status); err != nil {
		h.logger.Error("管理员管理群聊申请", zap.Error(err))
		response.SetFail(c, code.Cause(err).Message(), nil)
		return
	}

	response.Success(c, msg+"成功", nil)
}

// @Summary 将用户从群聊移除
// @Description 将用户从群聊移除
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @param request body model.RemoveUserFromGroupRequest true "request"
// @Success		200 {object} model.Response{}
// @Router /relation/group/admin/manage/remove [post]
func (h *Handler) removeUserFromGroup(c *gin.Context) {
	req := new(model.RemoveUserFromGroupRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	for _, d := range req.Member {
		if userID == d {
			response.SetFail(c, "不能将自己从群聊中移除", nil)
			return
		}
	}

	if err = h.svc.RemoveUserFromGroup(c, req.GroupID, userID, req.Member); err != nil {
		h.logger.Error("RemoveUserFromGroup Failed", zap.Error(err))
		response.SetFail(c, err.Error(), nil)
		return
	}

	response.SetSuccess(c, "移出群聊成功", nil)
}

// @Summary 退出群聊
// @Description 退出群聊
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @param request body model.QuitGroupRequest true "request"
// @Success		200 {object} model.Response{}
// @Router /relation/group/quit [post]
func (h *Handler) quitGroup(c *gin.Context) {
	req := new(model.QuitGroupRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	if err = h.svc.QuitGroup(c, req.GroupID, userID); err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	response.SetSuccess(c, "退出群聊成功", nil)
}

// @Summary 删除群聊申请记录
// @Description 删除群聊申请记录
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @param request body model.DeleteRecordRequest true "request"
// @Success		200 {object} model.Response{}
// @Router /relation/group/delete_request_record [post]
func (h *Handler) deleteGroupRequestRecord(c *gin.Context) {
	req := new(model.DeleteRecordRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID, err := http.ParseTokenReUid(c)
	if err != nil {
		h.logger.Error("token解析失败", zap.Error(err))
		response.SetFail(c, "token解析失败", nil)
		return
	}

	if err = h.svc.DeleteGroupFriendRecord(c, userID, req.ID); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "删除群聊申请记录成功", nil)
}
