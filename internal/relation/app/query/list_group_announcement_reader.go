package query

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type ListGroupAnnouncementRead struct {
	UserID         string
	GroupID        uint32
	AnnouncementID uint32
	PageNum        int
	PageSize       int
}

func (cmd *ListGroupAnnouncementRead) Validate() error {
	if cmd == nil || cmd.UserID == "" || cmd.GroupID == 0 {
		return code.InvalidParameter
	}
	return nil
}

type ListGroupAnnouncementReadResponse struct {
	List []*GroupAnnouncementReadUser
	//Total int64
	//Page  int32
}

type ListGroupAnnouncementReadHandler decorator.CommandHandler[*ListGroupAnnouncementRead, *ListGroupAnnouncementReadResponse]

func NewListGroupAnnouncementReadHandler(
	logger *zap.Logger,
	grd service.GroupRelationDomain,
	groupAnnouncementDomain service.GroupAnnouncementDomain,
	userService rpc.UserService,
) ListGroupAnnouncementReadHandler {
	return &listGroupAnnouncementReadHandler{
		logger:                  logger,
		grd:                     grd,
		groupAnnouncementDomain: groupAnnouncementDomain,
		userService:             userService,
	}
}

type listGroupAnnouncementReadHandler struct {
	logger *zap.Logger

	grd                     service.GroupRelationDomain
	groupAnnouncementDomain service.GroupAnnouncementDomain

	userService rpc.UserService
}

func (h *listGroupAnnouncementReadHandler) Handle(ctx context.Context, cmd *ListGroupAnnouncementRead) (*ListGroupAnnouncementReadResponse, error) {
	if err := h.grd.IsUserInActiveGroup(ctx, cmd.GroupID, cmd.UserID); err != nil {
		h.logger.Error("IsUserInActiveGroup", zap.Error(err))
		return nil, err
	}

	reads, err := h.groupAnnouncementDomain.ListAnnouncementRead(ctx, cmd.GroupID, cmd.AnnouncementID)
	if err != nil {
		h.logger.Error("ListAnnouncementRead", zap.Error(err))
		return nil, err
	}

	list, err := buildReadUserList(ctx, h.userService, reads)
	if err != nil {
		return nil, err
	}

	resp := &ListGroupAnnouncementReadResponse{
		List: list,
	}

	return resp, nil
}
