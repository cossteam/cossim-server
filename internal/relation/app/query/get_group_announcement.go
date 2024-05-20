package query

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type GetGroupAnnouncement struct {
	UserID         string
	GroupID        uint32
	AnnouncementID uint32
	PageNum        int
	PageSize       int
}

func (cmd *GetGroupAnnouncement) Validate() error {
	if cmd == nil || cmd.UserID == "" || cmd.GroupID == 0 {
		return code.InvalidParameter
	}
	return nil
}

type GetGroupAnnouncementHandler decorator.CommandHandler[*GetGroupAnnouncement, *GroupAnnouncement]

func NewGetGroupAnnouncementHandler(
	logger *zap.Logger,
	grd service.GroupRelationDomain,
	groupAnnouncementDomain service.GroupAnnouncementDomain,
	userService rpc.UserService,
) GetGroupAnnouncementHandler {
	return &getGroupAnnouncementHandler{
		logger:                  logger,
		grd:                     grd,
		groupAnnouncementDomain: groupAnnouncementDomain,
		userService:             userService,
	}
}

type getGroupAnnouncementHandler struct {
	logger *zap.Logger

	grd                     service.GroupRelationDomain
	groupAnnouncementDomain service.GroupAnnouncementDomain

	userService rpc.UserService
}

func (h *getGroupAnnouncementHandler) Handle(ctx context.Context, cmd *GetGroupAnnouncement) (*GroupAnnouncement, error) {
	if err := h.grd.IsUserInActiveGroup(ctx, cmd.GroupID, cmd.UserID); err != nil {
		h.logger.Error("IsUserInActiveGroup", zap.Error(err))
		return nil, err
	}

	announcement, err := h.groupAnnouncementDomain.GetAnnouncement(ctx, cmd.AnnouncementID)
	if err != nil {
		if errors.Is(err, code.NotFound) {
			return nil, code.RelationGroupErrGroupAnnouncementNotFoundFailed
		}
		h.logger.Error("GetAnnouncement", zap.Error(err))
		return nil, err
	}

	operatorInfo, err := h.userService.GetUserInfo(ctx, announcement.UserID)
	if err != nil {
		h.logger.Error("GetUserInfo", zap.Error(err))
		return nil, err
	}

	readers, err := h.groupAnnouncementDomain.ListAnnouncementRead(ctx, announcement.GroupID, announcement.ID)
	if err != nil {
		h.logger.Error("ListAnnouncementRead", zap.Error(err))
		return nil, err
	}

	readUserList, err := buildReadUserList(ctx, h.userService, readers)
	if err != nil {
		h.logger.Error("buildReadUserList", zap.Error(err))
		return nil, err
	}

	return &GroupAnnouncement{
		ID:       announcement.ID,
		GroupID:  announcement.GroupID,
		Title:    announcement.Title,
		Content:  announcement.Content,
		CreateAt: announcement.CreatedAt,
		UpdateAt: announcement.UpdatedAt,
		OperatorInfo: &ShortUserInfo{
			UserID:   operatorInfo.ID,
			CossID:   operatorInfo.CossID,
			Nickname: operatorInfo.Nickname,
			Avatar:   operatorInfo.Avatar,
		},
		ReadUserList: readUserList,
	}, nil
}
