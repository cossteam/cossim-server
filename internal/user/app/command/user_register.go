package command

import (
	"context"
	"fmt"
	"github.com/cossim/coss-server/internal/user/cache"
	"github.com/cossim/coss-server/internal/user/domain/entity"
	"github.com/cossim/coss-server/internal/user/domain/service"
	"github.com/cossim/coss-server/internal/user/infra/remote"
	"github.com/cossim/coss-server/internal/user/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/decorator"
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/workflow"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strings"
)

type UserRegister struct {
	Email       string
	Password    string
	ConfirmPass string
	Nickname    string
	PublicKey   string
}

type UserRegisterHandler decorator.CommandHandler[*UserRegister, string]

func NewUserRegisterHandler(
	logger *zap.Logger,
	dtmGrpcServer string,
	baseUrl string,
	emailEnable bool,
	userCache cache.UserCache,
	ud service.UserDomain,
	relationUserService rpc.RelationUserService,
	smtpService remote.SmtpService,
	storageService remote.StorageService,
) UserRegisterHandler {
	return &userRegisterHandler{
		logger:              logger,
		dtmGrpcServer:       dtmGrpcServer,
		baseUrl:             baseUrl,
		emailEnable:         emailEnable,
		userCache:           userCache,
		ud:                  ud,
		relationUserService: relationUserService,
		smtpService:         smtpService,
		storageService:      storageService,
	}
}

type userRegisterHandler struct {
	logger        *zap.Logger
	dtmGrpcServer string
	baseUrl       string
	emailEnable   bool
	userCache     cache.UserCache

	ud                  service.UserDomain
	relationUserService rpc.RelationUserService

	smtpService    remote.SmtpService
	storageService remote.StorageService
}

func (h *userRegisterHandler) Handle(ctx context.Context, cmd *UserRegister) (string, error) {
	h.logger.Info("userRegisterHandler", zap.Any("cmd", cmd))
	if cmd == nil || cmd.Email == "" || cmd.Password == "" || cmd.ConfirmPass == "" {
		return "", code.InvalidParameter
	}

	if cmd.Password != cmd.ConfirmPass {
		return "", code.InvalidParameter.CustomMessage("password and confirm password not match")
	}

	user, err := h.ud.GetUserWithOpts(ctx, entity.WithEmail(cmd.Email))
	if err != nil && user != nil {
		h.logger.Error("get user with email error", zap.Error(err))
		return "", code.UserErrEmailAlreadyRegistered
	}

	password := utils.HashString(cmd.Password)

	cmd.Nickname = strings.TrimSpace(cmd.Nickname)
	if cmd.Nickname == "" {
		cmd.Nickname = cmd.Email
	}

	avatarUrl, err := h.storageService.GenerateAvatar(ctx)
	if err != nil {
		h.logger.Error("generate avatar failed", zap.Error(err))
		return "", err
	}

	aUrl := fmt.Sprintf("%s/%s", h.baseUrl+constants.DownLoadAddress, avatarUrl)

	// 获取系统通知用户
	_, err = h.ud.GetUser(ctx, constants.SystemNotification)
	if err != nil {
		h.logger.Error("failed to register user", zap.Error(err))
		return "", err
	}

	var uid string

	workflow.InitGrpc(h.dtmGrpcServer, "", grpc.NewServer())
	gid := shortuuid.New()
	wfName := "register_user_workflow_" + gid
	if err := workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		userID, err := h.ud.UserRegister(ctx, &entity.UserRegister{
			Email:     cmd.Email,
			NickName:  cmd.Nickname,
			Password:  password,
			Avatar:    aUrl,
			PublicKey: cmd.PublicKey,
		})
		if err != nil {
			h.logger.Error("register user failed", zap.Error(err))
			return status.Error(codes.Aborted, err.Error())
		}

		uid = userID

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			return h.ud.DeleteUser(ctx, userID)
		})

		// 添加系统通知机器人好友
		if err := h.relationUserService.EstablishFriendship(wf.Context, userID, constants.SystemNotification); err != nil {
			return status.Error(codes.Aborted, err.Error())
		}

		return err
	}); err != nil {
		h.logger.Error("workflow.Register register user failed", zap.Error(err))
		return "", err
	}

	if err := workflow.Execute(wfName, gid, nil); err != nil {
		h.logger.Error("workflow.Execute register user failed", zap.Error(err))
		return "", err
	}

	if err := h.sendVerificationEmail(ctx, cmd.Email, uid); err != nil {
		h.logger.Error("send verification email failed", zap.Error(err))
		return "", err
	}

	return uid, nil
}

func (h *userRegisterHandler) sendVerificationEmail(ctx context.Context, email, userID string) error {
	if !h.emailEnable {
		return nil
	}

	emailVerificationCode := uuid.New().String()
	// 设置用户邮箱验证码

	if err := h.userCache.SetUserEmailVerificationCode(ctx, userID, emailVerificationCode, cache.UserEmailVerificationCodeExpireTime); err != nil {
		return err
	}

	// 发送邮件
	subject := "欢迎注册"
	content := h.smtpService.GenerateEmailVerificationContent(h.baseUrl, userID, emailVerificationCode)
	if err := h.smtpService.SendEmail(email, subject, content); err != nil {
		h.logger.Error("failed to send email", zap.Error(err))
		return err
	}

	return nil
}
