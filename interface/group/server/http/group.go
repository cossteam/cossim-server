package http

import (
	"github.com/cossim/coss-server/interface/group/api/model"
	"github.com/cossim/coss-server/pkg/code"
	pkghttp "github.com/cossim/coss-server/pkg/http"
	"github.com/cossim/coss-server/pkg/http/response"
	api "github.com/cossim/coss-server/service/group/api/v1"
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
func getGroupInfoByGid(c *gin.Context) {
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

	resp, err := svc.GetGroupInfoByGid(c, uint32(gidInt))
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取群聊信息成功", resp)
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

	resp, err := svc.GetBatchGroupInfoByIDs(c, ids)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "批量获取群聊信息成功", resp)
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

	if !model.IsValidGroupType(api.GroupType(req.Type)) {
		response.SetFail(c, "群聊类型错误", nil)
	}

	userID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	resp, err := svc.UpdateGroup(c, req, userID)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "更新群聊信息成功", gin.H{"group": resp})
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
		Member:          req.Member,
	}

	resp, err := svc.CreateGroup(c, group)
	if err != nil {
		logger.Error("创建群聊失败", zap.Error(err))
		response.SetFail(c, code.Cause(err).Message(), nil)
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

	groupId, err := svc.DeleteGroup(c, uint32(gidInt), thisId)
	if err != nil {
		logger.Error("删除群聊失败", zap.Error(err))
		response.SetFail(c, "删除群聊失败", nil)
		return
	}

	response.SetSuccess(c, "删除群聊成功", gin.H{"gid": groupId})
}
