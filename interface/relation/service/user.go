package service

import (
	"context"
	"errors"
	"fmt"
	msgconfig "github.com/cossim/coss-server/interface/msg/config"
	"github.com/cossim/coss-server/pkg/msg_queue"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	userApi "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/workflow"
	"github.com/lithammer/shortuuid/v3"
	"go.uber.org/zap"
)

func (s *Service) ManageFriend(ctx context.Context, userId, friendId string, action int32) (interface{}, error) {
	var dialogId uint32

	fmt.Println("action => ", action)

	switch action {
	case 1: // 同意好友申请
		_, err := s.handleAction1(ctx, userId, friendId, relationgrpcv1.RelationStatus_RELATION_STATUS_ADDED)
		if err != nil {
			return nil, err
		}

	default: // 同意申请之外的操作，修改状态
		if err := s.manageFriend3(ctx, userId, friendId, dialogId, relationgrpcv1.RelationStatus_RELATION_STATUS_REJECTED); err != nil {
			return nil, err
		}
	}

	// 向用户推送通知
	resp, err := s.sendFriendManagementNotification(ctx, userId, friendId, "", relationgrpcv1.RelationStatus(action))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *Service) handleAction1(ctx context.Context, userId, friendId string, status relationgrpcv1.RelationStatus) (uint32, error) {
	var dialogId uint32
	relation, err := s.userRelationClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: userId, FriendId: friendId})
	if err != nil {
		return 0, err
	}

	if relation != nil && relation.DialogId != 0 {
		err = s.manageFriend1(ctx, userId, friendId, status, relation.DialogId)
		if err != nil {
			return 0, err
		}
	} else {
		dialogId, err = s.manageFriend2(ctx, userId, friendId, status)
		if err != nil {
			return 0, err
		}
	}

	return dialogId, nil
}

// manageFriend1 已经存在关系，修改关系状态
func (s *Service) manageFriend1(ctx context.Context, userId, friendId string, status relationgrpcv1.RelationStatus, dialogId uint32) error {
	var err error

	// 创建 DTM 分布式事务工作流
	gid := shortuuid.New()
	wfName := "manage_friend_workflow_1_" + gid
	if err = workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			r1 := &relationgrpcv1.DeleteDialogByIdRequest{DialogId: dialogId}
			_, err = s.dialogClient.DeleteDialogById(ctx, r1)
			if err != nil {
				return err
			}
			return nil
		})
		_, err = s.dialogClient.JoinDialog(context.Background(), &relationgrpcv1.JoinDialogRequest{DialogId: dialogId, UserId: userId})
		if err != nil {
			return err
		}

		mfr := &relationgrpcv1.ManageFriendRequest{
			UserId:   userId,
			FriendId: friendId,
			DialogId: dialogId,
			Status:   status,
		}
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.userRelationClient.ManageFriendRevert(ctx, mfr)
			if err != nil {
				return err
			}
			return nil
		})
		if _, err = s.userRelationClient.ManageFriend(ctx, mfr); err != nil {
			fmt.Println("s.userRelationClient.ManageFriend err => ", err)
			return err
		}

		return nil
	}); err != nil {
		return err
	}
	// 执行 DTM 分布式事务工作流
	if err = workflow.Execute(wfName, gid, nil); err != nil {
		return err
	}

	return nil
}

// manageFriend2 不存在好友关系，创建新的关系
func (s *Service) manageFriend2(ctx context.Context, userId, friendId string, status relationgrpcv1.RelationStatus) (uint32, error) {
	var dialogId uint32
	// 创建 DTM 分布式事务工作流
	// 执行 DTM 分布式事务工作流
	gid := shortuuid.New()
	wfName := "manage_friend_workflow_2_" + gid
	if err := workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		r1 := &relationgrpcv1.ConfirmFriendAndJoinDialogRequest{
			OwnerId: userId,
			UserId:  friendId,
		}
		resp1, err := s.dialogClient.ConfirmFriendAndJoinDialog(ctx, r1)
		if err != nil {
			return err
		}
		dialogId = resp1.Id
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.dialogClient.ConfirmFriendAndJoinDialogRevert(ctx, &relationgrpcv1.ConfirmFriendAndJoinDialogRevertRequest{DialogId: resp1.Id, OwnerId: userId, UserId: friendId})
			return err
		})

		r2 := &relationgrpcv1.ManageFriendRequest{
			UserId:   userId,
			FriendId: friendId,
			Status:   status,
			DialogId: resp1.Id,
		}
		_, err = s.userRelationClient.ManageFriend(ctx, r2)
		if err != nil {
			return err
		}
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.userRelationClient.ManageFriendRevert(ctx, r2)
			if err != nil {
				return err
			}
			return nil
		})

		return nil
	}); err != nil {
		fmt.Printf("manageFriend2 1 err: %v", err)
		return 0, err
	}
	if err := workflow.Execute(wfName, gid, nil); err != nil {
		fmt.Printf("manageFriend2 2 err: %v", err)
		return 0, err
	}

	return dialogId, nil
}

// manageFriend3 只修改关系状态
func (s *Service) manageFriend3(ctx context.Context, userId, friendId string, dialogId uint32, status relationgrpcv1.RelationStatus) error {
	// 创建 DTM 分布式事务工作流
	gid := shortuuid.New()
	wfName := "manage_friend_workflow_3_" + gid
	var err error
	if err = workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		mfr := &relationgrpcv1.ManageFriendRequest{
			UserId:   userId,
			FriendId: friendId,
			DialogId: dialogId,
			Status:   status,
		}
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.userRelationClient.ManageFriendRevert(ctx, mfr)
			if err != nil {
				return err
			}
			return nil
		})
		if _, err = s.userRelationClient.ManageFriend(ctx, mfr); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return err
	}

	// 执行 DTM 分布式事务工作流
	if err = workflow.Execute(wfName, gid, nil); err != nil {
		return err
	}

	return nil
}

func (s *Service) sendFriendManagementNotification(ctx context.Context, userID, targetID, E2EPublicKey string, status relationgrpcv1.RelationStatus) (interface{}, error) {
	targetInfo, err := s.getUserInfo(ctx, targetID)
	if err != nil {
		return nil, err
	}

	myInfo, err := s.getUserInfo(ctx, userID)
	if err != nil {
		return nil, err
	}

	wsMsgData := map[string]interface{}{"user_id": userID, "status": status}
	msg := msgconfig.WsMsg{Uid: targetID, Event: msgconfig.ManageFriendEvent, Data: wsMsgData}
	var responseData interface{}

	if status == 1 {
		wsMsgData["target_info"] = myInfo
		wsMsgData["e2e_public_key"] = E2EPublicKey
		responseData = targetInfo
	}

	err = s.publishServiceMessage(ctx, msg)
	if err != nil {
		s.logger.Error("Failed to publish service message", zap.Error(err))
		return nil, err
	}

	return responseData, nil
}

func (s *Service) getUserInfo(ctx context.Context, userID string) (*userApi.UserInfoResponse, error) {
	req := &userApi.UserInfoRequest{UserId: userID}
	resp, err := s.userClient.UserInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *Service) publishServiceMessage(ctx context.Context, msg msgconfig.WsMsg) error {
	err := s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
	if err != nil {
		s.logger.Error("Failed to publish service message", zap.Error(err))
		return err
	}
	return nil
}

func (s *Service) convertDialogType(_type uint32) (relationgrpcv1.DialogType, error) {
	switch _type {
	case 0:
		return relationgrpcv1.DialogType_USER_DIALOG, nil
	case 1:
		return relationgrpcv1.DialogType_GROUP_DIALOG, nil
	default:
		return 0, errors.New("invalid dialog type")
	}
}

func (s *Service) convertStatusToRelationStatus(status uint32) (relationgrpcv1.RelationStatus, error) {
	switch status {
	case 0:
		return relationgrpcv1.RelationStatus_RELATION_STATUS_APPLYING, nil
	case 1:
		return relationgrpcv1.RelationStatus_RELATION_STATUS_PENDING, nil
	case 2:
		return relationgrpcv1.RelationStatus_RELATION_STATUS_ADDED, nil
	case 3:
		return relationgrpcv1.RelationStatus_RELATION_STATUS_REJECTED, nil
	case 4:
		return relationgrpcv1.RelationStatus_RELATION_STATUS_BLOCKED, nil
	case 5:
		return relationgrpcv1.RelationStatus_RELATION_STATUS_DELETED, nil
	default:
		return 0, errors.New("invalid status")
	}
}
