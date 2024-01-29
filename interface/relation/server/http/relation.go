package http

import (
	"github.com/cossim/coss-server/interface/relation/api/model"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/http"
	pkghttp "github.com/cossim/coss-server/pkg/http"
	"strconv"

	"github.com/cossim/coss-server/pkg/http/response"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary 黑名单
// @Description 黑名单
// @Tags UserRelation
// @Produce  json
// @Success		200 {object} model.Response{}
// @Router /relation/user/blacklist [get]
func blackList(c *gin.Context) {
	userID, err := http.ParseTokenReUid(c)
	if err != nil {
		logger.Error("token解析失败", zap.Error(err))
		response.SetFail(c, "token解析失败", nil)
		return
	}

	resp, err := svc.BlackList(c, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "获取黑名单列表成功", resp)
}

// @Summary 好友列表
// @Description 好友列表
// @Tags UserRelation
// @Produce  json
// @Success		200 {object} model.Response{}
// @Router /relation/user/friend_list [get]
func friendList(c *gin.Context) {
	userID, err := http.ParseTokenReUid(c)
	if err != nil {
		logger.Error("token解析失败", zap.Error(err))
		response.SetFail(c, "token解析失败", nil)
		return
	}

	resp, err := svc.FriendList(c, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "获取好友列表成功", resp)
}

// @Summary 群聊列表
// @Description 群聊列表
// @Tags GroupRelation
// @Produce  json
// @Success		200 {object} model.Response{data=[]usersorter.CustomGroupData} "status 0:正常状态；1:被封禁状态；2:被删除状态"
// @Router /relation/group/list [get]
func getUserGroupList(c *gin.Context) {
	userID, err := http.ParseTokenReUid(c)
	if err != nil {
		logger.Error("token解析失败", zap.Error(err))
		response.SetFail(c, "token解析失败", nil)
		return
	}

	resp, err := svc.GetUserGroupList(c, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "获取用户群聊列表成功", resp)
}

// @Summary 好友申请列表
// @Description 好友申请列表
// @Tags UserRelation
// @Produce  json
// @Success		200 {object} model.Response{data=[]model.UserRequestListResponse} "UserStatus 申请状态 (0=申请中, 1=待通过, 2=已添加, 3=被拒绝, 4=已删除, 5=已拒绝)"
// @Router /relation/user/request_list [get]
func userRequestList(c *gin.Context) {
	userID, err := http.ParseTokenReUid(c)
	if err != nil {
		logger.Error("token解析失败", zap.Error(err))
		response.SetFail(c, "token解析失败", nil)
		return
	}

	resp, err := svc.UserRequestList(c, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "获取好友申请列表成功", resp)
}

// @Summary 删除黑名单
// @Description 删除黑名单
// @Tags UserRelation
// @Accept  json
// @Produce  json
// @param request body model.DeleteBlacklistRequest true "request"
// @Success		200 {object} model.Response{}
// @Router /relation/delete_blacklist [post]
func deleteBlacklist(c *gin.Context) {
	req := new(model.DeleteBlacklistRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID, err := http.ParseTokenReUid(c)
	if err != nil {
		logger.Error("token解析失败", zap.Error(err))
		response.SetFail(c, "token解析失败", nil)
		return
	}

	_, err = svc.DeleteBlacklist(c, userID, req.UserID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "删除黑名单成功", nil)
}

// @Summary 添加黑名单
// @Description 添加黑名单
// @Tags UserRelation
// @Accept  json
// @Produce  json
// @param request body model.AddBlacklistRequest true "request"
// @Success		200 {object} model.Response{}
// @Router /relation/user/add_blacklist [post]
func addBlacklist(c *gin.Context) {
	req := new(model.AddBlacklistRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID, err := http.ParseTokenReUid(c)
	if err != nil {
		logger.Error("token解析失败", zap.Error(err))
		response.SetFail(c, "token解析失败", nil)
		return
	}

	_, err = svc.AddBlacklist(c, userID, req.UserID)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "添加到黑名单成功", nil)
}

// @Summary 删除好友
// @Description 删除好友
// @Tags UserRelation
// @Accept  json
// @Produce  json
// @param request body model.DeleteFriendRequest true "request"
// @Success		200 {object} model.Response{}
// @Router /relation/user/delete_friend [post]
func deleteFriend(c *gin.Context) {
	req := new(model.DeleteFriendRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID, err := http.ParseTokenReUid(c)
	if err != nil {
		logger.Error("token解析失败", zap.Error(err))
		response.SetFail(c, "token解析失败", nil)
		return
	}

	if err = svc.DeleteFriend(c, userID, req.UserID); err != nil {
		response.SetFail(c, "删除好友失败", nil)
		return
	}

	response.SetSuccess(c, "删除好友成功", nil)
}

// @Summary 管理好友请求
// @Description 管理好友请求  action (0=拒绝, 1=同意)
// @Tags UserRelation
// @Accept  json
// @Produce  json
// @param request body model.ManageFriendRequest true "request"
// @Success		200 {object} model.Response{}
// @Router /relation/user/manage_friend [post]
func manageFriend(c *gin.Context) {
	req := new(model.ManageFriendRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	if err := req.Validator(); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID, err := http.ParseTokenReUid(c)
	if err != nil {
		logger.Error("token解析失败", zap.Error(err))
		response.SetFail(c, "token解析失败", nil)
		return
	}

	responseData, err := svc.ManageFriend(c, userID, req.RequestID, int32(req.Action), req.E2EPublicKey)
	if err != nil {
		response.SetFail(c, code.Cause(err).Message(), nil)
		return
	}

	response.SetSuccess(c, "管理好友申请成功", responseData)
}

// @Summary 发送好友请求
// @Description 发送好友请求
// @Tags UserRelation
// @Accept  json
// @Produce  json
// @param request body model.SendFriendRequest true "request"
// @Success		200 {object} model.Response{}
// @Router /relation/user/add_friend [post]
func addFriend(c *gin.Context) {
	req := new(model.SendFriendRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	resp, err := svc.SendFriendRequest(c, thisId, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "发送好友请求成功", resp)
}

// @Summary 群聊成员列表
// @Description 群聊成员列表
// @Tags GroupRelation
// @Param group_id query integer true "群聊ID"
// @Produce  json
// @Success		200 {object} model.Response{}
// @Router /relation/group/member [get]
func getGroupMember(c *gin.Context) {
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

	resp, err := svc.GetGroupMember(c, uint32(gid))
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
// @Success		200 {object} model.Response{data=model.GroupRequestListResponse} "status (0=申请中, 1=待通过, 2=已加入, 3=已删除, 4=被拒绝, 5=被封禁)"
// @Router /relation/group/request_list [get]
func groupRequestList(c *gin.Context) {
	userID, err := http.ParseTokenReUid(c)
	if err != nil {
		logger.Error("token解析失败", zap.Error(err))
		response.SetFail(c, "token解析失败", nil)
		return
	}

	resp, err := svc.GroupRequestList(c, userID)
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
func inviteGroup(c *gin.Context) {
	req := new(model.InviteGroupRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	uid, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	if err = svc.InviteGroup(c, uid, req); err != nil {
		logger.Error("邀请好友加入群聊失败", zap.Error(err))
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
func joinGroup(c *gin.Context) {
	req := new(model.JoinGroupRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	uid, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	_, err = svc.JoinGroup(c, uid, req)
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
func manageJoinGroup(c *gin.Context) {
	req := new(model.ManageJoinGroupRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	if err := req.Validator(); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
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

	if err = svc.ManageJoinGroup(c, req.GroupID, req.ID, userID, status); err != nil {
		logger.Error("用户管理群聊申请", zap.Error(err))
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
func adminManageJoinGroup(c *gin.Context) {
	req := new(model.ManageJoinGroupRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	if err := req.Validator(); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
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

	if err = svc.AdminManageJoinGroup(c, req.ID, req.GroupID, adminID, status); err != nil {
		logger.Error("管理员管理群聊申请", zap.Error(err))
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
func removeUserFromGroup(c *gin.Context) {
	req := new(model.RemoveUserFromGroupRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	if userID == req.UserID {
		response.SetFail(c, "不能将自己从群聊中移除", nil)
		return
	}

	if err = svc.RemoveUserFromGroup(c, req.GroupID, userID, req.UserID); err != nil {
		logger.Error("RemoveUserFromGroup Failed", zap.Error(err))
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
func quitGroup(c *gin.Context) {
	req := new(model.QuitGroupRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	userID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	if err = svc.QuitGroup(c, req.GroupID, userID); err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	response.SetSuccess(c, "退出群聊成功", nil)
}

// @Summary 交换用户端到端公钥
// @Description 交换用户端到端公钥
// @Tags UserRelation
// @Accept json
// @Produce json
// @param request body model.SwitchUserE2EPublicKeyRequest true "request"
// @Security BearerToken
// @Success 200 {object} model.Response{}
// @Router /relation/user/switch/e2e/key [post]
func switchUserE2EPublicKey(c *gin.Context) {
	req := new(model.SwitchUserE2EPublicKeyRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	// 获取用户ID，可以从请求中的token中解析出来，前提是你的登录接口已经设置了正确的token
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	_, err = svc.SwitchUserE2EPublicKey(c, thisId, req.UserId, req.PublicKey)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "交换用户公钥成功", nil)
}

// @Summary 设置群聊静默通知
// @Description 设置群聊静默通知
// @Tags GroupRelation
// @Accept  json
// @Produce  json
// @param request body model.SetGroupSilentNotificationRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /relation/group/silent [post]
func setGroupSilentNotification(c *gin.Context) {
	req := new(model.SetGroupSilentNotificationRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
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

	_, err = svc.SetGroupSilentNotification(c, req.GroupId, thisId, req.IsSilent)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置成功", gin.H{"group_id": req.GroupId})
}

// @Summary 设置私聊静默通知
// @Description 设置私聊静默通知
// @Tags UserRelation
// @Accept  json
// @Produce  json
// @param request body model.SetUserSilentNotificationRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /relation/user/silent [post]
func setUserSilentNotification(c *gin.Context) {
	req := new(model.SetUserSilentNotificationRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
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

	_, err = svc.UserSilentNotification(c, thisId, req.UserId, req.IsSilent)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置成功", gin.H{"user_id": req.UserId})
}

// @Summary 关闭或打开对话(action: 0:关闭对话, 1:打开对话)
// @Description 关闭或打开对话(action: 0:关闭对话, 1:打开对话)
// @Tags Dialog
// @Accept  json
// @Produce  json
// @param request body model.CloseOrOpenDialogRequest true "request"
// @Success 200 {object} model.Response{}
// @Router /relation/dialog/show [post]
func closeOrOpenDialog(c *gin.Context) {
	req := new(model.CloseOrOpenDialogRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	if !model.IsValidOpenAction(req.Action) {
		response.SetFail(c, "打开或关闭对话失败", nil)
		return
	}

	_, err = svc.OpenOrCloseDialog(c, thisId, req)
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
func topOrCancelTopDialog(c *gin.Context) {
	req := new(model.TopOrCancelTopDialogRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	if !model.IsValidTopAction(req.Action) {
		response.SetFail(c, "置顶或取消置顶对话失败", nil)
		return
	}

	_, err = svc.TopOrCancelTopDialog(c, thisId, req)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "操作成功", gin.H{"dialog_id": req.DialogId})
}
