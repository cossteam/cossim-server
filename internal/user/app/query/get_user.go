package query

import (
	"context"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/infra/rpc"
	"github.com/cossim/coss-server/pkg/decorator"
	"go.uber.org/zap"
)

type GetUse struct {
	CurrentUser string
	TargetUser  string
}

type GetUseResponse struct {
}

type GetUserHandler decorator.CommandHandler[*GetUse, *entity.UserInfo]

type getUserHandler struct {
	logger *zap.Logger

	userService     rpc.UserService
	relationService rpc.RelationUserService
}

func NewGetUserHandler(logger *zap.Logger, userService rpc.UserService, relationService rpc.RelationUserService) GetUserHandler {
	return &getUserHandler{logger: logger, userService: userService, relationService: relationService}
}

func (h *getUserHandler) Handle(ctx context.Context, query *GetUse) (*entity.UserInfo, error) {
	var userInfo *entity.UserInfo

	userInfo, err := h.userService.GetUserInfo(ctx, query.TargetUser)
	if err != nil {
		return nil, err
	}

	userInfo = &entity.UserInfo{
		UserID:    userInfo.UserID,
		CossID:    userInfo.CossID,
		Nickname:  userInfo.Nickname,
		Email:     userInfo.Email,
		Tel:       userInfo.Tel,
		Avatar:    userInfo.Avatar,
		Signature: userInfo.Signature,
		Status:    userInfo.Status,
		//LoginNumber:    0,
		//Preferences:    nil,
		//NewDeviceLogin: false,
		//LastLoginTime:  0,
	}

	relation, err := h.relationService.GetUserRelation(ctx, query.TargetUser, query.CurrentUser)
	if err == nil && relation != nil {
		userInfo.RelationStatus = relation.Status
		userInfo.Preferences = &entity.Preferences{
			OpenBurnAfterReading:        relation.OpenBurnAfterReading,
			SilentNotification:          relation.SilentNotification,
			Remark:                      relation.Remark,
			OpenBurnAfterReadingTimeOut: relation.OpenBurnAfterReadingTimeOut,
		}
	}

	return userInfo, nil
}
