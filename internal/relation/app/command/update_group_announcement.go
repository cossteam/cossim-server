package command

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type UpdateGroupAnnouncement struct {
	GroupID        uint32
	AnnouncementID uint32
	UserID         string
	Title          string
	Content        string
}

func (cmd *UpdateGroupAnnouncement) Validate() error {
	if cmd == nil || cmd.UserID == "" || cmd.GroupID == 0 || cmd.Title == "" {
		return code.InvalidParameter
	}
	return nil
}

type UpdateGroupAnnouncementHandler decorator.CommandHandlerNoneResponse[*UpdateGroupAnnouncement]

func NewUpdateGroupAnnouncementHandler(
	logger *zap.Logger,
	groupRelationDomain service.GroupRelationDomain,
	groupAnnouncementDomain service.GroupAnnouncementDomain,
) UpdateGroupAnnouncementHandler {
	return &updateGroupAnnouncementHandler{
		logger:                  logger,
		groupRelationDomain:     groupRelationDomain,
		groupAnnouncementDomain: groupAnnouncementDomain,
	}
}

type updateGroupAnnouncementHandler struct {
	logger                  *zap.Logger
	groupRelationDomain     service.GroupRelationDomain
	groupAnnouncementDomain service.GroupAnnouncementDomain
}

func (h *updateGroupAnnouncementHandler) Handle(ctx context.Context, cmd *UpdateGroupAnnouncement) error {
	//if err := cmd.Validate(); err != nil {
	//	return err
	//}

	if err := h.groupRelationDomain.IsUserInActiveGroup(ctx, cmd.GroupID, cmd.UserID); err != nil {
		h.logger.Error("IsUserInActiveGroup", zap.Error(err))
		return err
	}

	isAdmin, err := h.groupRelationDomain.IsUserGroupAdmin(ctx, cmd.GroupID, cmd.UserID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return code.Forbidden
	}

	exist, err := h.groupAnnouncementDomain.IsExist(ctx, cmd.AnnouncementID)
	if err != nil {
		h.logger.Error("IsExist", zap.Error(err))
		return err
	}
	if !exist {
		return code.RelationGroupErrGroupAnnouncementNotFoundFailed
	}

	if err := h.groupAnnouncementDomain.Update(ctx, &service.UpdateGroupAnnouncement{
		ID:      cmd.AnnouncementID,
		Title:   cmd.Title,
		Content: cmd.Content,
	}); err != nil {
		if errors.Is(err, code.NotFound) {
			return code.RelationGroupErrGroupAnnouncementNotFoundFailed
		}
		h.logger.Error("Update", zap.Error(err))
		return err
	}

	return nil
}
