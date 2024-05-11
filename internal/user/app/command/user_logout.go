package command

import (
	"context"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/domain/service"
	"github.com/cossim/coss-server/internal/user/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/decorator"
	"github.com/cossim/coss-server/pkg/utils"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"github.com/golang/protobuf/ptypes/any"
	"go.uber.org/zap"
)

type UserLogout struct {
	UserID   string
	DriverID string
}

type UserLogoutHandler decorator.CommandHandlerNoneResponse[*UserLogout]

type userLogoutHandler struct {
	logger        *zap.Logger
	userCache     cache.UserCache
	dtmGrpcServer string

	ad  service.AuthDomain
	ud  service.UserDomain
	uld service.UserLoginDomain

	pushService rpc.PushService
}

func NewUserLogoutHandler(logger *zap.Logger, userCache cache.UserCache, dtmGrpcServer string, ad service.AuthDomain, ud service.UserDomain, uld service.UserLoginDomain, pushService rpc.PushService) UserLogoutHandler {
	return &userLogoutHandler{logger: logger, userCache: userCache, dtmGrpcServer: dtmGrpcServer, ad: ad, ud: ud, uld: uld, pushService: pushService}
}

func (h *userLogoutHandler) Handle(ctx context.Context, cmd *UserLogout) error {
	loginInfo, err := h.uld.GetByUserIDAndDriverID(ctx, cmd.UserID, cmd.DriverID)
	if err != nil {
		h.logger.Error("failed to get user login info", zap.Error(err))
		return code.NotFound
	}

	// 通知消息服务关闭ws
	if err := h.notifyMsgService(ctx, loginInfo); err != nil {
		h.logger.Error("failed to notify msg service", zap.Error(err))
	}

	// 删除客户端信息
	if err := h.uld.DeleteByUserIDAndDriverID(ctx, cmd.UserID, cmd.DriverID); err != nil {
		h.logger.Error("failed to delete user login info", zap.Error(err))
		return err
	}

	return nil
}

func (h *userLogoutHandler) notifyMsgService(ctx context.Context, loginInfo *entity.UserLogin) error {
	if loginInfo.Rid == "" {
		return nil
	}

	data := &constants.OfflineEventData{
		Rid: loginInfo.Rid,
	}
	toBytes, err := utils.StructToBytes(data)
	if err != nil {
		h.logger.Error("failed to struct to bytes", zap.Error(err))
		return err
	}

	msg := &pushgrpcv1.WsMsg{
		Uid:    loginInfo.UserID,
		Event:  pushgrpcv1.WSEventType_OfflineEvent,
		Data:   &any.Any{Value: toBytes},
		SendAt: ptime.Now(),
		Rid:    loginInfo.Rid,
	}
	toBytes2, err := utils.StructToBytes(msg)
	if err != nil {
		return err
	}

	if _, err := h.pushService.PushWS(ctx, toBytes2); err != nil {
		h.logger.Error("发送消息失败", zap.Error(err))
		return err
	}

	return nil
}
