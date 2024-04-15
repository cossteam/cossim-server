package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	groupgrpcv1 "github.com/cossim/coss-server/internal/group/api/grpc/v1"
	"github.com/cossim/coss-server/internal/group/api/http/model"
	pushgrpcv1 "github.com/cossim/coss-server/internal/push/api/grpc/v1"
	relationgrpcv1 "github.com/cossim/coss-server/internal/relation/api/grpc/v1"
	storagev1 "github.com/cossim/coss-server/internal/storage/api/grpc/v1"
	usergrpcv1 "github.com/cossim/coss-server/internal/user/api/grpc/v1"
	"github.com/cossim/coss-server/pkg/code"
	myminio "github.com/cossim/coss-server/pkg/storage/minio"
	"github.com/cossim/coss-server/pkg/utils"
	httputil "github.com/cossim/coss-server/pkg/utils/http"
	"github.com/cossim/coss-server/pkg/utils/time"
	"github.com/cossim/coss-server/pkg/utils/usersorter"
	"github.com/dtm-labs/client/dtmcli"
	"github.com/dtm-labs/client/dtmgrpc"
	"github.com/dtm-labs/client/workflow"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/google/uuid"
	"github.com/lithammer/shortuuid/v3"
	"github.com/minio/minio-go/v7"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"io/ioutil"
	"mime/multipart"
)

func (s *Service) CreateGroup(ctx context.Context, req *groupgrpcv1.Group) (*model.CreateGroupResponse, error) {
	var err error
	friends, err := s.relationUserService.GetUserRelationByUserIds(ctx, &relationgrpcv1.GetUserRelationByUserIdsRequest{UserId: req.CreatorId, FriendIds: req.Member})
	if err != nil {
		s.logger.Error("获取好友关系失败", zap.Error(err))
		return nil, code.RelationErrCreateGroupFailed
	}

	isUserInFriends := func(userID string, friends []*relationgrpcv1.GetUserRelationResponse) bool {
		for _, friend := range friends {
			if friend.FriendId == userID {
				return true
			}
		}
		return false
	}

	//if len(req.Member) != len(friends.Users) {
	//	return nil, code.RelationUserErrFriendRelationNotFound.CustomMessage("加入的成员有些不是用户好友")
	//}
	for _, memberID := range req.Member {
		if !isUserInFriends(memberID, friends.Users) {
			info, err := s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{UserId: memberID})
			if err != nil {
				return nil, err
			}
			return nil, code.RelationUserErrFriendRelationNotFound.CustomMessage(fmt.Sprintf("%s不是你的好友", info.NickName))
		}
	}
	for _, friend := range friends.Users {
		if friend.Status != relationgrpcv1.RelationStatus_RELATION_NORMAL {
			return nil, code.StatusNotAvailable.CustomMessage(fmt.Sprintf("%s不是你的好友", friend.Remark))
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
	workflow.InitGrpc(s.dtmGrpcServer, s.groupServiceAddr, grpc.NewServer())
	gid := shortuuid.New()
	wfName := "create_group_workflow_" + gid
	if err = workflow.Register(wfName, func(wf *workflow.Workflow, data []byte) error {
		// 创建群聊
		resp1, err = s.groupService.CreateGroup(wf.Context, r1)
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
		}

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.groupService.CreateGroupRevert(wf.Context, &groupgrpcv1.CreateGroupRequest{Group: &groupgrpcv1.Group{
				Id: resp1.Id,
			}})
			return err
		})

		groupID = resp1.Id
		r22.GroupId = groupID
		r22.Member = req.Member
		r22.UserID = req.CreatorId

		resp2, err := s.relationGroupService.CreateGroupAndInviteUsers(wf.Context, r22)
		if err != nil {
			return status.Error(codes.Aborted, err.Error())
		}

		DialogID = resp2.DialogId

		wf.NewBranch().OnRollback(func(bb *dtmcli.BranchBarrier) error {
			_, err = s.relationGroupService.CreateGroupAndInviteUsersRevert(wf.Context, r22)
			return err
		})

		return err
	}); err != nil {
		return nil, err
	}

	if err = workflow.Execute(wfName, gid, nil); err != nil {
		s.logger.Error("WorkFlow CreateGroup", zap.Error(err))
		return nil, code.RelationErrCreateGroupFailed
	}

	data := map[string]interface{}{"group_id": groupID, "inviter_id": req.CreatorId}

	toBytes, err := utils.StructToBytes(data)
	if err != nil {
		return nil, err
	}
	// 给被邀请的用户推送
	for _, id := range req.Member {
		msg := &pushgrpcv1.WsMsg{Uid: id, Event: pushgrpcv1.WSEventType_InviteJoinGroupEvent, Data: &any.Any{Value: toBytes}, SendAt: time.Now(), PushOffline: true}
		toBytes2, err := utils.StructToBytes(msg)

		_, err = s.pushService.Push(ctx, &pushgrpcv1.PushRequest{
			Type: pushgrpcv1.Type_Ws,
			Data: toBytes2,
		})
		if err != nil {
			s.logger.Error("推送消息失败", zap.Error(err))
		}
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
	_, err := s.groupService.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{
		Gid: groupID,
	})
	if err != nil {
		return 0, code.GroupErrGroupNotFound
	}

	sf, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		UserId:  userID,
		GroupId: groupID,
	})
	if err != nil {
		return 0, err
	}

	if sf.Identity != relationgrpcv1.GroupIdentity_IDENTITY_OWNER {
		return 0, code.Forbidden
	}

	dialog, err := s.relationDialogService.GetDialogByGroupId(ctx, &relationgrpcv1.GetDialogByGroupIdRequest{GroupId: groupID})
	if err != nil {
		return 0, err
	}

	//查询所有群员
	_, err = s.relationGroupService.GetGroupUserIDs(ctx, &relationgrpcv1.GroupIDRequest{
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
		if err = tcc.CallBranch(r1, s.relationServiceAddr+relationgrpcv1.DialogService_DeleteDialogUsersByDialogID_FullMethodName, "", s.relationServiceAddr+relationgrpcv1.DialogService_DeleteDialogUsersByDialogIDRevert_FullMethodName, r); err != nil {
			return err
		}
		// 删除对话
		if err = tcc.CallBranch(r2, s.relationServiceAddr+relationgrpcv1.DialogService_DeleteDialogById_FullMethodName, "", s.relationServiceAddr+relationgrpcv1.DialogService_DeleteDialogByIdRevert_FullMethodName, r); err != nil {
			return err
		}
		// 删除群聊成员
		if err = tcc.CallBranch(r3, s.relationServiceAddr+relationgrpcv1.GroupRelationService_DeleteGroupRelationByGroupId_FullMethodName, "", s.relationServiceAddr+relationgrpcv1.GroupRelationService_DeleteGroupRelationByGroupIdRevert_FullMethodName, r); err != nil {
			return err
		}
		// 删除群聊
		if err = tcc.CallBranch(r4, s.ac.GRPC.Addr()+groupgrpcv1.GroupService_DeleteGroup_FullMethodName, "", s.ac.GRPC.Addr()+groupgrpcv1.GroupService_DeleteGroupRevert_FullMethodName, r); err != nil {
			return err
		}
		return err
	}); err != nil {
		s.logger.Error("WorkFlow DeleteGroup", zap.Error(err))
		return 0, code.GroupErrDeleteGroupFailed
	}

	return groupID, err
}

func (s *Service) UpdateGroup(ctx context.Context, req *model.UpdateGroupRequest, userID string) (interface{}, error) {
	group, err := s.groupService.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{
		Gid: req.GroupId,
	})
	if err != nil {
		s.logger.Error("更新群聊信息失败", zap.Error(err))
		return nil, err
	}

	sf, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
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

	resp, err := s.groupService.UpdateGroup(ctx, &groupgrpcv1.UpdateGroupRequest{
		Group: group,
	})
	if err != nil {
		s.logger.Error("更新群聊信息失败", zap.Error(err))
		return nil, err
	}

	return resp, nil
}

func (s *Service) GetBatchGroupInfoByIDs(ctx context.Context, ids []uint32) (interface{}, error) {
	groups, err := s.groupService.GetBatchGroupInfoByIDs(ctx, &groupgrpcv1.GetBatchGroupInfoRequest{
		GroupIds: ids,
	})
	if err != nil {
		s.logger.Error("批量获取群聊信息失败", zap.Error(err))
		return nil, err
	}

	return groups.Groups, nil
}

func (s *Service) GetGroupInfoByGid(ctx context.Context, gid uint32, userID string) (interface{}, error) {
	group, err := s.groupService.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{
		Gid: gid,
	})
	if err != nil {
		return nil, err
	}

	id, err := s.relationDialogService.GetDialogByGroupId(ctx, &relationgrpcv1.GetDialogByGroupIdRequest{GroupId: gid})
	if err != nil {
		return nil, err
	}

	relation, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		GroupId: gid,
		UserId:  userID,
	})
	if err != nil {
		return nil, err
	}

	fmt.Println("relation => ", relation)

	per := &model.Preferences{}
	if relation != nil && relation.GroupId != 0 {
		per = &model.Preferences{
			OpenBurnAfterReading: model.OpenBurnAfterReadingType(relation.OpenBurnAfterReading),
			SilentNotification:   model.SilentNotification(relation.IsSilent),
			Remark:               relation.Remark,
			EntryMethod:          model.EntryMethod(relation.JoinMethod),
			Inviter:              relation.Inviter,
			JoinedAt:             relation.JoinTime,
			MuteEndTime:          relation.MuteEndTime,
			Identity:             model.GroupIdentity(relation.Identity),
		}
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

// 更新redis里的对话列表数据
func (s *Service) insertRedisUserDialogList(userID string, msg model.UserDialogListResponse) error {
	key := fmt.Sprintf("dialog:%s", userID)
	//判断key是否存在，存在才继续
	f, err := s.redisClient.ExistsKey(key)
	if err != nil {
		return err
	}
	if !f {
		return nil
	}

	err = s.redisClient.AddToListLeft(key, []interface{}{msg})
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) removeRedisUserDialogList(userID string, dialogID uint32) error {
	key := fmt.Sprintf("dialog:%s", userID)
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
				var dialog model.UserDialogListResponse
				err := json.Unmarshal([]byte(v), &dialog)
				if err != nil {
					fmt.Println("Error decoding cached data:", err)
					return err
				}
				if dialog.DialogId == dialogID {
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
			var dialog model.UserDialogListResponse
			err := json.Unmarshal([]byte(v), &dialog)
			if err != nil {
				fmt.Println("Error decoding cached data:", err)
				return err
			}
			if dialog.DialogId == dialogID {
				_, err := s.redisClient.PopListElement(key, int64(i))
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *Service) ModifyGroupAvatar(ctx context.Context, userID string, groupID uint32, avatar multipart.File) (string, error) {
	_, err := s.groupService.GetGroupInfoByGid(ctx, &groupgrpcv1.GetGroupInfoRequest{
		Gid: groupID,
	})
	if err != nil {
		return "", err
	}

	relation, err := s.relationGroupService.GetGroupRelation(ctx, &relationgrpcv1.GetGroupRelationRequest{
		UserId:  userID,
		GroupId: groupID,
	})
	if err != nil {
		return "", err
	}

	if relation.Identity != relationgrpcv1.GroupIdentity_IDENTITY_OWNER && relation.Identity != relationgrpcv1.GroupIdentity_IDENTITY_ADMIN {
		return "", code.Forbidden
	}

	_, err = s.userService.UserInfo(ctx, &usergrpcv1.UserInfoRequest{
		UserId: userID,
	})
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(avatar)
	if err != nil {
		return "", err
	}

	bucket, err := myminio.GetBucketName(int(storagev1.FileType_Other))
	if err != nil {
		return "", err
	}

	// 将字节数组转换为 io.Reader
	reader := bytes.NewReader(data)
	fileID := uuid.New().String()
	key := myminio.GenKey(bucket, fileID+".jpeg")
	err = s.sp.UploadAvatar(ctx, key, reader, reader.Size(), minio.PutObjectOptions{
		ContentType: "image/jpeg",
	})
	if err != nil {
		return "", err
	}

	aUrl := fmt.Sprintf("http://%s%s/%s", s.gatewayAddress, s.downloadURL, key)
	if s.ac.SystemConfig.Ssl {
		aUrl, err = httputil.ConvertToHttps(aUrl)
		if err != nil {
			return "", err
		}
	}

	_, err = s.groupService.UpdateGroup(ctx, &groupgrpcv1.UpdateGroupRequest{
		Group: &groupgrpcv1.Group{
			Id:     groupID,
			Avatar: aUrl,
		},
	})
	if err != nil {
		return "", err
	}

	return aUrl, nil
}
