package command

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type DeleteGroupAnnouncement struct {
	UserID         string
	GroupID        uint32
	AnnouncementID uint32
}

func (cmd *DeleteGroupAnnouncement) Validate() error {
	if cmd == nil || cmd.UserID == "" || cmd.GroupID == 0 {
		return code.InvalidParameter
	}
	return nil
}

type DeleteGroupAnnouncementHandler decorator.CommandHandlerNoneResponse[*DeleteGroupAnnouncement]

func NewDeleteGroupAnnouncementHandler(
	logger *zap.Logger,
	grd service.GroupRelationDomain,
	groupAnnouncementDomain service.GroupAnnouncementDomain,
) DeleteGroupAnnouncementHandler {
	return &deleteGroupAnnouncementHandler{
		logger:                  logger,
		grd:                     grd,
		groupAnnouncementDomain: groupAnnouncementDomain,
	}
}

type deleteGroupAnnouncementHandler struct {
	logger *zap.Logger

	grd                     service.GroupRelationDomain
	groupAnnouncementDomain service.GroupAnnouncementDomain
}

func (h *deleteGroupAnnouncementHandler) Handle(ctx context.Context, cmd *DeleteGroupAnnouncement) error {
	if err := h.grd.IsUserInActiveGroup(ctx, cmd.GroupID, cmd.UserID); err != nil {
		h.logger.Error("IsUserInActiveGroup", zap.Error(err))
		return err
	}

	isAdmin, err := h.grd.IsUserGroupAdmin(ctx, cmd.GroupID, cmd.UserID)
	if err != nil {
		h.logger.Error("IsUserGroupAdmin", zap.Error(err))
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

	if err := h.groupAnnouncementDomain.Delete(ctx, cmd.AnnouncementID); err != nil {
		if errors.Is(err, code.NotFound) {
			return code.RelationGroupErrGroupAnnouncementNotFoundFailed
		}
		h.logger.Error("GetAnnouncement", zap.Error(err))
		return err
	}

	return nil
}
