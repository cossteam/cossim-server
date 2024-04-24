package command

import (
	"context"
	"fmt"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	"github.com/cossim/coss-server/internal/group/domain/group"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/decorator"
	"github.com/cossim/coss-server/pkg/utils"
	ptime "github.com/cossim/coss-server/pkg/utils/time"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/workflow"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/lithammer/shortuuid/v3"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CreateGroup struct {
	Type        uint
	Encrypt     bool
	JoinApprove bool
	CreateID    string
	Name        string
	Avatar      string
	Member      []string
	MaxMember   int
}

type CreateGroupResponse struct {
	ID              uint32
	Type            uint32
	DialogID        uint32
	Status          int
	MaxMembersLimit int
	Avatar          string
	Name            string
	CreatorID       string
}

type CreateGroupHandler decorator.CommandHandler[CreateGroup, CreateGroupResponse]

type createGroupHandler struct {
	groupRepo            group.Repository
	relationUserService  RelationUserService
	relationGroupService RelationGroupService
	userService          UserService
	pushService          PushService
	logger               *zap.Logger

	dtmGrpcServer    string
	groupServiceAddr string
}

func NewCreateGroupHandler(
	repo group.Repository,
	logger *zap.Logger,
	dtmGrpcServer string,
	userService UserService,
	pushService PushService,
	relationUserService RelationUserService,
	relationGroupService RelationGroupService,
) decorator.CommandHandler[CreateGroup, CreateGroupResponse] {
	if repo == nil {
		panic("nil repo")
	}

	h := &createGroupHandler{
		groupRepo:            repo,
		logger:               logger,
		dtmGrpcServer:        dtmGrpcServer,
		userService:          userService,
		pushService:          pushService,
		relationUserService:  relationUserService,
		relationGroupService: relationGroupService,
	}
	return h

	//return decorator.ApplyCommandDecorators[CreateGroup, CreateGroupResponse](
	//	&createGroupHandler{
	//		groupRepo: repo,
	//	},
	//	logger,
	//	nil,
	//)
}

func (h *createGroupHandler) Handle(ctx context.Context, cmd CreateGroup) (CreateGroupResponse, error) {
	resp := CreateGroupResponse{}

	friends, err := h.relationUserService.GetUserRelationships(ctx, cmd.CreateID, cmd.Member)
	if err != nil {
		h.logger.Error("get user relationships failed", zap.Error(err))
		return resp, err
	}
	isUserInFriends := func(userID string, friends map[string]UserRelationship) bool {
		relationship, exists := friends[userID]
		return exists && relationship.Status == UserRelationNormal
	}
	for _, memberID := range cmd.Member {
		if !isUserInFriends(memberID, friends) {
			info, err := h.userService.GetUserInfo(ctx, memberID)
			if err != nil {
				h.logger.Error("get user info failed", zap.Error(err))
				return resp, err
			}
			return resp, code.RelationUserErrFriendRelationNotFound.CustomMessage(fmt.Sprintf("%s不是你的好友", info.NickName))
		}
	}

	// TODO 暂时写死，应该在 group grpc 定义类型大小
	var maxMembersLimit int
	switch groupgrpcv1.GroupType(cmd.Type) {
	case groupgrpcv1.GroupType_Public:
		maxMembersLimit = 1000
	default:
		maxMembersLimit = 500
	}

	var groupID uint32
	// 创建 DTM 分布式事务工作流
	workflow.InitGrpc(h.dtmGrpcServer, h.groupServiceAddr, grpc.NewServer())
	gid := shortuuid.New()
	wfName := "create_group_workflow_" + gid
	if err = workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		// 创建群聊
		if err := h.groupRepo.Create(ctx, &group.Group{
			Type:            group.Type(cmd.Type),
			Status:          group.StatusNormal,
			MaxMembersLimit: maxMembersLimit,
			CreatorID:       cmd.CreateID,
			Name:            cmd.Name,
			Avatar:          cmd.Avatar,
			JoinApprove:     cmd.JoinApprove,
			Encrypt:         cmd.Encrypt,
		}, func(e *group.Group) (*group.Group, error) {
			groupID = e.ID
			resp.ID = e.ID
			resp.Avatar = e.Avatar
			resp.Name = e.Name
			resp.Type = uint32(e.Type)
			resp.Status = int(e.Status)
			return e, nil
		}); err != nil {
			return errors.Wrap(err, "create group")
		}

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			return h.groupRepo.Delete(ctx, groupID)
		})

		dialogID, err := h.relationGroupService.CreateGroup(wf.Context, groupID, cmd.CreateID, cmd.Member)
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
		}

		resp.DialogID = dialogID

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			err := h.relationGroupService.CreateGroupRevert(wf.Context, groupID, cmd.CreateID, cmd.Member)
			return err
		})

		return err
	}); err != nil {
		h.logger.Error("WorkFlow Create", zap.Error(err))
		return resp, err
	}
	if err = workflow.Execute(wfName, gid, nil); err != nil {
		h.logger.Error("WorkFlow Create", zap.Error(err))
		return resp, code.RelationErrCreateGroupFailed
	}

	data := map[string]interface{}{"group_id": groupID, "inviter_id": cmd.CreateID}
	toBytes, err := utils.StructToBytes(data)
	if err != nil {
		return resp, err
	}
	// 给被邀请的用户推送
	for _, id := range cmd.Member {
		msg := &pushgrpcv1.WsMsg{
			Uid:         id,
			Event:       pushgrpcv1.WSEventType_InviteJoinGroupEvent,
			Data:        &any.Any{Value: toBytes},
			SendAt:      ptime.Now(),
			PushOffline: true,
		}
		toBytes2, err := utils.StructToBytes(msg)

		_, err = h.pushService.Push(ctx, 0, toBytes2)
		if err != nil {
			h.logger.Error("推送消息失败", zap.Error(err))
		}
	}

	return resp, nil
}

//func (h *createGroupHandler) HandlerGrpcClient(serviceName string, conn *grpc.ClientConn) {
//	addr := conn.Target()
//	switch serviceName {
//	case app.UserServiceName:
//		h.userService = adapters.NewUserGrpc(usergrpcv1.NewUserServiceClient(conn))
//	case app.RelationUserServiceName:
//		h.relationUserService = adapters.NewRelationUserGrpc(relationgrpcv1.NewUserRelationServiceClient(conn))
//		h.relationGroupService = adapters.NewRelationGroupGrpc(relationgrpcv1.NewGroupRelationServiceClient(conn))
//	case app.PushServiceName:
//		h.pushService = adapters.NewPushService(pushgrpcv1.NewPushServiceClient(conn))
//	default:
//	}
//	h.logger.Info("gRPC client service initialized", zap.String("service", serviceName), zap.String("addr", addr))
//}
