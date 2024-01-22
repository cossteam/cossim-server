package service

import (
	"context"
	"errors"
	"github.com/cossim/coss-server/interface/group/api/model"
	groupgrpcv1 "github.com/cossim/coss-server/service/group/api/v1"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/dtm-labs/client/workflow"
	"github.com/lithammer/shortuuid/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) CreateGroup(ctx context.Context, req *groupgrpcv1.Group) (*model.CreateGroupResponse, error) {
	var err error

	friends, err := s.relationUserClient.GetUserRelationByUserIds(ctx, &relationgrpcv1.GetUserRelationByUserIdsRequest{UserId: req.CreatorId, FriendIds: req.Member})
	if err != nil {
		return nil, err
	}

	if len(req.Member) != len(friends.Users) {
		return nil, errors.New("user not in group")
	}
	for _, friend := range friends.Users {
		if friend.Status != relationgrpcv1.RelationStatus_RELATION_STATUS_ADDED {
			return nil, errors.New("好友状态不可用")
		}
	}

	r1 := &groupgrpcv1.CreateGroupRequest{Group: &groupgrpcv1.Group{
		Type:            req.Type,
		MaxMembersLimit: req.MaxMembersLimit,
		CreatorId:       req.CreatorId,
		Name:            req.Name,
		Avatar:          req.Avatar,
		Member:          req.Member,
	}}
	//r2 := &relationgrpcv1.JoinGroupRequest{
	//	UserId:   req.CreatorId,
	//	Identify: relationgrpcv1.GroupIdentity_IDENTITY_OWNER,
	//}
	r22 := &relationgrpcv1.CreateGroupAndInviteUsersRequest{
		UserID: req.CreatorId,
		Member: req.Member,
	}
	r3 := &relationgrpcv1.CreateAndJoinDialogWithGroupRequest{
		OwnerId: req.CreatorId,
		UserIds: req.Member,
		Type:    uint32(relationgrpcv1.DialogType_GROUP_DIALOG),
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
		//r2.GroupId = resp1.Id
		//_, err = s.relationGroupClient.JoinGroup(ctx, r2)
		//if err != nil {
		//	return err
		//}
		//wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
		//	_, err = s.relationGroupClient.JoinGroupRevert(ctx, r2)
		//	return err
		//})

		r22.GroupId = resp1.Id
		_, err = s.relationGroupClient.CreateGroupAndInviteUsers(ctx, r22)
		if err != nil {
			return err
		}
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.relationGroupClient.CreateGroupAndInviteUsersRevert(ctx, r22)
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

func (s *Service) DeleteGroup(ctx context.Context, groupID uint32, userID string) (uint32, error) {
	_, err := s.groupClient.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{
		Gid: groupID,
	})
	if err != nil {
		return 0, err
	}
	sf, err := s.relationGroupClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		UserId:  userID,
		GroupId: groupID,
	})
	if err != nil {
		return 0, err
	}
	if sf.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return 0, errors.New("不允许操作")
	}

	dialog, err := s.relationDialogClient.GetDialogByGroupId(ctx, &relationgrpcv1.GetDialogByGroupIdRequest{GroupId: groupID})
	if err != nil {
		return 0, err
	}

	r1 := &relationgrpcv1.DeleteDialogUsersByDialogIDRequest{
		DialogId: dialog.DialogId,
	}
	r2 := &relationgrpcv1.DeleteDialogByIdRequest{
		DialogId: dialog.DialogId,
	}
	r3 := &relationgrpcv1.GroupIDRequest{
		GroupId: groupID,
	}
	r4 := &groupgrpcv1.DeleteGroupRequest{
		Gid: groupID,
	}
	gid := shortuuid.New()
	if err = dtmgrpc.TccGlobalTransaction(s.dtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		r := &emptypb.Empty{}
		// 删除对话用户
		if err = tcc.CallBranch(r1, s.relationGrpcServer+relationgrpcv1.DialogService_DeleteDialogUsersByDialogID_FullMethodName, "", s.relationGrpcServer+relationgrpcv1.DialogService_DeleteDialogUsersByDialogIDRevert_FullMethodName, r); err != nil {
			return err
		}
		// 删除对话
		if err = tcc.CallBranch(r2, s.relationGrpcServer+relationgrpcv1.DialogService_DeleteDialogById_FullMethodName, "", s.relationGrpcServer+relationgrpcv1.DialogService_DeleteDialogByIdRevert_FullMethodName, r); err != nil {
			return err
		}
		// 删除群聊成员
		if err = tcc.CallBranch(r3, s.relationGrpcServer+relationgrpcv1.GroupRelationService_DeleteGroupRelationByGroupId_FullMethodName, "", s.relationGrpcServer+relationgrpcv1.GroupRelationService_DeleteGroupRelationByGroupIdRevert_FullMethodName, r); err != nil {
			return err
		}
		// 删除群聊
		if err = tcc.CallBranch(r4, s.groupGrpcServer+groupgrpcv1.GroupService_DeleteGroup_FullMethodName, "", s.groupGrpcServer+groupgrpcv1.GroupService_DeleteGroupRevert_FullMethodName, r); err != nil {
			return err
		}
		return err
	}); err != nil {
		return 0, err
	}

	return groupID, err
}
