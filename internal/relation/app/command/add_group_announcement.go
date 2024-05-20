package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type AddGroupAnnouncement struct {
	GroupID uint32
	UserID  string
	Title   string
	Content string
}

func (cmd *AddGroupAnnouncement) Validate() error {
	if cmd == nil || cmd.UserID == "" || cmd.GroupID == 0 || cmd.Title == "" {
		return code.InvalidParameter
	}
	return nil
}

type AddGroupAnnouncementResponse struct {
	Id           uint32
	GroupId      uint32
	Title        string
	Content      string
	CreateAt     int64
	UpdateAt     int64
	OperatorInfo *ShortUserInfo
}

type ShortUserInfo struct {
	UserID   string
	CossID   string
	Avatar   string
	Nickname string
}
type AddGroupAnnouncementHandler decorator.CommandHandler[*AddGroupAnnouncement, *AddGroupAnnouncementResponse]

func NewAddGroupAnnouncementHandler(
	logger *zap.Logger,
	groupRelationDomain service.GroupRelationDomain,
	groupAnnouncementDomain service.GroupAnnouncementDomain,
	userService rpc.UserService,
) AddGroupAnnouncementHandler {
	return &addGroupAnnouncementHandler{
		logger:                  logger,
		groupRelationDomain:     groupRelationDomain,
		groupAnnouncementDomain: groupAnnouncementDomain,
		userService:             userService,
	}
}

type addGroupAnnouncementHandler struct {
	logger                  *zap.Logger
	groupRelationDomain     service.GroupRelationDomain
	groupAnnouncementDomain service.GroupAnnouncementDomain
	userService             rpc.UserService
}

func (h *addGroupAnnouncementHandler) Handle(ctx context.Context, cmd *AddGroupAnnouncement) (*AddGroupAnnouncementResponse, error) {
	//if err := cmd.Validate(); err != nil {
	//	return err
	//}

	if err := h.groupRelationDomain.IsUserInActiveGroup(ctx, cmd.GroupID, cmd.UserID); err != nil {
		h.logger.Error("IsUserInActiveGroup", zap.Error(err))
		return nil, err
	}

	isAdmin, err := h.groupRelationDomain.IsUserGroupAdmin(ctx, cmd.GroupID, cmd.UserID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, code.Forbidden
	}

	ann, err := h.groupAnnouncementDomain.Add(ctx, &service.AddGroupAnnouncement{
		GroupID: cmd.GroupID,
		UserID:  cmd.UserID,
		Title:   cmd.Title,
		Content: cmd.Content,
	})
	if err != nil {
		h.logger.Error("AddGroupAnnouncementDomain", zap.Error(err))
		return nil, err
	}

	operatorInfo, err := h.userService.GetUserInfo(ctx, ann.UserID)
	if err != nil {
		return nil, err
	}

	return &AddGroupAnnouncementResponse{
		Id:       ann.ID,
		GroupId:  ann.GroupID,
		Title:    ann.Title,
		Content:  ann.Content,
		CreateAt: ann.CreatedAt,
		UpdateAt: ann.UpdatedAt,
		OperatorInfo: &ShortUserInfo{
			UserID:   operatorInfo.Avatar,
			CossID:   operatorInfo.CossID,
			Avatar:   operatorInfo.Avatar,
			Nickname: operatorInfo.Nickname,
		},
	}, nil
}
