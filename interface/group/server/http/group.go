package http

import (
	"context"
	pkghttp "github.com/cossim/coss-server/pkg/http"
	"github.com/cossim/coss-server/pkg/http/response"
	api "github.com/cossim/coss-server/service/group/api/v1"
	msgApi "github.com/cossim/coss-server/service/msg/api/v1"
	rapi "github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/gin-gonic/gin"
	"strconv"
)

// @Summary 获取群聊信息
// @Description 获取群聊信息
// @Accept  json
// @Produce  json
// @Param gid query int32 true "群聊ID"
// @Success 200 {object} utils.Response{}
// @Router /group/info [get]
func GetGroupInfoByGid(c *gin.Context) {
	gid := c.Query("gid")
	if gid == "" {
		response.Fail(c, "群聊ID不能为空", nil)
		return
	}
	//转换类型
	gidInt, err := strconv.Atoi(gid)
	if err != nil {
		response.Fail(c, "群聊ID错误", nil)
		return
	}
	group, err := groupClient.GetGroupInfoByGid(c, &api.GetGroupInfoRequest{
		Gid: uint32(gidInt),
	})
	if err != nil {
		response.Fail(c, "获取群聊信息失败", nil)
		return
	}

	response.Success(c, "获取群聊信息成功", group)
}

// @Summary 批量获取群聊信息
// @Description 批量获取群聊信息
// @Accept  json
// @Produce  json
// @Param groupIds query []string true "群聊ID列表"
// @Success 200 {object} utils.Response{}
// @Router /group/getBatch [get]
func getBatchGroupInfoByIDs(c *gin.Context) {
	groupIds := c.QueryArray("groupIds")
	ids := make([]uint32, len(groupIds))
	//转换类型
	for i, groupId := range groupIds {
		id, err := strconv.Atoi(groupId)
		if err != nil {
			response.Fail(c, "群聊ID列表转换失败", nil)
			return
		}
		ids[i] = uint32(id)
	}

	groups, err := groupClient.GetBatchGroupInfoByIDs(c, &api.GetBatchGroupInfoRequest{
		GroupIds: ids,
	})
	if err != nil {
		response.Fail(c, "批量获取群聊信息失败", nil)
		return
	}
	response.Success(c, "批量获取群聊信息成功", gin.H{"groups": groups})
}

type updateGroupRequest struct {
	Type            uint32 `json:"type"`
	Status          uint32 `json:"status"`
	MaxMembersLimit uint32 `json:"max_members_limit"`
	CreatorID       string `json:"creator_id"`
	Name            string `json:"name"`
	Avatar          string `json:"avatar"`
	GroupId         uint32 `json:"group_id"`
}

// @Summary 更新群聊信息
// @Description 更新群聊信息
// @Accept  json
// @Produce  json
// @Param request body updateGroupRequest true "请求体"
// @Success 200 {object} utils.Response{}
// @Router /group/update/{gid} [post]
func updateGroup(c *gin.Context) {
	req := new(updateGroupRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数验证失败", nil)
		return
	}

	group, err := groupClient.GetGroupInfoByGid(context.Background(), &api.GetGroupInfoRequest{
		Gid: req.GroupId,
	})
	if err != nil {
		response.Fail(c, "未找到对应的群聊", nil)
		return
	}
	//todo 判断权限

	// 更新群聊信息
	group.Type = int32(req.Type)
	group.Status = int32(req.Status)
	group.MaxMembersLimit = int32(req.MaxMembersLimit)
	group.CreatorId = req.CreatorID
	group.Name = req.Name
	//todo 上传头像
	group.Avatar = req.Avatar

	updatedGroup, err := groupClient.UpdateGroup(context.Background(), &api.UpdateGroupRequest{
		Group: group,
	})
	if err != nil {
		response.Fail(c, "更新群聊信息失败", nil)
		return
	}

	response.Success(c, "更新群聊信息成功", gin.H{"group": updatedGroup})
}

type createGroupRequest struct {
	Type            uint32 `json:"type"`
	MaxMembersLimit uint32 `json:"max_members_limit"`
	Name            string `json:"name" binding:"required"`
	Avatar          string `json:"avatar"`
}

// @Summary 创建群聊
// @Description 创建群聊
// @Accept  json
// @Produce  json
// @Param request body createGroupRequest true "请求体"
// @Success 200 {object} utils.Response{}
// @Router /group/create [post]
func createGroup(c *gin.Context) {
	req := new(createGroupRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数验证失败", nil)
		return
	}

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.Fail(c, err.Error(), nil)
		return
	}
	group := &api.Group{
		Type:            int32(req.Type),
		MaxMembersLimit: int32(req.MaxMembersLimit),
		CreatorId:       thisId,
		Name:            req.Name,
		Avatar:          req.Avatar,
	}

	createdGroup, err := groupClient.InsertGroup(context.Background(), &api.InsertGroupRequest{
		Group: group,
	})
	if err != nil {
		response.Fail(c, "创建群聊失败", nil)
		return
	}

	_, err = userGroupClient.JoinGroup(context.Background(), &rapi.JoinGroupRequest{
		GroupId: createdGroup.Id,
		UserId:  thisId,
	})
	if err != nil {
		c.Error(err)
		return
	}
	//创建对话
	dialog, err := dialogClient.CreateDialog(context.Background(), &msgApi.CreateDialogRequest{OwnerId: thisId, Type: 0, GroupId: createdGroup.Id})
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

	response.Success(c, "创建群聊成功", createdGroup)
}

// @Summary 删除群聊
// @Description 删除群聊
// @Accept  json
// @Produce  json
// @Param gid query string true "群聊ID"
// @Success 200 {object} utils.Response{}
// @Router /group/delete [post]
func deleteGroup(c *gin.Context) {
	gid := c.Query("gid")
	if gid == "" {
		response.Fail(c, "群聊ID不能为空", nil)
		return
	}
	//转换类型
	gidInt, err := strconv.Atoi(gid)
	if gidInt == 0 {
		response.Fail(c, "群聊ID错误", nil)
	}
	//todo 判断是否群成员
	//1.删除群聊成员
	//2.删除群聊
	groupId, err := groupClient.DeleteGroup(context.Background(), &api.DeleteGroupRequest{
		Gid: uint32(gidInt),
	})
	if err != nil {
		response.Fail(c, "删除群聊失败", nil)
		return
	}

	response.Success(c, "删除群聊成功", gin.H{"gid": groupId})
}
