package service

import (
	"context"
	"errors"
	msgconfig "github.com/cossim/coss-server/interface/msg/config"
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

	//_, err = s.dialogClient.JoinDialog(context.Background(), &relationgrpcv1.JoinDialogRequest{DialogId: id.DialogId, UserId: userID})
	//if err != nil {
	//	return err
	//}
	//
	//if _, err = s.groupRelationClient.ManageJoinGroup(context.Background(), &relationgrpcv1.ManageJoinGroupRequest{UserId: userID, GroupId: groupID, Status: status}); err != nil {
	//	return err
	//}

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
