package service

import (
	"context"
	"encoding/json"
	"fmt"
	groupApi "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	"github.com/cossim/coss-server/internal/relation/api/http/model"
	userApi "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/utils"
	"github.com/cossim/coss-server/pkg/utils/time"
	"github.com/cossim/coss-server/pkg/utils/usersorter"
	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/lithammer/shortuuid/v3"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Service) AddGroupAdmin(ctx context.Context, uid string, req *model.AddGroupAdminRequest) error {
	group, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupID})
	if err != nil {
		s.logger.Error("获取群聊成员失败", zap.Error(err))
		return err
	}

	if group.Status != groupApi.GroupStatus_GROUP_STATUS_NORMAL {
		return code.GroupErrGroupStatusNotAvailable
	}

	r3, err := s.relationGroupService.GetBatchGroupRelation(ctx, &relationgrpcv1.GetBatchGroupRelationRequest{GroupId: req.GroupID, UserIds: req.UserIDs})
	if err != nil {
		s.logger.Error("获取群聊成员失败", zap.Error(err))
		return err
	}

	marshal, err := json.Marshal(r3.GroupRelationResponses)
	if err != nil {
		return err
	}

	if len(req.UserIDs) != len(r3.GroupRelationResponses) {
		s.logger.Error("添加的用户数量与查询后的关系数量不一致", zap.Strings("req", req.UserIDs), zap.String("resp", string(marshal)))
		return code.RelationGroupErrRelationNotFound
	}

	_, err = s.relationGroupService.AddGroupAdmin(ctx, &relationgrpcv1.AddGroupAdminRequest{
		GroupID: req.GroupID,
		UserID:  uid,
		UserIDs: req.UserIDs,
	})
	if err != nil {
		s.logger.Error("添加群聊管理员失败", zap.Error(err))
		return err
	}

	return nil
}

func (s *Service) GetGroupMember(ctx context.Context, gid uint32, userID string) (interface{}, error) {
	group, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: gid})
	if err != nil {
		s.logger.Error("获取群聊成员失败", zap.Error(err))
		return nil, err
	}

	if group.Status != groupApi.GroupStatus_GROUP_STATUS_NORMAL {
		return nil, code.GroupErrGroupStatusNotAvailable
	}

	_, err = s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: gid, UserId: userID})
	if err != nil {
		return nil, err
	}

	groupRelation, err := s.relationGroupService.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{GroupId: gid})
	if err != nil {
		s.logger.Error("获取群聊成员失败", zap.Error(err))
		return nil, err
	}

	relation, err := s.relationGroupService.GetBatchGroupRelation(ctx, &relationgrpcv1.GetBatchGroupRelationRequest{GroupId: gid, UserIds: groupRelation.UserIds})
	if err != nil {
		return nil, err
	}

	resp, err := s.userService.GetBatchUserInfo(ctx, &userApi.GetBatchUserInfoRequest{UserIds: groupRelation.UserIds})
	if err != nil {
		s.logger.Error("获取群聊成员失败", zap.Error(err))
		return nil, err
	}

	var ids []string
	var data []*model.RequestListResponse
	for i, v := range resp.Users {
		ids = append(ids, v.UserId)
		data = append(data, &model.RequestListResponse{
			UserID:   v.UserId,
			Nickname: v.NickName,
			Avatar:   v.Avatar,
		})
		for _, v1 := range relation.GroupRelationResponses {
			if v1.UserId == v.UserId {
				data[i].Identity = model.GroupRelationIdentity(v1.Identity)
				data[i].Remark = v1.Remark
			}
		}
	}
	return data, nil
}

func (s *Service) JoinGroup(ctx context.Context, uid string, req *model.JoinGroupRequest) (interface{}, error) {
	group, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupID})
	if err != nil {
		s.logger.Error("获取群聊信息失败", zap.Error(err))
		return nil, err
	}

	// 查询是不是已经有邀请
	id, err := s.relationGroupJoinRequestService.GetGroupJoinRequestByGroupIdAndUserId(ctx, &relationgrpcv1.GetGroupJoinRequestByGroupIdAndUserIdRequest{
		GroupId: req.GroupID,
		UserId:  uid,
	})
	if err != nil {
		s.logger.Error("获取群聊信息失败", zap.Error(err))
	}

	if id != nil && id.GroupId != 0 {
		return nil, code.RelationErrGroupRequestAlreadyProcessed
	}

	if group.Status != groupApi.GroupStatus_GROUP_STATUS_NORMAL {
		return nil, code.GroupErrGroupStatusNotAvailable
	}
	//判断是否在群聊中
	relation, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		GroupId: req.GroupID,
		UserId:  uid,
	})
	if relation != nil && relation.GroupId != 0 {
		return nil, code.RelationGroupErrAlreadyInGroup
	}

	groupRelation, err := s.relationGroupService.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{GroupId: req.GroupID})
	if err != nil {
		s.logger.Error("获取群聊成员失败", zap.Error(err))
		return nil, err
	}
	if len(groupRelation.UserIds) >= int(group.MaxMembersLimit) {
		return nil, code.RelationGroupErrGroupFull
	}

	// 添加群聊申请记录
	_, err = s.relationGroupJoinRequestService.JoinGroup(ctx, &relationgrpcv1.JoinGroupRequest{
		UserId:  uid,
		GroupId: req.GroupID,
	})
	if err != nil {
		s.logger.Error("添加群聊申请失败", zap.Error(err))
		return nil, err
	}

	//查询所有管理员
	adminIds, err := s.relationGroupService.GetGroupAdminIds(ctx, &relationgrpcv1.GroupIDRequest{
		GroupId: req.GroupID,
	})
	data := &constants.JoinGroupEventData{
		GroupId: req.GroupID,
		UserId:  uid,
	}
	bytes, err := utils.StructToBytes(data)
	if err != nil {
		return nil, err
	}

	for _, id := range adminIds.UserIds {
		msg := &pushgrpcv1.WsMsg{Uid: id, Event: pushgrpcv1.WSEventType_JoinGroupEvent, PushOffline: true, Data: &any.Any{Value: bytes}, SendAt: time.Now()}
		toBytes, err := utils.StructToBytes(msg)
		if err != nil {
			return nil, err
		}
		//通知消息服务有消息需要发送
		_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{Data: toBytes, Type: pushgrpcv1.Type_Ws})
		if err != nil {
			s.logger.Error("加入群聊请求申请通知推送失败", zap.Error(err))
		}
	}
	return nil, nil
}

func (s *Service) GetUserGroupList(ctx context.Context, userID string) (interface{}, error) {

	// 获取用户群聊列表
	ids, err := s.relationGroupService.GetUserGroupIDs(context.Background(), &relationgrpcv1.GetUserGroupIDsRequest{UserId: userID})
	if err != nil {
		s.logger.Error("获取用户群聊列表失败", zap.Error(err))
		return nil, err
	}
	if len(ids.GroupId) == 0 {
		return []string{}, nil
	}

	ds, err := s.groupService.GetBatchGroupInfoByIDs(context.Background(), &groupApi.GetBatchGroupInfoRequest{GroupIds: ids.GroupId})
	if err != nil {
		s.logger.Error("获取群聊列表失败", zap.Error(err))
		return nil, err
	}
	//获取群聊对话信息
	dialogs, err := s.relationDialogService.GetDialogByGroupIds(context.Background(), &relationgrpcv1.GetDialogByGroupIdsRequest{GroupId: ids.GroupId})
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
	_, err := s.relationGroupService.GetGroupRelation(context.Background(), &relationgrpcv1.GetGroupRelationRequest{
		GroupId: gid,
		UserId:  uid,
	})
	if err != nil {
		s.logger.Error("获取群聊关系失败", zap.Error(err))
		return nil, err
	}

	_, err = s.relationGroupService.SetGroupSilentNotification(context.Background(), &relationgrpcv1.SetGroupSilentNotificationRequest{
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

func (s *Service) SetGroupBurnAfterReading(ctx context.Context, userId string, req *model.OpenGroupBurnAfterReadingRequest) (interface{}, error) {
	_, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupId})
	if err != nil {
		s.logger.Error("获取群聊信息失败", zap.Error(err))
		return nil, err
	}

	_, err = s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		UserId:  userId,
		GroupId: req.GroupId,
	})
	if err != nil {
		s.logger.Error("获取群聊关系失败", zap.Error(err))
		return nil, err
	}

	_, err = s.relationGroupService.SetGroupOpenBurnAfterReading(ctx, &relationgrpcv1.SetGroupOpenBurnAfterReadingRequest{
		UserId:               userId,
		GroupId:              req.GroupId,
		OpenBurnAfterReading: relationgrpcv1.OpenBurnAfterReadingType(req.Action),
	})
	if err != nil {
		s.logger.Error("设置群聊消息阅后即焚失败", zap.Error(err))
		return nil, err
	}
	return nil, nil
}

func (s *Service) GroupRequestList(ctx context.Context, userID string) (interface{}, error) {
	reqList, err := s.relationGroupJoinRequestService.GetGroupJoinRequestListByUserId(ctx, &relationgrpcv1.GetGroupJoinRequestListRequest{UserId: userID})
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
		//查询发送者接受者信息
		reinfo, err := s.userService.UserInfo(ctx, &userApi.UserInfoRequest{UserId: v.UserId})
		if err != nil {
			return nil, err
		}
		sendinfo := &userApi.UserInfoResponse{}
		if v.InviterId != "" {
			sendinfo, err = s.userService.UserInfo(ctx, &userApi.UserInfoRequest{UserId: v.InviterId})
			if err != nil {
				return nil, err
			}
		} else {
			sendinfo, err = s.userService.UserInfo(ctx, &userApi.UserInfoRequest{UserId: v.UserId})
			if err != nil {
				return nil, err
			}
		}

		data[i] = &model.GroupRequestListResponse{
			ID:       v.ID,
			GroupId:  v.GroupId,
			Remark:   v.Remark,
			CreateAt: int64(v.CreatedAt),
			Status:   SwitchGroupRequestStatus(userID, v.InviterId, reinfo.UserId, v.Status),
			ReceiverInfo: &model.UserInfo{
				UserID:     reinfo.UserId,
				UserName:   reinfo.NickName,
				UserAvatar: reinfo.Avatar,
			},
		}
		if v.InviterId != "" {
			data[i].SenderInfo = &model.UserInfo{
				UserID:     sendinfo.UserId,
				UserName:   sendinfo.NickName,
				UserAvatar: sendinfo.Avatar,
			}
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
		groupInfo, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: groupID})
		if err != nil {
			s.logger.Error("获取群聊信息失败", zap.Error(err))
			return nil, err
		}

		groupInfoMap[groupID] = groupInfo
	}

	_, err = s.userService.GetBatchUserInfo(ctx, &userApi.GetBatchUserInfoRequest{UserIds: uids})
	if err != nil {
		s.logger.Error("获取群聊成员信息失败", zap.Error(err))
		return nil, err
	}

	for _, v := range data {
		groupID := v.GroupId
		if groupInfo, ok := groupInfoMap[groupID]; ok {
			v.GroupName = groupInfo.Name
			v.GroupAvatar = groupInfo.Avatar
			v.GroupStatus = uint32(groupInfo.Status)
		}
	}
	return data, nil
}

func (s *Service) AdminManageJoinGroup(ctx context.Context, requestID, groupID uint32, userID string, status relationgrpcv1.GroupRequestStatus) error {
	group, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: groupID})
	if err != nil {
		s.logger.Error("get group info failed", zap.Error(err))
		return code.RelationGroupErrGetGroupInfoFailed
	}

	if group.Status != groupApi.GroupStatus_GROUP_STATUS_NORMAL {
		return code.GroupErrGroupStatusNotAvailable
	}

	relation1, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: groupID, UserId: userID})
	if err != nil {
		s.logger.Error("get group relation failed", zap.Error(err))
		return code.RelationGroupErrGetGroupInfoFailed
	}

	if relation1.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return code.Unauthorized
	}

	info, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: groupID})
	if err != nil {
		s.logger.Error("get group info failed", zap.Error(err))
		return code.RelationGroupErrGetGroupInfoFailed
	}

	if status == relationgrpcv1.GroupRequestStatus_Accepted {
		ds, err := s.relationGroupService.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{GroupId: groupID})
		if err != nil {
			return err
		}
		if info.MaxMembersLimit <= int32(len(ds.UserIds)) {
			return code.RelationGroupErrGroupFull
		}
	}

	r, err := s.relationGroupJoinRequestService.GetGroupJoinRequestByID(ctx, &relationgrpcv1.GetGroupJoinRequestByIDRequest{ID: requestID})
	if err != nil {
		return err
	}

	_, err = s.relationGroupJoinRequestService.ManageGroupJoinRequestByID(ctx, &relationgrpcv1.ManageGroupJoinRequestByIDRequest{
		ID:     requestID,
		Status: status,
	})
	if err != nil {
		return err
	}

	data := &constants.JoinGroupEventData{GroupId: groupID, Status: uint32(status)}
	bytes, err := utils.StructToBytes(data)
	if err != nil {
		return err
	}

	msg := &pushgrpcv1.WsMsg{
		Uid:    r.UserId,
		Event:  pushgrpcv1.WSEventType_JoinGroupEvent,
		Data:   &any.Any{Value: bytes},
		SendAt: time.Now(),
	}

	toBytes, err := utils.StructToBytes(msg)
	if err != nil {
		return err
	}

	_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{Data: toBytes, Type: pushgrpcv1.Type_Ws})
	if err != nil {
		s.logger.Error("发送消息失败", zap.Error(err))
	}

	return nil
}

func (s *Service) ManageJoinGroup(ctx context.Context, groupID uint32, requestID uint32, userID string, status relationgrpcv1.GroupRequestStatus) error {
	_, err := s.relationGroupJoinRequestService.GetGroupJoinRequestByGroupIdAndUserId(ctx, &relationgrpcv1.GetGroupJoinRequestByGroupIdAndUserIdRequest{GroupId: groupID, UserId: userID})
	if err != nil {
		s.logger.Error("找不到入群请求", zap.Error(err))
		return code.RelationErrManageFriendRequestFailed
	}

	info, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: groupID})
	if err != nil {
		s.logger.Error("get group info failed", zap.Error(err))
		return code.RelationGroupErrGetGroupInfoFailed
	}

	relation, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: groupID, UserId: userID})
	if err != nil {
		s.logger.Error("get group relation failed", zap.Error(err))
	}

	if relation != nil && relation.GroupId != 0 {
		return code.RelationGroupErrAlreadyInGroup
	}

	if status == relationgrpcv1.GroupRequestStatus_Accepted {
		ds, err := s.relationGroupService.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{GroupId: groupID})
		if err != nil {
			return err
		}
		if info.MaxMembersLimit <= int32(len(ds.UserIds)) {
			return code.RelationGroupErrGroupFull
		}
	}

	_, err = s.relationGroupJoinRequestService.GetGroupJoinRequestByID(ctx, &relationgrpcv1.GetGroupJoinRequestByIDRequest{ID: requestID})
	if err != nil {
		return err
	}

	_, err = s.relationGroupJoinRequestService.ManageGroupJoinRequestByID(ctx, &relationgrpcv1.ManageGroupJoinRequestByIDRequest{
		ID:     requestID,
		Status: status,
	})
	if err != nil {
		return err
	}

	data := constants.JoinGroupEventData{GroupId: groupID, Status: uint32(status)}
	bytes, err := utils.StructToBytes(data)
	if err != nil {
		return err
	}
	msg := &pushgrpcv1.WsMsg{
		Uid:         userID,
		Event:       pushgrpcv1.WSEventType_JoinGroupEvent,
		Data:        &any.Any{Value: bytes},
		SendAt:      time.Now(),
		PushOffline: true,
	}

	msgBytes, err := utils.StructToBytes(msg)
	if err != nil {
		return err
	}

	_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{Data: msgBytes, Type: pushgrpcv1.Type_Ws})
	if err != nil {
		s.logger.Error("发送消息失败", zap.Error(err))
	}

	return nil
}

func (s *Service) CreateGroupAnnouncement(ctx context.Context, userId string, req *model.CreateGroupAnnouncementRequest) (interface{}, error) {
	_, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupId})
	if err != nil {
		return nil, err
	}

	relation, err := s.relationGroupService.GetGroupRelation(context.Background(), &relationgrpcv1.GetGroupRelationRequest{UserId: userId, GroupId: req.GroupId})
	if err != nil {
		return nil, err
	}

	if relation.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return nil, code.Forbidden
	}

	resp, err := s.relationGroupAnnouncementService.CreateGroupAnnouncement(ctx, &relationgrpcv1.CreateGroupAnnouncementRequest{
		GroupId: req.GroupId,
		UserId:  userId,
		Content: req.Content,
		Title:   req.Title,
	})

	if err != nil {
		return nil, err
	}

	//查询发送者信息
	info, err := s.userService.UserInfo(ctx, &userApi.UserInfoRequest{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}

	//查询群成员
	ds, err := s.relationGroupService.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{GroupId: req.GroupId})
	if err != nil {
		return nil, err
	}

	data := &model.WsGroupRelationOperatorMsg{
		Id:      resp.ID,
		GroupId: req.GroupId,
		Title:   req.Title,
		Content: req.Content,
		OperatorInfo: model.SenderInfo{
			UserId: info.UserId,
			Avatar: info.Avatar,
			Name:   info.NickName,
		},
	}

	bytes, err := utils.StructToBytes(data)
	if err != nil {
		return nil, err
	}

	for _, id := range ds.UserIds {
		if id == userId {
			continue
		}

		msg := &pushgrpcv1.WsMsg{Uid: id, Event: pushgrpcv1.WSEventType_CreateGroupAnnouncementEvent, PushOffline: true, Data: &any.Any{Value: bytes}}
		//err = s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
		//if err != nil {
		//	s.logger.Error("通知消息服务有消息需要发送失败", zap.Error(err))
		//}
		toBytes, err := utils.StructToBytes(msg)
		if err != nil {
			return nil, err
		}
		_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{Data: toBytes, Type: pushgrpcv1.Type_Ws})
		if err != nil {
			s.logger.Error("发送消息失败", zap.Error(err))
		}
	}

	return model.CreateGroupAnnouncementResponse{
		Id:      resp.ID,
		GroupId: req.GroupId,
		Title:   req.Title,
		Content: req.Content,
		OperatorInfo: model.SenderInfo{
			UserId: info.UserId,
			Avatar: info.Avatar,
			Name:   info.NickName,
		},
	}, nil
}

func (s *Service) RemoveUserFromGroup(ctx context.Context, groupID uint32, adminID string, userIDs []string) error {
	gr1, err := s.relationGroupService.GetGroupRelation(context.Background(), &relationgrpcv1.GetGroupRelationRequest{UserId: adminID, GroupId: groupID})
	if err != nil {
		s.logger.Error("获取用户群组关系失败", zap.Error(err))
		return code.RelationGroupErrGroupRelationFailed
	}

	if gr1.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return code.Forbidden
	}

	relation, err := s.relationGroupService.GetBatchGroupRelation(ctx, &relationgrpcv1.GetBatchGroupRelationRequest{GroupId: groupID, UserIds: userIDs})
	if err != nil {
		return err
	}

	for _, v := range relation.GroupRelationResponses {
		if v.Identity != relationgrpcv1.GroupIdentity_IDENTITY_USER {
			return code.Forbidden
		}
	}

	//删除用户群聊关系
	_, err = s.relationGroupService.RemoveGroupRelationByGroupIdAndUserIDs(ctx, &relationgrpcv1.RemoveGroupRelationByGroupIdAndUserIDsRequest{GroupId: groupID, UserIDs: userIDs})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) QuitGroup(ctx context.Context, groupID uint32, userID string) error {
	//查询用户是否在群聊中
	relation, err := s.relationGroupService.GetGroupRelation(context.Background(), &relationgrpcv1.GetGroupRelationRequest{UserId: userID, GroupId: groupID})
	if err != nil {
		return code.RelationGroupErrGroupRelationFailed
	}

	if relation.Identity == relationgrpcv1.GroupIdentity_IDENTITY_OWNER {
		return code.RelationGroupErrGroupOwnerCantLeaveGroupFailed
	}

	dialog, err := s.relationDialogService.GetDialogByGroupId(ctx, &relationgrpcv1.GetDialogByGroupIdRequest{GroupId: groupID})
	if err != nil {
		return code.DialogErrGetDialogByIdFailed
	}

	r1 := &relationgrpcv1.DeleteDialogUserByDialogIDAndUserIDRequest{DialogId: dialog.DialogId, UserId: userID}
	r2 := &relationgrpcv1.LeaveGroupRequest{UserId: userID, GroupId: groupID}
	gid := shortuuid.New()
	err = dtmgrpc.TccGlobalTransaction(s.dtmGrpcServer, gid, func(tcc *dtmgrpc.TccGrpc) error {
		r := &emptypb.Empty{}
		if err = tcc.CallBranch(r1, s.relationServiceAddr+relationgrpcv1.DialogService_DeleteDialogUserByDialogIDAndUserID_FullMethodName, "", s.relationServiceAddr+relationgrpcv1.DialogService_DeleteDialogUserByDialogIDAndUserIDRevert_FullMethodName, r); err != nil {
			return err
		}
		err = tcc.CallBranch(r2, s.relationServiceAddr+relationgrpcv1.GroupRelationService_LeaveGroup_FullMethodName, "", s.relationServiceAddr+relationgrpcv1.GroupRelationService_LeaveGroupRevert_FullMethodName, r)
		return err
	})
	if err != nil {
		return code.RelationGroupErrLeaveGroupFailed
	}

	return nil
}

func (s *Service) InviteGroup(ctx context.Context, inviterId string, req *model.InviteGroupRequest) error {
	group, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupID})
	if err != nil {
		return code.GroupErrGetGroupInfoByGidFailed
	}

	if group.Status != groupApi.GroupStatus_GROUP_STATUS_NORMAL {
		return code.GroupErrGroupStatusNotAvailable
	}

	fmt.Println("InviteGroup 1")

	for _, s2 := range req.Member {
		//查询是不是已经有邀请
		id, err := s.relationGroupJoinRequestService.GetGroupJoinRequestByGroupIdAndUserId(ctx, &relationgrpcv1.GetGroupJoinRequestByGroupIdAndUserIdRequest{
			GroupId: req.GroupID,
			UserId:  s2,
		})
		if err != nil {
			s.logger.Error("获取群聊信息失败", zap.Error(err))
		}
		fmt.Println("id => ", id)
		if id != nil && id.GroupId != 0 && (id.Status == relationgrpcv1.GroupRequestStatus_Pending || id.Status == relationgrpcv1.GroupRequestStatus_Invite) {
			return code.RelationErrGroupRequestAlreadyProcessed
		}
	}

	relation1, err := s.relationGroupService.GetGroupRelation(context.Background(), &relationgrpcv1.GetGroupRelationRequest{GroupId: req.GroupID, UserId: inviterId})
	if err != nil {
		s.logger.Error("获取用户群组关系失败", zap.Error(err))
		return code.RelationGroupErrGroupRelationFailed
	}

	if relation1.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return code.Unauthorized
	}

	grs, err := s.relationGroupService.GetBatchGroupRelation(ctx, &relationgrpcv1.GetBatchGroupRelationRequest{GroupId: req.GroupID, UserIds: req.Member})
	if err != nil {
		s.logger.Error("获取用户群组关系失败", zap.Error(err))
		//if !errors.Is(code.Cause(err), code.RelationGroupErrRelationNotFound) {
		//	return code.RelationGroupErrInviteFailed
		//}
		return code.RelationGroupErrInviteFailed
	}

	for _, v := range grs.GroupRelationResponses {
		fmt.Println("v => ", v)
	}

	if len(grs.GroupRelationResponses) > 0 {
		return code.RelationGroupErrInviteFailed
	}

	//查询所有管理员
	adminIds, err := s.relationGroupService.GetGroupAdminIds(context.Background(), &relationgrpcv1.GroupIDRequest{
		GroupId: req.GroupID,
	})
	if err != nil {
		s.logger.Error("获取群聊管理员失败", zap.Error(err))
		return code.RelationGroupErrInviteFailed
	}

	_, err = s.relationGroupJoinRequestService.InviteJoinGroup(ctx, &relationgrpcv1.InviteJoinGroupRequest{
		GroupId:   req.GroupID,
		InviterId: inviterId,
		Member:    req.Member,
	})
	if err != nil {
		s.logger.Error("邀请用户加入群聊失败", zap.Error(err))
		return code.RelationGroupErrInviteFailed
	}

	data := map[string]interface{}{"group_id": req.GroupID, "user_id": inviterId}
	bytes, err := utils.StructToBytes(data)
	if err != nil {
		return err
	}
	var msgs []*pushgrpcv1.WsMsg
	for _, id := range adminIds.UserIds {
		msg := &pushgrpcv1.WsMsg{Uid: id, Event: pushgrpcv1.WSEventType_JoinGroupEvent, Data: &any.Any{Value: bytes}, SendAt: time.Now(), PushOffline: true}
		//通知消息服务有消息需要发送
		msgs = append(msgs, msg)
	}

	toBytes, err := utils.StructToBytes(msgs)
	if err != nil {
		return err
	}
	_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{Type: pushgrpcv1.Type_Ws_Batch_User, Data: toBytes})
	if err != nil {
		s.logger.Error("推送消息失败", zap.Error(err))
	}

	data2 := map[string]interface{}{"group_id": req.GroupID, "inviter_id": inviterId}
	// 给被邀请的用户推送
	bytes2, err := utils.StructToBytes(data2)
	if err != nil {
		return err
	}
	for _, id := range req.Member {
		msg := &pushgrpcv1.WsMsg{Uid: id, Event: pushgrpcv1.WSEventType_InviteJoinGroupEvent, Data: &any.Any{Value: bytes2}, SendAt: time.Now()}
		structToBytes, err := utils.StructToBytes(msg)
		if err != nil {
			return err
		}
		_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{Type: pushgrpcv1.Type_Ws, Data: structToBytes})
		if err != nil {
			s.logger.Error("推送消息失败", zap.Error(err))
		}
	}

	return nil
}

func (s *Service) GetGroupAnnouncementList(ctx context.Context, userId string, groupId uint32) (interface{}, error) {
	_, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: groupId})
	if err != nil {
		return nil, err
	}

	_, err = s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: groupId, UserId: userId})
	if err != nil {
		return nil, err
	}

	res, err := s.relationGroupAnnouncementService.GetGroupAnnouncementList(ctx, &relationgrpcv1.GetGroupAnnouncementListRequest{GroupId: groupId})
	if err != nil {
		return nil, err
	}

	var respList []model.GetGroupAnnouncementListResponse

	for _, item := range res.AnnouncementList {
		info, err := s.userService.UserInfo(ctx, &userApi.UserInfoRequest{
			UserId: item.UserId,
		})
		if err != nil {
			return nil, err
		}
		//查询每条已读人数
		users, err := s.relationGroupAnnouncementService.GetReadUsers(ctx, &relationgrpcv1.GetReadUsersRequest{
			GroupId:        item.GroupId,
			AnnouncementId: item.ID,
		})
		if err != nil {
			return nil, err
		}
		var readerInfos []*model.GetGroupAnnouncementReadUsersRequest
		for _, ru := range users.AnnouncementReadUsers {
			info2, err := s.userService.UserInfo(ctx, &userApi.UserInfoRequest{
				UserId: ru.UserId,
			})
			if err != nil {
				return nil, err
			}
			readerInfos = append(readerInfos, &model.GetGroupAnnouncementReadUsersRequest{
				UserId:         info2.UserId,
				ID:             ru.ID,
				ReadAt:         int64(ru.ReadAt),
				GroupId:        ru.GroupId,
				AnnouncementId: ru.AnnouncementId,
				ReaderInfo: model.SenderInfo{
					UserId: info2.UserId,
					Avatar: info2.Avatar,
					Name:   info2.NickName,
				},
			})
		}
		respList = append(respList, model.GetGroupAnnouncementListResponse{
			Id:       item.ID,
			Title:    item.Title,
			Content:  item.Content,
			GroupId:  item.GroupId,
			CreateAt: item.CreatedAt,
			UpdateAt: item.UpdatedAt,
			OperatorInfo: model.SenderInfo{
				UserId: info.UserId,
				Avatar: info.Avatar,
				Name:   info.NickName,
			},
			ReadUserList: readerInfos,
		})
	}

	return respList, nil
}

func (s *Service) GetGroupAnnouncementDetail(ctx context.Context, userId string, id, groupId uint32) (interface{}, error) {
	_, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: groupId})
	if err != nil {
		return nil, err
	}

	_, err = s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: groupId, UserId: userId})
	if err != nil {
		return nil, err
	}

	announcement, err := s.relationGroupAnnouncementService.GetGroupAnnouncement(ctx, &relationgrpcv1.GetGroupAnnouncementRequest{ID: id})
	if err != nil {
		return nil, err
	}

	info, err := s.userService.UserInfo(ctx, &userApi.UserInfoRequest{
		UserId: announcement.AnnouncementInfo.UserId,
	})
	if err != nil {
		return nil, err
	}

	//查询每条已读人数
	users, err := s.relationGroupAnnouncementService.GetReadUsers(ctx, &relationgrpcv1.GetReadUsersRequest{
		GroupId:        groupId,
		AnnouncementId: id,
	})
	if err != nil {
		return nil, err
	}
	var readerInfos []*model.GetGroupAnnouncementReadUsersRequest
	for _, ru := range users.AnnouncementReadUsers {
		info2, err := s.userService.UserInfo(ctx, &userApi.UserInfoRequest{
			UserId: ru.UserId,
		})
		if err != nil {
			return nil, err
		}
		readerInfos = append(readerInfos, &model.GetGroupAnnouncementReadUsersRequest{
			UserId:         info2.UserId,
			ID:             ru.ID,
			ReadAt:         int64(ru.ReadAt),
			GroupId:        ru.GroupId,
			AnnouncementId: ru.AnnouncementId,
			ReaderInfo: model.SenderInfo{
				UserId: info2.UserId,
				Avatar: info2.Avatar,
				Name:   info2.NickName,
			},
		})
	}

	return model.GetGroupAnnouncementListResponse{
		Id:       announcement.AnnouncementInfo.ID,
		Title:    announcement.AnnouncementInfo.Title,
		Content:  announcement.AnnouncementInfo.Content,
		GroupId:  announcement.AnnouncementInfo.GroupId,
		CreateAt: announcement.AnnouncementInfo.CreatedAt,
		UpdateAt: announcement.AnnouncementInfo.UpdatedAt,
		OperatorInfo: model.SenderInfo{
			UserId: info.UserId,
			Avatar: info.Avatar,
			Name:   info.NickName,
		},
		ReadUserList: readerInfos,
	}, nil
}

func (s *Service) UpdateGroupAnnouncement(ctx context.Context, userId string, req *model.UpdateGroupAnnouncementRequest) (interface{}, error) {
	_, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupId})
	if err != nil {
		return nil, err
	}

	relation, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: req.GroupId, UserId: userId})
	if err != nil {
		return nil, err
	}

	if relation.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return nil, code.Forbidden
	}

	an, err := s.relationGroupAnnouncementService.GetGroupAnnouncement(ctx, &relationgrpcv1.GetGroupAnnouncementRequest{ID: req.Id})
	if err != nil {
		return nil, err
	}

	if an.AnnouncementInfo.GroupId != req.GroupId {
		return nil, code.RelationGroupErrGroupAnnouncementNotFoundFailed
	}

	res, err := s.relationGroupAnnouncementService.UpdateGroupAnnouncement(ctx, &relationgrpcv1.UpdateGroupAnnouncementRequest{ID: req.Id, Title: req.Title, Content: req.Content})
	if err != nil {
		return nil, err
	}

	//查询发送者信息
	info, err := s.userService.UserInfo(ctx, &userApi.UserInfoRequest{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}

	//查询群成员
	ds, err := s.relationGroupService.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{GroupId: req.GroupId})
	if err != nil {
		return nil, err
	}
	data := &model.WsGroupRelationOperatorMsg{
		Id:      req.Id,
		GroupId: req.GroupId,
		Title:   req.Title,
		Content: req.Content,
		OperatorInfo: model.SenderInfo{
			UserId: info.UserId,
			Avatar: info.Avatar,
			Name:   info.NickName,
		},
	}

	bytes, err := utils.StructToBytes(data)
	if err != nil {
		return nil, err
	}

	var msgs []*pushgrpcv1.WsMsg
	for _, id := range ds.UserIds {
		if id == userId {
			continue
		}

		msg := &pushgrpcv1.WsMsg{Uid: id, Event: pushgrpcv1.WSEventType_UpdateGroupAnnouncementEvent, Data: &any.Any{Value: bytes}}
		msgs = append(msgs, msg)
		//err = s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
	}
	toBytes, err := utils.StructToBytes(msgs)
	if err != nil {
		return nil, err
	}

	_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{Type: pushgrpcv1.Type_Ws_Batch_User, Data: toBytes})
	if err != nil {
		s.logger.Error("发送消息失败", zap.Error(err))
	}

	return res, nil
}

func (s *Service) DeleteGroupAnnouncement(ctx context.Context, userId string, req *model.DeleteGroupAnnouncementRequest) (interface{}, error) {
	_, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupId})
	if err != nil {
		return nil, err
	}

	relation, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: req.GroupId, UserId: userId})
	if err != nil {
		return nil, err
	}

	if relation.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return nil, code.Forbidden
	}

	an, err := s.relationGroupAnnouncementService.GetGroupAnnouncement(ctx, &relationgrpcv1.GetGroupAnnouncementRequest{ID: req.Id})
	if err != nil {
		return nil, err
	}

	if an.AnnouncementInfo.GroupId != req.GroupId {
		return nil, code.RelationGroupErrGroupAnnouncementNotFoundFailed
	}

	res, err := s.relationGroupAnnouncementService.DeleteGroupAnnouncement(ctx, &relationgrpcv1.DeleteGroupAnnouncementRequest{ID: req.Id})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// 设置群聊公告为已读
func (s *Service) ReadGroupAnnouncement(ctx context.Context, userId string, req *model.ReadGroupAnnouncementRequest) (interface{}, error) {
	_, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupId})
	if err != nil {
		return nil, err
	}

	_, err = s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: req.GroupId, UserId: userId})
	if err != nil {
		return nil, err
	}

	an, err := s.relationGroupAnnouncementService.GetGroupAnnouncement(ctx, &relationgrpcv1.GetGroupAnnouncementRequest{ID: req.Id})
	if err != nil {
		return nil, err
	}

	if an.AnnouncementInfo.GroupId != req.GroupId {
		return nil, code.RelationGroupErrGroupAnnouncementNotFoundFailed
	}

	read, err := s.relationGroupAnnouncementService.MarkAnnouncementAsRead(ctx, &relationgrpcv1.MarkAnnouncementAsReadRequest{GroupId: req.GroupId, AnnouncementId: req.Id, UserIds: []string{userId}})
	if err != nil {
		return nil, err
	}

	return read, nil
}

// 获取读取群聊公告列表
func (s *Service) GetReadGroupAnnouncementUserList(ctx context.Context, userId string, aid, groupId uint32) (interface{}, error) {
	var response []*model.GetGroupAnnouncementReadUsersRequest
	_, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: groupId})
	if err != nil {
		return nil, err
	}

	_, err = s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: groupId, UserId: userId})
	if err != nil {
		return nil, err
	}

	an, err := s.relationGroupAnnouncementService.GetGroupAnnouncement(ctx, &relationgrpcv1.GetGroupAnnouncementRequest{ID: aid})
	if err != nil {
		return nil, err
	}

	if an.AnnouncementInfo.GroupId != groupId {
		return nil, code.RelationGroupErrGroupAnnouncementNotFoundFailed
	}

	users, err := s.relationGroupAnnouncementService.GetReadUsers(ctx, &relationgrpcv1.GetReadUsersRequest{GroupId: groupId, AnnouncementId: aid})
	if err != nil {
		return nil, err
	}

	var uids []string
	if len(users.AnnouncementReadUsers) > 0 {
		for _, user := range users.AnnouncementReadUsers {
			uids = append(uids, user.UserId)
		}
	}

	info, err := s.userService.GetBatchUserInfo(ctx, &userApi.GetBatchUserInfoRequest{UserIds: uids})
	if err != nil {
		return nil, err
	}

	for _, user := range users.AnnouncementReadUsers {
		for _, value := range info.Users {
			if user.UserId == value.UserId {
				response = append(response, &model.GetGroupAnnouncementReadUsersRequest{
					ID:             user.ID,
					AnnouncementId: user.AnnouncementId,
					GroupId:        user.GroupId,
					UserId:         user.UserId,
					ReadAt:         int64(user.ReadAt),
					ReaderInfo: model.SenderInfo{
						UserId: value.UserId,
						Avatar: value.Avatar,
						Name:   value.NickName,
					},
				})
				break
			}
		}
	}
	return response, nil
}

func SwitchGroupRequestStatus(thisId, senderId, receiverId string, status relationgrpcv1.GroupRequestStatus) model.GroupRequestStatus {
	result := model.GroupRequestStatus(status)
	if status == relationgrpcv1.GroupRequestStatus_Invite {
		if thisId == receiverId {
			result = model.InvitationReceived
		} else {
			result = model.InviteSender
		}
	}
	return result
}

func (s *Service) SetGroupOpenBurnAfterReadingTimeOut(ctx context.Context, userID string, req *model.SetGroupOpenBurnAfterReadingTimeOutRequest) error {
	_, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupId})
	if err != nil {
		s.logger.Error("获取群聊信息失败", zap.Error(err))
		return err
	}

	_, err = s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		UserId:  userID,
		GroupId: req.GroupId,
	})
	if err != nil {
		s.logger.Error("获取群聊关系失败", zap.Error(err))
		return err
	}

	_, err = s.relationGroupService.SetGroupOpenBurnAfterReadingTimeOut(ctx, &relationgrpcv1.SetGroupOpenBurnAfterReadingTimeOutRequest{
		UserId:                      userID,
		GroupId:                     req.GroupId,
		OpenBurnAfterReadingTimeOut: req.OpenBurnAfterReadingTimeOut,
	})
	if err != nil {
		s.logger.Error("设置群聊消息阅后即焚时间失败", zap.Error(err))
		return err
	}
	return nil
}

func (s *Service) SetGroupUserRemark(ctx context.Context, userID string, req *model.SetGroupUserRemarkRequest) error {
	_, err := s.groupService.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupId})
	if err != nil {
		s.logger.Error("获取群聊信息失败", zap.Error(err))
		return err
	}

	_, err = s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		UserId:  userID,
		GroupId: req.GroupId,
	})
	if err != nil {
		s.logger.Error("获取群聊关系失败", zap.Error(err))
		return err
	}

	_, err = s.relationGroupService.SetGroupUserRemark(ctx, &relationgrpcv1.SetGroupUserRemarkRequest{
		UserId:  userID,
		GroupId: req.GroupId,
		Remark:  req.Remark,
	})
	if err != nil {
		s.logger.Error("设置群聊用户备注失败", zap.Error(err))
		return err
	}

	return nil
}

func (s *Service) DeleteGroupFriendRecord(ctx context.Context, uid string, id uint32) error {
	gr, err := s.relationGroupJoinRequestService.GetGroupJoinRequestByID(ctx, &relationgrpcv1.GetGroupJoinRequestByIDRequest{ID: id})
	if err != nil {
		s.logger.Error("获取群聊申请记录失败", zap.Uint32("id", id), zap.String("uid", uid), zap.Error(err))
		return err
	}

	if gr.OwnerID != uid {
		return code.Forbidden
	}

	_, err = s.relationGroupJoinRequestService.DeleteGroupRecord(ctx, &relationgrpcv1.DeleteGroupRecordRequest{
		ID:     id,
		UserId: uid,
	})
	if err != nil {
		s.logger.Error("删除群聊申请记录失败", zap.Uint32("id", id), zap.String("uid", uid), zap.Error(err))
		return err
	}
	return nil
}
