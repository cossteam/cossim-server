package service

import (
	"context"
	"errors"
	"fmt"
	msgconfig "github.com/cossim/coss-server/interface/msg/config"
	"github.com/cossim/coss-server/interface/relation/api/model"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/msg_queue"
	groupApi "github.com/cossim/coss-server/service/group/api/v1"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/lithammer/shortuuid/v3"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	"time"
)

func (s *Service) ManageJoinGroup(ctx context.Context, groupID uint32, adminID, userID string, status relationgrpcv1.GroupRelationStatus) error {
	group, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: groupID})
	if err != nil {
		return err
	}

	if group.Status != groupApi.GroupStatus_GROUP_STATUS_NORMAL {
		return errors.New("群聊状态不可用")
	}

	relation1, err := s.groupRelationClient.GetGroupRelation(context.Background(), &relationgrpcv1.GetGroupRelationRequest{GroupId: groupID, UserId: adminID})
	if err != nil {
		return err
	}

	if relation1.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return errors.New("权限不足")
	}

	relation2, err := s.groupRelationClient.GetGroupRelation(context.Background(), &relationgrpcv1.GetGroupRelationRequest{GroupId: groupID, UserId: userID})
	if err != nil {
		return err
	}

	if err = s.validateGroupRelationStatus(relation2, status); err != nil {
		return err
	}

	id, err := s.dialogClient.GetDialogByGroupId(context.Background(), &relationgrpcv1.GetDialogByGroupIdRequest{GroupId: groupID})
	if err != nil {
		return err
	}

	r1 := &relationgrpcv1.JoinDialogRequest{DialogId: id.DialogId, UserId: userID}
	r2 := &relationgrpcv1.ManageJoinGroupRequest{UserId: userID, GroupId: groupID, Status: relation1.Status}
	gid := shortuuid.New()
	if err = dtmgrpc.TccGlobalTransaction(s.dtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		r := &emptypb.Empty{}
		if err = tcc.CallBranch(r1, s.dialogGrpcServer+relationgrpcv1.DialogService_JoinDialog_FullMethodName, "", s.dialogGrpcServer+relationgrpcv1.DialogService_JoinDialogRevert_FullMethodName, r); err != nil {
			return err
		}
		if err = tcc.CallBranch(r2, s.relationGrpcServer+relationgrpcv1.GroupRelationService_ManageJoinGroup_FullMethodName, "", s.relationGrpcServer+relationgrpcv1.GroupRelationService_ManageJoinGroupRevert_FullMethodName, r); err != nil {
			return err
		}
		return err
	}); err != nil {
		s.logger.Error("TCC ManageJoinGroup Failed", zap.Error(err))
		return err
	}

	msg := msgconfig.WsMsg{
		Uid:    userID,
		Event:  msgconfig.JoinGroupEvent,
		Data:   map[string]interface{}{"group_id": groupID, "status": status},
		SendAt: time.Now().Unix(),
	}
	err = s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
	if err != nil {
		s.logger.Error("通知消息服务有消息需要发送失败", zap.Error(err))
	}

	return nil
}

func (s *Service) RemoveUserFromGroup(ctx context.Context, groupID uint32, adminID, userID string) error {
	gr1, err := s.groupRelationClient.GetGroupRelation(context.Background(), &relationgrpcv1.GetGroupRelationRequest{UserId: adminID, GroupId: groupID})
	if err != nil {
		s.logger.Error("获取用户群组关系失败", zap.Error(err))
		return errors.New("获取用户群组关系失败")
	}

	if gr1.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return errors.New("没有权限操作")
	}

	gr2, err := s.groupRelationClient.GetGroupRelation(context.Background(), &relationgrpcv1.GetGroupRelationRequest{UserId: userID, GroupId: groupID})
	if err != nil {
		s.logger.Error("获取用户群组关系失败", zap.Error(err))
		return errors.New("获取用户群组关系失败")
	}

	dialog, err := s.dialogClient.GetDialogByGroupId(ctx, &relationgrpcv1.GetDialogByGroupIdRequest{GroupId: groupID})
	if err != nil {
		return errors.New("获取群聊会话失败")
	}

	r1 := &relationgrpcv1.DeleteDialogUserByDialogIDAndUserIDRequest{DialogId: dialog.DialogId, UserId: userID}
	r2 := &relationgrpcv1.DeleteGroupRelationByGroupIdAndUserIDRequest{GroupID: groupID, UserID: userID, Status: gr2.Status}

	gid := shortuuid.New()
	err = dtmgrpc.TccGlobalTransaction(s.dtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		r := &emptypb.Empty{}
		if err = tcc.CallBranch(r1, s.relationGrpcServer+relationgrpcv1.DialogService_DeleteDialogUserByDialogIDAndUserID_FullMethodName, "", s.relationGrpcServer+relationgrpcv1.DialogService_DeleteDialogUserByDialogIDAndUserIDRevert_FullMethodName, r); err != nil {
			return err
		}
		err = tcc.CallBranch(r2, s.relationGrpcServer+relationgrpcv1.GroupRelationService_DeleteGroupRelationByGroupIdAndUserID_FullMethodName, "", s.relationGrpcServer+relationgrpcv1.GroupRelationService_DeleteGroupRelationByGroupIdAndUserIDRevert_FullMethodName, r)
		return err
	})
	if err != nil {
		return errors.New("移出群聊失败")
	}

	//id, err := s.dialogClient.DeleteDialogUserByDialogIDAndUserID(ctx, &relationgrpcv1.DeleteDialogUserByDialogIDAndUserIDRequest{DialogId: dialog.DialogId, UserId: userID})
	//if err != nil {
	//	return err
	//}
	//
	//_, err = s.groupRelationClient.DeleteGroupRelationByGroupIdAndUserID(ctx, &relationgrpcv1.DeleteGroupRelationByGroupIdAndUserIDRequest{GroupID: groupID, UserID: userID})
	//if err != nil {
	//	return err
	//}

	return nil
}

func (s *Service) QuitGroup(ctx context.Context, groupID uint32, userID string) error {
	//查询用户是否在群聊中
	_, err := s.groupRelationClient.GetGroupRelation(context.Background(), &relationgrpcv1.GetGroupRelationRequest{UserId: userID, GroupId: groupID})
	if err != nil {
		return errors.New("用户群聊状态不可用")
	}

	dialog, err := s.dialogClient.GetDialogByGroupId(ctx, &relationgrpcv1.GetDialogByGroupIdRequest{GroupId: groupID})
	if err != nil {
		return errors.New("获取群聊会话失败")
	}

	r1 := &relationgrpcv1.DeleteDialogUserByDialogIDAndUserIDRequest{DialogId: dialog.DialogId, UserId: userID}
	r2 := &relationgrpcv1.LeaveGroupRequest{UserId: userID, GroupId: groupID}
	gid := shortuuid.New()
	err = dtmgrpc.TccGlobalTransaction(s.dtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		r := &emptypb.Empty{}
		if err = tcc.CallBranch(r1, s.relationGrpcServer+relationgrpcv1.DialogService_DeleteDialogUserByDialogIDAndUserID_FullMethodName, "", s.relationGrpcServer+relationgrpcv1.DialogService_DeleteDialogUserByDialogIDAndUserIDRevert_FullMethodName, r); err != nil {
			return err
		}
		err = tcc.CallBranch(r2, s.relationGrpcServer+relationgrpcv1.GroupRelationService_LeaveGroup_FullMethodName, "", s.relationGrpcServer+relationgrpcv1.GroupRelationService_LeaveGroupRevert_FullMethodName, r)
		return err
	})
	if err != nil {
		return errors.New("退出群聊失败")
	}

	return nil
}

func (s *Service) validateGroupRelationStatus(relation *relationgrpcv1.GetGroupRelationResponse, status relationgrpcv1.GroupRelationStatus) error {
	switch status {
	case relationgrpcv1.GroupRelationStatus_GroupStatusJoined:
		if relation.Status != relationgrpcv1.GroupRelationStatus_GroupStatusApplying {
			return errors.New("没有申请记录")
		}

		if relation.Status == relationgrpcv1.GroupRelationStatus_GroupStatusJoined {
			return errors.New("已经在群里了")
		}

	case relationgrpcv1.GroupRelationStatus_GroupStatusReject:
		if relation.Status != relationgrpcv1.GroupRelationStatus_GroupStatusJoined {
			return errors.New("用户没有在群组中或状态不可用")
		}
	default:
		return errors.New("用户状态异常")
	}

	return nil
}

func (s *Service) InviteGroup(ctx context.Context, adminID string, req *model.InviteGroupRequest) error {
	group, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupID})
	if err != nil {
		return err
	}

	if group.Status != groupApi.GroupStatus_GROUP_STATUS_NORMAL {
		return errors.New("群聊状态不可用")
	}

	relation1, err := s.groupRelationClient.GetGroupRelation(context.Background(), &relationgrpcv1.GetGroupRelationRequest{GroupId: req.GroupID, UserId: adminID})
	if err != nil {
		s.logger.Error("获取用户群组关系失败", zap.Error(err))
		return err
	}

	if relation1.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return errors.New("权限不足")
	}

	//relation2, err := s.groupRelationClient.GetGroupRelation(context.Background(), &relationgrpcv1.GetGroupRelationRequest{GroupId: req.GroupID, UserId: req.UserID})
	//if err != nil {
	//	return err
	//}

	grs, err := s.groupRelationClient.GetBatchGroupRelation(ctx, &relationgrpcv1.GetBatchGroupRelationRequest{GroupId: req.GroupID, UserIds: req.Member})
	fmt.Println("code => ", code.Cause(err).Code())
	if err != nil && !(code.Cause(err).Code() == 13112) {
		s.logger.Error("获取用户群组关系失败", zap.Error(err))
		return err
	}

	for _, gr := range grs.GroupRelationResponses {
		if gr.Status == relationgrpcv1.GroupRelationStatus_GroupStatusJoined {
			return errors.New("已经在群里了")
		}
	}

	for _, gr := range grs.GroupRelationResponses {
		//添加普通用户申请
		_, err = s.groupRelationClient.JoinGroup(context.Background(), &relationgrpcv1.JoinGroupRequest{UserId: gr.UserId, GroupId: req.GroupID, Identify: relationgrpcv1.GroupIdentity_IDENTITY_USER})
		if err != nil {
			return err
		}
	}

	//查询所有管理员
	adminIds, err := s.groupRelationClient.GetGroupAdminIds(context.Background(), &relationgrpcv1.GroupIDRequest{
		GroupId: req.GroupID,
	})
	for _, id := range adminIds.UserIds {
		msg := msgconfig.WsMsg{Uid: id, Event: msgconfig.JoinGroupEvent, Data: map[string]interface{}{"group_id": req.GroupID, "user_id": adminID}, SendAt: time.Now().Unix()}
		//通知消息服务有消息需要发送
		err = s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
		if err != nil {
			s.logger.Error("加入群聊请求申请通知推送失败", zap.Error(err))
		}
	}

	return nil
}
