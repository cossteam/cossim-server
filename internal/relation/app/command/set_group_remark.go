package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type SetGroupRemark struct {
	GroupID uint32
	UserID  string
	Remark  string
}

func (cmd *SetGroupRemark) Validate() error {
	if cmd == nil || cmd.UserID == "" || cmd.GroupID == 0 {
		return code.InvalidParameter
	}
	return nil
}

type SetGroupRemarkHandler decorator.CommandHandlerNoneResponse[*SetGroupRemark]

func NewSetGroupRemarkHandler(
	logger *zap.Logger,
	groupRelationDomain service.GroupRelationDomain,
	groupAnnouncementDomain service.GroupAnnouncementDomain,
) SetGroupRemarkHandler {
	return &setGroupRemarkHandler{
		logger:                  logger,
		groupRelationDomain:     groupRelationDomain,
		groupAnnouncementDomain: groupAnnouncementDomain,
	}
}

type setGroupRemarkHandler struct {
	logger                  *zap.Logger
	groupRelationDomain     service.GroupRelationDomain
	groupAnnouncementDomain service.GroupAnnouncementDomain
}

func (h *setGroupRemarkHandler) Handle(ctx context.Context, cmd *SetGroupRemark) error {
	//if err := cmd.Validate(); err != nil {
	//	return err
	//}

	if err := h.groupRelationDomain.IsUserInActiveGroup(ctx, cmd.GroupID, cmd.UserID); err != nil {
		h.logger.Error("IsUserInActiveGroup", zap.Error(err))
		return err
	}

	if err := h.groupRelationDomain.SetGroupRemark(ctx, cmd.GroupID, cmd.UserID, cmd.Remark); err != nil {
		h.logger.Error("set group remark error", zap.Error(err))
		return err
	}

	return nil
}
