package query

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type ListGroupMember struct {
	UserID   string
	GroupID  uint32
	PageNum  int
	PageSize int
}

func (cmd *ListGroupMember) Validate() error {
	if cmd == nil || cmd.UserID == "" || cmd.GroupID == 0 {
		return code.InvalidParameter
	}
	return nil
}

type ListGroupMemberResponse struct {
	List []*GroupMember
	//Total int64
	//Page  int32
}

type GroupMember struct {
	UserID   string `json:"user_id"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Remark   string `json:"remark"`
	Identity uint8  `json:"identity"`
}

type ListGroupMemberHandler decorator.CommandHandler[*ListGroupMember, *ListGroupMemberResponse]

func NewListGroupMemberHandler(
	logger *zap.Logger,
	grd service.GroupRelationDomain,
	userService rpc.UserService,
) ListGroupMemberHandler {
	return &listGroupMemberHandler{
		logger:      logger,
		grd:         grd,
		userService: userService,
	}
}

type listGroupMemberHandler struct {
	logger *zap.Logger

	grd service.GroupRelationDomain

	userService rpc.UserService
}

func (h *listGroupMemberHandler) Handle(ctx context.Context, cmd *ListGroupMember) (*ListGroupMemberResponse, error) {
	//if err := cmd.Validate(); err != nil {
	//	return nil, err
	//}

	if err := h.grd.IsUserInActiveGroup(ctx, cmd.GroupID, cmd.UserID); err != nil {
		h.logger.Error("IsUserInActiveGroup", zap.Error(err))
	}

	r, err := h.grd.ListGroupMember(ctx, cmd.GroupID, &service.ListGroupMemberOptions{})
	if err != nil {
		h.logger.Error("ListGroupMember", zap.Error(err))
		return nil, err
	}

	resp := &ListGroupMemberResponse{}
	for _, v := range r.List {
		info, err := h.userService.GetUserInfo(ctx, v.UserID)
		if err != nil {
			h.logger.Error("GetUserInfo", zap.Error(err))
			continue
		}
		resp.List = append(resp.List, &GroupMember{
			Remark:   v.Remark,
			Identity: uint8(v.Identity),
			UserID:   v.UserID,
			Nickname: info.Nickname,
			Avatar:   info.Avatar,
		})
	}

	return resp, nil
}
