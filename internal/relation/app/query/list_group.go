package query

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type ListGroup struct {
	UserID string
	//GroupID  uint32
	PageNum  int
	PageSize int
}

func (cmd *ListGroup) Validate() error {
	if cmd == nil || cmd.UserID == "" {
		return code.InvalidParameter
	}
	return nil
}

type ListGroupResponse struct {
	List []*GroupInfo
	//Total int64
	//Page  int32
}

type GroupInfo struct {
	Type     uint8
	Status   uint8
	ID       uint32
	DialogID uint32
	Name     string
	Avatar   string
}

type ListGroupHandler decorator.CommandHandler[*ListGroup, *ListGroupResponse]

func NewListGroupHandler(
	logger *zap.Logger,
	grd service.GroupRelationDomain,
	drd service.DialogRelationDomain,
	groupService rpc.GroupService,
) ListGroupHandler {
	return &listGroupHandler{
		logger:       logger,
		grd:          grd,
		drd:          drd,
		groupService: groupService,
	}
}

type listGroupHandler struct {
	logger *zap.Logger

	grd service.GroupRelationDomain
	drd service.DialogRelationDomain

	groupService rpc.GroupService
}

func (h *listGroupHandler) Handle(ctx context.Context, cmd *ListGroup) (*ListGroupResponse, error) {
	// 获取用户的组ID
	groupIDs, err := h.grd.ListUserGroupID(ctx, cmd.UserID)
	if err != nil {
		return nil, err
	}

	// 获取组的对话信息
	dialogs, err := h.drd.FindDialogsForGroups(ctx, groupIDs)
	if err != nil {
		return nil, err
	}

	// 获取组的详细信息
	groupInfos, err := h.groupService.GetGroupsInfo(ctx, groupIDs)
	if err != nil {
		return nil, err
	}

	// 创建一个映射来快速查找组信息
	groupInfoMap := make(map[uint32]*GroupInfo)
	for _, info := range groupInfos {
		groupInfoMap[info.ID] = &GroupInfo{
			ID:     info.ID,
			Type:   info.Type,
			Name:   info.Name,
			Avatar: info.Avatar,
			Status: info.Status,
		}
	}

	// 构建响应
	response := &ListGroupResponse{
		List: make([]*GroupInfo, 0, len(dialogs)),
	}
	for _, dialog := range dialogs {
		if info, ok := groupInfoMap[dialog.GroupId]; ok {
			infoCopy := *info // 创建一个GroupInfo副本，以免修改映射中的原始值
			infoCopy.DialogID = dialog.ID
			response.List = append(response.List, &infoCopy)
		}
	}

	return response, nil
}
