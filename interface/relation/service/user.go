package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cossim/coss-server/interface/relation/api/model"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
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
	ostime "time"
)

func (s *Service) FriendList(ctx context.Context, userID string) (interface{}, error) {
	//查询是否有缓存
	values, err := s.redisClient.GetAllListValues(fmt.Sprintf("friend:%s", userID))
	if err != nil {
		s.logger.Error("err:" + err.Error())
		return nil, code.RelationErrGetFriendListFailed
	}
	if len(values) > 0 {
		// 类型转换
		var responseList []usersorter.User
		for _, v := range values {
			var friend usersorter.CustomUserData
			err := json.Unmarshal([]byte(v), &friend)
			if err != nil {
				fmt.Println("Error decoding cached data:", err)
				continue
			}
			responseList = append(responseList, friend)
		}
		groupedUsers := usersorter.SortAndGroupUsers(responseList, "NickName")

		return groupedUsers, nil
	}

	// 获取好友列表
	friendListResp, err := s.userRelationClient.GetFriendList(ctx, &relationgrpcv1.GetFriendListRequest{UserId: userID})
	if err != nil {
		s.logger.Error("user service GetFriendList", zap.Error(err))
		return nil, err
	}
	if len(friendListResp.FriendList) == 0 {
		return []string{}, nil
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
					Remark:   friend.Remark,
				})
				break
			}
		}
	}

	var result []interface{}

	// Assuming data2 is a slice or array of a specific type
	for _, item := range data {
		result = append(result, item)
	}

	exists, err := s.redisClient.ExistsKey(fmt.Sprintf("friend:%s", userID))
	if err != nil {
		return nil, err
	}
	if !exists {
		//存储到缓存
		fmt.Println("len(result)", len(result))
		err = s.redisClient.AddToList(fmt.Sprintf("friend:%s", userID), result)
		if err != nil {
			s.logger.Error("err:" + err.Error())
			return nil, code.RelationErrGetFriendListFailed
		}

		//设置key过期时间
		err = s.redisClient.SetKeyExpiration(fmt.Sprintf("friend:%s", userID), 3*24*ostime.Hour)
		if err != nil {
			s.logger.Error("err:" + err.Error())
			return nil, code.RelationErrGetFriendListFailed
		}
	}

	// Sort and group by specified field
	groupedUsers := usersorter.SortAndGroupUsers(data, "NickName")
	fmt.Println("结果", len(groupedUsers))
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
				ReceiverId: v.ReceiverId,
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
				ReceiverId: v.ReceiverId,
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
		return nil, code.RelationErrSendFriendRequestFailed
	}
	_, err := s.userClient.UserInfo(ctx, &userApi.UserInfoRequest{UserId: req.UserId})
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err))
		return nil, code.UserErrNotExist
	}

	fRequest, err := s.userFriendRequestClient.GetFriendRequestByUserIdAndFriendId(ctx, &relationgrpcv1.GetFriendRequestByUserIdAndFriendIdRequest{
		UserId:   userID,
		FriendId: req.UserId,
	})
	if err != nil {
		if code.Code(code.RelationUserErrNoFriendRequestRecords.Code()) != code.Cause(err) {
			s.logger.Error("获取好友请求失败", zap.Error(err))
			return nil, err
		}
	}
	if fRequest != nil {
		if fRequest.Status == relationgrpcv1.FriendRequestStatus_FriendRequestStatus_PENDING {
			return nil, code.RelationErrFriendRequestAlreadyPending
		} else if fRequest.Status == relationgrpcv1.FriendRequestStatus_FriendRequestStatus_ACCEPT {
			return nil, code.RelationErrRequestAlreadyProcessed
		}
	}

	relation, err := s.userRelationClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: userID, FriendId: req.UserId})
	if err != nil {
		if code.Code(code.RelationUserErrFriendRelationNotFound.Code()) != code.Cause(err) {
			s.logger.Error("获取好友关系失败", zap.Error(err))
			return nil, err
		}
	}
	if relation != nil {
		if relation.DialogId != 0 {
			return nil, code.RelationErrAlreadyFriends
		}
	}

	//删除之前的
	_, err = s.userFriendRequestClient.DeleteFriendRequestByUserIdAndFriendId(ctx, &relationgrpcv1.DeleteFriendRequestByUserIdAndFriendIdRequest{
		UserId:   userID,
		FriendId: req.UserId,
	})
	if err != nil {
		return nil, err
	}

	resp, err := s.userFriendRequestClient.SendFriendRequest(ctx, &relationgrpcv1.SendFriendRequestStruct{
		SenderId:   userID,
		ReceiverId: req.UserId,
		Remark:     req.Remark,
	})
	if err != nil {
		s.logger.Error("添加好友失败", zap.Error(err))
		return nil, code.RelationErrSendFriendRequestFailed
	}

	wsMsgData := constants.AddFriendEventData{
		UserId:       userID,
		Msg:          req.Remark,
		E2EPublicKey: req.E2EPublicKey,
	}
	msg := constants.WsMsg{Uid: req.UserId, Event: constants.AddFriendEvent, Data: wsMsgData}

	if err = s.publishServiceMessage(ctx, msg); err != nil {
		s.logger.Error("Failed to publish service message", zap.Error(err))
	}

	return resp, nil
}

func (s *Service) ManageFriend(ctx context.Context, userId string, questId uint32, action int32, key string) (interface{}, error) {
	qs, err := s.userFriendRequestClient.GetFriendRequestById(ctx, &relationgrpcv1.GetFriendRequestByIdRequest{ID: questId})
	if err != nil {
		return nil, err
	}

	if qs == nil {
		return nil, code.RelationUserErrNoFriendRequestRecords
	}
	if qs.ReceiverId != userId {
		return nil, code.RelationUserErrNoFriendRequestRecords
	}

	if qs.Status != relationgrpcv1.FriendRequestStatus_FriendRequestStatus_PENDING {
		return nil, code.RelationErrRequestAlreadyProcessed
	}
	relation, err := s.userRelationClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: qs.SenderId, FriendId: userId})
	if err != nil {
		if code.Code(code.RelationUserErrFriendRelationNotFound.Code()) != code.Cause(err) {
			s.logger.Error("获取好友关系失败", zap.Error(err))
			return nil, err
		}
	}
	if relation != nil {
		if relation.DialogId != 0 {
			return nil, code.RelationErrAlreadyFriends
		}
	}

	_, err = s.userFriendRequestClient.ManageFriendRequest(ctx, &relationgrpcv1.ManageFriendRequestStruct{
		ID:     questId,
		Status: s.convertStatusToRelationStatus(uint32(action)),
	})
	if err != nil {
		return nil, err
	}

	if action == 1 && s.cache {
		relation2, err := s.userRelationClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: qs.SenderId, FriendId: userId})
		if err != nil {
			if code.Code(code.RelationUserErrFriendRelationNotFound.Code()) != code.Cause(err) {
				s.logger.Error("获取好友关系失败", zap.Error(err))
				return nil, code.RelationErrConfirmFriendFailed
			}
		}
		dialogInfo, err := s.dialogClient.GetDialogById(ctx, &relationgrpcv1.GetDialogByIdRequest{DialogId: relation2.DialogId})
		if err != nil {
			s.logger.Error("获取好友关系失败", zap.Error(err))
			return nil, code.RelationErrConfirmFriendFailed
		}

		info, err := s.userClient.UserInfo(ctx, &userApi.UserInfoRequest{UserId: relation2.UserId})
		if err != nil {
			s.logger.Error("获取好友关系失败", zap.Error(err))
			return nil, code.RelationErrConfirmFriendFailed
		}

		info2, err := s.userClient.UserInfo(ctx, &userApi.UserInfoRequest{UserId: relation2.FriendId})
		if err != nil {
			s.logger.Error("获取好友关系失败", zap.Error(err))
			return nil, code.RelationErrConfirmFriendFailed
		}

		re := model.UserDialogListResponse{
			DialogId:       relation2.DialogId,
			UserId:         info2.UserId,
			DialogType:     model.ConversationType(dialogInfo.Type),
			DialogName:     info2.NickName,
			DialogAvatar:   info2.Avatar,
			DialogCreateAt: dialogInfo.CreateAt,
		}
		//更新缓存
		err = s.insertRedisUserDialogList(info.UserId, re)
		if err != nil {
			s.logger.Error("获取好友关系失败", zap.Error(err))
			return nil, code.RelationErrConfirmFriendFailed
		}

		re.UserId = info.UserId
		re.DialogName = info.NickName
		re.DialogAvatar = info.Avatar
		err = s.insertRedisUserDialogList(info2.UserId, re)
		if err != nil {
			s.logger.Error("获取好友关系失败", zap.Error(err))
			return nil, code.RelationErrConfirmFriendFailed
		}

		err = s.updateRedisFriendList(userId, usersorter.CustomUserData{
			UserID:    info2.UserId,
			NickName:  info2.NickName,
			Email:     info2.Email,
			Tel:       info2.Tel,
			Avatar:    info2.Avatar,
			Signature: info2.Signature,
			Status:    uint(info2.Status),
			DialogId:  relation2.DialogId,
		})
		if err != nil {
			return nil, err
		}
	}

	// 向用户推送通知
	resp, err := s.sendFriendManagementNotification(ctx, userId, qs.SenderId, key, relationgrpcv1.RelationStatus(action))
	if err != nil {
		s.logger.Error("发送好友管理通知失败", zap.Error(err))
	}

	return resp, nil
}

func (s *Service) DeleteFriend(ctx context.Context, userID, friendID string) error {
	if s.cache {
		//查询是否有缓存
		values, err := s.redisClient.GetAllListValues(fmt.Sprintf("friend:%s", userID))
		if err != nil {
			s.logger.Error("err:" + err.Error())
			return code.RelationErrDeleteFriendFailed
		}
		if len(values) == 1 {
			err := s.redisClient.DelKey(fmt.Sprintf("friend:%s", userID))
			if err != nil {
				return code.RelationErrDeleteFriendFailed
			}
		}
	}

	relation, err := s.userRelationClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: userID, FriendId: friendID})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		return code.RelationErrRelationNotFound
	}

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

	if s.cache {
		err = s.removeRedisUserDialogList(relation.UserId, relation.DialogId)
		if err != nil {
			s.logger.Error("删除用户好友失败", zap.Error(err))
			return code.RelationErrDeleteFriendFailed
		}

		err = s.removeRedisFriendList(relation.UserId, relation.FriendId)
		if err != nil {
			s.logger.Error("删除用户好友失败", zap.Error(err))
			return code.RelationErrDeleteFriendFailed
		}
	}

	return nil
}

func (s *Service) AddBlacklist(ctx context.Context, userID, friendID string) (interface{}, error) {
	_, err := s.userRelationClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: userID, FriendId: friendID})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		return nil, err
	}

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
	msg := constants.WsMsg{Uid: friendID, Event: constants.PushE2EPublicKeyEvent, Data: reqm, SendAt: time.Now()}

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

	_, err = s.userRelationClient.SetFriendSilentNotification(ctx, &relationgrpcv1.SetFriendSilentNotificationRequest{
		UserId:   userID,
		FriendId: friendID,
		IsSilent: relationgrpcv1.UserSilentNotificationType(silent),
	})
	if err != nil {
		s.logger.Error("设置好友静默通知失败", zap.Error(err))
		return nil, err
	}
	return nil, nil
}

func (s *Service) SetUserBurnAfterReading(ctx context.Context, userId string, req *model.OpenUserBurnAfterReadingRequest) (interface{}, error) {
	_, err := s.userRelationClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
		UserId:   userId,
		FriendId: req.UserId,
	})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		return nil, err
	}

	_, err = s.userRelationClient.SetUserOpenBurnAfterReading(ctx, &relationgrpcv1.SetUserOpenBurnAfterReadingRequest{
		UserId:               userId,
		FriendId:             req.UserId,
		OpenBurnAfterReading: relationgrpcv1.OpenBurnAfterReadingType(req.Action),
	})
	if err != nil {
		s.logger.Error("设置用户消息阅后即焚失败", zap.Error(err))
		return nil, err
	}
	return nil, nil
}

// manageFriend2 不存在好友关系，创建新的关系
func (s *Service) manageFriend2(ctx context.Context, userId, friendId string, status relationgrpcv1.RelationStatus) (uint32, error) {
	var dialogId uint32
	// 创建 DTM 分布式事务工作流
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
		s.logger.Error("workflow RegisterGRPC manageFriend2", zap.Error(err))
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
		s.logger.Error("workflow RegisterGRPC manageFriend3", zap.Error(err))
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

	wsMsgData := constants.ManageFriendEventData{UserId: userID, Status: uint32(status)}
	msg := constants.WsMsg{Uid: targetID, Event: constants.ManageFriendEvent, Data: wsMsgData}
	var responseData interface{}

	if status == 1 {
		wsMsgData.E2EPublicKey = E2EPublicKey
		wsMsgData.TargetInfo = myInfo
		responseData = targetInfo
	}

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

func (s *Service) SetUserFriendRemark(ctx context.Context, userID string, req *model.SetUserFriendRemarkRequest) (interface{}, error) {
	_, err := s.userRelationClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
		UserId:   userID,
		FriendId: req.UserId,
	})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		return nil, err
	}

	_, err = s.userRelationClient.SetFriendRemark(ctx, &relationgrpcv1.SetFriendRemarkRequest{
		UserId:   userID,
		Remark:   req.Remark,
		FriendId: req.UserId,
	})
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *Service) publishServiceMessage(ctx context.Context, msg constants.WsMsg) error {
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

func (s *Service) convertStatusToRelationStatus(status uint32) relationgrpcv1.FriendRequestStatus {
	switch status {
	case 0:
		return relationgrpcv1.FriendRequestStatus_FriendRequestStatus_REJECT
	case 1:
		return relationgrpcv1.FriendRequestStatus_FriendRequestStatus_ACCEPT
	default:
		return relationgrpcv1.FriendRequestStatus_FriendRequestStatus_REJECT
	}
}

// 更新redis里的对话列表数据
func (s *Service) updateRedisUserDialogList(userID string, msg model.UserDialogListResponse) error {
	//判断key是否存在，存在才继续
	f, err := s.redisClient.ExistsKey(fmt.Sprintf("dialog:%s", userID))
	if err != nil {
		return err
	}
	if !f {
		return nil
	}
	//查询是否有缓存
	values, err := s.redisClient.GetAllListValues(fmt.Sprintf("dialog:%s", userID))
	if err != nil {
		return err
	}
	if len(values) > 0 {
		// 类型转换
		var responseList []model.UserDialogListResponse
		for _, v := range values {
			// 在这里根据实际的数据结构进行解析
			// 这里假设你的缓存数据是 JSON 字符串，需要解析为 UserDialogListResponse 类型
			var dialog model.UserDialogListResponse
			err := json.Unmarshal([]byte(v), &dialog)
			if err != nil {
				fmt.Println("Error decoding cached data:", err)
				continue
			}
			responseList = append(responseList, dialog)
		}

		for i, v := range responseList {
			if v.DialogId == msg.DialogId {
				//替换该位置的
				responseList[i] = msg
			}
		}

		//保存回redis
		// 创建一个新的[]interface{}类型的数组
		var interfaceList []interface{}

		// 遍历responseList数组，并将每个元素转换为interface{}类型后添加到interfaceList数组中
		for _, dialog := range responseList {
			interfaceList = append(interfaceList, dialog)
		}

		err := s.redisClient.DelKey(fmt.Sprintf("dialog:%s", userID))
		if err != nil {
			return err
		}

		//存储到缓存
		err = s.redisClient.AddToList(fmt.Sprintf("dialog:%s", userID), interfaceList)
		if err != nil {
			return err
		}
		//设置key过期时间
		err = s.redisClient.SetKeyExpiration(fmt.Sprintf("dialog:%s", userID), 3*24*ostime.Hour)
		if err != nil {
			return err
		}
	}
	return nil
}

// 添加redis里的对话列表数据
func (s *Service) insertRedisUserDialogList(userID string, msg model.UserDialogListResponse) error {
	//判断key是否存在，存在才继续
	f, err := s.redisClient.ExistsKey(fmt.Sprintf("dialog:%s", userID))
	if err != nil {
		return err
	}
	if !f {
		return nil
	}
	//查询是否有缓存
	values, err := s.redisClient.GetAllListValues(fmt.Sprintf("dialog:%s", userID))
	if err != nil {
		return err
	}
	if len(values) > 0 {
		// 类型转换
		var responseList []model.UserDialogListResponse
		responseList = append(responseList, msg)
		for _, v := range values {
			var dialog model.UserDialogListResponse
			err := json.Unmarshal([]byte(v), &dialog)
			if err != nil {
				fmt.Println("Error decoding cached data:", err)
				continue
			}
			responseList = append(responseList, dialog)
		}

		//保存回redis
		// 创建一个新的[]interface{}类型的数组
		var interfaceList []interface{}

		// 遍历responseList数组，并将每个元素转换为interface{}类型后添加到interfaceList数组中
		for _, dialog := range responseList {
			interfaceList = append(interfaceList, dialog)
		}

		err := s.redisClient.DelKey(fmt.Sprintf("dialog:%s", userID))
		if err != nil {
			return err
		}

		//存储到缓存
		err = s.redisClient.AddToList(fmt.Sprintf("dialog:%s", userID), interfaceList)
		if err != nil {
			return err
		}
		//设置key过期时间
		err = s.redisClient.SetKeyExpiration(fmt.Sprintf("dialog:%s", userID), 3*24*ostime.Hour)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) removeRedisUserDialogList(userID string, dialogID uint32) error {
	//判断key是否存在，存在才继续
	f, err := s.redisClient.ExistsKey(fmt.Sprintf("dialog:%s", userID))
	if err != nil {
		return err
	}
	if !f {
		return nil
	}

	//查询是否有缓存
	values, err := s.redisClient.GetAllListValues(fmt.Sprintf("dialog:%s", userID))
	if err != nil {
		return err
	}
	if len(values) > 0 {
		// 类型转换
		var responseList []model.UserDialogListResponse
		if len(values) == 1 {
			err := s.redisClient.DelKey(fmt.Sprintf("dialog:%s", userID))
			if err != nil {
				return err
			}
			return nil
		}
		for _, v := range values {
			var dialog model.UserDialogListResponse
			err := json.Unmarshal([]byte(v), &dialog)
			if err != nil {
				fmt.Println("Error decoding cached data:", err)
				return err
			}
			if dialog.DialogId != dialogID {
				responseList = append(responseList, dialog)
			}
		}
		if len(responseList) == 0 {
			return nil
		}
		//保存回redis
		// 创建一个新的[]interface{}类型的数组
		var interfaceList []interface{}

		// 遍历responseList数组，并将每个元素转换为interface{}类型后添加到interfaceList数组中
		for _, dialog := range responseList {
			interfaceList = append(interfaceList, dialog)
		}

		err = s.redisClient.DelKey(fmt.Sprintf("dialog:%s", userID))
		if err != nil {
			return err
		}
		//存储到缓存
		err = s.redisClient.AddToList(fmt.Sprintf("dialog:%s", userID), interfaceList)
		if err != nil {
			return err
		}
		//设置key过期时间
		err = s.redisClient.SetKeyExpiration(fmt.Sprintf("dialog:%s", userID), 3*24*ostime.Hour)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) SetUserOpenBurnAfterReadingTimeOut(ctx context.Context, userID string, req *model.SetUserOpenBurnAfterReadingTimeOutRequest) error {
	_, err := s.userRelationClient.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
		UserId:   userID,
		FriendId: req.FriendId,
	})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		return err
	}

	_, err = s.userRelationClient.SetUserOpenBurnAfterReadingTimeOut(ctx, &relationgrpcv1.SetUserOpenBurnAfterReadingTimeOutRequest{
		UserId:                      userID,
		FriendId:                    req.FriendId,
		OpenBurnAfterReadingTimeOut: req.OpenBurnAfterReadingTimeOut,
	})
	if err != nil {
		s.logger.Error("设置用户消息阅后即焚失败", zap.Error(err))
		return err
	}
	return nil
}

func (s *Service) updateRedisFriendList(userID string, msg usersorter.CustomUserData) error {
	key := fmt.Sprintf("friend:%s", userID)
	exists, err := s.redisClient.ExistsKey(key)
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}

	err = s.redisClient.AddToListLeft(key, []interface{}{msg})
	if err != nil {
		return err
	}
	return nil
}

func (s *Service) removeRedisFriendList(userID string, friendID string) error {
	key := fmt.Sprintf("friend:%s", userID)
	//判断key是否存在，存在才继续
	f, err := s.redisClient.ExistsKey(key)
	if err != nil {
		return err
	}
	if !f {
		return nil
	}

	length, err := s.redisClient.GetListLength(key)
	if err != nil {
		return err
	}

	if length > 10 {
		for i := int64(0); i < length; i += 10 {
			stop := i + 9
			if stop >= length {
				stop = length - 1
			}

			// 获取当前范围内的元素
			values, err := s.redisClient.GetList(key, i, stop)
			if err != nil {
				return err
			}

			// 遍历当前范围内的元素
			for j, v := range values {
				var user usersorter.CustomUserData
				err := json.Unmarshal([]byte(v), &user)
				if err != nil {
					fmt.Println("Error decoding cached data:", err)
					return err
				}
				if user.UserID == friendID {
					// 弹出指定位置的元素
					_, err := s.redisClient.PopListElement(key, i+int64(j))
					if err != nil {
						return err
					}
				}
			}
		}
	} else {
		values, err := s.redisClient.GetAllListValues(key)
		if err != nil {
			return err
		}
		for i, v := range values {
			var user usersorter.CustomUserData
			err := json.Unmarshal([]byte(v), &user)
			if err != nil {
				fmt.Println("Error decoding cached data:", err)
				return err
			}
			if user.UserID == friendID {
				_, err := s.redisClient.PopListElement(key, int64(i))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
