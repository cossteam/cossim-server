package interfaces

import (
	"fmt"
	v1 "github.com/cossim/coss-server/internal/relation/api/http/v1"
	"github.com/cossim/coss-server/internal/relation/app/command"
	"github.com/cossim/coss-server/internal/relation/app/query"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
	"github.com/mozillazg/go-pinyin"
	"reflect"
	"strings"
)

func (h *HttpServer) RemoveGroupMember(c *gin.Context, id uint32) {
	req := &v1.RemoveGroupMemberJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.SetFail(c, "参数错误", nil)
		return
	}

	if err := h.app.Commands.RemoveGroupMember.Handle(c, &command.RemoveGroupMember{
		GroupID:       id,
		CurrentUserID: c.Value(constants.UserID).(string),
		RemoveMember:  req.Member,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "移除群聊成员成功", nil)
}

func (h *HttpServer) AddGroupAdmin(c *gin.Context, id uint32) {
	req := &v1.AddGroupAdminJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.SetFail(c, "参数错误", nil)
		return
	}

	if err := h.app.Commands.AddGroupAdmin.Handle(c, &command.AddGroupAdmin{
		UserID:      c.Value(constants.UserID).(string),
		GroupID:     id,
		TargetUsers: req.UserIds,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "添加群聊管理员成功", nil)
}

func (h *HttpServer) SetGroupRemark(c *gin.Context, id uint32) {
	req := &v1.SetGroupRemarkJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.SetFail(c, "参数错误", nil)
		return
	}

	if err := h.app.Commands.SetGroupRemark.Handle(c, &command.SetGroupRemark{
		UserID:  c.Value(constants.UserID).(string),
		GroupID: id,
		Remark:  req.Remark,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置群聊昵称成功", nil)
}

func (h *HttpServer) SetGroupSilent(c *gin.Context, id uint32) {
	req := &v1.SetGroupSilentJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.SetFail(c, "参数错误", nil)
		return
	}

	if err := h.app.Commands.SetGroupSilent.Handle(c, &command.SetGroupSilent{
		UserID:  c.Value(constants.UserID).(string),
		GroupID: id,
		Silent:  req.Silent,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置群聊免打扰成功", nil)
}

func (h *HttpServer) QuitGroup(c *gin.Context, id uint32) {
	if err := h.app.Commands.QuitGroup.Handle(c, &command.QuitGroup{
		UserID:  c.Value(constants.UserID).(string),
		GroupID: id,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "退出群聊成功", nil)
}

func (h *HttpServer) ListGroup(c *gin.Context, params v1.ListGroupParams) {
	istGroup, err := h.app.Queries.ListGroup.Handle(c, &query.ListGroup{
		UserID:   c.Value(constants.UserID).(string),
		PageNum:  *params.PageNum,
		PageSize: *params.PageSize,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取群聊列表成功", listGroupToResponse(istGroup))
}

func sortAndGroupUsers1(groups []v1.GroupInfo, fieldName string) map[string][]v1.GroupInfo {
	groupedUsers := make(map[string][]v1.GroupInfo)
	keyMap := make(map[string]bool)

	for _, group := range groups {
		_ = reflect.ValueOf(group).FieldByName(fieldName)
		fieldValue := fieldOf(group, fieldName)
		name := fmt.Sprintf("%v", fieldValue.Interface())

		if isChinese(name) {
			pinyinSlice := pinyin.Pinyin(name, pinyin.NewArgs())
			firstChar := getFirstChar(pinyinSlice)
			name = strings.ToUpper(firstChar)
		} else if isSpecialChar(name) {
			name = "#"
		}

		k := strings.ToUpper(name[:1])
		groupedUsers[k] = append(groupedUsers[k], group)
		keyMap[k] = true
	}

	return groupedUsers
}

func fieldOf(i interface{}, name string) reflect.Value {
	val := reflect.ValueOf(i)
	field := reflect.Indirect(val).FieldByName(name)
	if !field.IsValid() {
		// 如果字段不存在，返回零值
		return reflect.Value{}
	}
	return field
}

func listGroupToResponse(group *query.ListGroupResponse) *v1.GroupList {
	r := &v1.GroupList{
		List:  make(map[string][]v1.GroupInfo),
		Total: len(group.List),
	}
	var groupList []v1.GroupInfo
	for _, v := range group.List {
		groupList = append(groupList, v1.GroupInfo{
			Avatar:   v.Avatar,
			DialogId: v.DialogID,
			Id:       v.ID,
			Name:     v.Name,
			Status:   v.Status,
			Type:     v.Type,
		})
	}
	groupedUsers := sortAndGroupUsers1(groupList, "Name")
	r.List = groupedUsers
	return r
}

func (h *HttpServer) ManageGroupRequest(c *gin.Context, id uint32) {
	if err := h.app.Commands.ManageGroupRequest.Handle(c, &command.ManageGroupRequest{
		UserID:    c.Value(constants.UserID).(string),
		RequestID: id,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "管理群聊请求成功", nil)
}

func (h *HttpServer) AddGroupRequest(c *gin.Context, id uint32) {
	req := &v1.AddGroupRequestJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.SetFail(c, "参数错误", nil)
		return
	}

	if err := h.app.Commands.AddGroupRequest.Handle(c, &command.AddGroupRequest{
		UserID:  c.Value(constants.UserID).(string),
		GroupID: id,
		Remark:  req.Remark,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "发送入群申请成功", nil)
}

func (h *HttpServer) InviteJoinGroup(c *gin.Context, id uint32) {
	req := &v1.InviteJoinGroupJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.SetFail(c, "参数错误", nil)
		return
	}

	if err := h.app.Commands.InviteJoinGroup.Handle(c, &command.InviteJoinGroup{
		UserID:     c.Value(constants.UserID).(string),
		GroupID:    id,
		TargetUser: req.Member,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "邀请成功", nil)
}

func (h *HttpServer) DeleteGroupRequest(c *gin.Context, id uint32) {
	if err := h.app.Commands.DeleteGroupRequest.Handle(c.Request.Context(), &command.DeleteGroupRequest{
		UserID: c.Value(constants.UserID).(string),
		//GroupID:   0,
		RequestID: id,
	}); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "删除群聊申请记录成功", nil)
}

func (h *HttpServer) ListGroupRequest(c *gin.Context, params v1.ListGroupRequestParams) {
	listGroupRequest, err := h.app.Queries.ListGroupRequest.Handle(c.Request.Context(), &query.ListGroupRequest{
		UserID: c.Value(constants.UserID).(string),
		//GroupID: id,
		PageNum:  *params.PageNum,
		PageSize: *params.PageSize,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, "获取群聊申请列表成功", listGroupRequestToResponse(listGroupRequest))
}

func listGroupRequestToResponse(request *query.ListGroupRequestResponse) *v1.GroupRequestList {
	r := &v1.GroupRequestList{}
	for _, v := range request.List {
		r.List = append(r.List, v1.GroupRequest{
			CreateAt:    &v.CreateAt,
			CreatorId:   &v.CreatorId,
			ExpiredAt:   &v.ExpiredAt,
			GroupAvatar: &v.GroupAvatar,
			GroupId:     &v.GroupId,
			GroupName:   &v.GroupName,
			GroupType:   v1.GroupRequestGroupType(v.GroupType),
			Id:          &v.ID,
			ReceiverInfo: &v1.ShortUserInfo{
				Avatar:   v.RecipientInfo.Avatar,
				Nickname: v.RecipientInfo.Nickname,
				UserId:   v.RecipientInfo.ID,
			},
			Remark: v.Remark,
			SenderInfo: &v1.ShortUserInfo{
				Avatar:   v.SenderInfo.Avatar,
				Nickname: v.SenderInfo.Nickname,
				UserId:   v.SenderInfo.ID,
			},
			Status: v1.GroupRequestStatus(v.Status),
		})
	}
	return r
}

func (h *HttpServer) ListGroupMember(c *gin.Context, id uint32) {
	listGroupMember, err := h.app.Queries.ListGroupMember.Handle(c.Request.Context(), &query.ListGroupMember{
		UserID:  c.Value(constants.UserID).(string),
		GroupID: id,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取群聊成员成功", listGroupMemberToResponse(listGroupMember))
}

func listGroupMemberToResponse(member *query.ListGroupMemberResponse) *v1.GroupMemberList {
	r := &v1.GroupMemberList{}
	for _, m := range member.List {
		r.List = append(r.List, v1.GroupMember{
			UserId:   m.UserID,
			Nickname: m.Nickname,
			Avatar:   m.Avatar,
			Remark:   m.Remark,
			Identity: v1.GroupMemberIdentity(m.Identity),
		})
	}
	return r
}
