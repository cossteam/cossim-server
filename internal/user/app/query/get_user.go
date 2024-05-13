package query

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/domain/service"
	"github.com/cossim/coss-server/internal/user/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type GetUse struct {
	CurrentUser string
	TargetUser  string
	TargetEmail string
}

type GetUseResponse struct {
}

type GetUserHandler decorator.CommandHandler[*GetUse, *entity.UserInfo]

type getUserHandler struct {
	logger *zap.Logger

	ud              service.UserDomain
	relationService rpc.RelationUserService
}

func NewGetUserHandler(logger *zap.Logger, ud service.UserDomain, relationService rpc.RelationUserService) GetUserHandler {
	return &getUserHandler{logger: logger, ud: ud, relationService: relationService}
}

func (h *getUserHandler) Handle(ctx context.Context, query *GetUse) (*entity.UserInfo, error) {
	var userInfo *entity.UserInfo
	var err error
	fmt.Println("query.TargetEmail => ", query.TargetEmail)

	// 根据查询类型获取用户信息
	switch {
	case query.TargetEmail != "":
		userInfo, err = h.getUserByEmail(ctx, query.TargetEmail)
	case query.TargetUser != "":
		userInfo, err = h.getUserByID(ctx, query.TargetUser)
	default:
		err = code.InvalidParameter
	}

	if err != nil {
		h.logger.Error("get user info error", zap.Error(err))
		return nil, err
	}

	// 获取用户关系
	if userInfo != nil {
		if err = h.populateUserRelation(ctx, query.CurrentUser, query.TargetUser, userInfo); err != nil {
			h.logger.Warn("get user relation error", zap.Error(err))
		}
	}

	return userInfo, nil
}

// 通过邮箱获取用户信息
func (h *getUserHandler) getUserByEmail(ctx context.Context, email string) (*entity.UserInfo, error) {
	user, err := h.ud.GetUserWithOpts(ctx, entity.WithEmail(email))
	if err != nil {
		return nil, err
	}
	return mapUserToUserInfo(user), nil
}

// 通过用户ID获取用户信息
func (h *getUserHandler) getUserByID(ctx context.Context, userID string) (*entity.UserInfo, error) {
	user, err := h.ud.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	return mapUserToUserInfo(user), nil
}

// 将用户实体映射为用户信息实体
func mapUserToUserInfo(user *entity.User) *entity.UserInfo {
	return &entity.UserInfo{
		UserID:    user.ID,
		CossID:    user.CossID,
		Nickname:  user.NickName,
		Email:     user.Email,
		Tel:       user.Tel,
		Avatar:    user.Avatar,
		Signature: user.Signature,
		Status:    user.Status,
	}
}

// 填充用户关系信息
func (h *getUserHandler) populateUserRelation(ctx context.Context, currentUser, targetUser string, userInfo *entity.UserInfo) error {
	relation, err := h.relationService.GetUserRelation(ctx, currentUser, targetUser)
	if err != nil {
		return err
	}

	if relation != nil {
		userInfo.RelationStatus = relation.Status
		userInfo.Preferences = &entity.Preferences{
			OpenBurnAfterReading:        relation.OpenBurnAfterReading,
			SilentNotification:          relation.SilentNotification,
			Remark:                      relation.Remark,
			OpenBurnAfterReadingTimeOut: relation.OpenBurnAfterReadingTimeOut,
		}
	}

	return nil
}
