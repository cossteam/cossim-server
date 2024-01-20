package http

import (
	"context"
	"github.com/cossim/coss-server/interface/group/api/model"
	pkghttp "github.com/cossim/coss-server/pkg/http"
	"github.com/cossim/coss-server/pkg/http/response"
	api "github.com/cossim/coss-server/service/group/api/v1"
	rapi "github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"strconv"
)

// @Summary 获取群聊信息
// @Description 获取群聊信息
// @Accept  json
// @Produce  json
// @Param gid query int32 true "群聊ID"
// @Success 200 {object} model.Response{}
// @Router /group/info [get]
func GetGroupInfoByGid(c *gin.Context) {
	gid := c.Query("gid")
	if gid == "" {
		response.SetFail(c, "群聊ID不能为空", nil)
		return
	}
	//转换类型
	gidInt, err := strconv.Atoi(gid)
	if err != nil {
		response.SetFail(c, "群聊ID错误", nil)
		return
	}
	group, err := groupClient.GetGroupInfoByGid(c, &api.GetGroupInfoRequest{
		Gid: uint32(gidInt),
	})
	if err != nil {
		response.SetFail(c, "获取群聊信息失败", nil)
		return
	}

	response.SetSuccess(c, "获取群聊信息成功", group)
}

// @Summary 批量获取群聊信息
// @Description 批量获取群聊信息
// @Accept  json
// @Produce  json
// @Param groupIds query []string true "群聊ID列表"
// @Success 200 {object} model.Response{}
// @Router /group/getBatch [get]
func getBatchGroupInfoByIDs(c *gin.Context) {
	groupIds := c.QueryArray("groupIds")
	ids := make([]uint32, len(groupIds))
	//转换类型
	for i, groupId := range groupIds {
		id, err := strconv.Atoi(groupId)
		if err != nil {
			response.SetFail(c, "群聊ID列表转换失败", nil)
			return
		}
		ids[i] = uint32(id)
	}

	groups, err := groupClient.GetBatchGroupInfoByIDs(c, &api.GetBatchGroupInfoRequest{
		GroupIds: ids,
	})
	if err != nil {
		response.SetFail(c, "批量获取群聊信息失败", nil)
		return
	}
	response.SetSuccess(c, "批量获取群聊信息成功", gin.H{"groups": groups})
}

// @Summary 更新群聊信息
// @Description 更新群聊信息
// @Accept  json
// @Produce  json
// @Param request body model.UpdateGroupRequest true "请求体"
// @Success 200 {object} model.Response{}
// @Router /group/update/{gid} [post]
func updateGroup(c *gin.Context) {
	req := new(model.UpdateGroupRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	group, err := groupClient.GetGroupInfoByGid(context.Background(), &api.GetGroupInfoRequest{
		Gid: req.GroupId,
	})
	if err != nil {
		response.SetFail(c, "未找到对应的群聊", nil)
		return
	}
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}
	if !model.IsValidGroupType(api.GroupType(req.Type)) {
		response.SetFail(c, "群聊类型错误", nil)
	}
	sf, err := userGroupClient.GetGroupRelation(context.Background(), &rapi.GetGroupRelationRequest{
		UserId:  thisId,
		GroupId: req.GroupId,
	})
	if err != nil {
		response.SetFail(c, "未找到对应的群聊", nil)
		return
	}
	if sf.Identity != rapi.GroupIdentity_IDENTITY_ADMIN && sf.Identity != rapi.GroupIdentity_IDENTITY_OWNER {
		response.SetFail(c, "没有权限", nil)
		return
	}

	// 更新群聊信息
	group.Type = api.GroupType(int32(req.Type))
	group.Status = api.GroupStatus(int32(req.Status))
	group.MaxMembersLimit = int32(req.MaxMembersLimit)
	group.CreatorId = req.CreatorID
	group.Name = req.Name
	group.Avatar = req.Avatar

	updatedGroup, err := groupClient.UpdateGroup(context.Background(), &api.UpdateGroupRequest{
		Group: group,
	})
	if err != nil {
		response.SetFail(c, "更新群聊信息失败", nil)
		return
	}

	response.SetSuccess(c, "更新群聊信息成功", gin.H{"group": updatedGroup})
}

// @Summary 创建群聊
// @Description 创建群聊
// @Accept  json
// @Produce  json
// @Param request body model.CreateGroupRequest true "请求体"
// @Success 200 {object} model.Response{}
// @Router /group/create [post]
func createGroup(c *gin.Context) {
	req := new(model.CreateGroupRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}
	//判断参数如果不属于枚举定义的则返回错误
	if !model.IsValidGroupType(api.GroupType(req.Type)) {
		response.SetFail(c, "群聊类型错误", nil)
		return
	}
	group := &api.Group{
		Type:            api.GroupType(int32(req.Type)),
		MaxMembersLimit: int32(req.MaxMembersLimit),
		CreatorId:       thisId,
		Name:            req.Name,
		Avatar:          req.Avatar,
	}
	//
	//createdGroup, err := groupClient.CreateGroup(context.Background(), &api.CreateGroupRequest{
	//	Group: group,
	//})
	//if err != nil {
	//	c.Error(err)
	//	return
	//}
	//
	//_, err = userGroupClient.JoinGroup(context.Background(), &rapi.JoinGroupRequest{
	//	GroupId:  createdGroup.Id,
	//	UserId:   thisId,
	//	Identify: rapi.GroupIdentity_IDENTITY_OWNER,
	//})
	//if err != nil {
	//	c.Error(err)
	//	return
	//}
	////创建对话
	//dialog, err := dialogClient.CreateDialog(context.Background(), &rapi.CreateDialogRequest{OwnerId: thisId, Type: 1, GroupId: createdGroup.Id})
	//if err != nil {
	//	c.Error(err)
	//	return
	//}
	////加入对话
	//_, err = dialogClient.JoinDialog(context.Background(), &rapi.JoinDialogRequest{DialogId: dialog.Id, UserId: thisId})
	//if err != nil {
	//	c.Error(err)
	//	return
	//}
	//resp := &model.CreateGroupResponse{
	//	Id:              createdGroup.Id,
	//	Avatar:          createdGroup.Avatar,
	//	Name:            createdGroup.Name,
	//	Type:            uint32(createdGroup.Type),
	//	Status:          int32(createdGroup.Status),
	//	MaxMembersLimit: createdGroup.MaxMembersLimit,
	//	CreatorId:       createdGroup.CreatorId,
	//	DialogId:        dialog.Id,
	//}

	resp, err := svc.CreateGroup(c, group)
	if err != nil {
		logger.Error("创建群聊失败", zap.Error(err))
		response.Fail(c, "创建群聊失败", nil)
		return
	}
	response.Success(c, "创建群聊成功", resp)
}

// @Summary 删除群聊
// @Description 删除群聊
// @Accept  json
// @Produce  json
// @Param gid query string true "群聊ID"
// @Success 200 {object} model.Response{}
// @Router /group/delete [post]
func deleteGroup(c *gin.Context) {
	gid := c.Query("gid")
	if gid == "" {
		response.SetFail(c, "群聊ID不能为空", nil)
		return
	}
	//转换类型
	gidInt, err := strconv.Atoi(gid)
	if gidInt == 0 {
		response.SetFail(c, "群聊ID错误", nil)
	}
	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}
	_, err = groupClient.GetGroupInfoByGid(context.Background(), &api.GetGroupInfoRequest{
		Gid: uint32(gidInt),
	})
	if err != nil {
		c.Error(err)
		return
	}
	sf, err := userGroupClient.GetGroupRelation(context.Background(), &rapi.GetGroupRelationRequest{
		UserId:  thisId,
		GroupId: uint32(gidInt),
	})
	if err != nil {
		c.Error(err)
		return
	}
	if sf.Identity != rapi.GroupIdentity_IDENTITY_ADMIN && sf.Identity != rapi.GroupIdentity_IDENTITY_OWNER {
		c.Error(err)
		return
	}
	//删除对话用户
	_, err = dialogClient.DeleteDialogUsersByDialogID(context.Background(), &rapi.DeleteDialogUsersByDialogIDRequest{
		DialogId: uint32(gidInt),
	})
	if err != nil {
		c.Error(err)
		return
	}
	//删除对话
	_, err = dialogClient.DeleteDialogById(context.Background(), &rapi.DeleteDialogByIdRequest{
		DialogId: uint32(gidInt),
	})
	if err != nil {
		c.Error(err)
		return
	}
	//1.删除群聊成员
	_, err = userGroupClient.DeleteGroupRelationByGroupId(context.Background(), &rapi.GroupIDRequest{
		GroupId: uint32(gidInt),
	})
	if err != nil {
		c.Error(err)
		return
	}
	//2.删除群聊
	groupId, err := groupClient.DeleteGroup(context.Background(), &api.DeleteGroupRequest{
		Gid: uint32(gidInt),
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "删除群聊成功", gin.H{"gid": groupId})
}
