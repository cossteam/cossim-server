package interfaces

import (
	v1 "github.com/cossim/coss-server/internal/relation/api/http/v1"
	"github.com/cossim/coss-server/internal/relation/app/command"
	"github.com/cossim/coss-server/internal/relation/app/query"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/http/response"
	"github.com/gin-gonic/gin"
)

func (h *HttpServer) AddGroupAnnouncement(c *gin.Context, id uint32) {
	req := &v1.AddGroupAnnouncementJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.SetFail(c, "参数错误", nil)
		return
	}

	addGroupAnnouncement, err := h.app.Commands.AddGroupAnnouncement.Handle(c, &command.AddGroupAnnouncement{
		GroupID: id,
		UserID:  c.Value(constants.UserID).(string),
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "添加群组公告成功", AddGroupAnnouncementToResponse(addGroupAnnouncement))
}

func AddGroupAnnouncementToResponse(announcement *command.AddGroupAnnouncementResponse) *v1.GroupAnnouncement {
	return &v1.GroupAnnouncement{
		Content:  announcement.Content,
		CreateAt: announcement.CreateAt,
		GroupId:  announcement.GroupId,
		Id:       announcement.Id,
		OperatorInfo: v1.ShortUserInfo{
			Avatar:   announcement.OperatorInfo.Avatar,
			CossId:   announcement.OperatorInfo.CossID,
			Nickname: announcement.OperatorInfo.Nickname,
			UserId:   announcement.OperatorInfo.UserID,
		},
		Title:    announcement.Title,
		UpdateAt: announcement.UpdateAt,
	}
}

func (h *HttpServer) DeleteGroupAnnouncement(c *gin.Context, id uint32, aid uint32) {
	if err := h.app.Commands.DeleteGroupAnnouncement.Handle(c, &command.DeleteGroupAnnouncement{
		UserID:         c.Value(constants.UserID).(string),
		GroupID:        id,
		AnnouncementID: aid,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "删除群组公告成功", nil)
}

func (h *HttpServer) UpdateGroupAnnouncement(c *gin.Context, id uint32, aid uint32) {
	req := &v1.UpdateGroupAnnouncementJSONRequestBody{}
	if err := c.ShouldBindJSON(req); err != nil {
		response.SetFail(c, "参数错误", nil)
		return
	}

	if err := h.app.Commands.UpdateGroupAnnouncement.Handle(c, &command.UpdateGroupAnnouncement{
		GroupID:        id,
		AnnouncementID: aid,
		UserID:         c.Value(constants.UserID).(string),
		Title:          req.Title,
		Content:        req.Content,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "更新群组公告成功", nil)
}

func (h *HttpServer) ListGroupAnnouncement(c *gin.Context, id uint32) {
	//req := &v1.ListGroupRequestJSONRequestBody{}
	//if err := c.ShouldBindJSON(req); err != nil {
	//	response.SetFail(c, "参数错误", nil)
	//	return
	//}

	listGroupAnnouncement, err := h.app.Queries.ListGroupAnnouncement.Handle(c, &query.ListGroupAnnouncement{
		UserID:  c.Value(constants.UserID).(string),
		GroupID: id,
		//PageNum:  0,
		//PageSize: 0,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取群组公告列表成功", listGroupAnnouncementToResponse(listGroupAnnouncement))
}

func listGroupAnnouncementToResponse(announcement *query.ListGroupAnnouncementResponse) *v1.GroupAnnouncementList {
	resp := &v1.GroupAnnouncementList{
		List: make([]v1.GroupAnnouncement, 0),
	}
	for _, a := range announcement.List {
		resp.List = append(resp.List, mapGroupAnnouncement(a))
	}
	return resp
}

func (h *HttpServer) GetGroupAnnouncement(c *gin.Context, id uint32, aid uint32) {
	getGroupAnnouncement, err := h.app.Queries.GetGroupAnnouncement.Handle(c, &query.GetGroupAnnouncement{
		UserID:         c.Value(constants.UserID).(string),
		GroupID:        id,
		AnnouncementID: aid,
		//PageNum:        0,
		//PageSize:       0,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取群组公告成功", mapGroupAnnouncement(getGroupAnnouncement))
}

func (h *HttpServer) SetGroupAnnouncementRead(c *gin.Context, id uint32, aid uint32) {
	if err := h.app.Commands.SetGroupAnnouncementRead.Handle(c, &command.SetGroupAnnouncementRead{
		UserID:         c.Value(constants.UserID).(string),
		GroupID:        id,
		AnnouncementID: aid,
	}); err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "设置群组公告已读成功", nil)
}

func (h *HttpServer) ListGroupAnnouncementRead(c *gin.Context, id uint32, aid uint32) {
	listGroupAnnouncementRead, err := h.app.Queries.ListGroupAnnouncementRead.Handle(c, &query.ListGroupAnnouncementRead{
		UserID:         c.Value(constants.UserID).(string),
		GroupID:        id,
		AnnouncementID: aid,
		//PageNum:        0,
		//PageSize:       0,
	})
	if err != nil {
		c.Error(err)
		return
	}

	response.SetSuccess(c, "获取群组公告已读列表成功", listGroupAnnouncementReadToResponse(listGroupAnnouncementRead))
}

func listGroupAnnouncementReadToResponse(read *query.ListGroupAnnouncementReadResponse) *v1.GroupAnnouncementReaderList {
	resp := &v1.GroupAnnouncementReaderList{}

	for _, r := range read.List {
		resp.List = append(resp.List, mapGroupAnnouncementReadUser(r))
	}
	return resp
}

func mapGroupAnnouncementReadUser(r *query.GroupAnnouncementReadUser) v1.GroupAnnouncementReadUser {
	return v1.GroupAnnouncementReadUser{
		AnnouncementId: r.AnnouncementID,
		GroupId:        r.GroupID,
		Id:             r.ID,
		ReadAt:         r.ReadAt,
		ReaderInfo: &v1.ShortUserInfo{
			Avatar:   r.ReaderInfo.Avatar,
			CossId:   r.ReaderInfo.CossID,
			Nickname: r.ReaderInfo.Nickname,
			UserId:   r.ReaderInfo.UserID,
		},
		UserId: r.UserID,
	}
}

func mapGroupAnnouncement(a *query.GroupAnnouncement) v1.GroupAnnouncement {
	var readUserList []v1.GroupAnnouncementReadUser
	if a.ReadUserList != nil {
		for _, r := range a.ReadUserList {
			readUserList = append(readUserList, mapGroupAnnouncementReadUser(r))
		}
	}
	resp := v1.GroupAnnouncement{
		Content:      a.Content,
		CreateAt:     a.CreateAt,
		GroupId:      a.GroupID,
		Id:           a.ID,
		ReadUserList: readUserList,
		Title:        a.Title,
		UpdateAt:     a.UpdateAt,
		OperatorInfo: v1.ShortUserInfo{
			Avatar:   a.OperatorInfo.Avatar,
			Nickname: a.OperatorInfo.Nickname,
			UserId:   a.OperatorInfo.UserID,
			CossId:   a.OperatorInfo.CossID,
		},
	}

	return resp
}
