package interfaces

import (
	v1 "github.com/cossim/coss-server/internal/relation/api/http/v1"
	"github.com/cossim/coss-server/internal/relation/app/command"
	"github.com/cossim/coss-server/internal/relation/app/query"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
	"github.com/mozillazg/go-pinyin"
	"reflect"
	"strings"
	"unicode"
)

func (h *HttpServer) DeleteFriend(c *gin.Context, id string) {
	if err := h.app.Commands.DeleteFriend.Handle(c, &command.DeleteFriend{
		CurrentUserID: c.Value(constants.UserID).(string),
		TargetUserID:  id,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "删除好友成功", nil)
}

func (h *HttpServer) SetUserBurn(c *gin.Context, id string) {
	req := &v1.SetUserBurnJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.SetFail(c, "参数错误", nil)
		return
	}

	if err := h.app.Commands.SetUserBurn.Handle(c, &command.SetUserBurn{
		CurrentUserID: c.Value(constants.UserID).(string),
		TargetUserID:  id,
		Burn:          req.Burn,
		Timeout:       req.Timeout,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置用户阅后即焚成功", nil)
}

func (h *HttpServer) SetUserRemark(c *gin.Context, id string) {
	req := &v1.SetUserRemarkJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.SetFail(c, "参数错误", nil)
		return
	}

	if err := h.app.Commands.SetUserRemark.Handle(c, &command.SetUserRemark{
		CurrentUserID: c.Value(constants.UserID).(string),
		TargetUserID:  id,
		Remark:        req.Remark,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置用户备注成功", nil)
}

func (h *HttpServer) SetUserSilent(c *gin.Context, id string) {
	req := &v1.SetUserSilentJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.SetFail(c, "参数错误", nil)
		return
	}

	if err := h.app.Commands.SetUserSilent.Handle(c, &command.SetUserSilent{
		CurrentUserID: c.Value(constants.UserID).(string),
		TargetUserID:  id,
		Silent:        req.Silent,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置用户免打扰成功", nil)
}

func (h *HttpServer) ExchangeE2EKey(c *gin.Context, id string) {
	req := &v1.ExchangeE2EKeyJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.SetFail(c, "参数错误", nil)
		return
	}

	if err := h.app.Commands.ExchangeE2EKey.Handle(c, &command.ExchangeE2EKey{
		CurrentUserID: c.Value(constants.UserID).(string),
		TargetUserID:  id,
		PublicKey:     req.PublicKey,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "交换端到端公钥成功", nil)
}

// Blacklist
// @Summary 黑名单
// @Description 黑名单
// @Tags userRelation
// @Produce  json
// @Success		200 {object} model.Response{}
// @Router /api/v1/relation/user/blacklist [get]
func (h *HttpServer) Blacklist(c *gin.Context, params v1.BlacklistParams) {
	blacklist, err := h.app.Queries.UserBlacklist.Handle(c.Request.Context(), &query.UserBlacklist{
		UserID:   c.Value(constants.UserID).(string),
		PageNum:  *params.PageNum,
		PageSize: *params.PageSize,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取黑名单列表成功", blacklistToResponse(blacklist))
}

func blacklistToResponse(blacklist *query.UserBlacklistResponse) *v1.Blacklist {
	var blacklistList []v1.Black
	for _, v := range blacklist.List {
		blacklistList = append(blacklistList, v1.Black{
			CossId:   v.CossID,
			Nickname: v.Nickname,
			UserId:   v.UserID,
			Avatar:   v.Avatar,
		})
	}
	return &v1.Blacklist{
		List:  blacklistList,
		Total: blacklist.Total,
		//Page:  blacklist.Page,
	}
}

// AddBlacklist
// @Summary 添加黑名单
// @Description 添加黑名单
// @Tags UserRelation
// @Accept  json
// @Produce  json
// @param request body v1.AddBlacklistJSONRequestBody true "request"
// @Success		200 {object} model.Response{}
// @Router /api/v1/relation/user/blacklist [post]
func (h *HttpServer) AddBlacklist(c *gin.Context) {
	req := &v1.AddBlacklistJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.SetFail(c, "参数错误", nil)
		return
	}

	if err := h.app.Commands.AddBlacklist.Handle(c, &command.AddBlacklist{
		CurrentUserID: c.Value(constants.UserID).(string),
		TargetUserID:  req.UserId,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "添加黑名单成功", nil)
}

// DeleteBlacklist
// @Summary 删除黑名单
// @Description 删除黑名单
// @Tags userRelation
// @Accept  json
// @Produce  json
// @Param id path string true "要移除黑名单的用户ID"
// @Success		200 {object} model.Response{}
// @Router /api/v1/relation/user/blacklist/{id} [delete]
func (h *HttpServer) DeleteBlacklist(c *gin.Context, id string) {
	if err := h.app.Commands.DeleteBlacklist.Handle(c, &command.DeleteBlacklist{
		CurrentUserID: c.Value(constants.UserID).(string),
		TargetUserID:  id,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "删除黑名单成功", nil)
}

// ListFriendRequest
// @Summary 好友申请列表
// @Description 好友申请列表
// @Tags userRelation
// @Produce  json
// @Param page_num query int false "页码"
// @Param page_size query int false "页大小"
// @Success		200 {object} model.Response{data=v1.UserFriendRequestList} "UserStatus 申请状态 (0=申请中, 1=已通过, 2=被拒绝)"
// @Router /api/v1/relation/user/friend_request [get]
func (h *HttpServer) ListFriendRequest(c *gin.Context, params v1.ListFriendRequestParams) {
	listFriendRequest, err := h.app.Queries.ListFriendRequest.Handle(c, &query.ListFriendRequest{
		UserID:   c.Value(constants.UserID).(string),
		PageNum:  *params.PageNum,
		PageSize: *params.PageSize,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取好友请求列表成功", listFriendRequestToResponse(listFriendRequest))
}

func listFriendRequestToResponse(request *query.ListFriendRequestResponse) *v1.UserFriendRequestList {
	listFriendRequestList := &v1.UserFriendRequestList{
		Total: request.Total,
	}
	for _, v := range request.List {
		listFriendRequestList.List = append(listFriendRequestList.List, v1.FriendRequest{
			Id:        v.ID,
			Remark:    v.Remark,
			Status:    v.Status,
			CreateAt:  v.CreateAt,
			ExpiredAt: v.ExpiredAt,
			SenderId:  v.SenderID,
			SenderInfo: &v1.FriendRequestUserInfo{
				Avatar:   v.SenderInfo.Avatar,
				CossId:   v.SenderInfo.CossID,
				Nickname: v.SenderInfo.Nickname,
				UserId:   v.SenderInfo.UserID,
			},
			RecipientId: v.RecipientID,
			RecipientInfo: &v1.FriendRequestUserInfo{
				Avatar:   v.RecipientInfo.Avatar,
				CossId:   v.RecipientInfo.CossID,
				Nickname: v.RecipientInfo.Nickname,
				UserId:   v.RecipientInfo.UserID,
			},
		})
	}
	return listFriendRequestList
}

// DeleteFriendRequest
// @Summary 删除好友申请记录
// @Description 删除好友申请记录
// @Tags userRelation
// @Accept  json
// @Produce  json
// @Param id path int true "好友请求记录ID"
// @Success		200 {object} model.Response{}
// @Router /api/v1/relation/user/friend_request/{id} [delete]
func (h *HttpServer) DeleteFriendRequest(c *gin.Context, id uint32) {
	if id == 0 {
		response.SetFail(c, "好友请求ID不能为空", nil)
		return
	}

	if err := h.app.Commands.DeleteFriendRequest.Handle(c, &command.DeleteFriendRequest{
		UserID: c.Value(constants.UserID).(string),
		ID:     id,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "删除好友请求成功", nil)
}

// ManageFriendRequest
// @Summary 管理好友请求
// @Description 管理好友请求  action (0=拒绝, 1=同意)
// @Tags userRelation
// @Accept  json
// @Produce  json
// @param request body v1.ManageFriendRequestJSONRequestBody true "request"
// @Success		200 {object} model.Response{}
// @Router /api/v1/relation/user/friend_request/{id} [PUT]
func (h *HttpServer) ManageFriendRequest(c *gin.Context, id uint32) {
	req := &v1.ManageFriendRequestJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.SetFail(c, "参数错误", nil)
		return
	}

	if err := h.app.Commands.ManageFriendRequest.Handle(c, &command.ManageFriendRequest{
		UserID: c.Value(constants.UserID).(string),
		ID:     id,
		Action: command.ManageFriendRequestAction(req.Action),
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "处理好友请求成功", nil)
}

// ListFriend
// @Summary 好友列表
// @Description 好友列表
// @Tags userRelation
// @Produce  json
// @Success		200 {object} model.Response{}
// @Router /api/v1/relation/user/friend [get]
func (h *HttpServer) ListFriend(c *gin.Context) {
	listFriend, err := h.app.Queries.ListFriend.Handle(c, &query.ListFriend{
		UserID: c.Value(constants.UserID).(string),
		//PageNum:  0,
		//PageSize: 0,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取好友列表成功", listFriendToResponse(listFriend))
}

func listFriendToResponse(listFriend *query.ListFriendResponse) *v1.UserFriendList {
	resp := &v1.UserFriendList{
		List:  make(map[string][]v1.UserInfo),
		Total: len(listFriend.List),
	}
	var userList []v1.UserInfo
	for _, v := range listFriend.List {
		userList = append(userList, v1.UserInfo{
			Avatar:   v.Avatar,
			CossId:   v.CossId,
			DialogId: v.DialogId,
			Email:    v.Email,
			Nickname: v.NickName,
			Preferences: v1.Preferences{
				OpenBurnAfterReading:        v.Preferences.OpenBurnAfterReading,
				OpenBurnAfterReadingTimeOut: v.Preferences.OpenBurnAfterReadingTimeOut,
				Remark:                      v.Preferences.Remark,
				SilentNotification:          v.Preferences.SilentNotification,
			},
			RelationStatus: v1.UserInfoRelationStatus(v.RelationStatus),
			Signature:      v.Signature,
			Status:         v.Status,
			Tel:            v.Tel,
			UserId:         v.UserID,
		})
	}
	groupedUsers := sortAndGroupUsers(userList, "Nickname")
	resp.List = groupedUsers
	return resp
}

func sortAndGroupUsers(users []v1.UserInfo, fieldName string) map[string][]v1.UserInfo {
	groupedUsers := make(map[string][]v1.UserInfo)
	keyMap := make(map[string]bool)

	for _, user := range users {
		name := getFieldValue(user, fieldName)
		groupKey := getGroupKey(name, user.Preferences.Remark)
		groupedUsers[groupKey] = append(groupedUsers[groupKey], user)
		keyMap[groupKey] = true
	}

	return groupedUsers
}

func getFieldValue(user v1.UserInfo, fieldName string) string {
	r := reflect.ValueOf(user)
	f := reflect.Indirect(r).FieldByName(fieldName)
	if f.IsValid() {
		return f.String()
	}
	return ""
}

func getGroupKey(name, remark string) string {
	if remark != "" {
		name = remark
	}

	if isChinese(name) {
		pinyinSlice := pinyin.Pinyin(name, pinyin.NewArgs())
		firstChar := getFirstChar(pinyinSlice)
		return strings.ToUpper(firstChar)
	} else if isSpecialChar(name) {
		return "#"
	}

	return strings.ToUpper(name[:1])
}

func isSpecialChar(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return true
		}
	}
	return false
}

func isChinese(s string) bool {
	for _, r := range s {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func getFirstChar(pinyinSlice [][]string) string {
	if len(pinyinSlice) > 0 && len(pinyinSlice[0]) > 0 {
		return pinyinSlice[0][0][:1]
	}
	return ""
}

// AddFriend
// @Summary 发送好友请求
// @Description 发送好友请求
// @Tags userRelation
// @Accept  json
// @Produce  json
// @param request body v1.AddFriendJSONRequestBody true "request"
// @Success		200 {object} model.Response{}
// @Router /api/v1/relation/user/friend [post]
func (h *HttpServer) AddFriend(c *gin.Context) {
	req := &v1.AddFriendJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.SetFail(c, "参数错误", nil)
		return
	}

	if err := h.app.Commands.AddFriend.Handle(c, &command.AddFriend{
		CurrentUserID: c.Value(constants.UserID).(string),
		TargetUserID:  req.UserId,
		Remark:        req.Remark,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "发送好友请求成功", nil)
}
