package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cossim/coss-server/interface/relation/api/model"
	groupApi "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	userApi "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/msg_queue"
	"github.com/cossim/coss-server/pkg/utils/time"
	"github.com/cossim/coss-server/pkg/utils/usersorter"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/lithammer/shortuuid/v3"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	ostime "time"
)

func (s *Service) GetGroupMember(ctx context.Context, gid uint32, userID string) (interface{}, error) {
	group, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: gid})
	if err != nil {
		s.logger.Error("获取群聊成员失败", zap.Error(err))
		return nil, err
	}

	if group.Status != groupApi.GroupStatus_GROUP_STATUS_NORMAL {
		return nil, code.GroupErrGroupStatusNotAvailable
	}

	_, err = s.groupRelationClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: gid, UserId: userID})
	if err != nil {
		return nil, err
	}

	groupRelation, err := s.groupRelationClient.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{GroupId: gid})
	if err != nil {
		s.logger.Error("获取群聊成员失败", zap.Error(err))
		return nil, err
	}

	relation, err := s.groupRelationClient.GetBatchGroupRelation(ctx, &relationgrpcv1.GetBatchGroupRelationRequest{GroupId: gid, UserIds: groupRelation.UserIds})
	if err != nil {
		return nil, err
	}

	resp, err := s.userClient.GetBatchUserInfo(ctx, &userApi.GetBatchUserInfoRequest{UserIds: groupRelation.UserIds})
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
			}
		}
	}
	return data, nil
}

func (s *Service) JoinGroup(ctx context.Context, uid string, req *model.JoinGroupRequest) (interface{}, error) {
	group, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupID})
	if err != nil {
		s.logger.Error("获取群聊信息失败", zap.Error(err))
		return nil, err
	}

	//查询是不是已经有邀请
	id, err := s.groupJoinRequestClient.GetGroupJoinRequestByGroupIdAndUserId(ctx, &relationgrpcv1.GetGroupJoinRequestByGroupIdAndUserIdRequest{
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
	relation, err := s.groupRelationClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		GroupId: req.GroupID,
		UserId:  uid,
	})
	if relation != nil {
		return nil, code.RelationGroupErrAlreadyInGroup
	}

	groupRelation, err := s.groupRelationClient.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{GroupId: req.GroupID})
	if err != nil {
		s.logger.Error("获取群聊成员失败", zap.Error(err))
		return nil, err
	}

	if len(groupRelation.UserIds) >= int(group.MaxMembersLimit) {
		return nil, code.RelationGroupErrGroupFull
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
		msg := constants.WsMsg{Uid: id, Event: constants.JoinGroupEvent, Data: constants.JoinGroupEventData{
			GroupId: req.GroupID,
			UserId:  uid,
		}, SendAt: time.Now()}
		//通知消息服务有消息需要发送
		err = s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
		if err != nil {
			s.logger.Error("加入群聊请求申请通知推送失败", zap.Error(err))
		}
	}
	return nil, nil
}

func (s *Service) GetUserGroupList(ctx context.Context, userID string) (interface{}, error) {
	if s.cache {
		//查询是否有缓存
		values, err := s.redisClient.GetAllListValues(fmt.Sprintf("group:%s", userID))
		if err != nil {
			s.logger.Error("err:" + err.Error())
			return nil, code.RelationErrGetFriendListFailed
		}
		if len(values) > 0 {
			// 类型转换
			var responseList []usersorter.Group
			for _, v := range values {
				var friend usersorter.CustomGroupData
				err := json.Unmarshal([]byte(v), &friend)
				if err != nil {
					fmt.Println("Error decoding cached data:", err)
					continue
				}
				responseList = append(responseList, friend)
			}
			groupedUsers := usersorter.SortAndGroupUsers(responseList, "Name")

			return groupedUsers, nil
		}
	}

	// 获取用户群聊列表
	ids, err := s.groupRelationClient.GetUserGroupIDs(context.Background(), &relationgrpcv1.GetUserGroupIDsRequest{UserId: userID})
	if err != nil {
		s.logger.Error("获取用户群聊列表失败", zap.Error(err))
		return nil, err
	}
	if len(ids.GroupId) == 0 {
		return []string{}, nil
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

	if s.cache {
		var result []interface{}

		// Assuming data2 is a slice or array of a specific type
		for _, item := range data {
			result = append(result, item)
		}

		//存储到缓存
		err = s.redisClient.AddToList(fmt.Sprintf("group:%s", userID), result)
		if err != nil {
			s.logger.Error("err:" + err.Error())
			return nil, code.RelationErrGetFriendListFailed
		}

		//设置key过期时间
		err = s.redisClient.SetKeyExpiration(fmt.Sprintf("group:%s", userID), 3*24*ostime.Hour)
		if err != nil {
			s.logger.Error("err:" + err.Error())
			return nil, code.RelationErrGetFriendListFailed
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

func (s *Service) SetGroupBurnAfterReading(ctx context.Context, userId string, req *model.OpenGroupBurnAfterReadingRequest) (interface{}, error) {
	_, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupId})
	if err != nil {
		s.logger.Error("获取群聊信息失败", zap.Error(err))
		return nil, err
	}

	_, err = s.groupRelationClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		UserId:  userId,
		GroupId: req.GroupId,
	})
	if err != nil {
		s.logger.Error("获取群聊关系失败", zap.Error(err))
		return nil, err
	}

	_, err = s.groupRelationClient.SetGroupOpenBurnAfterReading(ctx, &relationgrpcv1.SetGroupOpenBurnAfterReadingRequest{
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
		//查询发送者接受者信息
		reinfo, err := s.userClient.UserInfo(ctx, &userApi.UserInfoRequest{UserId: v.UserId})
		if err != nil {
			return nil, err
		}
		sendinfo := &userApi.UserInfoResponse{}
		if v.InviterId != "" {
			sendinfo, err = s.userClient.UserInfo(ctx, &userApi.UserInfoRequest{UserId: v.InviterId})
			if err != nil {
				return nil, err
			}
		}

		data[i] = &model.GroupRequestListResponse{
			ID:      v.ID,
			GroupId: v.GroupId,
			Remark:  v.Remark,
			Status:  SwitchGroupRequestStatus(userID, v.InviterId, reinfo.UserId, v.Status),
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
		groupInfo, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: groupID})
		if err != nil {
			s.logger.Error("获取群聊信息失败", zap.Error(err))
			return nil, err
		}

		groupInfoMap[groupID] = groupInfo
	}

	_, err = s.userClient.GetBatchUserInfo(ctx, &userApi.GetBatchUserInfoRequest{UserIds: uids})
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

	info, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: groupID})
	if err != nil {
		s.logger.Error("get group info failed", zap.Error(err))
		return code.RelationGroupErrGetGroupInfoFailed
	}

	if status == relationgrpcv1.GroupRequestStatus_Accepted {
		ds, err := s.groupRelationClient.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{GroupId: groupID})
		if err != nil {
			return err
		}
		if info.MaxMembersLimit <= int32(len(ds.UserIds)) {
			return code.RelationGroupErrGroupFull
		}
	}

	req, err := s.groupJoinRequestClient.GetGroupJoinRequestByID(ctx, &relationgrpcv1.GetGroupJoinRequestByIDRequest{ID: requestID})
	if err != nil {
		return err
	}

	_, err = s.groupJoinRequestClient.ManageGroupJoinRequestByID(ctx, &relationgrpcv1.ManageGroupJoinRequestByIDRequest{
		ID:     requestID,
		Status: status,
	})
	if err != nil {
		return err
	}

	if status == relationgrpcv1.GroupRequestStatus_Accepted && s.cache {
		dialogInfo, err := s.dialogClient.GetDialogByGroupId(ctx, &relationgrpcv1.GetDialogByGroupIdRequest{GroupId: groupID})
		if err != nil {
			s.logger.Error("获取群聊对话信息失败", zap.Error(err))
			return code.RelationGroupErrManageJoinFailed
		}

		re := model.UserDialogListResponse{
			DialogId:       dialogInfo.DialogId,
			GroupId:        groupID,
			DialogType:     model.ConversationType(dialogInfo.Type),
			DialogName:     info.Name,
			DialogAvatar:   info.Avatar,
			DialogCreateAt: dialogInfo.CreateAt,
		}

		//更新缓存
		err = s.insertRedisUserDialogList(req.UserId, re)
		if err != nil {
			s.logger.Error("获取好友关系失败", zap.Error(err))
			return code.RelationGroupErrManageJoinFailed
		}
		err = s.updateRedisGroupList(req.UserId, usersorter.CustomGroupData{
			GroupID:  groupID,
			Name:     info.Name,
			Avatar:   info.Avatar,
			Status:   uint(info.Status),
			DialogId: dialogInfo.DialogId,
		})
	}

	msg := constants.WsMsg{
		Uid:    req.UserId,
		Event:  constants.JoinGroupEvent,
		Data:   constants.JoinGroupEventData{GroupId: groupID, Status: uint32(status)},
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

	info, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: groupID})
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

	if status == relationgrpcv1.GroupRequestStatus_Accepted {
		ds, err := s.groupRelationClient.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{GroupId: groupID})
		if err != nil {
			return err
		}
		if info.MaxMembersLimit <= int32(len(ds.UserIds)) {
			return code.RelationGroupErrGroupFull
		}
	}

	req, err := s.groupJoinRequestClient.GetGroupJoinRequestByID(ctx, &relationgrpcv1.GetGroupJoinRequestByIDRequest{ID: requestID})
	if err != nil {
		return err
	}

	_, err = s.groupJoinRequestClient.ManageGroupJoinRequestByID(ctx, &relationgrpcv1.ManageGroupJoinRequestByIDRequest{
		ID:     requestID,
		Status: status,
	})
	if err != nil {
		return err
	}

	if status == relationgrpcv1.GroupRequestStatus_Accepted && s.cache {
		dialogInfo, err := s.dialogClient.GetDialogByGroupId(ctx, &relationgrpcv1.GetDialogByGroupIdRequest{GroupId: groupID})
		if err != nil {
			s.logger.Error("获取群聊对话信息失败", zap.Error(err))
			return code.RelationGroupErrManageJoinFailed
		}

		re := model.UserDialogListResponse{
			DialogId:       dialogInfo.DialogId,
			GroupId:        groupID,
			DialogType:     model.ConversationType(dialogInfo.Type),
			DialogName:     info.Name,
			DialogAvatar:   info.Avatar,
			DialogCreateAt: dialogInfo.CreateAt,
		}

		//更新缓存
		err = s.insertRedisUserDialogList(req.UserId, re)
		if err != nil {
			s.logger.Error("获取好友关系失败", zap.Error(err))
			return code.RelationGroupErrManageJoinFailed
		}

		err = s.updateRedisGroupList(req.UserId, usersorter.CustomGroupData{
			GroupID:  groupID,
			Name:     info.Name,
			Avatar:   info.Avatar,
			Status:   uint(info.Status),
			DialogId: dialogInfo.DialogId,
		})
	}

	msg := constants.WsMsg{
		Uid:    userID,
		Event:  constants.JoinGroupEvent,
		Data:   constants.JoinGroupEventData{GroupId: groupID, Status: uint32(status)},
		SendAt: time.Now(),
	}
	if err = s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg); err != nil {
		s.logger.Error("通知消息服务有消息需要发送失败", zap.Error(err))
	}

	return nil
}

func (s *Service) CreateGroupAnnouncement(ctx context.Context, userId string, req *model.CreateGroupAnnouncementRequest) (interface{}, error) {
	_, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupId})
	if err != nil {
		return nil, err
	}

	relation, err := s.groupRelationClient.GetGroupRelation(context.Background(), &relationgrpcv1.GetGroupRelationRequest{UserId: userId, GroupId: req.GroupId})
	if err != nil {
		return nil, err
	}

	if relation.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return nil, code.Forbidden
	}

	resp, err := s.groupAnnouncementClient.CreateGroupAnnouncement(ctx, &relationgrpcv1.CreateGroupAnnouncementRequest{
		GroupId: req.GroupId,
		UserId:  userId,
		Content: req.Content,
		Title:   req.Title,
	})

	if err != nil {
		return nil, err
	}

	//查询发送者信息
	info, err := s.userClient.UserInfo(ctx, &userApi.UserInfoRequest{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}

	//查询群成员
	ds, err := s.groupRelationClient.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{GroupId: req.GroupId})
	if err != nil {
		return nil, err
	}
	for _, id := range ds.UserIds {
		if id == userId {
			continue
		}
		msg := constants.WsMsg{Uid: id, Event: constants.CreateGroupAnnouncementEvent, Data: model.WsGroupRelationOperatorMsg{
			Id:      resp.ID,
			GroupId: req.GroupId,
			Title:   req.Title,
			Content: req.Content,
			OperatorInfo: model.SenderInfo{
				UserId: info.UserId,
				Avatar: info.Avatar,
				Name:   info.NickName,
			},
		}}
		err = s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
		if err != nil {
			s.logger.Error("通知消息服务有消息需要发送失败", zap.Error(err))
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
	gr1, err := s.groupRelationClient.GetGroupRelation(context.Background(), &relationgrpcv1.GetGroupRelationRequest{UserId: adminID, GroupId: groupID})
	if err != nil {
		s.logger.Error("获取用户群组关系失败", zap.Error(err))
		return code.RelationGroupErrGroupRelationFailed
	}

	if gr1.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return code.Forbidden
	}

	relation, err := s.groupRelationClient.GetBatchGroupRelation(ctx, &relationgrpcv1.GetBatchGroupRelationRequest{GroupId: groupID, UserIds: userIDs})
	if err != nil {
		return err
	}

	for _, v := range relation.GroupRelationResponses {
		if v.Identity != relationgrpcv1.GroupIdentity_IDENTITY_USER {
			return code.Forbidden
		}
	}

	for _, d := range userIDs {
		err = s.removeRedisGroupList(d, groupID)
		if err != nil {
			s.logger.Error("退出群聊失败", zap.Error(err))
			return code.RelationGroupErrGroupRelationFailed
		}
	}

	//删除用户群聊关系
	_, err = s.groupRelationClient.RemoveGroupRelationByGroupIdAndUserIDs(ctx, &relationgrpcv1.RemoveGroupRelationByGroupIdAndUserIDsRequest{GroupId: groupID, UserIDs: userIDs})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) QuitGroup(ctx context.Context, groupID uint32, userID string) error {
	//查询用户是否在群聊中
	relation, err := s.groupRelationClient.GetGroupRelation(context.Background(), &relationgrpcv1.GetGroupRelationRequest{UserId: userID, GroupId: groupID})
	if err != nil {
		return code.RelationGroupErrGroupRelationFailed
	}

	if relation.Identity == relationgrpcv1.GroupIdentity_IDENTITY_OWNER {
		return code.RelationGroupErrGroupOwnerCantLeaveGroupFailed
	}

	dialog, err := s.dialogClient.GetDialogByGroupId(ctx, &relationgrpcv1.GetDialogByGroupIdRequest{GroupId: groupID})
	if err != nil {
		return code.DialogErrGetDialogByIdFailed
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
		return code.RelationGroupErrLeaveGroupFailed
	}

	err = s.removeRedisUserDialogList(userID, dialog.DialogId)
	if err != nil {
		s.logger.Error("退出群聊失败", zap.Error(err))
		return code.RelationGroupErrLeaveGroupFailed
	}

	err = s.removeRedisGroupList(userID, groupID)
	if err != nil {
		s.logger.Error("退出群聊失败", zap.Error(err))
		return code.RelationGroupErrLeaveGroupFailed
	}

	return nil
}

func (s *Service) InviteGroup(ctx context.Context, inviterId string, req *model.InviteGroupRequest) error {
	group, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupID})
	if err != nil {
		return code.GroupErrGetGroupInfoByGidFailed
	}

	if group.Status != groupApi.GroupStatus_GROUP_STATUS_NORMAL {
		return code.GroupErrGroupStatusNotAvailable
	}

	for _, s2 := range req.Member {
		//查询是不是已经有邀请
		id, err := s.groupJoinRequestClient.GetGroupJoinRequestByGroupIdAndUserId(ctx, &relationgrpcv1.GetGroupJoinRequestByGroupIdAndUserIdRequest{
			GroupId: req.GroupID,
			UserId:  s2,
		})
		if err != nil {
			s.logger.Error("获取群聊信息失败", zap.Error(err))
		}
		if id != nil && id.GroupId != 0 {
			return code.RelationErrGroupRequestAlreadyProcessed
		}
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
		msg := constants.WsMsg{Uid: id, Event: constants.JoinGroupEvent, Data: map[string]interface{}{"group_id": req.GroupID, "user_id": inviterId}, SendAt: time.Now()}
		//通知消息服务有消息需要发送
		err = s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
		if err != nil {
			s.logger.Error("加入群聊请求申请通知推送失败", zap.Error(err))
		}
	}

	// 给被邀请的用户推送
	for _, id := range req.Member {
		msg := constants.WsMsg{Uid: id, Event: constants.InviteJoinGroupEvent, Data: map[string]interface{}{"group_id": req.GroupID, "inviter_id": inviterId}, SendAt: time.Now()}
		err = s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
	}

	return nil
}

func (s *Service) GetGroupAnnouncementList(ctx context.Context, userId string, groupId uint32) (interface{}, error) {
	_, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: groupId})
	if err != nil {
		return nil, err
	}

	_, err = s.groupRelationClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: groupId, UserId: userId})
	if err != nil {
		return nil, err
	}

	res, err := s.groupAnnouncementClient.GetGroupAnnouncementList(ctx, &relationgrpcv1.GetGroupAnnouncementListRequest{GroupId: groupId})
	if err != nil {
		return nil, err
	}

	var respList []model.GetGroupAnnouncementListResponse

	for _, item := range res.AnnouncementList {
		info, err := s.userClient.UserInfo(ctx, &userApi.UserInfoRequest{
			UserId: item.UserId,
		})
		if err != nil {
			return nil, err
		}
		//查询每条已读人数
		users, err := s.groupAnnouncementReadClient.GetReadUsers(ctx, &relationgrpcv1.GetReadUsersRequest{
			GroupId:        item.GroupId,
			AnnouncementId: item.ID,
		})
		if err != nil {
			return nil, err
		}
		var readerInfos []*model.GetGroupAnnouncementReadUsersRequest
		for _, ru := range users.AnnouncementReadUsers {
			info2, err := s.userClient.UserInfo(ctx, &userApi.UserInfoRequest{
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
	_, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: groupId})
	if err != nil {
		return nil, err
	}

	_, err = s.groupRelationClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: groupId, UserId: userId})
	if err != nil {
		return nil, err
	}

	announcement, err := s.groupAnnouncementClient.GetGroupAnnouncement(ctx, &relationgrpcv1.GetGroupAnnouncementRequest{ID: id})
	if err != nil {
		return nil, err
	}

	info, err := s.userClient.UserInfo(ctx, &userApi.UserInfoRequest{
		UserId: announcement.AnnouncementInfo.UserId,
	})
	if err != nil {
		return nil, err
	}

	//查询每条已读人数
	users, err := s.groupAnnouncementReadClient.GetReadUsers(ctx, &relationgrpcv1.GetReadUsersRequest{
		GroupId:        groupId,
		AnnouncementId: id,
	})
	if err != nil {
		return nil, err
	}
	var readerInfos []*model.GetGroupAnnouncementReadUsersRequest
	for _, ru := range users.AnnouncementReadUsers {
		info2, err := s.userClient.UserInfo(ctx, &userApi.UserInfoRequest{
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
	_, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupId})
	if err != nil {
		return nil, err
	}

	relation, err := s.groupRelationClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: req.GroupId, UserId: userId})
	if err != nil {
		return nil, err
	}

	if relation.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return nil, code.Forbidden
	}

	an, err := s.groupAnnouncementClient.GetGroupAnnouncement(ctx, &relationgrpcv1.GetGroupAnnouncementRequest{ID: req.Id})
	if err != nil {
		return nil, err
	}

	if an.AnnouncementInfo.GroupId != req.GroupId {
		return nil, code.RelationGroupErrGroupAnnouncementNotFoundFailed
	}

	res, err := s.groupAnnouncementClient.UpdateGroupAnnouncement(ctx, &relationgrpcv1.UpdateGroupAnnouncementRequest{ID: req.Id, Title: req.Title, Content: req.Content})
	if err != nil {
		return nil, err
	}

	//查询发送者信息
	info, err := s.userClient.UserInfo(ctx, &userApi.UserInfoRequest{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}

	//查询群成员
	ds, err := s.groupRelationClient.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{GroupId: req.GroupId})
	if err != nil {
		return nil, err
	}
	for _, id := range ds.UserIds {
		if id == userId {
			continue
		}
		msg := constants.WsMsg{Uid: id, Event: constants.UpdateGroupAnnouncementEvent, Data: model.WsGroupRelationOperatorMsg{
			Id:      req.Id,
			GroupId: req.GroupId,
			Title:   req.Title,
			Content: req.Content,
			OperatorInfo: model.SenderInfo{
				UserId: info.UserId,
				Avatar: info.Avatar,
				Name:   info.NickName,
			},
		}}
		err = s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
	}

	return res, nil
}

func (s *Service) DeleteGroupAnnouncement(ctx context.Context, userId string, req *model.DeleteGroupAnnouncementRequest) (interface{}, error) {
	_, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupId})
	if err != nil {
		return nil, err
	}

	relation, err := s.groupRelationClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: req.GroupId, UserId: userId})
	if err != nil {
		return nil, err
	}

	if relation.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return nil, code.Forbidden
	}

	an, err := s.groupAnnouncementClient.GetGroupAnnouncement(ctx, &relationgrpcv1.GetGroupAnnouncementRequest{ID: req.Id})
	if err != nil {
		return nil, err
	}

	if an.AnnouncementInfo.GroupId != req.GroupId {
		return nil, code.RelationGroupErrGroupAnnouncementNotFoundFailed
	}

	res, err := s.groupAnnouncementClient.DeleteGroupAnnouncement(ctx, &relationgrpcv1.DeleteGroupAnnouncementRequest{ID: req.Id})
	if err != nil {
		return nil, err
	}
	return res, nil
}

// 设置群聊公告为已读
func (s *Service) ReadGroupAnnouncement(ctx context.Context, userId string, req *model.ReadGroupAnnouncementRequest) (interface{}, error) {
	_, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupId})
	if err != nil {
		return nil, err
	}

	_, err = s.groupRelationClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: req.GroupId, UserId: userId})
	if err != nil {
		return nil, err
	}

	an, err := s.groupAnnouncementClient.GetGroupAnnouncement(ctx, &relationgrpcv1.GetGroupAnnouncementRequest{ID: req.Id})
	if err != nil {
		return nil, err
	}

	if an.AnnouncementInfo.GroupId != req.GroupId {
		return nil, code.RelationGroupErrGroupAnnouncementNotFoundFailed
	}

	read, err := s.groupAnnouncementReadClient.MarkAnnouncementAsRead(ctx, &relationgrpcv1.MarkAnnouncementAsReadRequest{GroupId: req.GroupId, AnnouncementId: req.Id, UserIds: []string{userId}})
	if err != nil {
		return nil, err
	}

	return read, nil
}

// 获取读取群聊公告列表
func (s *Service) GetReadGroupAnnouncementUserList(ctx context.Context, userId string, aid, groupId uint32) (interface{}, error) {
	var response []*model.GetGroupAnnouncementReadUsersRequest
	_, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: groupId})
	if err != nil {
		return nil, err
	}

	_, err = s.groupRelationClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{GroupId: groupId, UserId: userId})
	if err != nil {
		return nil, err
	}

	an, err := s.groupAnnouncementClient.GetGroupAnnouncement(ctx, &relationgrpcv1.GetGroupAnnouncementRequest{ID: aid})
	if err != nil {
		return nil, err
	}

	if an.AnnouncementInfo.GroupId != groupId {
		return nil, code.RelationGroupErrGroupAnnouncementNotFoundFailed
	}

	users, err := s.groupAnnouncementReadClient.GetReadUsers(ctx, &relationgrpcv1.GetReadUsersRequest{GroupId: groupId, AnnouncementId: aid})
	if err != nil {
		return nil, err
	}

	var uids []string
	if len(users.AnnouncementReadUsers) > 0 {
		for _, user := range users.AnnouncementReadUsers {
			uids = append(uids, user.UserId)
		}
	}

	info, err := s.userClient.GetBatchUserInfo(ctx, &userApi.GetBatchUserInfoRequest{UserIds: uids})
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
	_, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupId})
	if err != nil {
		s.logger.Error("获取群聊信息失败", zap.Error(err))
		return err
	}

	_, err = s.groupRelationClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		UserId:  userID,
		GroupId: req.GroupId,
	})
	if err != nil {
		s.logger.Error("获取群聊关系失败", zap.Error(err))
		return err
	}

	_, err = s.groupRelationClient.SetGroupOpenBurnAfterReadingTimeOut(ctx, &relationgrpcv1.SetGroupOpenBurnAfterReadingTimeOutRequest{
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
	_, err := s.groupClient.GetGroupInfoByGid(ctx, &groupApi.GetGroupInfoRequest{Gid: req.GroupId})
	if err != nil {
		s.logger.Error("获取群聊信息失败", zap.Error(err))
		return err
	}

	_, err = s.groupRelationClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		UserId:  userID,
		GroupId: req.GroupId,
	})
	if err != nil {
		s.logger.Error("获取群聊关系失败", zap.Error(err))
		return err
	}

	_, err = s.groupRelationClient.SetGroupUserRemark(ctx, &relationgrpcv1.SetGroupUserRemarkRequest{
		UserId:  userID,
		GroupId: req.GroupId,
		Remark:  req.Remark,
	})
	if err != nil {
		s.logger.Error("设置群聊用户备注失败", zap.Error(err))
		return err
	}

	if s.cache {
		err := s.redisClient.DelKey(fmt.Sprintf("dialog:%s", userID))
		if err != nil {
			s.logger.Error("设置群聊用户备注失败", zap.Error(err))
			return err
		}
	}
	return nil
}

func (s *Service) updateRedisGroupList(userID string, msg usersorter.CustomGroupData) error {
	key := fmt.Sprintf("group:%s", userID)
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

func (s *Service) removeRedisGroupList(userID string, groupID uint32) error {
	key := fmt.Sprintf("group:%s", userID)
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
				var group usersorter.CustomGroupData
				err := json.Unmarshal([]byte(v), &group)
				if err != nil {
					fmt.Println("Error decoding cached data:", err)
					return err
				}
				if group.GroupID == groupID {
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
			var group usersorter.CustomGroupData
			err := json.Unmarshal([]byte(v), &group)
			if err != nil {
				fmt.Println("Error decoding cached data:", err)
				return err
			}
			if group.GroupID == groupID {
				_, err := s.redisClient.PopListElement(key, int64(i))
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}
