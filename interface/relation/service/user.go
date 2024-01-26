package service

import (
	"context"
	"errors"
	"fmt"
	msgconfig "github.com/cossim/coss-server/interface/msg/config"
	"github.com/cossim/coss-server/interface/relation/api/model"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/msg_queue"
	"github.com/cossim/coss-server/pkg/utils/time"
	"github.com/cossim/coss-server/pkg/utils/usersorter"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	userApi "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/dtm-labs/client/workflow"
	"github.com/lithammer/shortuuid/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) FriendList(ctx context.Context, userID string) (interface{}, error) {
	// 获取好友列表
	friendListResp, err := s.userRelationClient.GetFriendList(ctx, &relationgrpcv1.GetFriendListRequest{UserId: userID})
	if err != nil {
		s.logger.Error("user service GetFriendList", zap.Error(err))
		return nil, err
	}

	var users []string
	for _, user := range friendListResp.FriendList {
		users = append(users, user.UserId)
	}

	userInfos, err := s.userClient.GetBatchUserInfo(ctx, &userApi.GetBatchUserInfoRequest{UserIds: users})
	if err != nil {
		s.logger.Error("user service GetBatchUserInfo", zap.Error(err))
		return nil, err
	}

	var data []usersorter.User
	for _, v := range userInfos.Users {
		for _, friend := range friendListResp.FriendList {
			if friend.UserId == v.UserId {
				data = append(data, usersorter.CustomUserData{
					UserID:   v.UserId,
					NickName: v.NickName,
					Email:    v.Email,
					Tel:      v.Tel,
					Avatar:   v.Avatar,
					Status:   uint(v.Status),
					DialogId: friend.DialogId,
				})
				break
			}
		}

	}

	// Sort and group by specified field
	groupedUsers := usersorter.SortAndGroupUsers(data, "NickName")
	return groupedUsers, nil
}

func (s *Service) BlackList(ctx context.Context, userID string) (interface{}, error) {
	// 获取黑名单列表
	blacklistResp, err := s.userRelationClient.GetBlacklist(ctx, &relationgrpcv1.GetBlacklistRequest{UserId: userID})
	if err != nil {
		s.logger.Error("user service GetBlacklist", zap.Error(err))
		return nil, err
	}

	var users []string
	for _, user := range blacklistResp.Blacklist {
		users = append(users, user.UserId)
	}

	blacklist, err := s.userClient.GetBatchUserInfo(ctx, &userApi.GetBatchUserInfoRequest{UserIds: users})
	if err != nil {
		s.logger.Error("user service GetBatchUserInfo", zap.Error(err))
		return nil, err
	}

	return blacklist, nil
}

func (s *Service) UserRequestList(ctx context.Context, userID string) (interface{}, error) {

	reqList, err := s.userFriendRequestClient.GetFriendRequestList(ctx, &relationgrpcv1.GetFriendRequestListRequest{UserId: userID})
	if err != nil {
		s.logger.Error("user service GetFriendRequestList", zap.Error(err))
		return nil, err
	}

	var data []*model.UserRequestListResponse
	for _, v := range reqList.FriendRequestList {
		if v.SenderId == userID {
			info, err := s.userClient.UserInfo(ctx, &userApi.UserInfoRequest{UserId: v.ReceiverId})
			if err != nil {
				return nil, err
			}
			data = append(data, &model.UserRequestListResponse{
				ID:         v.ID,
				ReceiverId: info.UserId,
				Remark:     v.Remark,
				RequestAt:  v.CreateAt,
				Status:     uint32(v.Status),
				SenderId:   v.SenderId,
				ReceiverInfo: &model.UserInfo{
					UserID:     info.UserId,
					UserName:   info.NickName,
					UserAvatar: info.Avatar,
				},
			})
		} else {
			info, err := s.userClient.UserInfo(ctx, &userApi.UserInfoRequest{UserId: v.SenderId})
			if err != nil {
				return nil, err
			}
			data = append(data, &model.UserRequestListResponse{
				ID:         v.ID,
				ReceiverId: info.UserId,
				Remark:     v.Remark,
				RequestAt:  v.CreateAt,
				Status:     uint32(v.Status),
				SenderId:   v.SenderId,
				ReceiverInfo: &model.UserInfo{
					UserID:     info.UserId,
					UserName:   info.NickName,
					UserAvatar: info.Avatar,
				},
			})
		}
	}
	return data, nil
}
func (s *Service) SendFriendRequest(ctx context.Context, userID string, req *model.SendFriendRequest) (interface{}, error) {
	if req.UserId == userID {
		err := errors.New("不能添加自己为好友")
		return nil, err
	}
	_, err := s.userClient.UserInfo(ctx, &userApi.UserInfoRequest{UserId: req.UserId})
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err))
		return nil, code.UserErrNotExist
	}

	_, err = s.userFriendRequestClient.SendFriendRequest(ctx, &relationgrpcv1.SendFriendRequestStruct{
		SenderId:   userID,
		ReceiverId: req.UserId,
		Remark:     req.Remark,
	})
	if err != nil {
		s.logger.Error("添加好友失败", zap.Error(err))
		err := errors.New("添加好友失败")
		return nil, err
	}
	return nil, nil
}

func (s *Service) ManageFriend(ctx context.Context, userID, friendID string, action int32, key string) (interface{}, error) {
	//var dialogId uint32
	//
	//// 检查要操作的用户是否存在
	//_, err := s.userClient.UserInfo(context.Background(), &userApi.UserInfoRequest{UserId: friendID})
	//if err != nil {
	//	return nil, err
	//}
	//
	//switch action {
	//case 1: // 同意好友申请
	//	_, err = s.handleAction1(ctx, userID, friendID, relationgrpcv1.RelationStatus_RELATION_STATUS_ADDED)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//default: // 同意申请之外的操作，修改状态
	//	if err = s.manageFriend3(ctx, userID, friendID, dialogId, relationgrpcv1.RelationStatus_RELATION_STATUS_REJECTED); err != nil {
	//		return nil, err
	//	}
	//}
	//
	//// 向用户推送通知
	resp, err := s.sendFriendManagementNotification(ctx, userID, friendID, key, relationgrpcv1.RelationStatus(action))
	if err != nil {
		s.logger.Error("发送好友管理通知失败", zap.Error(err))
	}

	return resp, nil
}

func (s *Service) handleAction1(ctx context.Context, userId, friendId string, status relationgrpcv1.RelationStatus) (uint32, error) {
	var dialogId uint32
	relation, err := s.userRelationClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: userId, FriendId: friendId})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		return 0, err
	}

	if relation != nil && relation.DialogId != 0 {
		err = s.manageFriend1(ctx, userId, friendId, status, relation.DialogId)
		if err != nil {
			s.logger.Error("修改好友关系失败", zap.Error(err))
			return 0, err
		}
	} else {
		dialogId, err = s.manageFriend2(ctx, userId, friendId, status)
		if err != nil {
			s.logger.Error("添加好友关系失败", zap.Error(err))
			return 0, err
		}
	}

	return dialogId, nil
}

// manageFriend1 已经存在关系，修改关系状态
func (s *Service) manageFriend1(ctx context.Context, userId, friendId string, status relationgrpcv1.RelationStatus, dialogId uint32) error {
	var err error
	// 创建 DTM 分布式事务工作流
	workflow.InitGrpc(s.dtmGrpcServer, s.relationGrpcServer, grpc.NewServer())
	gid := shortuuid.New()
	wfName := "manage_friend_workflow_1_" + gid
	if err = workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			r1 := &relationgrpcv1.DeleteDialogByIdRequest{DialogId: dialogId}
			_, err = s.dialogClient.DeleteDialogById(ctx, r1)
			if err != nil {
				s.logger.Error("删除对话失败", zap.Error(err))
				return err
			}
			return nil
		})
		_, err = s.dialogClient.JoinDialog(context.Background(), &relationgrpcv1.JoinDialogRequest{DialogId: dialogId, UserId: userId})
		if err != nil {
			s.logger.Error("加入对话失败", zap.Error(err))
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
		s.logger.Error("workflow.Register err => ", zap.Error(err))
		return code.RelationErrConfirmFriendFailed
	}
	// 执行 DTM 分布式事务工作流
	if err = workflow.Execute(wfName, gid, nil); err != nil {
		return code.RelationErrConfirmFriendFailed
	}

	return nil
}

func (s *Service) DeleteFriend(ctx context.Context, userID, friendID string) error {
	// 检查删除的用户是否存在
	//_, err := s.userClient.UserInfo(ctx, &userApi.UserInfoRequest{UserId: friendID})
	//if err != nil {
	//	s.logger.Error("获取用户信息失败", zap.Error(err))
	//	return err
	//}

	relation, err := s.userRelationClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: userID, FriendId: friendID})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		return err
	}
	//TODO
	//if relation.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
	//	return code.RelationUserErrFriendRelationNotFound
	//}

	r1 := &relationgrpcv1.DeleteDialogUserByDialogIDAndUserIDRequest{DialogId: relation.DialogId, UserId: userID}
	r2 := &relationgrpcv1.DeleteFriendRequest{UserId: userID, FriendId: friendID}
	gid := shortuuid.New()
	// tcc
	if err = dtmgrpc.TccGlobalTransaction(s.dtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		r := &emptypb.Empty{}
		err = tcc.CallBranch(r1, s.dialogGrpcServer+relationgrpcv1.DialogService_DeleteDialogUserByDialogIDAndUserID_FullMethodName, "", s.dialogGrpcServer+relationgrpcv1.DialogService_DeleteDialogUserByDialogIDAndUserIDRevert_FullMethodName, r)
		if err != nil {
			return err
		}
		err = tcc.CallBranch(r2, s.relationGrpcServer+relationgrpcv1.UserRelationService_DeleteFriend_FullMethodName, "", s.relationGrpcServer+relationgrpcv1.UserRelationService_DeleteFriendRevert_FullMethodName, r)
		return err
	}); err != nil {
		s.logger.Error("TCC DeleteFriend", zap.Error(err))
		return err
	}

	return nil
}

func (s *Service) AddBlacklist(ctx context.Context, userID, friendID string) (interface{}, error) {
	_, err := s.userRelationClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: userID, FriendId: friendID})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		return nil, err
	}
	//TODO
	//if relation.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
	//	return nil, code.RelationUserErrFriendRelationNotFound
	//}

	// 进行添加黑名单操作
	_, err = s.userRelationClient.AddBlacklist(ctx, &relationgrpcv1.AddBlacklistRequest{UserId: userID, FriendId: friendID})
	if err != nil {
		s.logger.Error("添加黑名单失败", zap.Error(err))
		return nil, err
	}
	return nil, nil
}

func (s *Service) DeleteBlacklist(ctx context.Context, userID, friendID string) (interface{}, error) {
	relation, err := s.userRelationClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: userID, FriendId: friendID})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		return nil, err
	}

	if relation.Status != relationgrpcv1.RelationStatus_RELATION_STATUS_BLOCKED {
		return nil, code.RelationErrNotInBlacklist
	}

	// 进行删除黑名单操作
	_, err = s.userRelationClient.DeleteBlacklist(context.Background(), &relationgrpcv1.DeleteBlacklistRequest{UserId: userID, FriendId: friendID})
	if err != nil {
		s.logger.Error("删除黑名单失败", zap.Error(err))
		return nil, err
	}
	return nil, nil
}

func (s *Service) SwitchUserE2EPublicKey(ctx context.Context, userID string, friendID string, publicKey string) (interface{}, error) {
	reqm := model.SwitchUserE2EPublicKeyRequest{
		UserId:    userID,
		PublicKey: publicKey,
	}
	msg := msgconfig.WsMsg{Uid: friendID, Event: msgconfig.PushE2EPublicKeyEvent, Data: reqm, SendAt: time.Now()}

	//通知消息服务有消息需要发送
	if err := s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg); err != nil {
		s.logger.Error("交换用户端到端公钥通知推送失败", zap.Error(err))
		return nil, code.UserErrSwapPublicKeyFailed
	}
	return nil, nil
}

func (s *Service) UserSilentNotification(ctx context.Context, userID string, friendID string, silent model.SilentNotificationType) (interface{}, error) {
	_, err := s.userRelationClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
		UserId:   userID,
		FriendId: friendID,
	})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		return nil, err
	}

	//TODO
	//if relation.Status != relationgrpcv1.RelationStatus_RELATION_STATUS_ADDED {
	//	return nil, code.RelationUserErrFriendRelationNotFound
	//}

	_, err = s.userRelationClient.SetFriendSilentNotification(ctx, &relationgrpcv1.SetFriendSilentNotificationRequest{
		UserId:   userID,
		FriendId: friendID,
		IsSilent: relationgrpcv1.UserSilentNotificationType(silent),
	})
	if err != nil {
		s.logger.Error("设置好友静音通知失败", zap.Error(err))
		return nil, err
	}
	return nil, nil
}

// manageFriend2 不存在好友关系，创建新的关系
func (s *Service) manageFriend2(ctx context.Context, userId, friendId string, status relationgrpcv1.RelationStatus) (uint32, error) {
	var dialogId uint32
	// 创建 DTM 分布式事务工作流
	fmt.Println("s.dtmGrpcServer => ", s.dtmGrpcServer)
	fmt.Println("s.relationGrpcServer => ", s.relationGrpcServer)

	workflow.InitGrpc(s.dtmGrpcServer, s.relationGrpcServer, grpc.NewServer())
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
		s.logger.Error("workflow Register manageFriend2", zap.Error(err))
		return 0, code.RelationErrConfirmFriendFailed
	}
	if err := workflow.Execute(wfName, gid, nil); err != nil {
		s.logger.Error("workflow manageFriend2", zap.Error(err))
		return 0, code.RelationErrConfirmFriendFailed
	}

	return dialogId, nil
}

// manageFriend3 只修改关系状态 现在只是拒绝操作
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
		s.logger.Error("workflow Register manageFriend3", zap.Error(err))
		return code.RelationErrRejectFriendFailed
	}

	// 执行 DTM 分布式事务工作流
	if err = workflow.Execute(wfName, gid, nil); err != nil {
		s.logger.Error("workflow manageFriend3", zap.Error(err))
		return code.RelationErrRejectFriendFailed
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
	fmt.Println("msg:=>", msg)

	if err = s.publishServiceMessage(ctx, msg); err != nil {
		s.logger.Error("Failed to publish service message", zap.Error(err))
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
	return s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
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

//func (s *Service) convertStatusToRelationStatus(status uint32) (relationgrpcv1.RelationStatus, error) {
//	switch status {
//	case 0:
//		return relationgrpcv1.RelationStatus_RELATION_STATUS_APPLYING, nil
//	case 1:
//		return relationgrpcv1.RelationStatus_RELATION_STATUS_PENDING, nil
//	case 2:
//		return relationgrpcv1.RelationStatus_RELATION_STATUS_ADDED, nil
//	case 3:
//		return relationgrpcv1.RelationStatus_RELATION_STATUS_REJECTED, nil
//	case 4:
//		return relationgrpcv1.RelationStatus_RELATION_STATUS_BLOCKED, nil
//	case 5:
//		return relationgrpcv1.RelationStatus_RELATION_STATUS_DELETED, nil
//	default:
//		return 0, errors.New("invalid status")
//	}
//}
