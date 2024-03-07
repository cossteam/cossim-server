package service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cossim/coss-server/interface/group/api/model"
	"github.com/cossim/coss-server/pkg/code"
	"github.com/cossim/coss-server/pkg/constants"
	"github.com/cossim/coss-server/pkg/msg_queue"
	"github.com/cossim/coss-server/pkg/utils/time"
	"github.com/cossim/coss-server/pkg/utils/usersorter"
	groupgrpcv1 "github.com/cossim/coss-server/service/group/api/v1"
	relationgrpcv1 "github.com/cossim/coss-server/service/relation/api/v1"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/dtm-labs/client/workflow"
	"github.com/lithammer/shortuuid/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"sync"
	ostime "time"
)

func (s *Service) CreateGroup(ctx context.Context, req *groupgrpcv1.Group) (*model.CreateGroupResponse, error) {
	var err error
	friends, err := s.relationUserClient.GetUserRelationByUserIds(ctx, &relationgrpcv1.GetUserRelationByUserIdsRequest{UserId: req.CreatorId, FriendIds: req.Member})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		return nil, code.RelationErrCreateGroupFailed
	}

	if len(req.Member) != len(friends.Users) {
		return nil, code.RelationUserErrFriendRelationNotFound
	}
	for _, friend := range friends.Users {
		if friend.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
			return nil, code.StatusNotAvailable
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

	r22 := &relationgrpcv1.CreateGroupAndInviteUsersRequest{
		UserID: req.CreatorId,
		Member: req.Member,
	}

	resp1 := &groupgrpcv1.Group{}
	var groupID uint32
	var DialogID uint32
	// 创建 DTM 分布式事务工作流
	workflow.InitGrpc(s.dtmGrpcServer, s.relationGrpcServer, grpc.NewServer())
	gid := shortuuid.New()
	wfName := "create_group_workflow_" + gid
	if err = workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		// 创建群聊
		resp1, err = s.groupClient.CreateGroup(wf.Context, r1)
		if err != nil {
			return err
		}
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.groupClient.CreateGroupRevert(wf.Context, &groupgrpcv1.CreateGroupRequest{Group: &groupgrpcv1.Group{
				Id: resp1.Id,
			}})
			return err
		})
		groupID = resp1.Id
		r22.GroupId = groupID
		r22.Member = req.Member
		r22.UserID = req.CreatorId
		resp2, err := s.relationGroupClient.CreateGroupAndInviteUsers(wf.Context, r22)
		if err != nil {
			return err
		}
		DialogID = resp2.DialogId
		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.relationGroupClient.CreateGroupAndInviteUsersRevert(wf.Context, r22)
			return err
		})

		return err
	}); err != nil {
		s.logger.Error("WorkFlow CreateGroup", zap.Error(err))
		return nil, code.RelationErrCreateGroupFailed
	}
	if err = workflow.Execute(wfName, gid, nil); err != nil {
		s.logger.Error("WorkFlow CreateGroup", zap.Error(err))
		return nil, code.RelationErrCreateGroupFailed
	}

	err = s.insertRedisGroupList(req.CreatorId, usersorter.CustomGroupData{
		GroupID:  groupID,
		Name:     resp1.Name,
		Avatar:   resp1.Avatar,
		Status:   uint(resp1.Status),
		DialogId: DialogID,
	})
	if err != nil {
		s.logger.Error("CreateGroup", zap.Error(err))
		return nil, code.RelationErrCreateGroupFailed
	}

	dialog, err := s.relationDialogClient.GetDialogById(ctx, &relationgrpcv1.GetDialogByIdRequest{DialogId: DialogID})
	if err != nil {
		s.logger.Error("CreateGroup", zap.Error(err))
		return nil, code.RelationErrCreateGroupFailed
	}

	err = s.insertRedisUserDialogList(req.CreatorId, model.UserDialogListResponse{
		DialogId:          DialogID,
		GroupId:           groupID,
		DialogType:        model.ConversationType(dialog.Type),
		DialogName:        req.Name,
		DialogAvatar:      req.Avatar,
		DialogUnreadCount: 0,
		LastMessage:       model.Message{},
		DialogCreateAt:    dialog.CreateAt,
		TopAt:             0,
	})
	if err != nil {
		s.logger.Error("CreateGroup", zap.Error(err))
		return nil, code.RelationErrCreateGroupFailed
	}

	// 给被邀请的用户推送
	for _, id := range req.Member {
		msg := constants.WsMsg{Uid: id, Event: constants.InviteJoinGroupEvent, Data: map[string]interface{}{"group_id": groupID, "inviter_id": req.CreatorId}, SendAt: time.Now()}
		//通知消息服务有消息需要发送
		err = s.rabbitMQClient.PublishServiceMessage(msg_queue.RelationService, msg_queue.MsgService, msg_queue.Service_Exchange, msg_queue.SendMessage, msg)
	}

	return &model.CreateGroupResponse{
		Id:              resp1.Id,
		Avatar:          resp1.Avatar,
		Name:            resp1.Name,
		Type:            uint32(resp1.Type),
		Status:          int32(resp1.Status),
		MaxMembersLimit: resp1.MaxMembersLimit,
		CreatorId:       resp1.CreatorId,
		DialogId:        DialogID,
	}, nil
}

func (s *Service) DeleteGroup(ctx context.Context, groupID uint32, userID string) (uint32, error) {
	_, err := s.groupClient.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{
		Gid: groupID,
	})
	if err != nil {
		return 0, code.GroupErrGroupNotFound
	}
	sf, err := s.relationGroupClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		UserId:  userID,
		GroupId: groupID,
	})
	if err != nil {
		return 0, err
	}
	if sf.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return 0, code.Forbidden
	}
	dialog, err := s.relationDialogClient.GetDialogByGroupId(ctx, &relationgrpcv1.GetDialogByGroupIdRequest{GroupId: groupID})
	if err != nil {
		return 0, err
	}

	//查询所有群员
	relation, err := s.relationGroupClient.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{
		GroupId: groupID,
	})
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
		s.logger.Error("WorkFlow DeleteGroup", zap.Error(err))
		return 0, code.GroupErrDeleteGroupFailed
	}

	for _, res := range relation.UserIds {
		err := s.removeRedisGroupList(res, groupID)
		if err != nil {
			return 0, err
		}

		err = s.removeRedisUserDialogList(res, dialog.DialogId)
		if err != nil {
			return 0, err
		}
	}

	return groupID, err
}

func (s *Service) UpdateGroup(ctx context.Context, req *model.UpdateGroupRequest, userID string) (interface{}, error) {
	group, err := s.groupClient.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{
		Gid: req.GroupId,
	})
	if err != nil {
		s.logger.Error("更新群聊信息失败", zap.Error(err))
		return nil, err
	}

	sf, err := s.relationGroupClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		UserId:  userID,
		GroupId: req.GroupId,
	})
	if err != nil {
		s.logger.Error("获取用户群聊关系失败", zap.Error(err))
		return nil, err
	}

	if sf.Identity == relationgrpcv1.GroupIdentity_IDENTITY_USER {
		return nil, code.Unauthorized
	}

	group.Type = groupgrpcv1.GroupType(req.Type)
	group.Name = req.Name
	group.Avatar = req.Avatar
	group.Id = req.GroupId
	switch req.Type {
	case uint32(groupgrpcv1.GroupType_TypeEncrypted):
		group.MaxMembersLimit = model.EncryptedGroup
	default:
		group.MaxMembersLimit = model.DefaultGroup
	}

	resp, err := s.groupClient.UpdateGroup(ctx, &groupgrpcv1.UpdateGroupRequest{
		Group: group,
	})
	if err != nil {
		s.logger.Error("更新群聊信息失败", zap.Error(err))
		return nil, err
	}

	//查询所有群员
	relation, err := s.relationGroupClient.GetBatchGroupRelation(ctx, &relationgrpcv1.GetBatchGroupRelationRequest{
		GroupId: group.Id,
	})
	if err != nil {
		return 0, err
	}

	wg := sync.WaitGroup{}
	for _, res := range relation.GroupRelationResponses {
		go func(id string) {
			defer wg.Done()
			wg.Add(1)
			err := s.removeRedisGroupList(id, group.Id)
			if err != nil {
				return
			}
		}(res.UserId)
	}

	wg.Wait()
	return resp, nil
}

func (s *Service) GetBatchGroupInfoByIDs(ctx context.Context, ids []uint32) (interface{}, error) {
	groups, err := s.groupClient.GetBatchGroupInfoByIDs(ctx, &groupgrpcv1.GetBatchGroupInfoRequest{
		GroupIds: ids,
	})
	if err != nil {
		s.logger.Error("批量获取群聊信息失败", zap.Error(err))
		return nil, err
	}

	return groups.Groups, nil
}

func (s *Service) GetGroupInfoByGid(ctx context.Context, gid uint32, userID string) (interface{}, error) {
	group, err := s.groupClient.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{
		Gid: gid,
	})
	if err != nil {
		return nil, err
	}

	id, err := s.relationDialogClient.GetDialogByGroupId(ctx, &relationgrpcv1.GetDialogByGroupIdRequest{GroupId: gid})
	if err != nil {
		return nil, err
	}

	relation, err := s.relationGroupClient.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		GroupId: gid,
		UserId:  userID,
	})
	if err != nil {
		return nil, err
	}

	per := &model.Preferences{
		OpenBurnAfterReading: model.OpenBurnAfterReadingType(relation.OpenBurnAfterReading),
		SilentNotification:   model.SilentNotification(relation.IsSilent),
		Remark:               relation.Remark,
		EntryMethod:          model.EntryMethod(relation.JoinMethod),
		GroupNickname:        relation.GroupNickname,
		Inviter:              relation.Inviter,
		JoinedAt:             relation.JoinTime,
		MuteEndTime:          relation.MuteEndTime,
		Identity:             model.GroupIdentity(relation.Identity),
	}

	return &model.GroupInfo{
		Id:              group.Id,
		Avatar:          group.Avatar,
		Name:            group.Name,
		Type:            uint32(group.Type),
		Status:          int32(group.Status),
		MaxMembersLimit: group.MaxMembersLimit,
		CreatorId:       group.CreatorId,
		DialogId:        id.DialogId,
		Preferences:     per,
	}, nil
}

func (s *Service) insertRedisGroupList(userID string, msg usersorter.CustomGroupData) error {
	exists, err := s.redisClient.ExistsKey(fmt.Sprintf("group:%s", userID))
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	//查询是否有缓存
	values, err := s.redisClient.GetAllListValues(fmt.Sprintf("group:%s", userID))
	if err != nil {
		return err
	}
	if len(values) > 0 {
		// 类型转换
		var responseList []usersorter.Group
		responseList = append(responseList, msg)
		for _, v := range values {
			var friend usersorter.CustomGroupData
			err := json.Unmarshal([]byte(v), &friend)
			if err != nil {
				fmt.Println("Error decoding cached data:", err)
				continue
			}
			responseList = append(responseList, friend)
		}

		var result []interface{}

		// Assuming data2 is a slice or array of a specific type
		for _, item := range responseList {
			result = append(result, item)
		}

		err := s.redisClient.DelKey(fmt.Sprintf("group:%s", userID))
		if err != nil {
			return err
		}

		//存储到缓存
		err = s.redisClient.AddToList(fmt.Sprintf("group:%s", userID), result)
		if err != nil {
			return err
		}

		//设置key过期时间
		err = s.redisClient.SetKeyExpiration(fmt.Sprintf("group:%s", userID), 3*24*ostime.Hour)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) removeRedisGroupList(userID string, groupID uint32) error {
	exists, err := s.redisClient.ExistsKey(fmt.Sprintf("group:%s", userID))
	if err != nil {
		return err
	}
	if !exists {
		return nil
	}
	//查询是否有缓存
	values, err := s.redisClient.GetAllListValues(fmt.Sprintf("group:%s", userID))
	if err != nil {
		return err
	}
	if len(values) > 0 {
		// 类型转换
		var responseList []usersorter.Group
		if len(values) == 1 {
			err := s.redisClient.DelKey(fmt.Sprintf("group:%s", userID))
			if err != nil {
				return err
			}
			return nil
		}
		for _, v := range values {
			var group usersorter.CustomGroupData
			err := json.Unmarshal([]byte(v), &group)
			if err != nil {
				fmt.Println("Error decoding cached data:", err)
				continue
			}
			if group.GroupID != groupID {
				responseList = append(responseList, group)
			}
		}

		var result []interface{}

		// Assuming data2 is a slice or array of a specific type
		for _, item := range responseList {
			result = append(result, item)
		}

		err := s.redisClient.DelKey(fmt.Sprintf("group:%s", userID))
		if err != nil {
			return err
		}

		//存储到缓存
		err = s.redisClient.AddToList(fmt.Sprintf("group:%s", userID), result)
		if err != nil {
			return err
		}

		//设置key过期时间
		err = s.redisClient.SetKeyExpiration(fmt.Sprintf("group:%s", userID), 3*24*ostime.Hour)
		if err != nil {
			return err
		}
	}
	return nil
}

// 更新redis里的对话列表数据
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
