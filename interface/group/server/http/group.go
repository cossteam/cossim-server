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
// @Param group_id query int32 true "群聊ID"
// @Success 200 {object} model.Response{}
// @Router /group/info [get]
func (h *Handler) getGroupInfoByGid(c *gin.Context) {
	gid := c.Query("group_id")
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

	userID, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	resp, err := h.svc.GetGroupInfoByGid(c, uint32(gidInt), userID)
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
// @Param group_ids query []string true "群聊ID列表"
// @Success 200 {object} model.Response{}
// @Router /group/getBatch [get]
func (h *Handler) getBatchGroupInfoByIDs(c *gin.Context) {
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

	resp, err := h.svc.GetBatchGroupInfoByIDs(c, ids)
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
// @Router /group/update/ [post]
func (h *Handler) updateGroup(c *gin.Context) {
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

	resp, err := h.svc.UpdateGroup(c, req, userID)
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
func (h *Handler) createGroup(c *gin.Context) {
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
		Type:      api.GroupType(int32(req.Type)),
		CreatorId: thisId,
		Name:      req.Name,
		Avatar:    req.Avatar,
		Member:    req.Member,
	}
	if len(req.Member) == 0 {
		group.Member = make([]string, 0)
	}

	switch group.Type {
	case api.GroupType_TypeEncrypted:
		group.MaxMembersLimit = model.EncryptedGroup
	default:
		group.MaxMembersLimit = model.DefaultGroup
	}

	resp, err := h.svc.CreateGroup(c, group)
	if err != nil {
		h.logger.Error("创建群聊失败", zap.Error(err))
		response.SetFail(c, code.Cause(err).Message(), nil)
		return
	}
	response.Success(c, "创建群聊成功", resp)
}

// @Summary 删除群聊
// @Description 删除群聊
// @Accept  json
// @Produce  json
// @Param request body model.DeleteGroupRequest true "群聊id"
// @Success 200 {object} model.Response{}
// @Router /group/delete [post]
func (h *Handler) deleteGroup(c *gin.Context) {
	req := new(model.DeleteGroupRequest)
	if err := c.ShouldBindJSON(&req); err != nil {
		response.SetFail(c, "参数验证失败", nil)
		return
	}

	thisId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	groupId, err := h.svc.DeleteGroup(c, req.GroupId, thisId)
	if err != nil {
		h.logger.Error("删除群聊失败", zap.Error(err))
		response.SetFail(c, "删除群聊失败", nil)
		return
	}

	response.SetSuccess(c, "删除群聊成功", gin.H{"group_id": groupId})
}

// @Summary 修改群聊头像
// @Description 修改群聊头像
// @Accept  json
// @Produce  json
// @param file formData file true "头像文件"
// @param group_id formData int64 true "群聊id"
// @Success		200 {object} model.Response{}
// @Router /group/avatar/modify [post]
func (h *Handler) modifyGroupAvatar(c *gin.Context) {
	gid, _ := c.GetPostForm("group_id")

	if gid == "" {
		response.SetFail(c, "群聊ID不能为空", nil)
		return
	}

	groupId, err := strconv.Atoi(gid)
	if err != nil {
		response.SetFail(c, "群聊ID错误", nil)
		return
	}

	userId, err := pkghttp.ParseTokenReUid(c)
	if err != nil {
		response.SetFail(c, err.Error(), nil)
		return
	}

	// Parse form data
	if err := c.Request.ParseMultipartForm(25 << 20); // 25 MB limit
	err != nil {
		response.SetFail(c, "Failed to parse form data", nil)
		return
	}

	// Get the file from the form data
	file, handler, err := c.Request.FormFile("file")
	if err != nil {
		response.SetFail(c, "Error retrieving the file", nil)
		return
	}
	defer file.Close()

	// Check file type
	contentType := handler.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" {
		response.SetFail(c, "Unsupported file type. Only JPEG and PNG are allowed.", nil)
		return
	}

	// Check file size
	if handler.Size > 25<<20 { // 25 MB limit
		response.SetFail(c, "File size exceeds the limit. Maximum allowed size is 25 MB.", nil)
		return
	}

	url, err := h.svc.ModifyGroupAvatar(c, userId, uint32(groupId), file)
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "修改成功", gin.H{"group_id": userId, "avatar": url})
}
