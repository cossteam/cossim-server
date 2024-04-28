package service

import (
	"context"
	"errors"
	"fmt"
	msggrpcv1 "github.com/cossim/coss-server/internal/msg/api/grpc/v1"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/api/http/model"
	userApi "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/cossim/coss-server/pkg/utils/time"
	"github.com/cossim/coss-server/pkg/utils/usersorter"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/workflow"
	any2 "github.com/golang/protobuf/ptypes/any"
	"github.com/lithammer/shortuuid/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Service) FriendList(ctx context.Context, userID string) (interface{}, error) {
	// 获取好友列表
	friendListResp, err := s.relationUserService.GetFriendList(ctx, &relationgrpcv1.GetFriendListRequest{UserId: userID})
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

	userInfos, err := s.userService.GetBatchUserInfo(ctx, &userApi.GetBatchUserInfoRequest{UserIds: users})
	if err != nil {
		s.logger.Error("user service GetBatchUserInfo", zap.Error(err))
		return nil, err
	}

	var data []usersorter.User
	for _, v := range userInfos.Users {
		for _, friend := range friendListResp.FriendList {
			if friend.UserId == v.UserId {
				pre := &usersorter.Preferences{
					OpenBurnAfterReading:        uint32(friend.OpenBurnAfterReading),
					OpenBurnAfterReadingTimeOut: friend.OpenBurnAfterReadingTimeOut,
					SilentNotification:          uint32(friend.IsSilent),
					Remark:                      friend.Remark,
				}
				data = append(data, usersorter.CustomUserData{
					UserID:         v.UserId,
					NickName:       v.NickName,
					CossId:         v.CossId,
					Email:          v.Email,
					Tel:            v.Tel,
					Avatar:         v.Avatar,
					Status:         uint(v.Status),
					DialogId:       friend.DialogId,
					RelationStatus: uint32(friend.Status),
					Signature:      v.Signature,
					Preferences:    pre,
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

	// Sort and group by specified field
	groupedUsers := usersorter.SortAndGroupUsers(data, "NickName")
	return groupedUsers, nil
}

func (s *Service) BlackList(ctx context.Context, userID string) (interface{}, error) {
	// 获取黑名单列表
	blacklistResp, err := s.relationUserService.GetBlacklist(ctx, &relationgrpcv1.GetBlacklistRequest{UserId: userID})
	if err != nil {
		s.logger.Error("user service GetBlacklist", zap.Error(err))
		return nil, err
	}

	var users []string
	for _, user := range blacklistResp.Blacklist {
		users = append(users, user.UserId)
	}

	blacklist, err := s.userService.GetBatchUserInfo(ctx, &userApi.GetBatchUserInfoRequest{UserIds: users})
	if err != nil {
		s.logger.Error("user service GetBatchUserInfo", zap.Error(err))
		return nil, err
	}

	return blacklist, nil
}

func (s *Service) UserRequestList(ctx context.Context, userID string, pageSize int, pageNum int) (interface{}, error) {

	reqList, err := s.relationUserFriendRequestService.GetFriendRequestList(ctx, &relationgrpcv1.GetFriendRequestListRequest{UserId: userID, PageSize: uint32(pageSize), PageNum: uint32(pageNum)})
	if err != nil {
		s.logger.Error("user service GetFriendRequestList", zap.Error(err))
		return nil, err
	}

	fmt.Println("reqList => ", reqList)

	var data []*model.UserRequestListResponse
	for _, v := range reqList.FriendRequestList {
		if v.SenderId == userID {
			info, err := s.userService.UserInfo(ctx, &userApi.UserInfoRequest{UserId: v.ReceiverId})
			if err != nil {
				return nil, err
			}
			data = append(data, &model.UserRequestListResponse{
				ID:         v.ID,
				ReceiverId: v.ReceiverId,
				Remark:     v.Remark,
				Status:     uint32(v.Status),
				SenderId:   v.SenderId,
				CreateAt:   int64(v.CreateAt),
				ReceiverInfo: &model.UserInfo{
					UserID:     info.UserId,
					UserName:   info.NickName,
					UserAvatar: info.Avatar,
				},
			})
		} else {
			info, err := s.userService.UserInfo(ctx, &userApi.UserInfoRequest{UserId: v.SenderId})
			if err != nil {
				return nil, err
			}
			data = append(data, &model.UserRequestListResponse{
				ID:         v.ID,
				ReceiverId: v.ReceiverId,
				Remark:     v.Remark,
				Status:     uint32(v.Status),
				SenderId:   v.SenderId,
				CreateAt:   int64(v.CreateAt),
				ReceiverInfo: &model.UserInfo{
					UserID:     info.UserId,
					UserName:   info.NickName,
					UserAvatar: info.Avatar,
				},
			})
		}
	}
	return &model.UserRequestListResponseList{
		List:        data,
		Total:       int64(reqList.Total),
		CurrentPage: int32(pageNum),
	}, nil
}
func (s *Service) SendFriendRequest(ctx context.Context, userID string, req *model.SendFriendRequest) (interface{}, error) {
	friendID := req.UserId
	if friendID == userID {
		return nil, code.MyCustomErrorCode.CustomMessage("不能添加自己为好友")
	}
	_, err := s.userService.UserInfo(ctx, &userApi.UserInfoRequest{UserId: friendID})
	if err != nil {
		s.logger.Error("获取用户信息失败", zap.Error(err))
		return nil, code.UserErrNotExist
	}

	fRequest, err := s.relationUserFriendRequestService.GetFriendRequestByUserIdAndFriendId(ctx, &relationgrpcv1.GetFriendRequestByUserIdAndFriendIdRequest{
		UserId:   userID,
		FriendId: friendID,
	})
	if err != nil {
		if code.Code(code.RelationUserErrNoFriendRequestRecords.Code()) != code.Cause(err) {
			s.logger.Error("获取好友请求失败", zap.Error(err))
			return nil, err
		}
	}
	if fRequest != nil {
		if fRequest.Status == relationgrpcv1.FriendRequestStatus_FriendRequestStatus_PENDING && fRequest.ID != 0 {
			return nil, code.RelationErrFriendRequestAlreadyPending
		} else if fRequest.Status == relationgrpcv1.FriendRequestStatus_FriendRequestStatus_ACCEPT {
			return nil, code.RelationErrRequestAlreadyProcessed
		}
	}

	relation, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: userID, FriendId: friendID})
	if err != nil {
		if code.Code(code.RelationUserErrFriendRelationNotFound.Code()) != code.Cause(err) {
			s.logger.Error("获取好友关系失败", zap.Error(err))
			return nil, err
		}
	}
	if relation != nil && relation.DialogId != 0 {
		return nil, code.RelationErrAlreadyFriends
	}

	relation2, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: friendID, FriendId: userID})
	if err != nil {
		if code.Code(code.RelationUserErrFriendRelationNotFound.Code()) != code.Cause(err) {
			s.logger.Error("获取好友关系失败", zap.Error(err))
			return nil, err
		}
	}

	// 单删之后再添加
	if relation2.Status == relationgrpcv1.RelationStatus_RELATION_NORMAL {
		_, err := s.relationUserService.AddFriendAfterDelete(ctx, &relationgrpcv1.AddFriendAfterDeleteRequest{
			UserId:   userID,
			FriendId: friendID,
		})
		if err != nil {
			s.logger.Error("单删添加好友失败", zap.Error(err))
			return nil, err
		}
		return nil, nil
	}

	// 删除之前的
	_, err = s.relationUserFriendRequestService.DeleteFriendRequestByUserIdAndFriendId(ctx, &relationgrpcv1.DeleteFriendRequestByUserIdAndFriendIdRequest{
		UserId:   userID,
		FriendId: friendID,
	})
	if err != nil {
		if code.Code(code.RelationUserErrNoFriendRequestRecords.Code()) != code.Cause(err) {
			return nil, err
		}
	}

	// 创建好友申请
	// 被拉黑了只创建自己的好友申请记录，对方是没有的
	resp, err := s.relationUserFriendRequestService.SendFriendRequest(ctx, &relationgrpcv1.SendFriendRequestStruct{
		SenderId:   userID,
		ReceiverId: friendID,
		Remark:     req.Remark,
	})
	if err != nil {
		s.logger.Error("添加好友失败", zap.Error(err))
		return nil, code.RelationErrSendFriendRequestFailed
	}

	// 被拉黑了不推送消息
	if relation2.Status == relationgrpcv1.RelationStatus_RELATION_STATUS_BLOCKED {
		return resp, nil
	}

	wsMsgData := constants.AddFriendEventData{
		UserId:       userID,
		Msg:          req.Remark,
		E2EPublicKey: req.E2EPublicKey,
	}
	bytes, err := utils.StructToBytes(wsMsgData)
	if err != nil {
		return nil, err
	}

	msg := &pushgrpcv1.WsMsg{Uid: friendID, Event: pushgrpcv1.WSEventType_AddFriendEvent, Data: &any2.Any{Value: bytes}}

	toBytes, err := utils.StructToBytes(msg)
	if err != nil {
		return nil, err
	}

	_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{Data: toBytes})
	if err != nil {
		s.logger.Error("Failed to push message", zap.Error(err))
	}

	return resp, nil
}

func (s *Service) ManageFriend(ctx context.Context, userId string, questId uint32, action int32, key string) (interface{}, error) {
	qs, err := s.relationUserFriendRequestService.GetFriendRequestById(ctx, &relationgrpcv1.GetFriendRequestByIdRequest{ID: questId})
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
	relation, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: qs.SenderId, FriendId: userId})
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

	_, err = s.relationUserFriendRequestService.ManageFriendRequest(ctx, &relationgrpcv1.ManageFriendRequestStruct{
		ID:     questId,
		Status: s.convertStatusToRelationStatus(uint32(action)),
	})
	if err != nil {
		return nil, err
	}

	// 向用户推送通知
	resp, err := s.sendFriendManagementNotification(ctx, userId, qs.SenderId, key, relationgrpcv1.RelationStatus(action))
	if err != nil {
		s.logger.Error("发送好友管理通知失败", zap.Error(err))
	}

	return resp, nil
}

func (s *Service) DeleteFriend(ctx context.Context, userID, friendID string) error {

	relation, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: userID, FriendId: friendID})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		if relation == nil || relation.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
			return code.RelationUserErrFriendRelationNotFound
		}
		return err
	}

	r1 := &relationgrpcv1.DeleteDialogUserByDialogIDAndUserIDRequest{DialogId: relation.DialogId, UserId: userID}
	r2 := &relationgrpcv1.DeleteFriendRequest{UserId: userID, FriendId: friendID}
	r3 := &msggrpcv1.DeleteUserMsgByDialogIdRequest{DialogId: relation.DialogId}
	gid := shortuuid.New()
	workflow.InitGrpc(s.dtmGrpcServer, s.relationServiceAddr, grpc.NewServer())
	wfName := "delete_relation_workflow_" + gid
	if err := workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		_, err := s.relationDialogService.DeleteDialogUserByDialogIDAndUserID(wf.Context, r1)
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
		}

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := s.relationDialogService.DeleteDialogUserByDialogIDAndUserIDRevert(wf.Context, r1)
			return err
		})

		_, err = s.relationUserService.DeleteFriend(wf.Context, r2)
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
		}

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err := s.relationUserService.DeleteFriendRevert(wf.Context, r2)
			return err
		})

		_, err = s.msgClient.ConfirmDeleteUserMessageByDialogId(wf.Context, r3)
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
		}

		return err
	}); err != nil {
		return err
	}

	if err := workflow.Execute(wfName, gid, nil); err != nil {
		return code.RelationErrDeleteFriendFailed
	}

	return nil
}

func (s *Service) AddBlacklist(ctx context.Context, userID, friendID string) (interface{}, error) {
	_, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: userID, FriendId: friendID})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		return nil, err
	}

	// 进行添加黑名单操作
	_, err = s.relationUserService.AddBlacklist(ctx, &relationgrpcv1.AddBlacklistRequest{UserId: userID, FriendId: friendID})
	if err != nil {
		s.logger.Error("添加黑名单失败", zap.Error(err))
		return nil, err
	}

	return nil, nil
}

func (s *Service) DeleteBlacklist(ctx context.Context, userID, friendID string) (interface{}, error) {
	relation, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{UserId: userID, FriendId: friendID})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		return nil, err
	}

	if relation.Status != relationgrpcv1.RelationStatus_RELATION_STATUS_BLOCKED {
		return nil, code.RelationErrNotInBlacklist
	}

	// 进行删除黑名单操作
	_, err = s.relationUserService.DeleteBlacklist(context.Background(), &relationgrpcv1.DeleteBlacklistRequest{UserId: userID, FriendId: friendID})
	if err != nil {
		s.logger.Error("删除黑名单失败", zap.Error(err))
		return nil, err
	}

	return nil, nil
}

func (s *Service) SwitchUserE2EPublicKey(ctx context.Context, userID string, friendID string, publicKey string) (interface{}, error) {
	reqm := &model.SwitchUserE2EPublicKeyRequest{
		UserId:    userID,
		PublicKey: publicKey,
	}
	bytes, err := utils.StructToBytes(reqm)
	if err != nil {
		return nil, err
	}
	msg := &pushgrpcv1.WsMsg{Uid: friendID, Event: pushgrpcv1.WSEventType_PushE2EPublicKeyEvent, Data: &any2.Any{Value: bytes}, SendAt: time.Now()}
	toBytes, err := utils.StructToBytes(msg)
	if err != nil {
		return nil, err
	}
	////通知消息服务有消息需要发送
	//if err := s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg); err != nil {
	//	s.logger.Error("交换用户端到端公钥通知推送失败", zap.Error(err))
	//	return nil, code.UserErrSwapPublicKeyFailed
	//}
	_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{
		Type: pushgrpcv1.Type_Ws,
		Data: toBytes,
	})
	if err != nil {
		s.logger.Error("发送消息失败", zap.Error(err))
	}
	return nil, nil
}

func (s *Service) UserSilentNotification(ctx context.Context, userID string, friendID string, silent model.SilentNotificationType) (interface{}, error) {
	_, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
		UserId:   userID,
		FriendId: friendID,
	})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		return nil, err
	}

	_, err = s.relationUserService.SetFriendSilentNotification(ctx, &relationgrpcv1.SetFriendSilentNotificationRequest{
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
	_, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
		UserId:   userId,
		FriendId: req.UserId,
	})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		return nil, err
	}

	_, err = s.relationUserService.SetUserOpenBurnAfterReading(ctx, &relationgrpcv1.SetUserOpenBurnAfterReadingRequest{
		UserId:               userId,
		FriendId:             req.UserId,
		OpenBurnAfterReading: relationgrpcv1.OpenBurnAfterReadingType(req.Action),
		TimeOut:              req.TimeOut,
	})
	if err != nil {
		s.logger.Error("设置用户消息阅后即焚失败", zap.Error(err))
		return nil, err
	}

	return nil, nil
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

	var responseData interface{}

	if status == 1 {
		wsMsgData.E2EPublicKey = E2EPublicKey
		wsMsgData.TargetInfo = myInfo
		responseData = targetInfo
	}
	bytes, err := utils.StructToBytes(wsMsgData)
	if err != nil {
		return nil, err
	}
	msg := &pushgrpcv1.WsMsg{Uid: targetID, Event: pushgrpcv1.WSEventType_ManageFriendEvent, Data: &any2.Any{Value: bytes}}
	toBytes, err := utils.StructToBytes(msg)
	if err != nil {
		return nil, err
	}

	_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{Data: toBytes, Type: pushgrpcv1.Type_Ws})
	if err != nil {
		s.logger.Error("发送消息失败", zap.Error(err))
	}

	return responseData, nil
}

func (s *Service) getUserInfo(ctx context.Context, userID string) (*userApi.UserInfoResponse, error) {
	req := &userApi.UserInfoRequest{UserId: userID}
	resp, err := s.userService.UserInfo(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *Service) SetUserFriendRemark(ctx context.Context, userID string, req *model.SetUserFriendRemarkRequest) (interface{}, error) {
	_, err := s.relationUserService.GetUserRelation(ctx, &relationgrpcv1.GetUserRelationRequest{
		UserId:   userID,
		FriendId: req.UserId,
	})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		return nil, err
	}

	_, err = s.relationUserService.SetFriendRemark(ctx, &relationgrpcv1.SetFriendRemarkRequest{
		UserId:   userID,
		Remark:   req.Remark,
		FriendId: req.UserId,
	})
	if err != nil {
		return nil, err
	}

	return nil, nil
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

func (s *Service) DeleteUserFriendRecord(ctx context.Context, uid string, id uint32) error {
	// 打印请求进来的日志
	s.logger.Debug("开始处理删除用户好友申请记录请求", zap.String("uid", uid), zap.Uint32("recordID", id))

	fr, err := s.relationUserFriendRequestService.GetFriendRequestById(ctx, &relationgrpcv1.GetFriendRequestByIdRequest{ID: id})
	if err != nil {
		s.logger.Debug("获取好友申请记录失败", zap.String("uid", uid), zap.Uint32("recordID", id), zap.Error(err))
		return err
	}

	if fr.OwnerID != uid {
		s.logger.Warn("用户没有权限删除该好友申请记录", zap.String("uid", uid), zap.Uint32("recordID", id))
		return code.Forbidden
	}

	_, err = s.relationUserFriendRequestService.DeleteFriendRecord(ctx, &relationgrpcv1.DeleteFriendRecordRequest{ID: id, UserId: uid})
	if err != nil {
		s.logger.Error("删除好友申请记录失败", zap.Error(err))
		return err
	}

	s.logger.Debug("成功删除好友申请记录", zap.String("uid", uid), zap.Stringer("FriendRecord", fr))

	return nil
}
