package command

import (
	"context"
	"github.com/cossim/coss-server/internal/relation/domain/service"
	"github.com/cossim/coss-server/internal/relation/infra/rpc"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/workflow"
	"github.com/lithammer/shortuuid/v4"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type DeleteFriend struct {
	CurrentUserID string
	TargetUserID  string
}

func (cmd *DeleteFriend) Validate() error {
	if cmd == nil || cmd.CurrentUserID == "" || cmd.TargetUserID == "" {
		return code.InvalidParameter
	}

	return nil
}

type DeleteFriendHandler decorator.CommandHandlerNoneResponse[*DeleteFriend]

func NewDeleteFriendHandler(
	logger *zap.Logger,
	dtmGrpcServer string,
	userRelationService service.UserRelationDomain,
	msgService rpc.MsgService,
) DeleteFriendHandler {
	return &deleteFriendHandler{
		logger:             logger,
		dtmGrpcServer:      dtmGrpcServer,
		userRelationDomain: userRelationService,
		msgService:         msgService,
	}
}

type deleteFriendHandler struct {
	logger        *zap.Logger
	dtmGrpcServer string

	userRelationDomain service.UserRelationDomain
	msgService         rpc.MsgService
}

func (h *deleteFriendHandler) Handle(ctx context.Context, cmd *DeleteFriend) error {
	if err := cmd.Validate(); err != nil {
		return err
	}

	isBlack, err := h.userRelationDomain.IsInBlacklist(ctx, cmd.CurrentUserID, cmd.TargetUserID)
	if err != nil {
		h.logger.Error("check blacklist error", zap.Error(err))
		return err
	}
	if isBlack {
		return code.StatusNotAvailable.CustomMessage("请先将当前用户移出黑名单，再删除")
	}

	isFriend, err := h.userRelationDomain.IsFriend(ctx, cmd.CurrentUserID, cmd.TargetUserID)
	if err != nil {
		h.logger.Error("check friend error", zap.Error(err))
		return err
	}
	if !isFriend {
		return code.RelationUserErrFriendRelationNotFound
	}

	dialogID, err := h.userRelationDomain.GetDialogID(ctx, cmd.CurrentUserID, cmd.TargetUserID)
	if err != nil {
		h.logger.Error("get dialog id error", zap.Error(err))
		return err
	}

	gid := shortuuid.New()
	workflow.InitGrpc(h.dtmGrpcServer, "", grpc.NewServer())
	wfName := "delete_relation_workflow_" + gid
	if err := workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		if err := h.userRelationDomain.DeleteFriend(ctx, cmd.CurrentUserID, cmd.TargetUserID); err != nil {
			return status.Error(codes.Aborted, err.Error())
		}

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			return h.userRelationDomain.DeleteFriendRollback(wf.Context, cmd.CurrentUserID, cmd.TargetUserID)
		})

		if err := h.msgService.CleanDialogMessages(wf.Context, dialogID); err != nil {
			return status.Error(codes.Aborted, err.Error())
		}

		return err
	}); err != nil {
		return err
	}

	if err := workflow.Execute(wfName, gid, nil); err != nil {
		h.logger.Error("delete friend error", zap.Error(err))
		return code.RelationErrDeleteFriendFailed
	}

	return nil
}
