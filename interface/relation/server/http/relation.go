package http

import (
	"context"
	"fmt"
	msgconfig "github.com/cossim/coss-server/interface/msg/config"
	"github.com/cossim/coss-server/pkg/http"
	pkghttp "github.com/cossim/coss-server/pkg/http"
	"github.com/cossim/coss-server/pkg/msg_queue"
	"time"

	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/cossim/coss-server/pkg/utils/usersorter"
	msgApi "github.com/cossim/coss-server/service/msg/api/v1"
	relationApi "github.com/cossim/coss-server/service/relation/api/v1"
	userApi "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// @Summary 黑名单
// @Description 黑名单
// @Accept  json
// @Produce  json
// @Success		200 {object} utils.Response{}
// @Router /relation/blacklist [get]
func blackList(c *gin.Context) {
	userID, err := http.ParseTokenReUid(c)
	if err != nil {
		logger.Error("token解析失败", zap.Error(err))
		response.Fail(c, "token解析失败", nil)
		return
	}

	// 检查用户是否存在
	user, err := userClient.UserInfo(context.Background(), &userApi.UserInfoRequest{UserId: userID})
	if err != nil {
		c.Error(err)
		return
	}

	if user == nil {
		response.Fail(c, "用户不存在", nil)
		return
	}

	// 获取黑名单列表
	blacklistResp, err := relationClient.GetBlacklist(context.Background(), &relationApi.GetBlacklistRequest{UserId: userID})
	if err != nil {
		c.Error(err)
		return
	}

	var users []string
	for _, user := range blacklistResp.Blacklist {
		users = append(users, user.UserId)
	}

	blacklist, err := userClient.GetBatchUserInfo(context.Background(), &userApi.GetBatchUserInfoRequest{UserIds: users})
	if err != nil {
		return
	}

	response.Success(c, "获取黑名单列表成功", blacklist)
}

// @Summary 好友列表
// @Description 好友列表
// @Accept  json
// @Produce  json
// @Success		200 {object} utils.Response{}
// @Router /relation/friend_list [get]
func friendList(c *gin.Context) {
	userID, err := http.ParseTokenReUid(c)
	if err != nil {
		logger.Error("token解析失败", zap.Error(err))
		response.Fail(c, "token解析失败", nil)
		return
	}
	// 检查用户是否存在
	user, err := userClient.UserInfo(context.Background(), &userApi.UserInfoRequest{UserId: userID})
	if err != nil {
		logger.Error("user service UserInfo", zap.Error(err))
		c.Error(err)
		return
	}

	if user == nil {
		response.Fail(c, "用户不存在", nil)
		return
	}

	// 获取好友列表
	friendListResp, err := relationClient.GetFriendList(context.Background(), &relationApi.GetFriendListRequest{UserId: userID})
	if err != nil {
		logger.Error("user service GetFriendList", zap.Error(err))
		c.Error(err)
		return
	}
	var users []string
	for _, user := range friendListResp.FriendList {
		users = append(users, user.UserId)
	}

	userInfos, err := userClient.GetBatchUserInfo(context.Background(), &userApi.GetBatchUserInfoRequest{UserIds: users})
	if err != nil {
		logger.Error("user service GetBatchUserInfo", zap.Error(err))
		c.Error(err)
		return
	}

	var data []usersorter.User
	for _, v := range userInfos.Users {
		data = append(data, usersorter.CustomUserData{
			UserID:   v.UserId,
			NickName: v.NickName,
			Email:    v.Email,
			Tel:      v.Tel,
			Avatar:   v.Avatar,
			Status:   uint(v.Status),
		})
	}

	// Sort and group by specified field
	groupedUsers := usersorter.SortAndGroupUsers(data, "NickName")

	response.Success(c, "获取好友列表成功", groupedUsers)
}

type deleteBlacklistRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// @Summary 删除黑名单
// @Description 删除黑名单
// @Accept  json
// @Produce  json
// @param request body deleteBlacklistRequest true "request"
// @Success		200 {object} utils.Response{}
// @Router /relation/delete_blacklist [post]
func deleteBlacklist(c *gin.Context) {
	req := new(deleteBlacklistRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.Fail(c, "参数验证失败", nil)
		return
	}

	userID, err := http.ParseTokenReUid(c)
	if err != nil {
		logger.Error("token解析失败", zap.Error(err))
		response.Fail(c, "token解析失败", nil)
		return
	}

	// 检查用户是否存在
	user, err := userClient.UserInfo(context.Background(), &userApi.UserInfoRequest{UserId: userID})
	if err != nil {
		c.Error(err)
		return
	}

	if user == nil {
		response.Fail(c, "用户不存在", nil)
		return
	}

	// 检查要删除的黑名单用户是否存在
	user2, err := userClient.UserInfo(context.Background(), &userApi.UserInfoRequest{UserId: userID})
	if err != nil {
		c.Error(err)
		return
	}

	if user2 == nil {
		response.Fail(c, "用户不存在", nil)
		return
	}

	// 进行删除黑名单操作
	if _, err = relationClient.DeleteBlacklist(context.Background(), &relationApi.DeleteBlacklistRequest{UserId: userID, FriendId: req.UserID}); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "删除黑名单成功", nil)
}

type addBlacklistRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// @Summary 添加黑名单
// @Description 添加黑名单
// @Accept  json
// @Produce  json
// @param request body addBlacklistRequest true "request"
// @Success		200 {object} utils.Response{}
// @Router /relation/add_blacklist [post]
func addBlacklist(c *gin.Context) {
	req := new(addBlacklistRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.Fail(c, "参数验证失败", nil)
		return
	}

	userID, err := http.ParseTokenReUid(c)
	if err != nil {
		logger.Error("token解析失败", zap.Error(err))
		response.Fail(c, "token解析失败", nil)
		return
	}

	// 检查用户是否存在
	user, err := userClient.UserInfo(context.Background(), &userApi.UserInfoRequest{UserId: userID})
	if err != nil {
		c.Error(err)
		return
	}

	if user == nil {
		response.Fail(c, "用户不存在", nil)
		return
	}

	// 检查添加黑名单的用户是否存在
	user2, err := userClient.UserInfo(context.Background(), &userApi.UserInfoRequest{UserId: req.UserID})
	if err != nil {
		c.Error(err)
		return
	}

	if user2 == nil {
		response.Fail(c, "用户不存在", nil)
		return
	}

	// 进行添加黑名单操作
	if _, err = relationClient.AddBlacklist(context.Background(), &relationApi.AddBlacklistRequest{UserId: userID, FriendId: req.UserID}); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "添加到黑名单成功", nil)
}

type deleteFriendRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

// @Summary 删除好友
// @Description 删除好友
// @Accept  json
// @Produce  json
// @param request body deleteFriendRequest true "request"
// @Success		200 {object} utils.Response{}
// @Router /relation/delete_friend [post]
func deleteFriend(c *gin.Context) {
	req := new(deleteFriendRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.Fail(c, "参数验证失败", nil)
		return
	}

	userID, err := http.ParseTokenReUid(c)
	if err != nil {
		logger.Error("token解析失败", zap.Error(err))
		response.Fail(c, "token解析失败", nil)
		return
	}

	// 检查用户是否存在
	user, err := userClient.UserInfo(context.Background(), &userApi.UserInfoRequest{UserId: userID})
	if err != nil {
		c.Error(err)
		return
	}

	if user == nil {
		response.Fail(c, "用户不存在", nil)
		return
	}

	// 检查删除的用户是否存在
	user2, err := userClient.UserInfo(context.Background(), &userApi.UserInfoRequest{UserId: req.UserID})
	if err != nil {
		c.Error(err)
		return
	}

	if user2 == nil {
		response.Fail(c, "要删除的用户不存在", nil)
		return
	}

	// 进行删除好友操作
	if _, err = relationClient.DeleteFriend(context.Background(), &relationApi.DeleteFriendRequest{UserId: userID, FriendId: req.UserID}); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "删除好友成功", nil)
}

type confirmFriendRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	P2PublicKey string `json:"p2public_key"`
}

// @Summary 确认添加好友
// @Description 确认添加好友
// @Accept  json
// @Produce  json
// @param request body confirmFriendRequest true "request"
// @Success		200 {object} utils.Response{}
// @Router /relation/confirm_friend [post]
func confirmFriend(c *gin.Context) {
	req := new(confirmFriendRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.Fail(c, "参数验证失败", nil)
		return
	}

	userID, err := http.ParseTokenReUid(c)
	if err != nil {
		logger.Error("token解析失败", zap.Error(err))
		response.Fail(c, "token解析失败", nil)
		return
	}

	// 检查用户是否存在
	user, err := userClient.UserInfo(context.Background(), &userApi.UserInfoRequest{UserId: userID})
	if err != nil {
		c.Error(err)
		return
	}

	if user == nil {
		response.Fail(c, "用户不存在", nil)
		return
	}

	// 检查要添加的用户是否存在
	user2, err := userClient.UserInfo(context.Background(), &userApi.UserInfoRequest{UserId: req.UserID})
	if err != nil {
		c.Error(err)
		return
	}

	if user2 == nil {
		response.Fail(c, "用户不存在", nil)
		return
	}

	// 进行确认好友操作
	if _, err = relationClient.ConfirmFriend(context.Background(), &relationApi.ConfirmFriendRequest{UserId: userID, FriendId: req.UserID}); err != nil {
		c.Error(err)
		return
	}
	msg := msgconfig.WsMsg{Uid: req.UserID, Event: msgconfig.AddFriendEvent, Data: req}
	// todo 记录离线推送
	err = rabbitMQClient.PublishMessage(req.UserID, msg)
	if err != nil {
		fmt.Println("发布消息失败：", err)
		response.Fail(c, "同意好友请求失败", nil)
		return
	}
	response.Success(c, "确认好友成功", nil)
}

type addFriendRequest struct {
	UserID      string `json:"user_id" binding:"required"`
	Msg         string `json:"msg"`
	P2PublicKey string `json:"p2public_key"`
}

// @Summary 添加好友
// @Description 添加好友
// @Accept  json
// @Produce  json
// @param request body addFriendRequest true "request"
// @Success		200 {object} utils.Response{}
// @Router /relation/add_friend [post]
func addFriend(c *gin.Context) {
	req := new(addFriendRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.Fail(c, "参数验证失败", nil)
		return
	}
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.Fail(c, err.Error(), nil)
		return
	}
	// 检查用户是否存在
	user, err := userClient.UserInfo(context.Background(), &userApi.UserInfoRequest{UserId: thisId})
	if err != nil {
		c.Error(err)
		return
	}
	if user == nil {
		response.Fail(c, "用户不存在", nil)
		return
	}

	// 检查添加的用户是否存在
	user2, err := userClient.UserInfo(context.Background(), &userApi.UserInfoRequest{UserId: req.UserID})
	if err != nil {
		c.Error(err)
		return
	}
	if user2 == nil {
		response.Fail(c, "用户不存在", nil)
		return
	}
	if _, err := relationClient.AddFriend(context.Background(), &relationApi.AddFriendRequest{UserId: thisId, FriendId: req.UserID}); err != nil {
		c.Error(err)
		return
	}
	//创建对话
	dialog, err := dialogClient.CreateDialog(context.Background(), &msgApi.CreateDialogRequest{OwnerId: thisId, Type: 0, GroupId: 0})
	if err != nil {
		c.Error(err)
		return
	}
	//加入对话
	_, err = dialogClient.JoinDialog(context.Background(), &msgApi.JoinDialogRequest{DialogId: dialog.Id, UserId: thisId})
	if err != nil {
		c.Error(err)
		return
	}
	//todo 对方加入对话
	_, err = dialogClient.JoinDialog(context.Background(), &msgApi.JoinDialogRequest{DialogId: dialog.Id, UserId: req.UserID})
	if err != nil {
		c.Error(err)
		return
	}
	msg := msgconfig.WsMsg{Uid: req.UserID, Event: msgconfig.AddFriendEvent, Data: req, SendAt: time.Now().Unix()}

	//通知消息服务有消息需要发送
	err = rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
	if err != nil {
		return
	}

	err = rabbitMQClient.PublishMessage(req.UserID, msg)
	if err != nil {
		fmt.Println("发布消息失败：", err)
		response.Fail(c, "发送好友请求失败", nil)
		return
	}
	response.Success(c, "发送好友请求成功", nil)
}

type joinGroupRequest struct {
	GroupID uint32 `json:"group_id" binding:"required"`
}

// @Summary 加入群聊
// @Description 加入群聊
// @Accept  json
// @Produce  json
// @param request body joinGroupRequest true "request"
// @Success		200 {object} utils.Response{}
// @Router /relation/group/join [post]
func joinGroup(c *gin.Context) {
	req := new(joinGroupRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Error("参数验证失败", zap.Error(err))
		response.Fail(c, "参数验证失败", nil)
		return
	}

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.Fail(c, err.Error(), nil)
		return
	}

	_, err = userGroupClient.InsertUserGroup(context.Background(), &relationApi.UserGroupRequest{UserId: thisId, GroupId: req.GroupID})
	if err != nil {
		return
	}

	response.Success(c, "发送好友请求成功", nil)
}
