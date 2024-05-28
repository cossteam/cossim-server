package http

import (
	v1 "github.com/cossim/coss-server/internal/group/api/http/v1"
	"github.com/cossim/coss-server/internal/group/app/command"
	"github.com/cossim/coss-server/internal/group/app/query"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
)

// SearchGroup
// @Summary 搜索群聊
// @Description 搜索群聊
// @Tags group
// @Accept json
// @Produce json
// @Param Authorization header string true "Authentication header"
// @Param keyword query string false "关键词"
// @Param page_num query int false "页码"
// @Param page_size query int false "页大小"
// @Security ApiKeyAuth
// @Success 200 {object} v1.Group "成功创建群聊"
// @Router /api/v1/group/search [get]
func (h *HttpServer) SearchGroup(c *gin.Context, params v1.SearchGroupParams) {
	var page, pageSize int32 = 1, 10
	if params.Page != nil {
		page = *params.Page
	}
	if params.PageSize != nil {
		pageSize = *params.PageSize
	}

	searchGroup, err := h.app.Queries.SearchGroup.Handle(c, &query.SearchGroup{
		UserID:   c.Value(constants.UserID).(string),
		Keyword:  params.Keyword,
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		c.Error(err)
		return
	}
	response.SetSuccess(c, "搜索群聊成功", searchGroupToResponse(searchGroup))
}

func searchGroupToResponse(group []*query.Group) []*v1.Group {
	var resp []*v1.Group
	for _, g := range group {
		resp = append(resp, &v1.Group{
			Avatar:          g.Avatar,
			Id:              g.Id,
			MaxMembersLimit: g.MaxMembersLimit,
			Member:          g.Member,
			Name:            g.Name,
		})
	}

	return resp
}

// CreateGroup
// @Summary 创建群聊
// @Description 创建一个新的群聊
// @Tags group
// @Produce json
// @Param Authorization header string true "Authentication header"
// @Param keyword path integer true "群聊名称或ID" string
// @Security ApiKeyAuth
// @Success 200 {object} v1.Group "成功创建群聊"
// @Router /api/v1/group/search [get]
func (h *HttpServer) CreateGroup(c *gin.Context) {
	req := &v1.CreateGroupRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.Error(err)
		return
	}

	uid := c.Value(constants.UserID).(string)
	createGroup, err := h.app.Commands.CreateGroup.Handle(c, command.CreateGroup{
		CreateID:    uid,
		Name:        req.Name,
		Avatar:      req.Avatar,
		Type:        uint(req.Type),
		Member:      req.Member,
		Encrypt:     req.Encrypt,
		JoinApprove: req.JoinApprove,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "创建群聊成功", createGroupToResponse(&createGroup))
}

func createGroupToResponse(createGroup *command.CreateGroupResponse) *v1.CreateGroupResponse {
	return &v1.CreateGroupResponse{
		Avatar:          createGroup.Avatar,
		CreatorId:       createGroup.CreatorID,
		DialogId:        createGroup.DialogID,
		Id:              createGroup.ID,
		MaxMembersLimit: createGroup.MaxMembersLimit,
		Name:            createGroup.Name,
		Status:          createGroup.Status,
		Type:            createGroup.Type,
	}
}

// DeleteGroup
// @Summary 删除群聊
// @Description 群主解散创建的群聊
// @Tags group
// @Accept json
// @Produce json
// @Param id path integer true "要删除的群聊ID" format(uint32)
// @Security ApiKeyAuth
// @Success 200 {string} string "成功删除群聊"
// @Router /api/v1/group/{id} [delete]
func (h *HttpServer) DeleteGroup(c *gin.Context, id uint32) {
	uid := c.Value(constants.UserID).(string)
	resp, err := h.app.Commands.DeleteGroup.Handle(c, command.DeleteGroup{
		ID:     id,
		UserID: uid,
	})
	if err != nil {
		c.Error(err)
		return
	}
	response.SetSuccess(c, "删除群聊成功", resp)
}

// GetGroup
// @Summary 获取群聊信息
// @Description 根据群聊ID获取群聊的详细信息
// @Tags group
// @Param id path integer true "群聊ID" format(uint32)
// @Success 200 {object} v1.GroupInfo "返回群聊信息"
// @Router /api/v1/group/{id} [get]
func (h *HttpServer) GetGroup(c *gin.Context, id uint32) {
	getGroup, err := h.app.Queries.GetGroup.Handle(c, query.GetGroup{
		ID:     id,
		UserID: c.Value(constants.UserID).(string),
	})
	if err != nil {
		c.Error(err)
		return
	}
	response.SetSuccess(c, "获取群聊信息成功", getGroupToResponse(getGroup))
}

func getGroupToResponse(getGroup *query.GroupInfo) *v1.GroupInfo {
	return &v1.GroupInfo{
		Avatar:          getGroup.Avatar,
		CreatorId:       getGroup.CreatorId,
		DialogId:        getGroup.DialogId,
		Id:              getGroup.Id,
		MaxMembersLimit: int(getGroup.MaxMembersLimit),
		Name:            getGroup.Name,
		Status:          getGroup.Status,
		Type:            uint8(getGroup.Type),
		SilenceTime:     getGroup.SilenceTime,
		Encrypt:         getGroup.Encrypt,
		JoinApprove:     getGroup.JoinApprove,
		Preferences: &v1.Preferences{
			EntryMethod: getGroup.Preferences.EntryMethod,
			Identity:    getGroup.Preferences.Identity,
			Inviter:     getGroup.Preferences.Inviter,
			JoinedAt:    getGroup.Preferences.JoinedAt,
			MuteEndTime: getGroup.Preferences.MuteEndTime,
			Remark:      getGroup.Preferences.Remark,
			Silent:      getGroup.Preferences.SilentNotification,
		},
	}
}

// UpdateGroup
// @Summary 更新群聊信息
// @Description 更新现有群聊的信息
// @Tags group
// @Accept json
// @Produce json
// @Param id path uint32 true "要更新的群聊ID"
// @Param requestBody body v1.UpdateGroupRequest true "更新现有群聊的信息"
// @Success 200 {object} string "成功更新群聊信息"
// @Router /api/v1/group/{id} [put]
func (h *HttpServer) UpdateGroup(c *gin.Context, id uint32) {
	req := &v1.UpdateGroupRequest{}
	if err := c.ShouldBindJSON(req); err != nil {
		c.Error(err)
		return
	}

	uid := c.Value(constants.UserID).(string)
	var t *uint
	if req.Type != nil {
		tt := uint(*req.Type)
		t = &tt
	}
	resp, err := h.app.Commands.UpdateGroup.Handle(c, command.UpdateGroup{
		ID:          id,
		Type:        t,
		UserID:      uid,
		Name:        req.Name,
		Avatar:      req.Avatar,
		SilenceTime: req.SilenceTime,
		Encrypt:     req.Encrypt,
		JoinApprove: req.JoinApprove,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "更新群聊信息成功", resp)
}

func (h *HttpServer) GetAllGroup(c *gin.Context) {

}
