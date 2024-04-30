package query

import (
	"context"
	"github.com/cossim/coss-server/internal/group/app/command"
	"github.com/cossim/coss-server/internal/group/domain/repository"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
	"strconv"
)

type SearchGroup struct {
	UserID   string
	Keyword  string
	Page     int32
	PageSize int32
}

type SearchGroupHandler decorator.CommandHandler[*SearchGroup, []*Group]

var _ decorator.CommandHandler[*SearchGroup, []*Group] = &searchGroupHandler{}

type searchGroupHandler struct {
	groupRepo             repository.Repository
	relationGroupService  command.RelationGroupService
	relationDialogService command.RelationDialogService
	logger                *zap.Logger

	dtmGrpcServer string
}

func NewSearchGroupHandler(
	repo repository.Repository,
	logger *zap.Logger,
	dtmGrpcServer string,
	relationGroupService command.RelationGroupService,
	relationDialogService command.RelationDialogService,
) SearchGroupHandler {
	return &searchGroupHandler{
		groupRepo:             repo,
		relationGroupService:  relationGroupService,
		relationDialogService: relationDialogService,
		logger:                logger,
		dtmGrpcServer:         dtmGrpcServer,
	}
}

func (h *searchGroupHandler) Handle(ctx context.Context, cmd *SearchGroup) ([]*Group, error) {
	h.logger.Info("search group handler", zap.Any("cmd", cmd))
	if cmd.Page == 0 {
		cmd.Page = 1
	}
	if cmd.PageSize == 0 {
		cmd.PageSize = 10
	}
	query := repository.Query{
		Limit:  int(cmd.PageSize),
		Offset: int((cmd.Page - 1) * cmd.PageSize),
	}
	groupID, err := strconv.ParseUint(cmd.Keyword, 10, 32)
	if err == nil {
		query.ID = []uint32{uint32(groupID)}
	}
	query.Name = cmd.Keyword

	find, err := h.groupRepo.Find(ctx, query)
	if err != nil {
		return nil, err
	}

	var groups []*Group
	for _, group := range find {
		members, err := h.relationGroupService.GetGroupMembers(ctx, group.ID)
		if err != nil {
			h.logger.Error("get group members error", zap.Error(err))
			continue
		}
		groups = append(groups, &Group{
			Id:              group.ID,
			Avatar:          group.Avatar,
			Name:            group.Name,
			Type:            uint8(group.Type),
			Status:          int(group.Status),
			Member:          len(members),
			MaxMembersLimit: group.MaxMembersLimit,
			CreatorID:       group.CreatorID,
		})
	}

	return groups, nil
}
