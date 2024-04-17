package command

import (
	"context"
	"github.com/cossim/coss-server/internal/group/domain/group"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/workflow"
	"github.com/lithammer/shortuuid/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type DeleteGroup struct {
	ID     uint32 `json:"id"`
	UserID string `json:"user_id"`
}

type DeleteGroupResponse struct {
}

type DeleteGroupHandler decorator.CommandHandler[DeleteGroup, *DeleteGroupResponse]

type deleteGroupHandler struct {
	groupRepo             group.Repository
	relationGroupService  RelationGroupService
	relationDialogService RelationDialogService
	logger                *zap.Logger

	dtmGrpcServer string
}

func NewDeleteGroupHandler(
	repo group.Repository,
	logger *zap.Logger,
	dtmGrpcServer string,
	relationGroupService RelationGroupService,
	relationDialogService RelationDialogService,
) DeleteGroupHandler {
	if repo == nil {
		panic("nil repo")
	}

	h := &deleteGroupHandler{
		groupRepo:             repo,
		logger:                logger,
		dtmGrpcServer:         dtmGrpcServer,
		relationGroupService:  relationGroupService,
		relationDialogService: relationDialogService,
	}
	return h
}

func (h *deleteGroupHandler) Handle(ctx context.Context, cmd DeleteGroup) (*DeleteGroupResponse, error) {
	resp := &DeleteGroupResponse{}

	data, err := h.groupRepo.Get(ctx, cmd.ID)
	if err != nil || data == nil {
		return nil, code.GroupErrGroupNotFound
	}

	isOwner, err := h.relationGroupService.IsGroupOwner(ctx, cmd.ID, cmd.UserID)
	if err != nil {
		return nil, err
	}

	if !isOwner {
		return nil, code.Forbidden
	}

	dialogID, err := h.relationDialogService.GetGroupDialogID(ctx, cmd.ID)
	if err != nil {
		h.logger.Error("get group dialog id failed", zap.Error(err))
		return nil, err
	}

	// 创建 DTM 分布式事务工作流
	workflow.InitGrpc(h.dtmGrpcServer, "", grpc.NewServer())
	gid := shortuuid.New()
	wfName := "delete_group_workflow_" + gid
	if err := workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		// 删除对话用户
		if err := h.relationDialogService.DeleteUserDialog(ctx, dialogID); err != nil {
			return err
		}

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			return h.relationDialogService.DeleteUserDialogRevert(ctx, dialogID)
		})

		// 删除对话
		if err := h.relationDialogService.DeleteDialog(ctx, dialogID); err != nil {
			return err
		}

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			return h.relationDialogService.DeleteDialogRevert(ctx, dialogID)
		})

		// 删除群聊成员
		if err := h.relationGroupService.DeleteGroupRelations(ctx, cmd.ID); err != nil {
			return err
		}

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			return h.relationGroupService.DeleteGroupRelationsRevert(ctx, cmd.ID)
		})

		// 删除群聊
		if err := h.groupRepo.Delete(ctx, cmd.ID); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return resp, err
	}
	if err := workflow.Execute(wfName, gid, nil); err != nil {
		return nil, err
	}

	return resp, nil
}

//func (h *deleteGroupHandler) HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) {
//	addr := conn.Target()
//	switch serviceName {
//	case app.UserServiceName:
//		//h.userService = adapters.NewUserGrpc(usergrpcv1.NewUserServiceClient(conn))
//	case app.RelationUserServiceName:
//		h.relationDialogService = adapters.NewRelationDialogGrpc(relationgrpcv1.NewDialogServiceClient(conn))
//		h.relationGroupService = adapters.NewRelationGroupGrpc(relationgrpcv1.NewGroupRelationServiceClient(conn))
//	case app.PushServiceName:
//		//h.pushService = adapters.NewPushService(pushgrpcv1.NewPushServiceClient(conn))
//	default:
//	}
//	h.logger.Info("gRPC client service initialized", zap.String("service", serviceName), zap.String("addr", addr))
//}
