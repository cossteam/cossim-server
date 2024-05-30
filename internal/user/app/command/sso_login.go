package command

import (
	"context"
	"fmt"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/domain/service"
	"github.com/cossim/coss-server/internal/user/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/decorator"
	"github.com/cossim/coss-server/pkg/utils"
	httputil "github.com/cossim/coss-server/pkg/utils/http"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/workflow"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/lithammer/shortuuid/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

type SSOLogin struct {
	UserID      string
	DriverID    string
	ClientIP    string
	DriverToken string
	Platform    string
}

type SSOLoginResponse struct {
	Token          string
	UserID         string
	NewDeviceLogin bool
	LastLoginTime  int64
}

type SSOLoginHandler decorator.CommandHandler[*SSOLogin, *SSOLoginResponse]

func NewSSOLoginHandler(
	logger *zap.Logger,
	userCache cache.UserCache,
	dtmGrpcServer string,
	ad service.AuthDomain,
	ud service.UserDomain,
	uld service.UserLoginDomain,
	relationUserService rpc.RelationUserService,
	dialogService rpc.RelationDialogService,
	msgService rpc.MsgService,
	pushService rpc.PushService) SSOLoginHandler {
	return &ssoLoginHandler{
		logger:              logger,
		userCache:           userCache,
		dtmGrpcServer:       dtmGrpcServer,
		ad:                  ad,
		ud:                  ud,
		uld:                 uld,
		relationUserService: relationUserService,
		dialogService:       dialogService,
		msgService:          msgService,
		pushService:         pushService,
	}
}

type ssoLoginHandler struct {
	logger        *zap.Logger
	userCache     cache.UserCache
	dtmGrpcServer string

	ad  service.AuthDomain
	ud  service.UserDomain
	uld service.UserLoginDomain

	relationUserService rpc.RelationUserService
	dialogService       rpc.RelationDialogService
	msgService          rpc.MsgService
	pushService         rpc.PushService
}

func (h *ssoLoginHandler) Handle(ctx context.Context, cmd *SSOLogin) (*SSOLoginResponse, error) {
	user, err := h.ud.GetUser(ctx, cmd.UserID)
	if err != nil {
		return nil, code.MyCustomErrorCode.CustomMessage(code.UserErrNotExist.Message())
	}

	// 登录是否受限，例如账户未激活、达到设备限制等
	if err := h.uld.IsLoginRestricted(ctx, user.ID); err != nil {
		return nil, err
	}

	token, err := h.ad.GenerateUserToken(ctx, &entity.AuthClaims{
		UserID:   user.ID,
		Email:    user.Email,
		DriverID: cmd.DriverID,
	})
	if err != nil {
		h.logger.Error("生成用户token失败", zap.Error(err))
		return nil, err
	}

	lastLoginTime, err := h.uld.LastLoginTime(ctx, user.ID)
	if err != nil {
		h.logger.Error("获取用户最近一次登录时间失败", zap.Error(err))
		//return nil, err
	}

	isNewDevice, err := h.uld.IsNewDeviceLogin(ctx, user.ID, cmd.DriverID)
	if err != nil {
		return nil, err
	}

	users, err := h.uld.List(ctx, user.ID)
	if err != nil {
		return nil, err
	}

	var index = len(users) + 1

	cacheData := entity.UserLogin{
		UserID:      user.ID,
		Token:       token,
		CreatedAt:   ptime.Now(),
		ClientIP:    cmd.ClientIP,
		DriverID:    cmd.DriverID,
		DriverToken: cmd.DriverToken,
		Platform:    cmd.Platform,
	}

	workflow.InitGrpc(h.dtmGrpcServer, "", grpc.NewServer())
	gid := shortuuid.New()
	wfName := "login_user_workflow_" + gid
	if err := workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		if err := h.userCache.SetUserLoginInfo(wf.Context, user.ID, cmd.DriverID, &cacheData, cache.UserLoginExpireTime); err != nil {
			h.logger.Error("failed to set user login info", zap.Error(err))
			return status.Error(codes.Aborted, err.Error())
		}

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			err = h.userCache.DeleteUserLoginInfo(wf.Context, user.ID, cmd.DriverID)
			return err
		})

		if isNewDevice && user.Bot != 1 {
			_, msgId, err := h.pushFirstLogin(ctx, user.ID, cmd.ClientIP, cmd.DriverID)
			if err != nil {
				return status.Error(codes.Aborted, err.Error())
			}

			wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
				return h.msgService.DeleteMessage(ctx, msgId)
			})
		}

		e1 := &entity.UserLogin{
			UserID:      user.ID,
			DriverID:    cmd.DriverID,
			Token:       token,
			DriverToken: cmd.DriverToken,
			Platform:    cmd.Platform,
			LoginCount:  uint(index),
		}
		err := h.uld.Create(ctx, e1)
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
		}

		return nil
	}); err != nil {
		h.logger.Error("failed to register workflow", zap.Error(err))
		return nil, err
	}

	if err := workflow.Execute(wfName, gid, nil); err != nil {
		if strings.Contains(err.Error(), "用户不存在或密码错误") {
			return nil, code.UserErrNotExistOrPassword
		}
		return nil, code.UserErrLoginFailed
	}

	return &SSOLoginResponse{
		Token:  token,
		UserID: user.ID,
		//CossID:         user.CossID,
		//Nickname:       user.NickName,
		//Email:          user.Email,
		//Tel:            user.Tel,
		//Avatar:         user.Avatar,
		//Signature:      user.Signature,
		//Status:         uint8(user.Status),
		NewDeviceLogin: isNewDevice,
		LastLoginTime:  lastLoginTime,
	}, nil
}

func (h *ssoLoginHandler) pushFirstLogin(ctx context.Context, userID, clientIp, driverId string) (uint32, uint32, error) {
	if clientIp == "127.0.0.1" {
		clientIp = httputil.GetMyPublicIP()
	}
	info := httputil.OnlineIpInfo(clientIp)

	result := fmt.Sprintf("您在新设备登录，IP地址为：%s\n位置为：%s %s %s", clientIp, info.Country, info.RegionName, info.City)
	if info.RegionName == info.City {
		result = fmt.Sprintf("您在新设备登录，IP地址为：%s\n位置为：%s %s", clientIp, info.Country, info.City)
	}

	// 查询系统通知信息
	systemInfo, err2 := h.ud.GetUser(ctx, constants.SystemNotification)
	if err2 != nil {
		return 0, 0, err2
	}

	// 查询与系统通知的对话id
	relation, err := h.relationUserService.GetUserRelation(ctx, userID, constants.SystemNotification)
	if err != nil {
		return 0, 0, err
	}

	// 查询与系统的对话是否关闭
	closed, err := h.dialogService.IsDialogClosed(ctx, relation.DialogID, userID)
	if err != nil {
		return 0, 0, err
	}

	// 如果关闭，则打开
	if closed {
		if err := h.dialogService.ToggleDialog(ctx, relation.DialogID, userID, true); err != nil {
			return 0, 0, err
		}
	}

	// 插入消息
	msgID, err := h.msgService.SendUserTextMessage(ctx, relation.DialogID, constants.SystemNotification, userID, result)
	if err != nil {
		return 0, 0, err
	}

	rmsg := &pushgrpcv1.SendWsUserMsg{
		MsgType:    uint32(msggrpcv1.MessageType_Text),
		DialogId:   relation.DialogID,
		Content:    result,
		SenderId:   constants.SystemNotification,
		ReceiverId: userID,
		SendAt:     ptime.Now(),
		MsgId:      msgID,
		SenderInfo: &pushgrpcv1.SenderInfo{
			UserId: systemInfo.ID,
			Avatar: systemInfo.Avatar,
			Name:   systemInfo.NickName,
		},
	}

	toBytes, err := utils.StructToBytes(rmsg)
	if err != nil {
		return 0, 0, err
	}

	msg := &pushgrpcv1.WsMsg{Uid: userID, DriverId: driverId, Event: pushgrpcv1.WSEventType_SendUserMessageEvent, PushOffline: true, SendAt: ptime.Now(), Data: &any.Any{Value: toBytes}}

	toBytes2, err := utils.StructToBytes(msg)
	if err != nil {
		return 0, 0, err
	}

	_, err = h.pushService.PushWS(ctx, toBytes2)
	if err != nil {
		h.logger.Error("发送消息失败", zap.Error(err))
	}
	return relation.DialogID, msgID, nil
}
