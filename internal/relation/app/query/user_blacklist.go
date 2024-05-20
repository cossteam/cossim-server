package query

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type UserBlacklist struct {
	UserID   string
	PageNum  int
	PageSize int
}

type UserBlacklistResponse struct {
	List  []*Black
	Total int64
	Page  int32
}

type Black struct {
	UserID   string
	CossID   string
	Avatar   string
	Nickname string
}

type UserBlacklistHandler decorator.CommandHandler[*UserBlacklist, *UserBlacklistResponse]

func NewUserBlacklistHandler(
	logger *zap.Logger,
	urd service.UserRelationDomain,
	userService rpc.UserService,
) UserBlacklistHandler {
	return &userBlacklistHandler{
		logger:      logger,
		urd:         urd,
		userService: userService,
	}
}

type userBlacklistHandler struct {
	logger      *zap.Logger
	urd         service.UserRelationDomain
	userService rpc.UserService
}

func (h *userBlacklistHandler) Handle(ctx context.Context, cmd *UserBlacklist) (*UserBlacklistResponse, error) {
	if cmd == nil {
		return nil, code.InvalidParameter
	}

	blacklist, err := h.urd.Blacklist(ctx, &service.BlacklistOptions{
		UserID:   cmd.UserID,
		PageSize: cmd.PageNum,
		PageNum:  cmd.PageSize,
	})
	if err != nil {
		h.logger.Error("failed to get user blacklist", zap.Error(err))
		return nil, err
	}

	var list []*Black
	for _, b := range blacklist.List {
		user, err := h.userService.GetUserInfo(ctx, b.UserID)
		if err != nil {
			h.logger.Error("failed to get user info", zap.Error(err))
			return nil, err
		}
		list = append(list, &Black{
			Avatar:   user.Avatar,
			CossID:   user.CossID,
			Nickname: user.Nickname,
			UserID:   user.ID,
		})
	}

	return &UserBlacklistResponse{
		List:  list,
		Total: blacklist.Total,
		Page:  blacklist.Page,
	}, nil
}
