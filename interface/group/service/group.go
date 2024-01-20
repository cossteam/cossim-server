package service

import (
	"context"
	"github.com/cossim/coss-server/interface/group/api/model"
	api "github.com/cossim/coss-server/service/group/api/v1"
	groupgrpcv1 "github.com/cossim/coss-server/service/group/api/v1"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/workflow"
	"github.com/lithammer/shortuuid/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (s *Service) CreateGroup(ctx context.Context, req *api.Group) (*model.CreateGroupResponse, error) {
	var err error
	r1 := &groupgrpcv1.CreateGroupRequest{Group: &groupgrpcv1.Group{
		Type:            req.Type,
		MaxMembersLimit: req.MaxMembersLimit,
		CreatorId:       req.CreatorId,
		Name:            req.Name,
		Avatar:          req.Avatar,
	}}
	r2 := &relationgrpcv1.JoinGroupRequest{
		UserId:   req.CreatorId,
		Identify: relationgrpcv1.GroupIdentity_IDENTITY_OWNER,
	}
	r3 := &relationgrpcv1.CreateAndJoinDialogWithGroupRequest{
		OwnerId: req.CreatorId,
		Type:    uint32(req.Type),
	}
	resp1 := &groupgrpcv1.Group{}
	resp2 := &relationgrpcv1.CreateAndJoinDialogWithGroupResponse{}

	// 创建 DTM 分布式事务工作流
	workflow.InitGrpc(s.dtmGrpcServer, s.relationGrpcServer, grpc.NewServer())
	gid := shortuuid.New()
	wfName := "create_group_workflow_" + gid
	if err = workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		// 创建群聊
		resp1, err = s.groupClient.CreateGroup(ctx, r1)
		if err != nil {
			return err
		}
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.groupClient.CreateGroupRevert(ctx, &groupgrpcv1.CreateGroupRequest{Group: &groupgrpcv1.Group{
				Id: resp1.Id,
			}})
			return err
		})

		// 加入群聊
		r2.GroupId = resp1.Id
		_, err = s.relationGroupClient.JoinGroup(ctx, r2)
		if err != nil {
			return err
		}
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.relationGroupClient.JoinGroupRevert(ctx, r2)
			return err
		})

		// 创建群聊会话并加入
		r3.GroupId = resp1.Id
		_, err = s.relationDialogClient.CreateAndJoinDialogWithGroup(ctx, r3)
		if err != nil {
			return err
		}
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.relationDialogClient.CreateAndJoinDialogWithGroupRevert(ctx, r3)
			return err
		})

		return err
	}); err != nil {
		s.logger.Error("WorkFlow CreateGroup", zap.Error(err))
		return nil, err
	}
	if err = workflow.Execute(wfName, gid, nil); err != nil {
		s.logger.Error("WorkFlow CreateGroup", zap.Error(err))
		return nil, err
	}

	//gid := shortuuid.New()
	//if err = dtmgrpc.TccGlobalTransaction(s.dtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
	//	r := &emptypb.Empty{}
	//	// 创建群聊
	//	if err = tcc.CallBranch(r1, s.groupGrpcServer+groupgrpcv1.GroupService_CreateGroup_FullMethodName, "", s.groupGrpcServer+groupgrpcv1.GroupService_CreateGroupRevert_FullMethodName, resp1); err != nil {
	//		return err
	//	}
	//	r2.GroupId = resp1.Id
	//	// 加入群聊
	//	if err = tcc.CallBranch(r2, s.relationGrpcServer+relationgrpcv1.GroupRelationService_JoinGroup_FullMethodName, "", s.relationGrpcServer+relationgrpcv1.GroupRelationService_JoinGroupRevert_FullMethodName, r); err != nil {
	//		return err
	//	}
	//	r3.GroupId = resp1.Id
	//	// 创建群聊会话并加入
	//	if err = tcc.CallBranch(r3, s.relationGrpcServer+relationgrpcv1.DialogService_CreateAndJoinDialogWithGroup_FullMethodName, "", s.relationGrpcServer+relationgrpcv1.DialogService_CreateAndJoinDialogWithGroupRevert_FullMethodName, resp2); err != nil {
	//		return err
	//	}
	//	return err
	//}); err != nil {
	//	s.logger.Error("TCC CreateGroup", zap.Error(err))
	//	return nil, err
	//}

	return &model.CreateGroupResponse{
		Id:              resp1.Id,
		Avatar:          resp1.Avatar,
		Name:            resp1.Name,
		Type:            uint32(resp1.Type),
		Status:          int32(resp1.Status),
		MaxMembersLimit: resp1.MaxMembersLimit,
		CreatorId:       resp1.CreatorId,
		DialogId:        resp2.Id,
	}, nil
}
