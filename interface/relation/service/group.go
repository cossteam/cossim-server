package service

import (
	"context"
	"errors"
	msgconfig "github.com/cossim/coss-server/interface/msg/config"
	"github.com/cossim/coss-server/interface/relation/api/model"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/msg_queue"
	"github.com/cossim/coss-server/pkg/utils/time"
	"github.com/cossim/coss-server/pkg/utils/usersorter"
	groupApi "github.com/cossim/coss-server/service/group/api/v1"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	userApi "github.com/cossim/coss-server/service/user/api/v1"
	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/lithammer/shortuuid/v3"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) GetGroupMember(ctx context.Context, gid uint32) (interface{}, error) {
	group, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: gid})
	if err != nil {
		s.logger.Error("获取群聊成员失败", zap.Error(err))
		return nil, err
	}

	if group.Status != groupApi.GroupStatus_GROUP_STATUS_NORMAL {
		return nil, code.GroupErrGroupStatusNotAvailable
	}

	groupRelation, err := s.groupRelationClient.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{GroupId: gid})
	if err != nil {
		s.logger.Error("获取群聊成员失败", zap.Error(err))
		return nil, err
	}

	resp, err := s.userClient.GetBatchUserInfo(ctx, &userApi.GetBatchUserInfoRequest{UserIds: groupRelation.UserIds})
	if err != nil {
		s.logger.Error("获取群聊成员失败", zap.Error(err))
		return nil, err
	}

	type requestListResponse struct {
		UserID   string `json:"user_id"`
		Nickname string `json:"nickname"`
		Avatar   string `json:"avatar"`
	}

	var ids []string
	var data []*requestListResponse
	for _, v := range resp.Users {
		ids = append(ids, v.UserId)
		data = append(data, &requestListResponse{
			UserID:   v.UserId,
			Nickname: v.NickName,
			Avatar:   v.Avatar,
		})
	}
	return data, nil
}

func (s *Service) JoinGroup(ctx context.Context, uid string, req *model.JoinGroupRequest) (interface{}, error) {
	group, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupID})
	if err != nil {
		s.logger.Error("获取群聊信息失败", zap.Error(err))
		return nil, err
	}

	if group.Status != groupApi.GroupStatus_GROUP_STATUS_NORMAL {
		return nil, code.GroupErrGroupStatusNotAvailable
	}
	//判断是否在群聊中
	relation, err := s.groupRelationClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		GroupId: req.GroupID,
		UserId:  uid,
	})
	if relation != nil {
		return nil, code.RelationGroupErrAlreadyInGroup
	}

	//添加普通用户申请
	_, err = s.groupJoinRequestClient.JoinGroup(context.Background(), &relationgrpcv1.JoinGroupRequest{UserId: uid, GroupId: req.GroupID})
	if err != nil {
		s.logger.Error("添加群聊申请失败", zap.Error(err))
		return nil, err
	}
	//查询所有管理员
	adminIds, err := s.groupRelationClient.GetGroupAdminIds(context.Background(), &relationgrpcv1.GroupIDRequest{
		GroupId: req.GroupID,
	})
	for _, id := range adminIds.UserIds {
		msg := msgconfig.WsMsg{Uid: id, Event: msgconfig.JoinGroupEvent, Data: map[string]interface{}{"group_id": req.GroupID, "user_id": uid}, SendAt: time.Now()}
		//通知消息服务有消息需要发送
		err = s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
		if err != nil {
			s.logger.Error("加入群聊请求申请通知推送失败", zap.Error(err))
		}
	}
	return nil, nil
}

func (s *Service) GetUserGroupList(ctx context.Context, userID string) (interface{}, error) {
	// 获取用户群聊列表
	ids, err := s.groupRelationClient.GetUserGroupIDs(context.Background(), &relationgrpcv1.GetUserGroupIDsRequest{UserId: userID})
	if err != nil {
		s.logger.Error("获取用户群聊列表失败", zap.Error(err))
		return nil, err
	}

	ds, err := s.groupClient.GetBatchGroupInfoByIDs(context.Background(), &groupApi.GetBatchGroupInfoRequest{GroupIds: ids.GroupId})
	if err != nil {
		s.logger.Error("获取群聊列表失败", zap.Error(err))
		return nil, err
	}
	//获取群聊对话信息
	dialogs, err := s.dialogClient.GetDialogByGroupIds(context.Background(), &relationgrpcv1.GetDialogByGroupIdsRequest{GroupId: ids.GroupId})
	if err != nil {
		s.logger.Error("获取群聊对话列表失败", zap.Error(err))
		return nil, err
	}

	var data []usersorter.Group
	for _, group := range ds.Groups {
		for _, dialog := range dialogs.Dialogs {
			if dialog.GroupId == group.Id {
				data = append(data, usersorter.CustomGroupData{
					GroupID:  group.Id,
					Avatar:   group.Avatar,
					Status:   uint(group.Status),
					DialogId: dialog.DialogId,
					Name:     group.Name,
				})
				break
			}
		}

	}

	return usersorter.SortAndGroupUsers(data, "Name"), nil
}

func (s *Service) SetGroupSilentNotification(ctx context.Context, gid uint32, uid string, silent model.SilentNotificationType) (interface{}, error) {
	_, err := s.groupRelationClient.GetGroupRelation(context.Background(), &relationgrpcv1.GetGroupRelationRequest{
		GroupId: gid,
		UserId:  uid,
	})
	if err != nil {
		s.logger.Error("获取群聊关系失败", zap.Error(err))
		return nil, err
	}

	_, err = s.groupRelationClient.SetGroupSilentNotification(context.Background(), &relationgrpcv1.SetGroupSilentNotificationRequest{
		GroupId:  gid,
		UserId:   uid,
		IsSilent: relationgrpcv1.GroupSilentNotificationType(silent),
	})
	if err != nil {
		s.logger.Error("设置群聊静默通知失败", zap.Error(err))
		return nil, err
	}

	return nil, nil
}

func (s *Service) GroupRequestList(ctx context.Context, userID string) (interface{}, error) {
	reqList, err := s.groupJoinRequestClient.GetGroupJoinRequestListByUserId(ctx, &relationgrpcv1.GetGroupJoinRequestListRequest{UserId: userID})
	if err != nil {
		s.logger.Error("获取群聊申请列表失败", zap.Error(err))
		return nil, err
	}

	gids := make([]uint32, len(reqList.GroupJoinRequestResponses))
	uids := make([]string, len(reqList.GroupJoinRequestResponses))
	data := make([]*model.GroupRequestListResponse, len(reqList.GetGroupJoinRequestResponses()))

	for i, v := range reqList.GroupJoinRequestResponses {
		gids[i] = v.GroupId
		uids[i] = v.UserId

		data[i] = &model.GroupRequestListResponse{
			ID:          v.ID,
			GroupId:     v.GroupId,
			UserID:      v.UserId,
			Msg:         v.Remark,
			GroupStatus: uint32(v.Status),
		}
	}

	groupMap := make(map[uint32]*model.GroupRequestListResponse)
	for _, d := range data {
		groupMap[d.GroupId] = d
	}

	groupIDs := make([]uint32, 0, len(groupMap))
	for groupID := range groupMap {
		groupIDs = append(groupIDs, groupID)
	}

	groupInfoMap := make(map[uint32]*groupApi.Group)
	for _, groupID := range groupIDs {
		groupInfo, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: groupID})
		if err != nil {
			s.logger.Error("获取群聊信息失败", zap.Error(err))
			return nil, err
		}

		groupInfoMap[groupID] = groupInfo
	}

	users, err := s.userClient.GetBatchUserInfo(ctx, &userApi.GetBatchUserInfoRequest{UserIds: uids})
	if err != nil {
		s.logger.Error("获取群聊成员失败", zap.Error(err))
		return nil, err
	}

	for _, v := range data {
		groupID := v.GroupId
		if groupInfo, ok := groupInfoMap[groupID]; ok {
			v.GroupName = groupInfo.Name
			v.GroupAvatar = groupInfo.Avatar
		}

		for _, u := range users.Users {
			if v.UserID == u.UserId {
				v.UserName = u.NickName
				v.UserAvatar = u.Avatar
				break
			}
		}
	}
	return data, nil
}

func (s *Service) AdminManageJoinGroup(ctx context.Context, requestID, groupID uint32, userID string, status relationgrpcv1.GroupRequestStatus) error {
	group, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: groupID})
	if err != nil {
		s.logger.Error("get group info failed", zap.Error(err))
		return code.RelationGroupErrGetGroupInfoFailed
	}

	if group.Status != groupApi.GroupStatus_GROUP_STATUS_NORMAL {
		return code.GroupErrGroupStatusNotAvailable
	}

	relation1, err := s.groupRelationClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: groupID, UserId: userID})
	if err != nil {
		s.logger.Error("get group relation failed", zap.Error(err))
		return code.RelationGroupErrGetGroupInfoFailed
	}

	if relation1.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return code.Unauthorized
	}

	_, err = s.groupJoinRequestClient.ManageGroupJoinRequestByID(ctx, &relationgrpcv1.ManageGroupJoinRequestByIDRequest{
		ID:     requestID,
		Status: status,
	})
	if err != nil {
		return err
	}

	msg := msgconfig.WsMsg{
		Uid:    userID,
		Event:  msgconfig.JoinGroupEvent,
		Data:   map[string]interface{}{"group_id": groupID, "status": status},
		SendAt: time.Now(),
	}
	err = s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
	if err != nil {
		s.logger.Error("通知消息服务有消息需要发送失败", zap.Error(err))
	}

	return nil
}

func (s *Service) ManageJoinGroup(ctx context.Context, groupID uint32, requestID uint32, userID string, status relationgrpcv1.GroupRequestStatus) error {

	_, err := s.groupJoinRequestClient.GetGroupJoinRequestByGroupIdAndUserId(ctx, &relationgrpcv1.GetGroupJoinRequestByGroupIdAndUserIdRequest{GroupId: groupID, UserId: userID})
	if err != nil {
		s.logger.Error("找不到入群请求", zap.Error(err))
		return code.RelationErrManageFriendRequestFailed
	}

	_, err = s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: groupID})
	if err != nil {
		s.logger.Error("get group info failed", zap.Error(err))
		return code.RelationGroupErrGetGroupInfoFailed
	}

	relation, err := s.groupRelationClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: groupID, UserId: userID})
	if err != nil {
		s.logger.Error("get group relation failed", zap.Error(err))
	}

	if relation != nil {
		return code.RelationGroupErrAlreadyInGroup
	}

	_, err = s.groupJoinRequestClient.ManageGroupJoinRequestByID(ctx, &relationgrpcv1.ManageGroupJoinRequestByIDRequest{
		ID:     requestID,
		Status: status,
	})
	if err != nil {
		return err
	}

	msg := msgconfig.WsMsg{
		Uid:    userID,
		Event:  msgconfig.JoinGroupEvent,
		Data:   map[string]interface{}{"group_id": groupID, "status": status},
		SendAt: time.Now(),
	}
	if err = s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg); err != nil {
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

	_, err = s.groupRelationClient.GetGroupRelation(context.Background(), &relationgrpcv1.GetGroupRelationRequest{UserId: userID, GroupId: groupID})
	if err != nil {
		s.logger.Error("获取用户群组关系失败", zap.Error(err))
		return errors.New("获取用户群组关系失败")
	}

	dialog, err := s.dialogClient.GetDialogByGroupId(ctx, &relationgrpcv1.GetDialogByGroupIdRequest{GroupId: groupID})
	if err != nil {
		return errors.New("获取群聊会话失败")
	}

	r1 := &relationgrpcv1.DeleteDialogUserByDialogIDAndUserIDRequest{DialogId: dialog.DialogId, UserId: userID}
	r2 := &relationgrpcv1.DeleteGroupRelationByGroupIdAndUserIDRequest{GroupID: groupID, UserID: userID}

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

//func (s *Service) validateAdminGroupRelationStatus(relation *relationgrpcv1.GetGroupRelationResponse, status relationgrpcv1.GroupRelationStatus) error {
//	switch status {
//	case relationgrpcv1.GroupRelationStatus_GroupStatusJoined:
//		if relation.Status != relationgrpcv1.GroupRelationStatus_GroupStatusApplying {
//			return errors.New("没有申请记录")
//		}
//
//		if relation.Status == relationgrpcv1.GroupRelationStatus_GroupStatusJoined {
//			return errors.New("已经在群里了")
//		}
//
//	case relationgrpcv1.GroupRelationStatus_GroupStatusReject:
//		if relation.Status != relationgrpcv1.GroupRelationStatus_GroupStatusApplying {
//			return errors.New("没有处于申请中")
//		}
//	default:
//		return errors.New("用户状态异常")
//	}
//
//	return nil
//}
//
//func (s *Service) validateGroupRelationStatus(relation *relationgrpcv1.GetGroupRelationResponse, status relationgrpcv1.GetGroupRelationRequest) error {
//	switch status {
//	case relationgrpcv1.GroupRelationStatus_GroupStatusJoined:
//		fmt.Println("relation => ", relation)
//		if relation.Status == relationgrpcv1.GroupRelationStatus_GroupStatusJoined {
//			return code.RelationGroupErrAlreadyInGroup
//		}
//		if relation.Status != relatio  ngrpcv1.GroupRelationStatus_GroupStatusPending {
//			return errors.New("没有待处理的群聊申请记录")
//		}
//
//	//case relationgrpcv1.GroupRelationStatus_GroupStatusReject:
//	//	if relation.Status != relationgrpcv1.GroupRelationStatus_GroupStatusPending {
//	//		return errors.New("群聊申请状态异常")
//	//	}
//	default:
//		return errors.New("群聊申请状态异常")
//	}
//
//	return nil
//}

func (s *Service) InviteGroup(ctx context.Context, inviterId string, req *model.InviteGroupRequest) error {
	group, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupID})
	if err != nil {
		return code.GroupErrGetGroupInfoByGidFailed
	}

	if group.Status != groupApi.GroupStatus_GROUP_STATUS_NORMAL {
		return code.GroupErrGroupStatusNotAvailable
	}

	relation1, err := s.groupRelationClient.GetGroupRelation(context.Background(), &relationgrpcv1.GetGroupRelationRequest{GroupId: req.GroupID, UserId: inviterId})
	if err != nil {
		s.logger.Error("获取用户群组关系失败", zap.Error(err))
		return code.RelationGroupErrGroupRelationFailed
	}

	if relation1.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return code.Unauthorized
	}

	grs, err := s.groupRelationClient.GetBatchGroupRelation(ctx, &relationgrpcv1.GetBatchGroupRelationRequest{GroupId: req.GroupID, UserIds: req.Member})
	if err != nil {
		s.logger.Error("获取用户群组关系失败", zap.Error(err))
		//if !errors.Is(code.Cause(err), code.RelationGroupErrRelationNotFound) {
		//	return code.RelationGroupErrInviteFailed
		//}
		return code.RelationGroupErrInviteFailed
	}

	if len(grs.GroupRelationResponses) > 0 {
		return code.RelationGroupErrInviteFailed
	}
	//TODO 添加群聊配置，（是否邀请入群需要管理员权限）
	_, err = s.groupJoinRequestClient.InviteJoinGroup(ctx, &relationgrpcv1.InviteJoinGroupRequest{
		GroupId:   req.GroupID,
		InviterId: inviterId,
		Member:    req.Member,
	})
	if err != nil {
		s.logger.Error("邀请用户加入群聊失败", zap.Error(err))
		return code.RelationGroupErrInviteFailed
	}

	//查询所有管理员
	adminIds, err := s.groupRelationClient.GetGroupAdminIds(context.Background(), &relationgrpcv1.GroupIDRequest{
		GroupId: req.GroupID,
	})
	for _, id := range adminIds.UserIds {
		msg := msgconfig.WsMsg{Uid: id, Event: msgconfig.JoinGroupEvent, Data: map[string]interface{}{"group_id": req.GroupID, "user_id": inviterId}, SendAt: time.Now()}
		//通知消息服务有消息需要发送
		err = s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
		if err != nil {
			s.logger.Error("加入群聊请求申请通知推送失败", zap.Error(err))
		}
	}

	// 给被邀请的用户推送
	//for _, id := range req.Member {
	//	msg := msgconfig.WsMsg{Uid: id, Event: msgconfig.InviteJoinGroupEvent, Data: map[string]interface{}{"group_id": req.GroupID, "inviter_id": inviterId}, SendAt: time.Now()}
	//	//通知消息服务有消息需要发送

	//	err = s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
	//}

	return nil
}
