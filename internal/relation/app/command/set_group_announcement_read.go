package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type SetGroupAnnouncementRead struct {
	UserID         string
	GroupID        uint32
	AnnouncementID uint32
}

func (cmd *SetGroupAnnouncementRead) Validate() error {
	if cmd == nil || cmd.UserID == "" || cmd.GroupID == 0 || cmd.AnnouncementID == 0 {
		return code.InvalidParameter
	}
	return nil
}

type SetGroupAnnouncementReadHandler decorator.CommandHandlerNoneResponse[*SetGroupAnnouncementRead]

func NewSetGroupAnnouncementReadHandler(
	logger *zap.Logger,
	groupRelationDomain service.GroupRelationDomain,
	groupAnnouncementDomain service.GroupAnnouncementDomain,
) SetGroupAnnouncementReadHandler {
	return &setGroupAnnouncementReadHandler{
		logger:              logger,
		groupRelationDomain: groupRelationDomain,
	}
}

type setGroupAnnouncementReadHandler struct {
	logger                  *zap.Logger
	groupRelationDomain     service.GroupRelationDomain
	groupAnnouncementDomain service.GroupAnnouncementDomain
}

func (h *setGroupAnnouncementReadHandler) Handle(ctx context.Context, cmd *SetGroupAnnouncementRead) error {
	//if err := cmd.Validate(); err != nil {
	//	return err
	//}

	if err := h.groupRelationDomain.IsUserInActiveGroup(ctx, cmd.GroupID, cmd.UserID); err != nil {
		h.logger.Error("user is not in group", zap.Error(err))
		return err
	}

	if err := h.groupAnnouncementDomain.SetAnnouncementRead(ctx, cmd.GroupID, cmd.AnnouncementID, cmd.UserID); err != nil {
		h.logger.Error("set announcement read failed", zap.Error(err))
		return err
	}

	return nil
}
